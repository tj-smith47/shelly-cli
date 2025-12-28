// Package status provides the matter status command.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the matter status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "info"},
		Short:   "Show Matter status",
		Long: `Show Matter connectivity status for a Shelly device.

Displays:
- Whether Matter is enabled
- Commissionable status (can be added to a fabric)
- Number of paired fabrics
- Network information when connected`,
		Example: `  # Show Matter status
  shelly matter status living-room

  # Output as JSON
  shelly matter status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "text", "text", "json")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	status, err := svc.Wireless().FetchMatterStatus(ctx, opts.Device, ios)
	if err != nil {
		return err
	}

	if opts.Format == "json" {
		return term.OutputMatterStatusJSON(ios, status)
	}

	term.DisplayMatterStatus(ios, status, opts.Device)
	return nil
}
