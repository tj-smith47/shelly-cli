// Package config provides the zwave config command.
package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/zwave"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Factory *cmdutil.Factory
}

// NewCommand creates the zwave config command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"params", "parameters"},
		Short:   "Show common configuration parameters",
		Long: `Show common Z-Wave configuration parameters for Wave devices.

These parameters can be configured via your Z-Wave gateway's
configuration interface. Actual parameters vary by device model.`,
		Example: `  # Show common parameters
  shelly zwave config

  # JSON output
  shelly zwave config -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	params := zwave.CommonConfigParameters()

	return cmdutil.PrintListResult(ios, params, func(ios *iostreams.IOStreams, items []zwave.ConfigurationParameter) {
		ios.Title("Common Z-Wave Configuration Parameters")
		ios.Println()

		for _, p := range items {
			ios.Printf("  %s %s\n",
				theme.Highlight().Render(fmt.Sprintf("[%d]", p.Number)),
				theme.Bold().Render(p.Name))
			ios.Printf("      %s\n", p.Description)
			ios.Printf("      %s: %d  %s: %d-%d  %s: %d\n",
				theme.Dim().Render("Default"), p.DefaultValue,
				theme.Dim().Render("Range"), p.MinValue, p.MaxValue,
				theme.Dim().Render("Size"), p.Size)
			ios.Println()
		}

		ios.Info("Actual parameters vary by device model. Consult device documentation.")
	})
}
