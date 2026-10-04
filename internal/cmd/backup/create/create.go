// Package create provides the backup create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory       *cmdutil.Factory
	All           bool
	Device        string
	Dir           string
	Encrypt       string
	EncryptStdin  bool
	FilePath      string
	SkipScripts   bool
	SkipSchedules bool
	SkipWebhooks  bool
}

// NewCommand creates the backup create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "create [device] [file]",
		Aliases: []string{"new", "make"},
		Short:   "Create a device backup",
		Long: `Create a complete backup of a Shelly device, or of every registered
device with --all.

The backup includes configuration, scripts, schedules, and webhooks.
Backups are written as JSON. If no file is specified, the backup is saved
to ~/.config/shelly/backups/ (or --dir) with a name based on the device,
its MAC address and the date. Use "-" as the file to write to stdout.

With --all, every registered device is backed up to its own auto-named
file. A device that fails does not stop the others; each device's result
and a summary are printed, and the command exits non-zero when any device
failed. --encrypt and the --skip-* flags apply to every device.

Use --encrypt to AES-encrypt the backup with a password; restore the file
with 'shelly backup restore <device> <file> --decrypt <password>'.`,
		Example: `  # Create backup (auto-saved to ~/.config/shelly/backups/)
  shelly backup create living-room

  # Create backup to specific file
  shelly backup create living-room backup.json

  # Create auto-named backup in a directory
  shelly backup create living-room --dir ./backups

  # Back up every registered device
  shelly backup create --all

  # Back up every registered device into a dated directory, quietly
  shelly backup create --all --dir ./backups/$(date +%Y-%m-%d) -q

  # Create backup to stdout
  shelly backup create living-room -

  # Create encrypted backup
  shelly backup create living-room backup.json --encrypt mysecret

  # Read the encryption password from stdin, keeping it out of shell history
  shelly backup create living-room backup.json --encrypt-stdin < ~/.shelly-backup-password

  # Skip scripts in backup
  shelly backup create living-room backup.json --skip-scripts`,
		Annotations: cmdutil.DashIsOutputAnnotation(),
		Args:        cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case opts.All && len(args) > 0:
				return fmt.Errorf("--all backs up every registered device; do not pass a device or file")
			case !opts.All && len(args) == 0:
				return fmt.Errorf("requires a device argument, or --all")
			case len(args) == 2 && opts.Dir != "":
				return fmt.Errorf("--dir sets the folder for an auto-named backup; do not pass a file as well")
			}
			if len(args) > 0 {
				opts.Device = args[0]
			}
			if len(args) > 1 {
				opts.FilePath = args[1]
			}
			return run(cmd.Context(), opts)
		},
	}

	flags.AddAllFlag(cmd, &opts.All)
	cmd.Flags().StringVar(&opts.Dir, "dir", "", "Directory for auto-named backups (default ~/.config/shelly/backups/, created if missing)")
	cmdutil.AddSecretFlagsP(cmd, &opts.Encrypt, &opts.EncryptStdin, "encrypt", "e",
		"Password to AES-encrypt the backup", "Read the password to AES-encrypt the backup from stdin")
	cmd.Flags().BoolVar(&opts.SkipScripts, "skip-scripts", false, "Exclude scripts from backup")
	cmd.Flags().BoolVar(&opts.SkipSchedules, "skip-schedules", false, "Exclude schedules from backup")
	cmd.Flags().BoolVar(&opts.SkipWebhooks, "skip-webhooks", false, "Exclude webhooks from backup")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	if err := cmdutil.ResolveSecret(ios, &opts.Encrypt, opts.EncryptStdin, "encrypt", "Backup password"); err != nil {
		return err
	}
	svc := opts.Factory.ShellyService()

	backupOpts := backup.Options{
		SkipScripts:   opts.SkipScripts,
		SkipSchedules: opts.SkipSchedules,
		SkipWebhooks:  opts.SkipWebhooks,
	}

	if opts.All {
		cfg, err := opts.Factory.Config()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		dir, err := backup.EnsureDir(opts.Dir)
		if err != nil {
			return err
		}

		ios.StartProgress(fmt.Sprintf("Backing up %d devices...", len(cfg.Devices)))
		results := shelly.NewBackupExporter(svc).ExportAll(ctx, cfg.Devices, shelly.BackupExportOptions{
			Directory:  dir,
			Parallel:   config.GetGlobalMaxConcurrent(),
			BackupOpts: backupOpts,
			Encrypt:    opts.Encrypt,
			AutoName:   true,
		})
		ios.StopProgress()

		term.DisplayBackupCreateResults(ios, results, dir)
		if _, failed := shelly.CountBackupResults(results); failed > 0 {
			return fmt.Errorf("%d of %d device backups failed", failed, len(results))
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*3)
	defer cancel()

	var bkp *backup.DeviceBackup
	err := cmdutil.RunWithSpinner(ctx, ios, "Creating backup...", func(ctx context.Context) error {
		var err error
		bkp, err = svc.CreateBackup(ctx, opts.Device, backupOpts)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	data, err := export.EncodeBackup(bkp, opts.Encrypt)
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	if opts.FilePath == "-" {
		ios.Printf("%s\n", data)
		return nil
	}

	if opts.FilePath == "" {
		autoPath, pathErr := backup.AutoSavePathIn(opts.Dir, opts.Device, bkp, "json")
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
