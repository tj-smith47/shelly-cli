// Package create provides the backup create subcommand.
package create

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

var (
	formatFlag        string
	encryptFlag       string
	skipScriptsFlag   bool
	skipSchedulesFlag bool
	skipWebhooksFlag  bool
)

// NewCommand creates the backup create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <device> [file]",
		Aliases: []string{"new", "make"},
		Short:   "Create a device backup",
		Long: `Create a complete backup of a Shelly device.

The backup includes configuration, scripts, schedules, and webhooks.
If no file is specified, backup is written to stdout.

Use --encrypt to password-protect the backup (password verification only,
sensitive data is not encrypted in the file).`,
		Example: `  # Create backup to file
  shelly backup create living-room backup.json

  # Create YAML backup
  shelly backup create living-room backup.yaml --format yaml

  # Create backup to stdout
  shelly backup create living-room

  # Create encrypted backup
  shelly backup create living-room backup.json --encrypt mysecret

  # Skip scripts in backup
  shelly backup create living-room backup.json --skip-scripts`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]
			filePath := ""
			if len(args) > 1 {
				filePath = args[1]
			}
			return run(cmd.Context(), f, device, filePath)
		},
	}

	cmd.Flags().StringVarP(&formatFlag, "format", "f", "json", "Output format (json, yaml)")
	cmd.Flags().StringVarP(&encryptFlag, "encrypt", "e", "", "Password to protect backup")
	cmd.Flags().BoolVar(&skipScriptsFlag, "skip-scripts", false, "Exclude scripts from backup")
	cmd.Flags().BoolVar(&skipSchedulesFlag, "skip-schedules", false, "Exclude schedules from backup")
	cmd.Flags().BoolVar(&skipWebhooksFlag, "skip-webhooks", false, "Exclude webhooks from backup")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*3)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	opts := shelly.BackupOptions{
		SkipScripts:   skipScriptsFlag,
		SkipSchedules: skipSchedulesFlag,
		SkipWebhooks:  skipWebhooksFlag,
		Password:      encryptFlag,
	}

	var backup *shelly.DeviceBackup
	err := cmdutil.RunWithSpinner(ctx, ios, "Creating backup...", func(ctx context.Context) error {
		var err error
		backup, err = svc.CreateBackup(ctx, device, opts)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Format the output
	data, err := export.MarshalBackup(backup, formatFlag)
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	// Write to file or stdout
	if filePath == "" || filePath == "-" {
		ios.Printf("%s\n", data)
	} else {
		if err := os.WriteFile(filePath, data, 0o600); err != nil {
			return fmt.Errorf("failed to write backup file: %w", err)
		}
		ios.Success("Backup created: %s", filePath)
		term.DisplayBackupSummary(ios, backup)
	}

	return nil
}
