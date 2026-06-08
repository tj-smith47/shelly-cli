// Package migrate provides migration commands.
package migrate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate/diff"
	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate/validate"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// aliasMig is the migrate command alias.
const aliasMig = "mig"

// Options holds command options.
type Options struct {
	Factory       *cmdutil.Factory
	Source        string
	Target        string
	DryRun        bool
	Force         bool
	Yes           bool
	ResetSource   bool
	SkipAuth      bool
	SkipNetwork   bool
	SkipScripts   bool
	SkipSchedules bool
	SkipWebhooks  bool
	StaticIP      string
	Gateway       string
	Netmask       string
	DNS           string
	Name          string

	// resetSourceExplicit tracks whether --reset-source was explicitly set.
	resetSourceExplicit bool
}

// NewCommand creates the migrate command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "migrate <source-device> <target-device>",
		Aliases: []string{aliasMig},
		Short:   "Migrate configuration between devices",
		Long: `Migrate configuration from one Shelly device to another.

Reads the current configuration from the source device and applies it to
the target device. By default, everything is migrated including network
and authentication settings.

When network settings are migrated, the source device is factory reset
after a successful migration to prevent IP conflicts on the network.
Use --skip-network to keep both devices online with their current
network settings, or --reset-source=false to skip the factory reset
(warning: this may cause IP conflicts).

Use --dry-run to preview what would change without applying.`,
		Example: `  # Preview migration (dry run)
  shelly migrate living-room bedroom --dry-run

  # Full migration (factory resets source after)
  shelly migrate living-room bedroom --yes

  # Migrate without network config (no factory reset needed)
  shelly migrate living-room bedroom --skip-network

  # Migrate network but skip factory reset (may cause IP conflict)
  shelly migrate living-room bedroom --reset-source=false

  # Force migration between different device types
  shelly migrate living-room bedroom --force --yes

  # Clone config onto a new bulb with a distinct static IP (keeps both online,
  # source is not reset since there is no IP conflict)
  shelly migrate master-bath-1 new-bulb \
    --static-ip 10.23.47.221 --gateway 10.23.47.1 --netmask 255.255.254.0`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Source = args[0]
			opts.Target = args[1]
			opts.resetSourceExplicit = cmd.Flags().Changed("reset-source")
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be changed without applying")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force migration between different device types")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.ResetSource, "reset-source", true, "Factory reset source device after migration")
	cmd.Flags().BoolVar(&opts.SkipAuth, "skip-auth", false, "Skip authentication configuration")
	cmd.Flags().BoolVar(&opts.SkipNetwork, "skip-network", false, "Skip network configuration (WiFi, Ethernet)")
	cmd.Flags().BoolVar(&opts.SkipScripts, "skip-scripts", false, "Skip script migration")
	cmd.Flags().BoolVar(&opts.SkipSchedules, "skip-schedules", false, "Skip schedule migration")
	cmd.Flags().BoolVar(&opts.SkipWebhooks, "skip-webhooks", false, "Skip webhook migration")
	cmd.Flags().StringVar(&opts.StaticIP, "static-ip", "", "Assign this static IPv4 to the target instead of copying the source's IP")
	cmd.Flags().StringVar(&opts.Gateway, "gateway", "", "Static IPv4 default gateway (with --static-ip)")
	cmd.Flags().StringVar(&opts.Netmask, "netmask", "", "Static IPv4 subnet mask (with --static-ip)")
	cmd.Flags().StringVar(&opts.DNS, "dns", "", "Static IPv4 nameserver (optional, with --static-ip)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Override the target device name (defaults to the target identifier when it is a friendly alias)")
	cmd.MarkFlagsRequiredTogether("static-ip", "gateway", "netmask")

	cmd.AddCommand(validate.NewCommand(f))
	cmd.AddCommand(diff.NewCommand(f))

	return cmd
}

// shouldResetSource determines whether the source device should be factory reset.
// If the user explicitly set --reset-source, use that value.
// Otherwise, auto-compute: reset when network is being migrated.
func (o *Options) shouldResetSource() bool {
	if o.resetSourceExplicit {
		return o.ResetSource
	}
	// A static-IP override gives the target a distinct address, so the source
	// keeps its own IP without conflict and need not be reset.
	if o.StaticIP != "" {
		return false
	}
	return !o.SkipNetwork
}

// previewMigration renders the dry-run preview, noting any network override and
// whether the source will be factory reset.
func (o *Options) previewMigration(ctx context.Context, bkp *backup.DeviceBackup, override *backup.NetworkOverride) error {
	ios := o.Factory.IOStreams()
	d, err := o.Factory.ShellyService().CompareBackup(ctx, o.Target, bkp)
	if err != nil {
		return fmt.Errorf("failed to compare: %w", err)
	}
	term.DisplayMigrationPreview(ios, o.Source, string(backup.SourceDevice), o.Target, d)
	if override != nil {
		ios.Info("Target %q will get static IP %s; source %q keeps its own address", o.Target, override.StaticIP, o.Source)
	}
	if o.shouldResetSource() {
		ios.Warning("Source device %q will be factory reset after migration", o.Source)
	}
	return nil
}

// networkOverride builds a NetworkOverride from the static-IP flags, or nil when
// no static IP was requested.
func (o *Options) networkOverride() *backup.NetworkOverride {
	if o.StaticIP == "" {
		return nil
	}
	return &backup.NetworkOverride{
		StaticIP: o.StaticIP,
		Gateway:  o.Gateway,
		Netmask:  o.Netmask,
		DNS:      o.DNS,
	}
}

// confirmMigration prompts the user for confirmation unless --yes was passed.
// Returns (true, nil) to proceed, (false, nil) if cancelled.
func (o *Options) confirmMigration(resetSource bool) (bool, error) {
	if o.Yes {
		return true, nil
	}
	msg := fmt.Sprintf("Migrate %q -> %q", o.Source, o.Target)
	if resetSource {
		msg += fmt.Sprintf(" (source %q will be factory reset)", o.Source)
	}
	confirmed, err := o.Factory.ConfirmAction(msg+"?", false)
	if err != nil {
		return false, fmt.Errorf("confirmation failed: %w", err)
	}
	if !confirmed {
		o.Factory.IOStreams().Info("Migration cancelled")
	}
	return confirmed, nil
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*5)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	override := opts.networkOverride()
	if override != nil && opts.SkipNetwork {
		return fmt.Errorf("--static-ip cannot be used with --skip-network")
	}

	// Back up source device
	var bkp *backup.DeviceBackup
	err := cmdutil.RunWithSpinner(ctx, ios, "Reading source device...", func(ctx context.Context) error {
		var backupErr error
		bkp, backupErr = svc.CreateBackup(ctx, opts.Source, backup.Options{})
		return backupErr
	})
	if err != nil {
		return fmt.Errorf("failed to read source device: %w", err)
	}

	// Check target device compatibility
	if err := svc.CheckMigrationCompatibility(ctx, bkp, opts.Target, opts.Force); err != nil {
		term.DisplayCompatibilityError(ios, err)
		return err
	}

	if opts.DryRun {
		return opts.previewMigration(ctx, bkp, override)
	}

	resetSource := opts.shouldResetSource()

	// Warn about IP conflict if migrating network without resetting source.
	// A static-IP override gives the target a distinct address, so no conflict.
	if !opts.SkipNetwork && !resetSource && override == nil {
		ios.Warning("Migrating network settings without factory-resetting the source device")
		ios.Warning("This may cause IP conflicts on your network")
	}

	// Confirm before proceeding
	if confirmed, err := opts.confirmMigration(resetSource); err != nil || !confirmed {
		return err
	}

	// Perform migration
	restoreOpts := backup.RestoreOptions{
		SkipAuth:        opts.SkipAuth,
		SkipNetwork:     opts.SkipNetwork,
		SkipScripts:     opts.SkipScripts,
		SkipSchedules:   opts.SkipSchedules,
		SkipWebhooks:    opts.SkipWebhooks,
		NetworkOverride: override,
		Name:            cmdutil.DeviceDisplayName(opts.Name, opts.Target),
	}
	var result *backup.RestoreResult
	err = cmdutil.RunWithSpinner(ctx, ios, "Migrating configuration...", func(ctx context.Context) error {
		var restoreErr error
		result, restoreErr = svc.RestoreBackup(ctx, opts.Target, bkp, restoreOpts)
		return restoreErr
	})
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	ios.Success("Migration completed!")
	term.DisplayMigrationResult(ios, result)

	// Factory reset source device if needed
	if resetSource {
		err = cmdutil.RunWithSpinner(ctx, ios, "Factory resetting source device...", func(ctx context.Context) error {
			return svc.DeviceFactoryReset(ctx, opts.Source)
		})
		if err != nil {
			ios.Warning("Migration succeeded but factory reset of source failed: %v", err)
			ios.Info("You may need to manually factory reset %q to avoid IP conflicts", opts.Source)
			return nil
		}
		ios.Success("Source device %q has been factory reset", opts.Source)
	}

	return nil
}
