// Package configimport provides the config import subcommand.
// Named configimport to avoid conflict with Go's import keyword.
package configimport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory   *cmdutil.Factory
	Device    string
	FilePath  string
	DryRun    bool
	Merge     bool
	Overwrite bool
}

// NewCommand creates the config import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "import <device> <file>",
		Aliases: []string{"restore", "load"},
		Short:   "Import configuration from a file",
		Long: `Import device configuration from a JSON or YAML file.

By default, only specified keys are updated (merge mode). Use --overwrite
to replace the entire configuration.`,
		Example: `  # Import configuration (merge mode)
  shelly config import living-room config-backup.json

  # Dry run - show what would change without applying
  shelly config import living-room config.json --dry-run

  # Overwrite entire configuration
  shelly config import living-room config.json --overwrite`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.FilePath = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be changed without applying")
	cmd.Flags().BoolVar(&opts.Merge, "merge", true, "Merge with existing configuration (default)")
	cmd.Flags().BoolVar(&opts.Overwrite, "overwrite", false, "Overwrite entire configuration")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	// Read and parse file
	fileData, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var config map[string]any

	// Try JSON first, then YAML
	if err := json.Unmarshal(fileData, &config); err != nil {
		if err := yaml.Unmarshal(fileData, &config); err != nil {
			return fmt.Errorf("failed to parse file as JSON or YAML: %w", err)
		}
	}

	svc := opts.Factory.ShellyService()
	ios := opts.Factory.IOStreams()

	if opts.DryRun {
		// Get current config and show diff
		var currentConfig map[string]any
		err = cmdutil.RunWithSpinner(ctx, ios, "Getting current configuration...", func(ctx context.Context) error {
			var getErr error
			currentConfig, getErr = svc.GetConfig(ctx, opts.Device)
			return getErr
		})
		if err != nil {
			return fmt.Errorf("failed to get current configuration: %w", err)
		}

		ios.Title("Dry run - changes that would be applied")
		term.DisplayConfigMapDiff(ios, currentConfig, config)
		return nil
	}

	err = cmdutil.RunWithSpinner(ctx, ios, "Importing configuration...", func(ctx context.Context) error {
		return svc.SetConfig(ctx, opts.Device, config)
	})
	if err != nil {
		return fmt.Errorf("failed to import configuration: %w", err)
	}

	ios.Success("Configuration imported to %s", opts.Device)
	return nil
}
