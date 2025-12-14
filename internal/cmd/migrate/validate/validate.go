// Package validate provides the migrate validate subcommand.
package validate

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the migrate validate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate <backup-file>",
		Aliases: []string{"check", "verify"},
		Short:   "Validate a backup file",
		Long: `Validate a backup file for structural integrity.

Checks that the backup file is properly formatted and contains
all required fields.`,
		Example: `  # Validate a backup file
  shelly migrate validate backup.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(f, args[0])
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, filePath string) error {
	ios := f.IOStreams()

	// Read backup file
	data, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is user-provided CLI argument
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Validate backup
	backup, err := shelly.ValidateBackup(data)
	if err != nil {
		ios.Error("Validation failed: %v", err)
		return err
	}

	ios.Success("Backup file is valid")
	ios.Println()
	ios.Printf("  Version:   %d\n", backup.Version)
	ios.Printf("  Device:    %s (%s)\n", backup.Device().ID, backup.Device().Model)
	ios.Printf("  Firmware:  %s\n", backup.Device().FWVersion)
	ios.Printf("  Created:   %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
	ios.Printf("  Config:    %d keys\n", len(backup.Config))
	ios.Printf("  Scripts:   %d\n", len(backup.Scripts))
	ios.Printf("  Schedules: %d\n", len(backup.Schedules))
	ios.Printf("  Webhooks:  %d\n", len(backup.Webhooks))

	if backup.Encrypted() {
		ios.Printf("  Encrypted: yes\n")
	}

	return nil
}
