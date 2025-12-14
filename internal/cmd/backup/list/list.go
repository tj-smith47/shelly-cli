// Package list provides the backup list subcommand.
package list

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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

// backupFileInfo holds information about a backup file.
type backupFileInfo struct {
	Filename    string `json:"filename"`
	DeviceID    string `json:"device_id"`
	DeviceModel string `json:"device_model"`
	FWVersion   string `json:"fw_version"`
	CreatedAt   string `json:"created_at"`
	Encrypted   bool   `json:"encrypted"`
	Size        int64  `json:"size"`
}

func run(f *cmdutil.Factory, dir string) error {
	ios := f.IOStreams()

	dir, err := resolveBackupDir(dir)
	if err != nil {
		return err
	}

	exists, err := validateDirectory(ios, dir)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	backups, err := scanBackupFiles(dir)
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		ios.Info("No backup files found in %s", dir)
		return nil
	}

	return outputBackups(ios, backups)
}

func resolveBackupDir(dir string) (string, error) {
	if dir != "" {
		return dir, nil
	}
	configDir, err := config.Dir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}
	return filepath.Join(configDir, "backups"), nil
}

func validateDirectory(ios *iostreams.IOStreams, dir string) (bool, error) {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		ios.Info("No backups directory found at %s", dir)
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s is not a directory", dir)
	}
	return true, nil
}

func scanBackupFiles(dir string) ([]backupFileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	backups := make([]backupFileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isBackupFile(name) {
			continue
		}

		filePath := filepath.Join(dir, name)
		info, err := parseBackupFile(filePath)
		if err != nil {
			continue
		}
		info.Filename = name
		backups = append(backups, info)
	}
	return backups, nil
}

func isBackupFile(name string) bool {
	return strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

func outputBackups(ios *iostreams.IOStreams, backups []backupFileInfo) error {
	// Handle structured output (JSON/YAML) via global -o flag
	if cmdutil.WantsStructured() {
		return cmdutil.FormatOutput(ios, backups)
	}

	// Default table output
	printBackupsTable(ios, backups)
	return nil
}

func parseBackupFile(filePath string) (backupFileInfo, error) {
	var info backupFileInfo

	data, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is derived from directory listing
	if err != nil {
		return info, err
	}

	backup, err := shelly.ValidateBackup(data)
	if err != nil {
		return info, err
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return info, err
	}

	info.DeviceID = backup.Device().ID
	info.DeviceModel = backup.Device().Model
	info.FWVersion = backup.Device().FWVersion
	info.CreatedAt = backup.CreatedAt.Format("2006-01-02 15:04:05")
	info.Encrypted = backup.Encrypted()
	info.Size = stat.Size()

	return info, nil
}

func printBackupsTable(ios *iostreams.IOStreams, backups []backupFileInfo) {
	data := make([][]string, len(backups))
	for i, b := range backups {
		encrypted := ""
		if b.Encrypted {
			encrypted = "yes"
		}
		data[i] = []string{
			b.Filename,
			b.DeviceID,
			b.DeviceModel,
			b.CreatedAt,
			encrypted,
			formatSize(b.Size),
		}
	}

	headers := []string{"FILENAME", "DEVICE", "MODEL", "CREATED", "ENCRYPTED", "SIZE"}
	if err := output.PrintTableTo(ios.Out, headers, data); err != nil {
		ios.DebugErr("print table", err)
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
