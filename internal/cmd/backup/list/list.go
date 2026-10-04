// Package list provides the backup list subcommand.
package list

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the options for the list command.
type Options struct {
	Factory *cmdutil.Factory
	Dir     string
}

// NewCommand creates the backup list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			if len(args) > 0 {
				opts.Dir = args[0]
			}
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Resolve directory
	dir := opts.Dir
	if dir == "" {
		configDir, err := config.Dir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}
		dir = filepath.Join(configDir, "backups")
	}

	var backups []model.BackupFileInfo
	emptyMsg := fmt.Sprintf("No backup files found in %s", dir)
	info, err := config.Fs().Stat(dir)
	switch {
	case os.IsNotExist(err):
		emptyMsg = fmt.Sprintf("No backups directory found at %s", dir)
	case err != nil:
		return fmt.Errorf("failed to access directory: %w", err)
	case !info.IsDir():
		return fmt.Errorf("%s is not a directory", dir)
	default:
		if backups, err = export.ScanBackupFiles(dir); err != nil {
			return err
		}
	}

	return cmdutil.PrintList(ios, backups, term.DisplayBackupsTable, func() {
		ios.Info("%s", emptyMsg)
	})
}
