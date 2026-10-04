package cmdutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/reprovision"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// APRestorer restores a backup onto a device at its WiFi access point, hopping
// the host's WiFi to reach the device's AP. *shelly.Service satisfies it, as do
// the migrate and restore command service interfaces.
type APRestorer interface {
	RestoreToAP(ctx context.Context, apSSID, apHostIP, registryName string, bkp *backup.DeviceBackup, opts backup.RestoreOptions) (*backup.RestoreResult, string, error)
}

// RunAtAP runs action, which hops the host onto a device's factory WiFi access
// point, under a spinner labelled label. The spinner names each stage the
// operation reports through shelly.WithStepReporter.
func RunAtAP(ctx context.Context, ios *iostreams.IOStreams, label string, action func(context.Context) error) error {
	return RunWithSpinner(ctx, ios, label+" (hopping host WiFi)...", func(ctx context.Context) error {
		return action(shelly.WithStepReporter(ctx, func(step string) {
			ios.UpdateProgress(fmt.Sprintf("%s: %s...", label, step))
		}))
	})
}

// RestoreAtAP runs a host-WiFi-hopping restore onto a device at its AP, reports
// the outcome with the supplied reporter, and surfaces the device's post-restore
// LAN address.
//
// It returns the reporter's error so a partial section rejection (a result with
// Success=false and a nil top-level error) still yields a non-zero exit — never a
// false success — while always printing the recovered LAN address when the device
// rejoined, so a partial failure can be finished by hand. errPrefix labels a
// transport-level failure of the restore call itself.
//
// migrate and restore share this exact AP-hop sequence; centralizing it keeps the
// partial-failure handling (which has a history of device-stranding bugs) in one
// place so a fix lands for both commands at once. Callers differ only in the
// reporter (migration vs restore messaging) and the error prefix.
func RestoreAtAP(
	ctx context.Context,
	ios *iostreams.IOStreams,
	svc APRestorer,
	apSSID, apHostIP, name string,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
	errPrefix string,
	report func(ios *iostreams.IOStreams, target string, result *backup.RestoreResult) error,
) error {
	var (
		result  *backup.RestoreResult
		newAddr string
	)
	err := RunAtAP(ctx, ios, fmt.Sprintf("Restoring onto %s at AP %s", name, apSSID),
		func(ctx context.Context) error {
			var restoreErr error
			result, newAddr, restoreErr = svc.RestoreToAP(ctx, apSSID, apHostIP, name, bkp, opts)
			return restoreErr
		})
	if err != nil {
		return fmt.Errorf("%s: %w", errPrefix, err)
	}

	reportErr := report(ios, name, result)
	if newAddr != "" {
		ios.Info("%s is live at %s", name, newAddr)
	}
	return reportErr
}

// ValidateStaticIPFlags checks the --static-ip, --gateway, --netmask and --dns
// flags. --static-ip may stand alone: an empty gateway, netmask or DNS is taken
// from the backup's own static settings, and the restore refuses an address left
// without a gateway or netmask. --gateway, --netmask and --dns without
// --static-ip are refused.
func ValidateStaticIPFlags(staticIP, gateway, netmask, dns string) error {
	if (gateway != "" || netmask != "" || dns != "") && staticIP == "" {
		return fmt.Errorf("--gateway, --netmask and --dns only apply with --static-ip")
	}
	return nil
}

// NetworkFlags are the WiFi settings a restore or migrate writes in place of the
// backup's: --ssid, --password, --open and the static address flags.
type NetworkFlags struct {
	SSID     string
	Password string
	Open     bool
	StaticIP string
	Gateway  string
	Netmask  string
	DNS      string
}

// Validate refuses flag combinations that cannot apply. With skipNetwork no
// network setting is written, so each network flag is refused by name.
func (n *NetworkFlags) Validate(skipNetwork bool) error {
	if err := ValidateStaticIPFlags(n.StaticIP, n.Gateway, n.Netmask, n.DNS); err != nil {
		return err
	}
	if !skipNetwork {
		return nil
	}
	for _, f := range []struct {
		name string
		set  bool
	}{
		{"static-ip", n.StaticIP != ""},
		{"ssid", n.SSID != ""},
		{"password", n.Password != ""},
		{"open", n.Open},
	} {
		if f.set {
			return fmt.Errorf("--%s cannot be used with --skip-network", f.name)
		}
	}
	return nil
}

// Override returns the network override the flags describe, or nil when none
// was given.
func (n *NetworkFlags) Override() *backup.NetworkOverride {
	if n.StaticIP == "" && n.SSID == "" && n.Password == "" && !n.Open {
		return nil
	}
	return &backup.NetworkOverride{
		SSID:     n.SSID,
		Password: n.Password,
		Open:     n.Open,
		StaticIP: n.StaticIP,
		Gateway:  n.Gateway,
		Netmask:  n.Netmask,
		DNS:      n.DNS,
	}
}

// AddStaticIPFlags registers --static-ip with staticIPUsage, and --gateway,
// --netmask and --dns with the help every such command shares; from names
// where an omitted one is taken from ("the backup's").
func AddStaticIPFlags(cmd *cobra.Command, staticIP, gateway, netmask, dns *string, staticIPUsage, from string) {
	cmd.Flags().StringVar(staticIP, "static-ip", "", staticIPUsage)
	cmd.Flags().StringVar(gateway, "gateway", "", "Static IPv4 default gateway (with --static-ip; default: "+from+")")
	cmd.Flags().StringVar(netmask, "netmask", "", "Static IPv4 subnet mask (with --static-ip; default: "+from+")")
	cmd.Flags().StringVar(dns, "dns", "", "Static IPv4 nameserver (with --static-ip; default: "+from+")")
}

// AddWiFiPasswordFlag registers --password, the WiFi passphrase of the network
// the device joins, and --password-stdin, which reads it from stdin, with the
// help text every such command shares.
func AddWiFiPasswordFlag(cmd *cobra.Command, password *string, fromStdin *bool) {
	AddSecretFlags(cmd, password, fromStdin, wifiPasswordFlag,
		"WiFi password for the network (when omitted and one is needed, the passphrase stored on this host for it is used)",
		"Read the WiFi password from stdin")
}

// ReadWiFiPasswordStdin sets *password from stdin when the --password-stdin
// flag AddWiFiPasswordFlag registers was given.
func ReadWiFiPasswordStdin(ios *iostreams.IOStreams, password *string, fromStdin bool) error {
	return ResolveSecret(ios, password, fromStdin, wifiPasswordFlag, "WiFi password")
}

// wifiPasswordFlag is the name of the WiFi password flag.
const wifiPasswordFlag = "password"

// AddOpenFlag registers --open, for joining a network that has no password, and
// makes it exclusive with the command's --password and --password-stdin flags.
func AddOpenFlag(cmd *cobra.Command, open *bool) {
	cmd.Flags().BoolVar(open, "open", false, "Join a network that has no password")
	cmd.MarkFlagsMutuallyExclusive("open", wifiPasswordFlag)
	cmd.MarkFlagsMutuallyExclusive("open", wifiPasswordFlag+StdinFlagSuffix)
}

// WiFiPasswordPrompts holds the questions ResolveWiFiPassword may ask. A nil
// field falls back to the terminal prompt.
type WiFiPasswordPrompts struct {
	Password func(message string) (string, error)
	Confirm  func(message string, defaultValue bool) (bool, error)
}

// ResolveWiFiPassword finds the password for ssid when none was given: this
// host's stored passphrase first, then a prompt. An empty answer means an open
// network only when the user confirms it; otherwise, and whenever ios cannot
// prompt, it returns the passphrase error naming ssid.
func ResolveWiFiPassword(
	ctx context.Context,
	ios *iostreams.IOStreams,
	lookup func(ctx context.Context, ssid string) (string, error),
	ssid string,
	prompts WiFiPasswordPrompts,
) (password string, open bool, err error) {
	if pass, lookupErr := lookup(ctx, ssid); lookupErr == nil && pass != "" {
		ios.Success("WiFi password for %q recovered from this host", ssid)
		return pass, false, nil
	}
	if !ios.CanPrompt() && prompts.Password == nil {
		return "", false, shelly.PassphraseError(ssid)
	}

	ask := prompts.Password
	if ask == nil {
		ask = iostreams.Password
	}
	pass, err := ask("WiFi password:")
	if err != nil {
		return "", false, fmt.Errorf("password input failed: %w", err)
	}
	if pass != "" {
		return pass, false, nil
	}

	confirm := prompts.Confirm
	if confirm == nil {
		confirm = ios.Confirm
	}
	ok, err := confirm(fmt.Sprintf("Join %s as an open network with no password?", ssid), false)
	if err != nil {
		return "", false, fmt.Errorf("confirmation failed: %w", err)
	}
	if !ok {
		return "", false, shelly.PassphraseError(ssid)
	}
	return "", true, nil
}

// RestoreNameDescriber states the device-name decision of a LAN restore.
// *shelly.Service satisfies it.
type RestoreNameDescriber interface {
	DescribeRestoreName(ctx context.Context, device string, bkp *backup.DeviceBackup, opts backup.RestoreOptions) (string, error)
}

// PrintRestoreName prints the dry-run line for the device name a restore of
// bkp onto device writes, or a warning when the target cannot be read.
func PrintRestoreName(
	ctx context.Context, ios *iostreams.IOStreams, d RestoreNameDescriber, device string, bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
) {
	line, err := d.DescribeRestoreName(ctx, device, bkp, opts)
	if err != nil {
		ios.Warning("name: %v", err)
		return
	}
	ios.Info("%s", line)
}

// LANStationPlanner decides the WiFi station passphrase for a LAN restore.
// *shelly.Service satisfies it.
type LANStationPlanner interface {
	PlanLANStation(
		ctx context.Context, device string, bkp *backup.DeviceBackup, ov *backup.NetworkOverride, skipNetwork bool,
	) (*backup.NetworkOverride, shelly.LANStationPlan, error)
}

// PlanLANStation decides, before anything is written, what a LAN restore of
// bkp onto device writes as the station passphrase, prints that decision, and
// stores the override and secondary-station write to restore with in opts. A
// missing passphrase or a network flag with no network to apply it to is always
// an error. On a dry run a device that cannot be read is only a warning, so a
// backup file can still be previewed offline.
func PlanLANStation(
	ctx context.Context,
	ios *iostreams.IOStreams,
	planner LANStationPlanner,
	device string,
	bkp *backup.DeviceBackup,
	opts *backup.RestoreOptions,
) error {
	ov, plan, err := planner.PlanLANStation(ctx, device, bkp, opts.NetworkOverride, opts.SkipNetwork)
	if err != nil {
		if opts.DryRun && !errors.Is(err, reprovision.ErrNoPassphrase) && !errors.Is(err, types.ErrInvalidParam) {
			ios.Warning("WiFi station: %v", err)
			return nil
		}
		return err
	}
	opts.NetworkOverride = ov
	opts.Station1 = plan.Station1Write()
	ios.Info("%s", plan)
	return nil
}

// APStationPlanner decides the WiFi station a --to-ap restore writes, before
// any hop. *shelly.Service satisfies it.
type APStationPlanner interface {
	PlanAPStation(ctx context.Context, bkp *backup.DeviceBackup, ov *backup.NetworkOverride) (shelly.LANStationPlan, error)
}

// PreviewAPRestore prints the dry-run plan of a --to-ap restore of bkp onto
// the device at access point apSSID: the name it gets, the WiFi station it
// joins, the Gen1 firmware update when one may run, and the hop that would
// carry it all out. It never reads the device. Of this host's WiFi it reads
// only the current network and the stored passphrase the station needs, so
// nothing scans, joins, leaves a network or writes.
func PreviewAPRestore(
	ctx context.Context, ios *iostreams.IOStreams, planner APStationPlanner, apSSID string,
	bkp *backup.DeviceBackup, opts backup.RestoreOptions,
) error {
	plan, err := planner.PlanAPStation(ctx, bkp, opts.NetworkOverride)
	if err != nil {
		return err
	}
	ios.Info("%s", shelly.DescribeAPRestoreName(apSSID, bkp, opts))
	ios.Info("%s", plan)
	if dev := bkp.Device(); dev.Generation == 1 && !opts.AllowFirmwareDowngrade {
		ios.Info("Would update Gen1 firmware to match the backup (%s) if the device runs older firmware", dev.FWVersion)
	}
	ios.Info("Would hop onto AP %q from this host; nothing was written", apSSID)
	return nil
}
