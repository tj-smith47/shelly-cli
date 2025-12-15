// Package mock provides the mock command for mock device mode.
package mock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Device represents a mock device configuration.
type Device struct {
	Name     string                 `json:"name"`
	Model    string                 `json:"model"`
	Firmware string                 `json:"firmware"`
	MAC      string                 `json:"mac"`
	State    map[string]interface{} `json:"state"`
}

// NewCommand creates the mock command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mock",
		Aliases: []string{"simulate", "test"},
		Short:   "Mock device mode for testing",
		Long: `Mock device mode for testing without real hardware.

Create and manage mock devices for testing CLI commands
and automation scripts without physical Shelly devices.

Subcommands:
  create    - Create a new mock device
  list      - List mock devices
  delete    - Delete a mock device
  scenario  - Load a test scenario`,
		Example: `  # Create a mock device
  shelly mock create kitchen-light --model "Plus 1PM"

  # List mock devices
  shelly mock list

  # Load test scenario
  shelly mock scenario home-setup`,
	}

	cmd.AddCommand(newCreateCommand(f))
	cmd.AddCommand(newListCommand(f))
	cmd.AddCommand(newDeleteCommand(f))
	cmd.AddCommand(newScenarioCommand(f))

	return cmd
}

func newCreateCommand(f *cmdutil.Factory) *cobra.Command {
	var model, firmware string

	cmd := &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"add", "new"},
		Short:   "Create a mock device",
		Long:    `Create a new mock device for testing.`,
		Example: `  # Create a mock Plus 1PM
  shelly mock create kitchen --model "Plus 1PM"

  # Create with specific firmware
  shelly mock create bedroom --model "Plus 2PM" --firmware "1.0.8"`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runCreate(f, args[0], model, firmware)
		},
	}

	cmd.Flags().StringVar(&model, "model", "Plus 1PM", "Device model")
	cmd.Flags().StringVar(&firmware, "firmware", "1.0.0", "Firmware version")

	return cmd
}

func newListCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List mock devices",
		Long:    `List all configured mock devices.`,
		Example: `  # List mock devices
  shelly mock list`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runList(f)
		},
	}
}

func newDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a mock device",
		Long:    `Delete a mock device configuration.`,
		Example: `  # Delete mock device
  shelly mock delete kitchen`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runDelete(f, args[0])
		},
	}
}

func newScenarioCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
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
		RunE: func(_ *cobra.Command, args []string) error {
			return runScenario(f, args[0])
		},
	}
}

func getMockDir() (string, error) {
	configDir, err := config.Dir()
	if err != nil {
		return "", err
	}
	mockDir := filepath.Join(configDir, "mock")
	if err := os.MkdirAll(mockDir, 0o700); err != nil {
		return "", err
	}
	return mockDir, nil
}

func runCreate(f *cmdutil.Factory, name, model, firmware string) error {
	ios := f.IOStreams()

	mockDir, err := getMockDir()
	if err != nil {
		return err
	}

	device := Device{
		Name:     name,
		Model:    model,
		Firmware: firmware,
		MAC:      generateMAC(name),
		State: map[string]interface{}{
			"switch:0": map[string]interface{}{
				"output": false,
				"apower": 0.0,
			},
		},
	}

	data, err := json.MarshalIndent(device, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(mockDir, name+".json")
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return err
	}

	ios.Success("Created mock device: %s", name)
	ios.Printf("  Model: %s\n", model)
	ios.Printf("  Firmware: %s\n", firmware)
	ios.Printf("  MAC: %s\n", device.MAC)
	ios.Println("")
	ios.Info("Add to config with: shelly config device add %s --address mock://%s", name, name)

	return nil
}

func runList(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	mockDir, err := getMockDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(mockDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		ios.Info("No mock devices configured")
		ios.Info("Create one with: shelly mock create <name>")
		return nil
	}

	ios.Printf("Mock Devices:\n\n")

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filename := filepath.Join(mockDir, entry.Name())
		data, err := os.ReadFile(filename) //nolint:gosec // Mock dir is from config
		if err != nil {
			continue
		}

		var device Device
		if err := json.Unmarshal(data, &device); err != nil {
			continue
		}

		ios.Printf("  %s\n", device.Name)
		ios.Printf("    Model: %s, Firmware: %s\n", device.Model, device.Firmware)
		ios.Printf("    MAC: %s\n", device.MAC)
	}

	return nil
}

func runDelete(f *cmdutil.Factory, name string) error {
	ios := f.IOStreams()

	mockDir, err := getMockDir()
	if err != nil {
		return err
	}

	filename := filepath.Join(mockDir, name+".json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("mock device not found: %s", name)
	}

	if err := os.Remove(filename); err != nil {
		return err
	}

	ios.Success("Deleted mock device: %s", name)
	return nil
}

func runScenario(f *cmdutil.Factory, scenario string) error {
	ios := f.IOStreams()

	scenarios := map[string][]Device{
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

	mockDir, err := getMockDir()
	if err != nil {
		return err
	}

	ios.Info("Loading scenario: %s", scenario)

	for _, device := range devices {
		device.MAC = generateMAC(device.Name)
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

func generateMAC(name string) string {
	// Generate a deterministic MAC based on the name
	// Prefix with Shelly's OUI-like prefix
	hash := 0
	for _, c := range name {
		hash = hash*31 + int(c)
	}
	return fmt.Sprintf("AA:BB:CC:%02X:%02X:%02X",
		(hash>>16)&0xFF,
		(hash>>8)&0xFF,
		hash&0xFF,
	)
}
