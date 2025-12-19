// Package list provides the backup list subcommand.
package list

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the backup list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [directory]",
		Aliases: []string{"ls"},
		Short:   "List saved backups",
		Long: `List backup files in a directory.

By default, looks in the config directory's backups folder. Backup files
contain full device configuration snapshots that can be used to restore
device settings or migrate configurations between devices.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: Filename, Device, Model, Created, Encrypted, Size`,
		Example: `  # List backups in default location
  shelly backup list

  # List backups in specific directory
  shelly backup list /path/to/backups

  # Output as JSON
  shelly backup list -o json

  # Find backups for a specific device model
  shelly backup list -o json | jq '.[] | select(.device_model | contains("Plus"))'

  # Get most recent backup filename
  shelly backup list -o json | jq -r 'sort_by(.created_at) | last | .filename'

  # Short form
  shelly backup ls`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			dir := ""
			if len(args) > 0 {
				dir = args[0]
			}
			return run(f, dir)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, dir string) error {
	ios := f.IOStreams()

	// Resolve directory
	if dir == "" {
		configDir, err := config.Dir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}
		dir = filepath.Join(configDir, "backups")
	}

	// Validate directory exists
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		ios.Info("No backups directory found at %s", dir)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Scan for backup files
	backups, err := export.ScanBackupFiles(dir)
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		ios.Info("No backup files found in %s", dir)
		return nil
	}

	// Handle structured output (JSON/YAML) via global -o flag
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, backups)
	}

	// Default table output
	term.DisplayBackupsTable(ios, backups)
	return nil
}
