// Package create provides the backup create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory       *cmdutil.Factory
	Device        string
	Encrypt       string
	FilePath      string
	SkipScripts   bool
	SkipSchedules bool
	SkipWebhooks  bool
}

// NewCommand creates the backup create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "create <device> [file]",
		Aliases: []string{"new", "make"},
		Short:   "Create a device backup",
		Long: `Create a complete backup of a Shelly device.

The backup includes configuration, scripts, schedules, and webhooks.
Backups are written as JSON. If no file is specified, the backup is saved
to ~/.config/shelly/backups/ with a name based on the device and date. Use
"-" as the file to write to stdout.

Use --encrypt to AES-encrypt the backup with a password; restore the file
with 'shelly backup restore --decrypt <password>'.`,
		Example: `  # Create backup (auto-saved to ~/.config/shelly/backups/)
  shelly backup create living-room

  # Create backup to specific file
  shelly backup create living-room backup.json

  # Create backup to stdout
  shelly backup create living-room -

  # Create encrypted backup
  shelly backup create living-room backup.json --encrypt mysecret

  # Skip scripts in backup
  shelly backup create living-room backup.json --skip-scripts`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if len(args) > 1 {
				opts.FilePath = args[1]
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Encrypt, "encrypt", "e", "", "Password to AES-encrypt the backup")
	cmd.Flags().BoolVar(&opts.SkipScripts, "skip-scripts", false, "Exclude scripts from backup")
	cmd.Flags().BoolVar(&opts.SkipSchedules, "skip-schedules", false, "Exclude schedules from backup")
	cmd.Flags().BoolVar(&opts.SkipWebhooks, "skip-webhooks", false, "Exclude webhooks from backup")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*3)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	backupOpts := backup.Options{
		SkipScripts:   opts.SkipScripts,
		SkipSchedules: opts.SkipSchedules,
		SkipWebhooks:  opts.SkipWebhooks,
	}

	var bkp *backup.DeviceBackup
	err := cmdutil.RunWithSpinner(ctx, ios, "Creating backup...", func(ctx context.Context) error {
		var err error
		bkp, err = svc.CreateBackup(ctx, opts.Device, backupOpts)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// --encrypt wraps the plaintext backup in an AES-256-GCM envelope; otherwise
	// the backup is written as plaintext JSON.
	var data []byte
	if opts.Encrypt != "" {
		data, err = backup.Encrypt(bkp, opts.Encrypt)
	} else {
		data, err = export.MarshalBackup(bkp)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	// Write to stdout if "-"
	if opts.FilePath == "-" {
		ios.Printf("%s\n", data)
		return nil
	}

	// Auto-generate file path if not specified
	if opts.FilePath == "" {
		autoPath, pathErr := backup.AutoSavePath(opts.Device, bkp, "json")
		if pathErr != nil {
			return pathErr
		}
		opts.FilePath = autoPath
	}

	if err := afero.WriteFile(config.Fs(), opts.FilePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}
	ios.Success("Backup created: %s", opts.FilePath)
	term.DisplayBackupSummary(ios, bkp)

	return nil
}
