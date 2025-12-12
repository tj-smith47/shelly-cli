// Package export provides the config export subcommand.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"gopkg.in/yaml.v3"
)

var formatFlag string

// NewCommand creates the config export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <device> <file>",
		Short: "Export device configuration to a file",
		Long: `Export the complete device configuration to a file.

The configuration is saved in JSON format by default. Use --format=yaml
for YAML output.`,
		Example: `  # Export to JSON file
  shelly config export living-room config-backup.json

  # Export to YAML file
  shelly config export living-room config-backup.yaml --format=yaml

  # Export to stdout
  shelly config export living-room -`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], args[1])
		},
	}

	cmd.Flags().StringVarP(&formatFlag, "format", "f", "json", "Output format (json, yaml)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := f.ShellyService()

	spin := iostreams.NewSpinner("Getting configuration...")
	spin.Start()

	config, err := svc.GetConfig(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// Marshal based on format
	var data []byte
	switch formatFlag {
	case "yaml", "yml":
		data, err = yaml.Marshal(config)
	default:
		data, err = json.MarshalIndent(config, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file or stdout
	if filePath == "-" {
		fmt.Println(string(data))
	} else {
		if err := os.WriteFile(filePath, data, 0o644); err != nil { //nolint:gosec // G306: 0o644 is acceptable for config exports
			return fmt.Errorf("failed to write file: %w", err)
		}
		iostreams.Success("Configuration exported to %s", filePath)
	}

	return nil
}
