// Package configimport provides the config import subcommand.
// Named configimport to avoid conflict with Go's import keyword.
package configimport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	FilePath string
	DryRun   bool
}

// NewCommand creates the config import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "import <device> <file>",
		Aliases: []string{"restore", "load"},
		Short:   "Import configuration from a file",
		Long: `Import device configuration from a JSON or YAML file.

Keys present in the file are applied to the device; keys absent from the file
are left unchanged (the device merges the update — there is no whole-config
replace primitive). Capture a file in this format with 'shelly device config export'.

WiFi stations in the file: when the file's sys.device.mac is the device's own,
a station on the device's current network is written without a password (the
device keeps its key), and a changed network takes the password stored on this
host. A file from another device, or one with no MAC, never copies its station
address (ip, netmask, gw, nameserver, ipv4mode), and its network is written
only when the device is not already on it and this host has its password. A
station that cannot be written that way is left out with a warning; set it
with 'shelly wifi set'. --dry-run shows each station's planned write without
its password.`,
		Example: `  # Import configuration
  shelly device config import living-room config-backup.json

  # Dry run - show what would change without applying
  shelly device config import living-room config.json --dry-run`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.FilePath = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be changed without applying")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	// Read and parse file
	fileData, err := afero.ReadFile(config.Fs(), opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var deviceConfig map[string]any

	// Try JSON first, then YAML
	if err := json.Unmarshal(fileData, &deviceConfig); err != nil {
		if err := yaml.Unmarshal(fileData, &deviceConfig); err != nil {
			return fmt.Errorf("failed to parse file as JSON or YAML: %w", err)
		}
	}

	svc := opts.Factory.ShellyService()
	ios := opts.Factory.IOStreams()

	var changes, warnings []string
	msg := "Importing configuration..."
	if opts.DryRun {
		msg = "Comparing configurations..."
	}
	err = cmdutil.RunWithSpinner(ctx, ios, msg, func(ctx context.Context) error {
		var importErr error
		changes, warnings, importErr = svc.ImportConfig(ctx, opts.Device, deviceConfig, opts.DryRun)
		return importErr
	})
	if err != nil {
		return fmt.Errorf("failed to import configuration: %w", err)
	}
	for _, w := range warnings {
		ios.Warning("%s", w)
	}

	if opts.DryRun {
		if len(changes) == 0 {
			ios.Info("No changes would be made")
			return nil
		}
		ios.Title("Dry run - changes that would be applied")
		for _, change := range changes {
			ios.Printf("  %s\n", change)
		}
		return nil
	}

	ios.Success("Configuration imported to %s", opts.Device)
	return nil
}
