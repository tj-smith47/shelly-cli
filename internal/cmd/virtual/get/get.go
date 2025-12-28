// Package get provides the virtual get command.
package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Device  string
	Key     string
	Factory *cmdutil.Factory
}

// NewCommand creates the virtual get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "get <device> <key>",
		Aliases: []string{"show", "status"},
		Short:   "Get a virtual component value",
		Long: `Get the current value of a virtual component.

The key format is "type:id", for example "boolean:200" or "number:201".`,
		Example: `  # Get a boolean component
  shelly virtual get kitchen boolean:200

  # Get a number component with JSON output
  shelly virtual get kitchen number:201 -o json`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Key = args[1]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate key format
	if _, _, err := shelly.ParseVirtualKey(opts.Key); err != nil {
		return err
	}

	// Fetch the component from the list to get full details
	var component *shelly.VirtualComponent
	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching virtual component...", func(ctx context.Context) error {
		components, fetchErr := svc.ListVirtualComponents(ctx, opts.Device)
		if fetchErr != nil {
			return fetchErr
		}
		for i := range components {
			if components[i].Key == opts.Key {
				component = &components[i]
				return nil
			}
		}
		return fmt.Errorf("virtual component %s not found", opts.Key)
	})
	if err != nil {
		return err
	}

	return cmdutil.PrintResult(ios, component, func(ios *iostreams.IOStreams, c *shelly.VirtualComponent) {
		term.DisplayVirtualComponent(ios, c)
	})
}
