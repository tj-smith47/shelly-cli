// Package inclusion provides the zwave inclusion command.
package inclusion

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/profiles"
	"github.com/tj-smith47/shelly-go/zwave"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Model   string
	Mode    string
	Factory *cmdutil.Factory
}

// NewCommand creates the zwave inclusion command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "inclusion <model>",
		Aliases: []string{"include", "pair", "add"},
		Short:   "Show inclusion instructions",
		Long: `Show Z-Wave inclusion (pairing) instructions for a device.

Inclusion modes:
  smart_start - Automatic inclusion via QR code (recommended)
  button      - Manual inclusion using the S button
  switch      - Manual inclusion using the connected switch`,
		Example: `  # Show SmartStart inclusion (default)
  shelly zwave inclusion SNSW-001P16ZW

  # Show button-based inclusion
  shelly zwave inclusion SNSW-001P16ZW --mode button

  # Show switch-based inclusion
  shelly zwave inclusion SNSW-001P16ZW --mode switch`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Model = args[0]
			return run(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Mode, "mode", "smart_start", "Inclusion mode (smart_start, button, switch)")

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

	var mode zwave.InclusionMode
	switch strings.ToLower(opts.Mode) {
	case "smart_start", "smartstart", "qr":
		mode = zwave.InclusionSmartStart
	case "button", "s":
		mode = zwave.InclusionButton
	case "switch":
		mode = zwave.InclusionSwitch
	default:
		return fmt.Errorf("invalid mode %q, must be one of: smart_start, button, switch", opts.Mode)
	}

	info := zwave.GetInclusionInfo(device, mode)

	ios.Title("Z-Wave Inclusion Instructions")
	ios.Println()

	ios.Printf("  %s: %s\n", theme.Dim().Render("Device"), theme.Highlight().Render(device.Name()))
	ios.Printf("  %s: %s\n", theme.Dim().Render("Mode"), string(info.Mode))

	if info.DSKRequired {
		ios.Printf("  %s: %s\n", theme.Dim().Render("DSK PIN"), theme.SemanticWarning().Render("Required (5-digit code from device label)"))
	}

	ios.Println()
	ios.Printf("  %s:\n", theme.Dim().Render("Steps"))

	for _, step := range info.Instructions {
		ios.Printf("    %s\n", step)
	}

	ios.Println()
	return nil
}
