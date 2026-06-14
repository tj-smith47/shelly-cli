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
	SkipState     bool
	SkipMeters    bool
	StaticIP      string
	Gateway       string
	Netmask       string
	DNS           string
	Name          string
	ToAP          string
	APIP          string
	SSID          string
	Password      string

	AllowFirmwareDowngrade bool
	FirmwareURL            string

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
    --static-ip 10.23.47.221 --gateway 10.23.47.1 --netmask 255.255.254.0

  # Clone a live sibling straight onto a brand-new device at its factory WiFi AP:
  # hops the host onto the AP, applies the config + static IP, the device joins
  # the LAN, and the source is left untouched (target name = "fr")
  shelly migrate sr fr --to-ap ShellyBulbDuo-D0DCFF \
    --static-ip 10.23.47.227 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1`,
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
	cmd.Flags().BoolVar(&opts.SkipState, "skip-state", false, "Skip migrating live component state (color temperature, brightness); apply configuration only")
	cmd.Flags().BoolVar(&opts.SkipMeters, "skip-meters", false, "Skip migrating meter/energy-meter configuration (e.g. overpower limits)")
	cmd.Flags().StringVar(&opts.StaticIP, "static-ip", "", "Assign this static IPv4 to the target instead of copying the source's IP")
	cmd.Flags().StringVar(&opts.Gateway, "gateway", "", "Static IPv4 default gateway (with --static-ip)")
	cmd.Flags().StringVar(&opts.Netmask, "netmask", "", "Static IPv4 subnet mask (with --static-ip)")
	cmd.Flags().StringVar(&opts.DNS, "dns", "", "Static IPv4 nameserver (optional, with --static-ip)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Override the target device name (defaults to the target identifier when it is a friendly alias)")
	cmd.Flags().StringVar(&opts.ToAP, "to-ap", "", "Migrate onto a target at its factory WiFi AP with this SSID (hops host WiFi; source is never reset)")
	cmd.Flags().StringVar(&opts.APIP, "ap-ip", "", "Static host IP to use on the target's AP subnet during --to-ap (default 192.168.33.133)")
	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "Override the WiFi SSID the target joins (defaults to the source's network)")
	cmd.Flags().StringVar(&opts.Password, "password", "", "WiFi passphrase for the target network (optional: derived from this host's stored credentials when omitted; set to override or when derivation fails)")
	cmd.Flags().BoolVar(&opts.AllowFirmwareDowngrade, "allow-firmware-downgrade", false, "Force the older-firmware config write instead of the automatic firmware update (Gen1; the target is updated to matched firmware by default when the source is newer — this skips that and accepts the reboot-loop risk)")
	cmd.Flags().StringVar(&opts.FirmwareURL, "firmware-url", "", "Firmware image for the automatic downgrade-recovery update (default: derived from the source device model)")
	cmd.MarkFlagsRequiredTogether("static-ip", "gateway", "netmask")

	cmd.AddCommand(validate.NewCommand(f))
	cmd.AddCommand(diff.NewCommand(f))

	return cmd
}

// shouldResetSource determines whether the source device should be factory reset.
// If the user explicitly set --reset-source, use that value.
// Otherwise, auto-compute: reset when network is being migrated.
func (o *Options) shouldResetSource() bool {
	// A --to-ap target is a different physical device reached through its own AP,
	// so the source is never reset (it keeps its address and stays online).
	if o.ToAP != "" {
		return false
	}
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

// validateFlags rejects incompatible flag combinations before any device I/O.
func (o *Options) validateFlags(override *backup.NetworkOverride) error {
	if override != nil && o.SkipNetwork {
		return fmt.Errorf("--static-ip cannot be used with --skip-network")
	}
	if o.ToAP != "" {
		if o.SkipNetwork {
			return fmt.Errorf("--to-ap cannot be used with --skip-network (the device needs WiFi to leave its AP)")
		}
		if o.DryRun {
			return fmt.Errorf("--to-ap cannot be combined with --dry-run (the target is not reachable until it joins the network)")
		}
	}
	if o.APIP != "" && o.ToAP == "" {
		return fmt.Errorf("--ap-ip only applies with --to-ap")
	}
	return nil
}

// networkOverride builds a NetworkOverride from the static-IP flags, or nil when
// no static IP was requested.
func (o *Options) networkOverride() *backup.NetworkOverride {
	if o.StaticIP == "" && o.SSID == "" && o.Password == "" {
		return nil
	}
	return &backup.NetworkOverride{
		SSID:     o.SSID,
		Password: o.Password,
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

// migrateViaAP clones the source backup onto a target sitting at its factory
// WiFi AP, moving it onto the LAN in one step. Network settings are always
// applied (they are what take the device off its AP), the source is never reset
// (the target is a different physical device), and compatibility is not
// pre-checked since the target is unreachable until it joins the network.
func (o *Options) migrateViaAP(
	ctx context.Context,
	svc *shelly.Service,
	bkp *backup.DeviceBackup,
	override *backup.NetworkOverride,
) error {
	ios := o.Factory.IOStreams()

	if confirmed, err := o.confirmMigration(false); err != nil || !confirmed {
		return err
	}

	restoreOpts := backup.RestoreOptions{
		SkipAuth:        o.SkipAuth,
		SkipScripts:     o.SkipScripts,
		SkipSchedules:   o.SkipSchedules,
		SkipWebhooks:    o.SkipWebhooks,
		SkipState:       o.SkipState,
		SkipMeters:      o.SkipMeters,
		NetworkOverride: override,
		Name:            cmdutil.DeviceDisplayName(o.Name, o.Target),

		AllowFirmwareDowngrade: o.AllowFirmwareDowngrade,
		FirmwareURL:            o.FirmwareURL,
	}

	var (
		result  *backup.RestoreResult
		newAddr string
	)
	err := cmdutil.RunWithSpinner(ctx, ios,
		fmt.Sprintf("Restoring onto %s at AP %s (hopping host WiFi)...", o.Target, o.ToAP),
		func(ctx context.Context) error {
			var restoreErr error
			result, newAddr, restoreErr = svc.RestoreToAP(ctx, o.ToAP, o.APIP, o.Target, bkp, restoreOpts)
			return restoreErr
		})
	if err != nil {
		return fmt.Errorf("migration via AP failed: %w", err)
	}

	ios.Success("Migration completed!")
	if newAddr != "" {
		ios.Info("%s is live at %s", o.Target, newAddr)
	}
	term.DisplayMigrationResult(ios, result)
	return nil
}

func run(ctx context.Context, opts *Options) error {
	// --to-ap performs a WiFi hop, a full restore at the AP, and a LAN-rejoin
	// poll, which together far exceed a normal migration's budget.
	timeout := shelly.DefaultTimeout * 5
	if opts.ToAP != "" {
		timeout = shelly.DefaultTimeout * 30
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	override := opts.networkOverride()
	if err := opts.validateFlags(override); err != nil {
		return err
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

	// --to-ap: target sits at its factory AP, unreachable until provisioned, so
	// the on-network compatibility/dry-run paths are skipped; the Gen-aware
	// restore handles the device directly at the AP.
	if opts.ToAP != "" {
		return opts.migrateViaAP(ctx, svc, bkp, override)
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
		SkipState:       opts.SkipState,
		SkipMeters:      opts.SkipMeters,
		NetworkOverride: override,
		Name:            cmdutil.DeviceDisplayName(opts.Name, opts.Target),

		AllowFirmwareDowngrade: opts.AllowFirmwareDowngrade,
		FirmwareURL:            opts.FirmwareURL,
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
