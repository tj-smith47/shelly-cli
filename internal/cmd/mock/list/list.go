// Package list provides the mock list command.
package list

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the mock list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List mock devices",
		Long:    `List all configured mock devices.`,
		Example: `  # List mock devices
  shelly mock list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}
	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	mockDir, err := mock.Dir()
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

		var device mock.Device
		if err := json.Unmarshal(data, &device); err != nil {
			continue
		}

		ios.Printf("  %s\n", device.Name)
		ios.Printf("    Model: %s, Firmware: %s\n", device.Model, device.Firmware)
		ios.Printf("    MAC: %s\n", device.MAC)
	}

	return nil
}
