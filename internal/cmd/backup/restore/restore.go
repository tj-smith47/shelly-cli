// Package restore provides the backup restore subcommand.
package restore

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// restoreService is the subset of *shelly.Service the restore command drives.
// Depending on this narrow interface instead of the concrete service keeps the
// command testable with a stub: the concrete service reaches a real device (and,
// for --to-ap, hops the host's WiFi), which a unit test cannot exercise.
type restoreService interface {
	RestoreBackup(ctx context.Context, identifier string, deviceBackup *backup.DeviceBackup, opts backup.RestoreOptions) (*backup.RestoreResult, error)
	RestoreToAP(ctx context.Context, apSSID, apHostIP, registryName string, bkp *backup.DeviceBackup, opts backup.RestoreOptions) (*backup.RestoreResult, string, error)
	cmdutil.LANStationPlanner
	cmdutil.APStationPlanner
	cmdutil.RestoreNameDescriber
}

// Options holds the command options.
type Options struct {
	Factory                *cmdutil.Factory
	Decrypt                string
	DecryptStdin           bool
	Device                 string
	DryRun                 bool
	FilePath               string
	SkipAuth               bool
	SkipNetwork            bool
	SkipScripts            bool
	SkipSchedules          bool
	SkipWebhooks           bool
	SkipState              bool
	SkipMeters             bool
	StaticIP               string
	Gateway                string
	Netmask                string
	DNS                    string
	Name                   string
	ToAP                   string
	APIP                   string
	SSID                   string
	Password               string
	PasswordStdin          bool
	Open                   bool
	AllowFirmwareDowngrade bool
	FirmwareURL            string
	TraceFile              string

	// svc, when non-nil, overrides the service resolved from the Factory. It is the
	// test injection seam; production leaves it nil and uses Factory.ShellyService().
	svc restoreService
}

// service returns the injected restoreService when set, otherwise the concrete
// service from the Factory.
func (o *Options) service() restoreService {
	if o.svc != nil {
		return o.svc
	}
	return o.Factory.ShellyService()
}

// NewCommand creates the backup restore command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "restore <device> <file>",
		Aliases: []string{"apply", "load"},
		Short:   "Restore a device from backup",
		Long: `Restore a Shelly device from a backup file.

By default, everything from the backup is restored including network
and authentication settings. Use --skip-* flags to exclude specific
sections.`,
		Example: `  # Full restore from backup
  shelly backup restore living-room backup.json

  # Dry run - show what would change
  shelly backup restore living-room backup.json --dry-run

  # Restore without network config (keep current WiFi)
  shelly backup restore living-room backup.json --skip-network

  # Restore without auth config
  shelly backup restore living-room backup.json --skip-auth

  # Restore encrypted backup
  shelly backup restore living-room backup.json --decrypt mysecret

  # Read the decryption password from stdin
  shelly backup restore living-room backup.json --decrypt-stdin < ~/.shelly-backup-password

  # Skip scripts during restore
  shelly backup restore living-room backup.json --skip-scripts

  # Give a clone of another bulb its own address; the gateway, netmask and DNS
  # are the backup's
  shelly backup restore new-bulb master-bath-1.json --static-ip 10.23.47.221

  # Restore onto a device that joins a network with no password
  shelly backup restore guest-plug plug.json --ssid GuestWiFi --open

  # Clone another bulb's backup onto this device with a different static IP
  # (identity — MAC, serial, device ID — is never overwritten by restore)
  shelly backup restore new-bulb master-bath-1.json \
    --static-ip 10.23.47.221 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1

  # Preview a restore at the factory WiFi AP without hopping the host's WiFi
  shelly backup restore fr sr.json --to-ap ShellyBulbDuo-D0DCFF --dry-run

  # Restore a sibling's backup straight onto a brand-new device at its factory
  # WiFi AP: hops the host onto the AP, applies the config + static IP, and the
  # device joins the LAN — no separate provisioning step (target name = "fr")
  shelly backup restore fr sr.json --to-ap ShellyBulbDuo-D0DCFF \
    --static-ip 10.23.47.227 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1

  # If the target runs older firmware than the backup, it is updated automatically
  # before the restore so the configuration lands on matched firmware and cannot
  # reboot-loop — no flag needed. With --to-ap the update runs AT the factory AP
  # (where the device is stable): the image is fetched before the hop and re-served
  # on the AP subnet, then the full restore proceeds once the device joins the LAN.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.FilePath = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be restored without applying")
	cmd.Flags().BoolVar(&opts.SkipAuth, "skip-auth", false, "Skip authentication configuration")
	cmd.Flags().BoolVar(&opts.SkipNetwork, "skip-network", false, "Skip network configuration (WiFi, Ethernet)")
	cmd.Flags().BoolVar(&opts.SkipScripts, "skip-scripts", false, "Skip script restoration")
	cmd.Flags().BoolVar(&opts.SkipSchedules, "skip-schedules", false, "Skip schedule restoration")
	cmd.Flags().BoolVar(&opts.SkipWebhooks, "skip-webhooks", false, "Skip webhook restoration")
	cmd.Flags().BoolVar(&opts.SkipState, "skip-state", false, "Skip restoring live component state (color temperature, brightness); apply configuration only")
	cmd.Flags().BoolVar(&opts.SkipMeters, "skip-meters", false, "Skip restoring meter/energy-meter configuration (e.g. overpower limits)")
	cmdutil.AddSecretFlagsP(cmd, &opts.Decrypt, &opts.DecryptStdin, "decrypt", "d",
		"Password to decrypt backup", "Read the password to decrypt the backup from stdin")
	cmdutil.AddStaticIPFlags(cmd, &opts.StaticIP, &opts.Gateway, &opts.Netmask, &opts.DNS,
		"Override the backup's WiFi with this static IPv4 address (--gateway, --netmask and --dns default to the backup's)",
		"the backup's")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Set the device name (default: a backup of this same device keeps the name it recorded; a backup of another device takes the target's alias)")
	cmd.Flags().StringVar(&opts.ToAP, "to-ap", "", "Restore onto a device at its factory WiFi AP with this SSID (hops host WiFi; the network override moves it onto the LAN)")
	cmd.Flags().StringVar(&opts.APIP, "ap-ip", "", "Static host IP to use on the device's AP subnet during --to-ap (default 192.168.33.133)")
	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "Override the WiFi SSID the device joins (defaults to the backup's network)")
	cmdutil.AddWiFiPasswordFlag(cmd, &opts.Password, &opts.PasswordStdin)
	cmdutil.AddOpenFlag(cmd, &opts.Open)
	cmd.Flags().BoolVar(&opts.AllowFirmwareDowngrade, "allow-firmware-downgrade", false, "Force the older-firmware config write instead of the automatic firmware update (Gen1; the device is updated to matched firmware by default when the backup is newer — this skips that and accepts the reboot-loop risk)")
	cmd.Flags().StringVar(&opts.FirmwareURL, "firmware-url", "", "Firmware image for the automatic downgrade-recovery update (default: derived from the backup's device model)")
	cmd.Flags().StringVar(&opts.TraceFile, "trace-file", "", "Write a per-step Gen1 restore diagnostic (which setting destabilizes the device) to this file")
	if err := cmd.Flags().MarkHidden("trace-file"); err != nil {
		// MarkHidden only fails on an unknown flag name; the flag is defined above.
		panic(err)
	}

	return cmd
}

// attachTrace wires the --trace-file diagnostic sink onto restoreOpts when the
// flag is set, returning a cleanup that closes the file. It returns a no-op
// cleanup (never nil) when no trace was requested, so callers can defer it
// unconditionally.
func (o *Options) attachTrace(restoreOpts *backup.RestoreOptions) (func(), error) {
	if o.TraceFile == "" {
		return func() {}, nil
	}
	ios := o.Factory.IOStreams()
	traceFile, err := config.Fs().Create(o.TraceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open trace file: %w", err)
	}
	restoreOpts.StepTrace = traceFile
	ios.Info("Writing per-step restore trace to %s", o.TraceFile)
	return func() {
		if closeErr := traceFile.Close(); closeErr != nil {
			ios.DebugErr("close trace file", closeErr)
		}
	}, nil
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

// previewRestore prints what a dry run would restore: the static address, the
// device name and the WiFi station that would be written and, for --to-ap, the
// firmware update and access point hop that would run.
func (o *Options) previewRestore(ctx context.Context, svc restoreService, bkp *backup.DeviceBackup, restoreOpts backup.RestoreOptions) error {
	ios := o.Factory.IOStreams()
	ios.Title("Dry run - Restore preview")
	ios.Println()
	term.DisplayRestorePreview(ios, bkp, restoreOpts)
	if override := restoreOpts.NetworkOverride; override.IsStatic() {
		static, err := bkp.StaticNetwork(override)
		if err != nil {
			return err
		}
		ios.Info("WiFi station IP will be overridden to %s (gateway %s, netmask %s)", static.IP, static.Gateway, static.Netmask)
	}
	if o.ToAP != "" {
		return cmdutil.PreviewAPRestore(ctx, ios, svc, o.ToAP, bkp, restoreOpts)
	}
	cmdutil.PrintRestoreName(ctx, ios, svc, o.Device, bkp, restoreOpts)
	return cmdutil.PlanLANStation(ctx, ios, svc, o.Device, bkp, &restoreOpts)
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

func run(ctx context.Context, opts *Options) error {
	// --to-ap performs a WiFi hop, a full restore at the AP, and a LAN-rejoin
	// poll, which together far exceed a normal restore's budget.
	timeout := shelly.DefaultTimeout * 5
	if opts.ToAP != "" {
		timeout = shelly.DefaultTimeout * 30
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	if err := cmdutil.ReadWiFiPasswordStdin(ios, &opts.Password, opts.PasswordStdin); err != nil {
		return err
	}
	if err := cmdutil.ResolveSecret(ios, &opts.Decrypt, opts.DecryptStdin, "decrypt", "Backup password"); err != nil {
		return err
	}

	// Resolve file path (check backups dir if not found as-is)
	opts.FilePath = backup.ResolveFilePath(opts.FilePath)

	// Read backup file
	data, err := afero.ReadFile(config.Fs(), opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Load the backup, transparently decrypting an encrypted envelope when
	// --decrypt or --decrypt-stdin supplies the password.
	bkp, err := backup.Load(data, opts.Decrypt)
	if err != nil {
		if errors.Is(err, backup.ErrEncryptedNeedsPassword) {
			return errors.New("backup is encrypted: give its password with --decrypt <value>, or read it from stdin with --decrypt-stdin")
		}
		return fmt.Errorf("invalid backup file: %w", err)
	}

	if err := opts.validateFlags(); err != nil {
		return err
	}

	override := opts.network().Override()

	restoreOpts := backup.RestoreOptions{
		DryRun:                 opts.DryRun,
		SkipAuth:               opts.SkipAuth,
		SkipNetwork:            opts.SkipNetwork,
		SkipScripts:            opts.SkipScripts,
		SkipSchedules:          opts.SkipSchedules,
		SkipWebhooks:           opts.SkipWebhooks,
		SkipState:              opts.SkipState,
		SkipMeters:             opts.SkipMeters,
		NetworkOverride:        override,
		Name:                   opts.Name,
		AliasName:              cmdutil.DeviceDisplayName("", opts.Device),
		AllowFirmwareDowngrade: opts.AllowFirmwareDowngrade,
		FirmwareURL:            opts.FirmwareURL,
	}

	svc := opts.service()

	if opts.DryRun {
		return opts.previewRestore(ctx, svc, bkp, restoreOpts)
	}

	// The dry run refuses network flags that name no network, or a network
	// this host has no passphrase for, before any hop; the real run refuses
	// them the same way, before the trace file or the hop.
	if opts.ToAP != "" {
		if _, err := svc.PlanAPStation(ctx, bkp, restoreOpts.NetworkOverride); err != nil {
			return err
		}
	}

	// --trace-file streams a per-step Gen1 restore diagnostic to a file: which
	// setting each device tolerated and which one drove it into a reboot loop.
	cleanupTrace, err := opts.attachTrace(&restoreOpts)
	if err != nil {
		return err
	}
	defer cleanupTrace()

	// --to-ap: hop onto the device's factory AP, restore there, and let the
	// network override move it onto the LAN — provisioning and restore in one.
	if opts.ToAP != "" {
		return opts.restoreViaAP(ctx, svc, bkp, restoreOpts)
	}

	if err := cmdutil.PlanLANStation(ctx, ios, svc, opts.Device, bkp, &restoreOpts); err != nil {
		return err
	}

	var result *backup.RestoreResult
	err = cmdutil.RunWithSpinner(ctx, ios, "Restoring backup...", func(ctx context.Context) error {
		var restoreErr error
		result, restoreErr = svc.RestoreBackup(ctx, opts.Device, bkp, restoreOpts)
		return restoreErr
	})
	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// A restore can fail per-section while reporting no top-level error; gate the
	// success line and the exit code on the device actually accepting it.
	return term.ReportRestoreResult(ios, opts.Device, result)
}

// restoreViaAP restores the backup onto a device at its factory WiFi AP: it hops
// the host onto the AP, applies the config (with network + name overrides) at
// the AP address, and the restored station config moves the device onto the LAN.
func (o *Options) restoreViaAP(
	ctx context.Context,
	svc restoreService,
	bkp *backup.DeviceBackup,
	restoreOpts backup.RestoreOptions,
) error {
	ios := o.Factory.IOStreams()

	// The AP-hop restore-and-report sequence (including partial-failure handling)
	// is shared with the migrate command.
	return cmdutil.RestoreAtAP(ctx, ios, svc, o.ToAP, o.APIP, o.Device, bkp, restoreOpts,
		"failed to restore via AP", term.ReportRestoreResult)
}
