// Package create provides the mock create command.
package create

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
)

// NewCommand creates the mock create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], model, firmware)
		},
	}

	cmd.Flags().StringVar(&model, "model", "Plus 1PM", "Device model")
	cmd.Flags().StringVar(&firmware, "firmware", "1.0.0", "Firmware version")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, name, model, firmware string) error {
	ios := f.IOStreams()

	mockDir, err := mock.Dir()
	if err != nil {
		return err
	}

	dev := mock.Device{
		Name:     name,
		Model:    model,
		Firmware: firmware,
		MAC:      mock.GenerateMAC(name),
		State: map[string]interface{}{
			"switch:0": map[string]interface{}{
				"output": false,
				"apower": 0.0,
			},
		},
	}

	data, err := json.MarshalIndent(dev, "", "  ")
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
	ios.Printf("  MAC: %s\n", dev.MAC)
	ios.Println("")
	ios.Info("Add to config with: shelly config device add %s --address mock://%s", name, name)

	return nil
}
