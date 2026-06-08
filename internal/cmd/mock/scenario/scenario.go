// Package scenario provides the mock scenario command.
package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
)

// Scenario names and device defaults.
const (
	scenarioMinimal = "minimal"
	scenarioHome    = "home"
	scenarioOffice  = "office"

	defaultFirmware = "1.0.0"
	modelPlus1PM    = "Plus 1PM"
	modelPlus2PM    = "Plus 2PM"
	modelPlugS      = "Plug S"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Scenario string
}

// NewCommand creates the mock scenario command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "scenario <name>",
		Aliases: []string{"load", "setup"},
		Short:   "Load a test scenario",
		Long: `Load a pre-defined test scenario with multiple mock devices.

Built-in scenarios:
  home     - Basic home setup (3 devices)
  office   - Office setup (5 devices)
  minimal  - Single device for quick testing`,
		Example: `  # Load home scenario
  shelly mock scenario home

  # Load office scenario
  shelly mock scenario office`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Scenario = args[0]
			return run(cmd.Context(), opts)
		},
	}
	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	scenarios := map[string][]mock.Device{
		scenarioMinimal: {
			{Name: "test-switch", Model: modelPlus1PM, Firmware: defaultFirmware},
		},
		scenarioHome: {
			{Name: "living-room", Model: modelPlus1PM, Firmware: defaultFirmware},
			{Name: "bedroom", Model: modelPlus2PM, Firmware: defaultFirmware},
			{Name: "kitchen", Model: modelPlugS, Firmware: defaultFirmware},
		},
		scenarioOffice: {
			{Name: "desk-lamp", Model: modelPlus1PM, Firmware: defaultFirmware},
			{Name: "monitor", Model: modelPlugS, Firmware: defaultFirmware},
			{Name: "printer", Model: modelPlugS, Firmware: defaultFirmware},
			{Name: "air-purifier", Model: modelPlus2PM, Firmware: defaultFirmware},
			{Name: "heater", Model: modelPlus2PM, Firmware: defaultFirmware},
		},
	}

	devices, ok := scenarios[opts.Scenario]
	if !ok {
		return fmt.Errorf("unknown scenario: %s (available: minimal, home, office)", opts.Scenario)
	}

	mockDir, err := mock.Dir()
	if err != nil {
		return err
	}

	ios.Info("Loading scenario: %s", opts.Scenario)

	for _, device := range devices {
		device.MAC = mock.GenerateMAC(device.Name)
		device.State = map[string]interface{}{
			"switch:0": map[string]interface{}{
				"output": false,
				"apower": 0.0,
			},
		}

		data, err := json.MarshalIndent(device, "", "  ")
		if err != nil {
			ios.Warning("Failed to create %s: %v", device.Name, err)
			continue
		}

		filename := filepath.Join(mockDir, device.Name+".json")
		if err := afero.WriteFile(config.Fs(), filename, data, 0o600); err != nil {
			ios.Warning("Failed to write %s: %v", device.Name, err)
			continue
		}

		ios.Printf("  Created: %s (%s)\n", device.Name, device.Model)
	}

	ios.Println("")
	ios.Success("Scenario loaded: %d devices created", len(devices))
	ios.Info("List devices with: shelly mock list")

	return nil
}
