// Package restore provides the backup restore subcommand.
package restore

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory       *cmdutil.Factory
	Decrypt       string
	Device        string
	DryRun        bool
	FilePath      string
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

  # Skip scripts during restore
  shelly backup restore living-room backup.json --skip-scripts

  # Clone another bulb's backup onto this device with a different static IP
  # (identity — MAC, serial, device ID — is never overwritten by restore)
  shelly backup restore new-bulb master-bath-1.json \
    --static-ip 10.23.47.221 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1

  # Restore a sibling's backup straight onto a brand-new device at its factory
  # WiFi AP: hops the host onto the AP, applies the config + static IP, and the
  # device joins the LAN — no separate provisioning step (target name = "fr")
  shelly backup restore fr sr.json --to-ap ShellyBulbDuo-D0DCFF \
    --static-ip 10.23.47.227 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1`,
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
	cmd.Flags().StringVarP(&opts.Decrypt, "decrypt", "d", "", "Password to decrypt backup")
	cmd.Flags().StringVar(&opts.StaticIP, "static-ip", "", "Override the backup's WiFi with this static IPv4 address")
	cmd.Flags().StringVar(&opts.Gateway, "gateway", "", "Static IPv4 default gateway (with --static-ip)")
	cmd.Flags().StringVar(&opts.Netmask, "netmask", "", "Static IPv4 subnet mask (with --static-ip)")
	cmd.Flags().StringVar(&opts.DNS, "dns", "", "Static IPv4 nameserver (optional, with --static-ip)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Override the device name (defaults to the target identifier when it is a friendly alias)")
	cmd.Flags().StringVar(&opts.ToAP, "to-ap", "", "Restore onto a device at its factory WiFi AP with this SSID (hops host WiFi; the network override moves it onto the LAN)")
	cmd.Flags().StringVar(&opts.APIP, "ap-ip", "", "Static host IP to use on the device's AP subnet during --to-ap (default 192.168.33.133)")
	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "Override the WiFi SSID the device joins (defaults to the backup's network)")
	cmd.Flags().StringVar(&opts.Password, "password", "", "WiFi passphrase for the target network (optional: derived from this host's stored credentials when omitted; set to override or when derivation fails)")
	cmd.MarkFlagsRequiredTogether("static-ip", "gateway", "netmask")

	return cmd
}

// validateFlags rejects incompatible flag combinations before any device I/O.
func (o *Options) validateFlags() error {
	if o.StaticIP != "" && o.SkipNetwork {
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

	// Resolve file path (check backups dir if not found as-is)
	opts.FilePath = backup.ResolveFilePath(opts.FilePath)

	// Read backup file
	data, err := afero.ReadFile(config.Fs(), opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Validate backup
	bkp, err := backup.Validate(data)
	if err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Check encryption
	if bkp.Encrypted() && opts.Decrypt == "" {
		return fmt.Errorf("backup is encrypted, use --decrypt to provide password")
	}

	if err := opts.validateFlags(); err != nil {
		return err
	}

	var override *backup.NetworkOverride
	if opts.StaticIP != "" || opts.SSID != "" || opts.Password != "" {
		override = &backup.NetworkOverride{
			SSID:     opts.SSID,
			Password: opts.Password,
			StaticIP: opts.StaticIP,
			Gateway:  opts.Gateway,
			Netmask:  opts.Netmask,
			DNS:      opts.DNS,
		}
	}

	restoreOpts := backup.RestoreOptions{
		DryRun:          opts.DryRun,
		SkipAuth:        opts.SkipAuth,
		SkipNetwork:     opts.SkipNetwork,
		SkipScripts:     opts.SkipScripts,
		SkipSchedules:   opts.SkipSchedules,
		SkipWebhooks:    opts.SkipWebhooks,
		SkipState:       opts.SkipState,
		SkipMeters:      opts.SkipMeters,
		Password:        opts.Decrypt,
		NetworkOverride: override,
		Name:            cmdutil.DeviceDisplayName(opts.Name, opts.Device),
	}

	if opts.DryRun {
		ios.Title("Dry run - Restore preview")
		ios.Println()
		term.DisplayRestorePreview(ios, bkp, restoreOpts)
		if override != nil {
			ios.Info("WiFi station IP will be overridden to %s (gateway %s, netmask %s)", override.StaticIP, override.Gateway, override.Netmask)
		}
		return nil
	}

	svc := opts.Factory.ShellyService()

	// --to-ap: hop onto the device's factory AP, restore there, and let the
	// network override move it onto the LAN — provisioning and restore in one.
	if opts.ToAP != "" {
		return opts.restoreViaAP(ctx, svc, bkp, restoreOpts)
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

	// Print results
	ios.Success("Backup restored to %s", opts.Device)
	term.DisplayRestoreResult(ios, result)

	return nil
}

// restoreViaAP restores the backup onto a device at its factory WiFi AP: it hops
// the host onto the AP, applies the config (with network + name overrides) at
// the AP address, and the restored station config moves the device onto the LAN.
func (o *Options) restoreViaAP(
	ctx context.Context,
	svc *shelly.Service,
	bkp *backup.DeviceBackup,
	restoreOpts backup.RestoreOptions,
) error {
	ios := o.Factory.IOStreams()

	var (
		result  *backup.RestoreResult
		newAddr string
	)
	err := cmdutil.RunWithSpinner(ctx, ios,
		fmt.Sprintf("Restoring onto %s at AP %s (hopping host WiFi)...", o.Device, o.ToAP),
		func(ctx context.Context) error {
			var restoreErr error
			result, newAddr, restoreErr = svc.RestoreToAP(ctx, o.ToAP, o.APIP, o.Device, bkp, restoreOpts)
			return restoreErr
		})
	if err != nil {
		return fmt.Errorf("failed to restore via AP: %w", err)
	}

	ios.Success("Backup restored to %s", o.Device)
	if newAddr != "" {
		ios.Info("%s is live at %s", o.Device, newAddr)
	}
	term.DisplayRestoreResult(ios, result)
	return nil
}
