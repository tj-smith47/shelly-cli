// Package status provides the link status subcommand.
package status

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the options for the status command.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the link status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status [child-device]",
		Aliases: []string{"st", "s"},
		Short:   "Show link status with derived device state",
		Long: `Show the status of device links with resolved parent switch state.

When a linked child device is offline, its state is derived from the
parent switch state. If no device is specified, shows all links.`,
		Example: `  # Show status of all links
  shelly link status

  # Show status for a specific linked device
  shelly link status bulb-duo

  # Output as JSON
  shelly link status -o json`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.LinkedDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Device = args[0]
			}
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	links := config.ListLinks()
	if len(links) == 0 {
		ios.Info("No links defined")
		ios.Info("Use 'shelly link set <child> <parent>' to create a link")
		return nil
	}

	// Filter to single device if specified
	if opts.Device != "" {
		link, ok := config.GetLink(opts.Device)
		if !ok {
			return fmt.Errorf("no link found for device %q", opts.Device)
		}
		links = map[string]config.Link{opts.Device: link}
	}

	// Flatten map to slice for indexed concurrent access.
	type linkEntry struct {
		child string
		link  config.Link
	}
	entries := make([]linkEntry, 0, len(links))
	for child, link := range links {
		entries = append(entries, linkEntry{child: child, link: link})
	}

	statuses := make([]model.LinkStatus, len(entries))

	err := cmdutil.RunWithSpinner(ctx, ios, "Resolving link states...", func(ctx context.Context) error {
		var wg sync.WaitGroup
		for i, entry := range entries {
			idx := i
			e := entry
			wg.Go(func() {
				devCtx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
				defer cancel()

				ls := model.LinkStatus{
					ChildDevice:  e.child,
					ParentDevice: e.link.ParentDevice,
					SwitchID:     e.link.SwitchID,
				}

				switchStatus, switchErr := svc.SwitchStatus(devCtx, e.link.ParentDevice, e.link.SwitchID)
				if switchErr != nil {
					ls.ParentOnline = false
					ls.State = "Unknown"
				} else {
					ls.ParentOnline = true
					ls.SwitchOutput = switchStatus.Output
					if switchStatus.Output {
						ls.State = "On"
					} else {
						ls.State = "Off (switch off)"
					}
				}

				statuses[idx] = ls
			})
		}
		wg.Wait()
		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].ChildDevice < statuses[j].ChildDevice
	})

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, statuses)
	}

	term.DisplayLinkStatuses(ios, statuses)
	return nil
}
