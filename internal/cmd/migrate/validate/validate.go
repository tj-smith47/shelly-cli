// Package validate provides the migrate validate subcommand.
package validate

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// Options holds the options for the validate command.
type Options struct {
	Factory  *cmdutil.Factory
	FilePath string
}

// NewCommand creates the migrate validate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.FilePath = args[0]
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Read backup file
	data, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Validate backup
	bkp, err := backup.Validate(data)
	if err != nil {
		ios.Error("Validation failed: %v", err)
		return err
	}

	ios.Success("Backup file is valid")
	ios.Println()
	ios.Printf("  Version:   %d\n", bkp.Version)
	ios.Printf("  Device:    %s (%s)\n", bkp.Device().ID, bkp.Device().Model)
	ios.Printf("  Firmware:  %s\n", bkp.Device().FWVersion)
	ios.Printf("  Created:   %s\n", bkp.CreatedAt.Format("2006-01-02 15:04:05"))
	ios.Printf("  Config:    %d keys\n", len(bkp.Config))
	ios.Printf("  Scripts:   %d\n", len(bkp.Scripts))
	ios.Printf("  Schedules: %d\n", len(bkp.Schedules))
	ios.Printf("  Webhooks:  %d\n", len(bkp.Webhooks))

	if bkp.Encrypted() {
		ios.Printf("  Encrypted: yes\n")
	}

	return nil
}
