// Package migrate provides migration commands.
package migrate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate/diff"
	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate/validate"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// aliasMig is the migrate command alias.
const aliasMig = "mig"

// migrateService is the subset of *shelly.Service the migrate command drives.
// Depending on this narrow interface instead of the concrete service keeps the
// command testable with a stub: the concrete service reaches real devices (and,
// for --to-ap, hops the host's WiFi), which a unit test cannot exercise.
type migrateService interface {
	CreateBackup(ctx context.Context, identifier string, opts backup.Options) (*backup.DeviceBackup, error)
	CheckMigrationCompatibility(ctx context.Context, bkp *backup.DeviceBackup, target string, force bool) error
	CompareBackup(ctx context.Context, identifier string, deviceBackup *backup.DeviceBackup) (*model.BackupDiff, error)
	RestoreBackup(ctx context.Context, identifier string, deviceBackup *backup.DeviceBackup, opts backup.RestoreOptions) (*backup.RestoreResult, error)
	RestoreToAP(ctx context.Context, apSSID, apHostIP, registryName string, bkp *backup.DeviceBackup, opts backup.RestoreOptions) (*backup.RestoreResult, string, error)
	cmdutil.LANStationPlanner
	cmdutil.APStationPlanner
	cmdutil.RestoreNameDescriber
	DeviceFactoryReset(ctx context.Context, identifier string) error
}

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
	Open          bool

	AllowFirmwareDowngrade bool
	FirmwareURL            string

	// resetSourceExplicit tracks whether --reset-source was explicitly set.
	resetSourceExplicit bool

	// svc, when non-nil, overrides the service resolved from the Factory. It is the
	// test injection seam; production leaves it nil and uses Factory.ShellyService().
	svc migrateService
}

// service returns the injected migrateService when set, otherwise the concrete
// service from the Factory.
func (o *Options) service() migrateService {
	if o.svc != nil {
		return o.svc
	}
	return o.Factory.ShellyService()
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

The target's WiFi address: without --static-ip the target takes the source's
addressing as the source has it. A source with a static address hands that
address to the target; the source is then factory reset (the default), and
with --reset-source=false both devices hold the same address. --static-ip
gives the target its own address, with any of --gateway, --netmask and --dns
left off taken from the source. The WiFi station line printed before the
migration (and by --dry-run) names the address that will be written.

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

  # Give the target its own address; the gateway, netmask and DNS are the
  # source's
  shelly migrate master-bath-1 new-bulb --static-ip 10.23.47.221

  # Migrate onto a target that joins a network with no password
  shelly migrate guest-plug new-plug --ssid GuestWiFi --open

  # Clone config onto a new bulb with a distinct static IP (keeps both online,
  # source is not reset since there is no IP conflict)
  shelly migrate master-bath-1 new-bulb \
    --static-ip 10.23.47.221 --gateway 10.23.47.1 --netmask 255.255.254.0

  # Clone a live sibling straight onto a brand-new device at its factory WiFi AP:
  # hops the host onto the AP, applies the config + static IP, the device joins
  # the LAN, and the source is left untouched (target name = "fr")
  shelly migrate sr fr --to-ap ShellyBulbDuo-D0DCFF \
    --static-ip 10.23.47.227 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1

  # Preview that migration without hopping the host's WiFi
  shelly migrate sr fr --to-ap ShellyBulbDuo-D0DCFF --dry-run`,
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
	cmdutil.AddStaticIPFlags(cmd, &opts.StaticIP, &opts.Gateway, &opts.Netmask, &opts.DNS,
		"Assign this static IPv4 to the target instead of copying the source's IP (--gateway, --netmask and --dns default to the source device's)",
		"the source device's")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Set the target device name (default: the target's alias when the source is another device; a target migrated from its own backup keeps the name it recorded)")
	cmd.Flags().StringVar(&opts.ToAP, "to-ap", "", "Migrate onto a target at its factory WiFi AP with this SSID (hops host WiFi; source is never reset)")
	cmd.Flags().StringVar(&opts.APIP, "ap-ip", "", "Static host IP to use on the target's AP subnet during --to-ap (default 192.168.33.133)")
	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "Override the WiFi SSID the target joins (defaults to the source's network)")
	cmdutil.AddWiFiPasswordFlag(cmd, &opts.Password)
	cmdutil.AddOpenFlag(cmd, &opts.Open)
	cmd.Flags().BoolVar(&opts.AllowFirmwareDowngrade, "allow-firmware-downgrade", false, "Force the older-firmware config write instead of the automatic firmware update (Gen1; the target is updated to matched firmware by default when the source is newer — this skips that and accepts the reboot-loop risk)")
	cmd.Flags().StringVar(&opts.FirmwareURL, "firmware-url", "", "Firmware image for the automatic downgrade-recovery update (default: derived from the source device model)")

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

// previewMigration renders the dry-run preview, noting the target's static
// address, when one was given, the device name the target gets, and whether
// the source will be factory reset. A --to-ap target is unreachable until it
// joins the network, so its preview lists what the backup restores instead of
// a diff, and adds the station, firmware and access point hop plan.
func (o *Options) previewMigration(
	ctx context.Context, bkp *backup.DeviceBackup, static shellybackup.StaticNetwork, restoreOpts backup.RestoreOptions,
) error {
	ios := o.Factory.IOStreams()
	svc := o.service()
	if o.ToAP != "" {
		ios.Title("Migration Preview (dry run)")
		ios.Println()
		term.DisplayRestorePreview(ios, bkp, restoreOpts)
	} else {
		d, err := svc.CompareBackup(ctx, o.Target, bkp)
		if err != nil {
			return fmt.Errorf("failed to compare: %w", err)
		}
		term.DisplayMigrationPreview(ios, o.Source, string(backup.SourceDevice), o.Target, d)
		cmdutil.PrintRestoreName(ctx, ios, svc, o.Target, bkp, restoreOpts)
	}
	if static.IP != "" {
		ios.Info("Target %q will get static IP %s (gateway %s, netmask %s); source %q keeps its own address",
			o.Target, static.IP, static.Gateway, static.Netmask, o.Source)
	}
	if o.shouldResetSource() {
		ios.Warning("Source device %q will be factory reset after migration", o.Source)
	}
	if o.ToAP != "" {
		return cmdutil.PreviewAPRestore(ctx, ios, svc, o.ToAP, bkp, restoreOpts)
	}
	return nil
}

// validateFlags rejects incompatible flag combinations before any device I/O.
func (o *Options) validateFlags() error {
	if err := o.network().Validate(o.SkipNetwork); err != nil {
		return err
	}
	if o.ToAP != "" {
		if o.SkipNetwork {
			return fmt.Errorf("--to-ap cannot be used with --skip-network (the device needs WiFi to leave its AP)")
		}
	}
	if o.APIP != "" && o.ToAP == "" {
		return fmt.Errorf("--ap-ip only applies with --to-ap")
	}
	return nil
}

// network returns the WiFi override flags.
func (o *Options) network() *cmdutil.NetworkFlags {
	return &cmdutil.NetworkFlags{
		SSID:     o.SSID,
		Password: o.Password,
		Open:     o.Open,
		StaticIP: o.StaticIP,
		Gateway:  o.Gateway,
		Netmask:  o.Netmask,
		DNS:      o.DNS,
	}
}

// staticNetwork returns the static address the target gets, with any gateway,
// netmask or DNS left off the flags taken from the source device. Its IP is
// empty when --static-ip was not given.
func (o *Options) staticNetwork(bkp *backup.DeviceBackup) (shellybackup.StaticNetwork, error) {
	override := o.network().Override()
	if !override.IsStatic() {
		return shellybackup.StaticNetwork{}, nil
	}
	return bkp.StaticNetwork(override)
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

// restoreOptions builds the restore options for the target from the flags.
func (o *Options) restoreOptions(override *backup.NetworkOverride) backup.RestoreOptions {
	return backup.RestoreOptions{
		DryRun:          o.DryRun,
		SkipAuth:        o.SkipAuth,
		SkipNetwork:     o.SkipNetwork,
		SkipScripts:     o.SkipScripts,
		SkipSchedules:   o.SkipSchedules,
		SkipWebhooks:    o.SkipWebhooks,
		SkipState:       o.SkipState,
		SkipMeters:      o.SkipMeters,
		NetworkOverride: override,
		Name:            o.Name,
		AliasName:       cmdutil.DeviceDisplayName("", o.Target),

		AllowFirmwareDowngrade: o.AllowFirmwareDowngrade,
		FirmwareURL:            o.FirmwareURL,
	}
}

// migrateViaAP clones the source backup onto a target sitting at its factory
// WiFi AP, moving it onto the LAN in one step. Network settings are always
// applied (they are what take the device off its AP), the source is never reset
// (the target is a different physical device), and compatibility is not
// pre-checked since the target is unreachable until it joins the network.
func (o *Options) migrateViaAP(
	ctx context.Context,
	svc migrateService,
	bkp *backup.DeviceBackup,
	static shellybackup.StaticNetwork,
	override *backup.NetworkOverride,
) error {
	ios := o.Factory.IOStreams()
	restoreOpts := o.restoreOptions(override)
	if o.DryRun {
		return o.previewMigration(ctx, bkp, static, restoreOpts)
	}
	// The dry run refuses network flags that name no network, or a network
	// this host has no passphrase for; the real run refuses them the same way,
	// before the prompt and the hop.
	if _, err := svc.PlanAPStation(ctx, bkp, override); err != nil {
		return err
	}

	if confirmed, err := o.confirmMigration(false); err != nil || !confirmed {
		return err
	}

	// The AP-hop restore-and-report sequence (including partial-failure handling)
	// is shared with the restore command. No source reset on this path — the
	// target is a different physical device.
	return cmdutil.RestoreAtAP(ctx, ios, svc, o.ToAP, o.APIP, o.Target, bkp, restoreOpts,
		"migration via AP failed", term.ReportMigrationResult)
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
	svc := opts.service()

	if err := opts.validateFlags(); err != nil {
		return err
	}
	override := opts.network().Override()

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

	// An address the source cannot complete is refused before the target is
	// touched or the user is asked to confirm.
	static, err := opts.staticNetwork(bkp)
	if err != nil {
		return err
	}

	// --to-ap: target sits at its factory AP, unreachable until provisioned, so
	// the on-network compatibility check and diff are skipped; the Gen-aware
	// restore handles the device directly at the AP.
	if opts.ToAP != "" {
		return opts.migrateViaAP(ctx, svc, bkp, static, override)
	}

	restoreOpts, err := opts.prepareTarget(ctx, svc, bkp, static, override)
	if err != nil || opts.DryRun {
		return err
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

	var result *backup.RestoreResult
	err = cmdutil.RunWithSpinner(ctx, ios, "Migrating configuration...", func(ctx context.Context) error {
		var restoreErr error
		result, restoreErr = svc.RestoreBackup(ctx, opts.Target, bkp, restoreOpts)
		return restoreErr
	})
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// A restore can reject sections while reporting no top-level error. Abort
	// BEFORE the source factory-reset on any such failure — otherwise the target
	// is left misconfigured AND the source is wiped, destroying both ends.
	if reportErr := term.ReportMigrationResult(ios, opts.Target, result); reportErr != nil {
		return reportErr
	}

	opts.factoryResetSource(ctx, svc, resetSource)
	return nil
}

// prepareTarget checks the target can take the backup, previews the migration
// on a dry run, and settles which WiFi station key the restore writes.
func (o *Options) prepareTarget(ctx context.Context, svc migrateService, bkp *backup.DeviceBackup, static shellybackup.StaticNetwork, override *backup.NetworkOverride) (backup.RestoreOptions, error) {
	ios := o.Factory.IOStreams()
	restoreOpts := o.restoreOptions(override)
	if err := svc.CheckMigrationCompatibility(ctx, bkp, o.Target, o.Force); err != nil {
		term.DisplayCompatibilityError(ios, err)
		return restoreOpts, err
	}
	if o.DryRun {
		if err := o.previewMigration(ctx, bkp, static, restoreOpts); err != nil {
			return restoreOpts, err
		}
	}
	err := cmdutil.PlanLANStation(ctx, ios, svc, o.Target, bkp, &restoreOpts)
	return restoreOpts, err
}

// factoryResetSource factory-resets the source device after a successful
// migration when one was requested, freeing its IP/name so it cannot collide
// with the newly-migrated target. A reset failure is a warning, not fatal: the
// migration itself already succeeded. A no-op when resetSource is false.
func (o *Options) factoryResetSource(ctx context.Context, svc migrateService, resetSource bool) {
	if !resetSource {
		return
	}
	ios := o.Factory.IOStreams()
	if err := cmdutil.RunWithSpinner(ctx, ios, "Factory resetting source device...", func(ctx context.Context) error {
		return svc.DeviceFactoryReset(ctx, o.Source)
	}); err != nil {
		ios.Warning("Migration succeeded but factory reset of source failed: %v", err)
		ios.Info("You may need to manually factory reset %q to avoid IP conflicts", o.Source)
		return
	}
	ios.Success("Source device %q has been factory reset", o.Source)
}
