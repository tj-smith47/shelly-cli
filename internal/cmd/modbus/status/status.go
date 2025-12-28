// Package status provides the modbus status command.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/modbus"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the modbus status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "s"},
		Short:   "Show Modbus-TCP status",
		Long:    `Show the current Modbus-TCP server status and configuration.`,
		Example: `  # Show Modbus status
  shelly modbus status kitchen

  # JSON output
  shelly modbus status kitchen -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
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
	svc := opts.Factory.ModbusService()

	var status *modbus.Status
	var config *modbus.Config
	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching Modbus status...", func(ctx context.Context) error {
		var err error
		status, err = svc.GetStatus(ctx, opts.Device)
		if err != nil {
			return err
		}
		config, err = svc.GetConfig(ctx, opts.Device)
		return err
	})
	if err != nil {
		return err
	}

	result := struct {
		Enabled    bool `json:"enabled"`
		Configured bool `json:"configured"`
	}{
		Enabled:    status.Enabled,
		Configured: config.Enable,
	}

	return cmdutil.PrintResult(ios, result, func(ios *iostreams.IOStreams, r struct {
		Enabled    bool `json:"enabled"`
		Configured bool `json:"configured"`
	}) {
		ios.Title("Modbus-TCP Status")
		ios.Println()

		enabledStr := theme.StatusError().Render("Disabled")
		if r.Enabled {
			enabledStr = theme.StatusOK().Render("Enabled (port 502)")
		}
		ios.Printf("  %s: %s\n", theme.Dim().Render("Status"), enabledStr)

		configStr := theme.StatusError().Render("No")
		if r.Configured {
			configStr = theme.StatusOK().Render("Yes")
		}
		ios.Printf("  %s: %s\n", theme.Dim().Render("Configured"), configStr)

		ios.Println()
	})
}
