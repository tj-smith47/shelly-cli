// Package create provides the mock create command.
package create

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Name     string
	Model    string
	Firmware string
}

// NewCommand creates the mock create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Name = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Model, "model", "Plus 1PM", "Device model")
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "1.0.0", "Firmware version")

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	mockDir, err := mock.Dir()
	if err != nil {
		return err
	}

	dev := mock.Device{
		Name:     opts.Name,
		Model:    opts.Model,
		Firmware: opts.Firmware,
		MAC:      mock.GenerateMAC(opts.Name),
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

	filename := filepath.Join(mockDir, opts.Name+".json")
	if err := afero.WriteFile(config.Fs(), filename, data, 0o600); err != nil {
		return err
	}

	ios.Success("Created mock device: %s", opts.Name)
	ios.Printf("  Model: %s\n", opts.Model)
	ios.Printf("  Firmware: %s\n", opts.Firmware)
	ios.Printf("  MAC: %s\n", dev.MAC)
	ios.Println("")
	ios.Info("Add to config with: shelly config device add %s --address mock://%s", opts.Name, opts.Name)

	return nil
}
