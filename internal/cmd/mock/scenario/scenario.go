// Package scenario provides the mock scenario command.
package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
)

// NewCommand creates the mock scenario command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}
	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, scenario string) error {
	ios := f.IOStreams()

	scenarios := map[string][]mock.Device{
		"minimal": {
			{Name: "test-switch", Model: "Plus 1PM", Firmware: "1.0.0"},
		},
		"home": {
			{Name: "living-room", Model: "Plus 1PM", Firmware: "1.0.0"},
			{Name: "bedroom", Model: "Plus 2PM", Firmware: "1.0.0"},
			{Name: "kitchen", Model: "Plug S", Firmware: "1.0.0"},
		},
		"office": {
			{Name: "desk-lamp", Model: "Plus 1PM", Firmware: "1.0.0"},
			{Name: "monitor", Model: "Plug S", Firmware: "1.0.0"},
			{Name: "printer", Model: "Plug S", Firmware: "1.0.0"},
			{Name: "air-purifier", Model: "Plus 2PM", Firmware: "1.0.0"},
			{Name: "heater", Model: "Plus 2PM", Firmware: "1.0.0"},
		},
	}

	devices, ok := scenarios[scenario]
	if !ok {
		return fmt.Errorf("unknown scenario: %s (available: minimal, home, office)", scenario)
	}

	mockDir, err := mock.Dir()
	if err != nil {
		return err
	}

	ios.Info("Loading scenario: %s", scenario)

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
		if err := os.WriteFile(filename, data, 0o600); err != nil {
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
