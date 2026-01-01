// Package info provides the zwave info command.
package info

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/profiles"
	"github.com/tj-smith47/shelly-go/zwave"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Model   string
	Factory *cmdutil.Factory
}

// NewCommand creates the zwave info command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "info <model>",
		Aliases: []string{"show", "i"},
		Short:   "Show Z-Wave device information",
		Long: `Show Z-Wave device information for a Shelly Wave model.

Displays device capabilities, supported protocols, and network topology options.`,
		Example: `  # Show info for Wave 1PM
  shelly zwave info SNSW-001P16ZW

  # JSON output
  shelly zwave info SNSW-001P16ZW -o json`,
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Could provide model completion here
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Model = args[0]
			return run(opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

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

	result := struct {
		Model             string `json:"model"`
		Name              string `json:"name"`
		Generation        string `json:"generation"`
		IsZWave           bool   `json:"is_zwave"`
		HasEthernet       bool   `json:"has_ethernet"`
		HasWiFi           bool   `json:"has_wifi"`
		SupportsLongRange bool   `json:"supports_long_range"`
		IsPro             bool   `json:"is_pro"`
	}{
		Model:             device.Model(),
		Name:              device.Name(),
		Generation:        device.Generation().String(),
		IsZWave:           device.IsZWave(),
		HasEthernet:       device.HasEthernet(),
		HasWiFi:           device.HasWiFi(),
		SupportsLongRange: device.SupportsLongRange(),
		IsPro:             device.IsPro(),
	}

	return cmdutil.PrintResult(ios, result, func(ios *iostreams.IOStreams, r struct {
		Model             string `json:"model"`
		Name              string `json:"name"`
		Generation        string `json:"generation"`
		IsZWave           bool   `json:"is_zwave"`
		HasEthernet       bool   `json:"has_ethernet"`
		HasWiFi           bool   `json:"has_wifi"`
		SupportsLongRange bool   `json:"supports_long_range"`
		IsPro             bool   `json:"is_pro"`
	}) {
		ios.Title("Z-Wave Device Info")
		ios.Println()

		ios.Printf("  %s: %s\n", theme.Dim().Render("Model"), theme.Highlight().Render(r.Model))
		ios.Printf("  %s: %s\n", theme.Dim().Render("Name"), r.Name)
		ios.Printf("  %s: %s\n", theme.Dim().Render("Generation"), r.Generation)

		ios.Println()
		ios.Printf("  %s:\n", theme.Dim().Render("Protocols"))
		ios.Printf("    Z-Wave: %s\n", output.RenderYesNo(r.IsZWave, output.CaseTitle, theme.FalseError))
		ios.Printf("    Ethernet: %s\n", output.RenderYesNo(r.HasEthernet, output.CaseTitle, theme.FalseError))
		ios.Printf("    WiFi: %s\n", output.RenderYesNo(r.HasWiFi, output.CaseTitle, theme.FalseError))

		ios.Println()
		ios.Printf("  %s:\n", theme.Dim().Render("Features"))
		ios.Printf("    Z-Wave Long Range: %s\n", output.RenderYesNo(r.SupportsLongRange, output.CaseTitle, theme.FalseError))
		ios.Printf("    Pro Series: %s\n", output.RenderYesNo(r.IsPro, output.CaseTitle, theme.FalseError))

		ios.Println()
	})
}
