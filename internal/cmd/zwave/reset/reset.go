// Package reset provides the zwave reset command.
package reset

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/profiles"
	"github.com/tj-smith47/shelly-go/zwave"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Model   string
	Factory *cmdutil.Factory
}

// NewCommand creates the zwave reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "reset <model>",
		Aliases: []string{"factory-reset", "factory"},
		Short:   "Show factory reset instructions",
		Long: `Show factory reset instructions for a Z-Wave device.

WARNING: Factory reset should only be used when the gateway is missing
or inoperable. All custom parameters, associations, and routing
information will be lost.`,
		Example: `  # Show factory reset instructions
  shelly zwave reset SNSW-001P16ZW`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Model = args[0]
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	profile, ok := profiles.Get(opts.Model)
	if !ok {
		return fmt.Errorf("unknown device model: %s", opts.Model)
	}

	device := zwave.NewDevice(profile)
	if !device.IsZWave() {
		return fmt.Errorf("%s is not a Z-Wave device", opts.Model)
	}

	info := zwave.GetFactoryResetInfo(device)

	ios.Title("Z-Wave Factory Reset Instructions")
	ios.Println()

	ios.Printf("  %s: %s\n", theme.Dim().Render("Device"), theme.Highlight().Render(device.Name()))

	ios.Println()
	ios.Printf("  %s %s\n", theme.StatusError().Render("⚠"), theme.SemanticWarning().Render("WARNING:"))
	ios.Printf("    %s\n", info.Warning)

	ios.Println()
	ios.Printf("  %s:\n", theme.Dim().Render("Steps"))

	for _, step := range info.Instructions {
		ios.Printf("    %s\n", step)
	}

	ios.Println()
	return nil
}
