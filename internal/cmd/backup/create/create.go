// Package create provides the backup create subcommand.
package create

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
		Use:   "create <device> [file]",
		Short: "Create a device backup",
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

	ios.StartProgress("Creating backup...")

	backup, err := svc.CreateBackup(ctx, device, opts)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Format the output
	var data []byte
	format := strings.ToLower(formatFlag)
	switch format {
	case "yaml", "yml":
		data, err = yaml.Marshal(backup)
	default:
		data, err = json.MarshalIndent(backup, "", "  ")
	}
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
		printBackupSummary(ios, backup)
	}

	return nil
}

func printBackupSummary(ios *iostreams.IOStreams, backup *shelly.DeviceBackup) {
	ios.Println()
	ios.Printf("  Device:    %s (%s)\n", backup.Device().ID, backup.Device().Model)
	ios.Printf("  Firmware:  %s\n", backup.Device().FWVersion)
	ios.Printf("  Config:    %d keys\n", len(backup.Config))
	if len(backup.Scripts) > 0 {
		ios.Printf("  Scripts:   %d\n", len(backup.Scripts))
	}
	if len(backup.Schedules) > 0 {
		ios.Printf("  Schedules: %d\n", len(backup.Schedules))
	}
	if len(backup.Webhooks) > 0 {
		ios.Printf("  Webhooks:  %d\n", len(backup.Webhooks))
	}
	if backup.Encrypted() {
		ios.Printf("  Encrypted: yes\n")
	}
}
