// Package export provides the config export subcommand.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	FilePath string
	Format   string
}

// NewCommand creates the config export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "export <device> <file>",
		Aliases: []string{"backup", "save"},
		Short:   "Export device configuration to a file",
		Long: `Export the complete device configuration to a file.

The configuration is saved in JSON format by default. Use --format=yaml
for YAML output.`,
		Example: `  # Export to JSON file
  shelly config export living-room config-backup.json

  # Export to YAML file
  shelly config export living-room config-backup.yaml --format=yaml

  # Export to stdout
  shelly config export living-room -`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceThenFile(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.FilePath = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", "json", "Output format (json, yaml)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()
	ios := opts.Factory.IOStreams()

	var config map[string]any
	err := cmdutil.RunWithSpinner(ctx, ios, "Getting configuration...", func(ctx context.Context) error {
		var err error
		config, err = svc.GetConfig(ctx, opts.Device)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// Marshal based on format
	var data []byte
	switch opts.Format {
	case "yaml", "yml":
		data, err = yaml.Marshal(config)
	default:
		data, err = json.MarshalIndent(config, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file or stdout
	if opts.FilePath == "-" {
		ios.Printf("%s\n", data)
	} else {
		if err := os.WriteFile(opts.FilePath, data, 0o644); err != nil { //nolint:gosec // G306: 0o644 is acceptable for config exports
			return fmt.Errorf("failed to write file: %w", err)
		}
		ios.Success("Configuration exported to %s", opts.FilePath)
	}

	return nil
}
