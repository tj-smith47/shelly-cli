// Package ble provides BLE-based device provisioning.
package ble

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/provisioning"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	DeviceAddress string
	SSID          string
	Password      string
	PasswordStdin bool
	Open          bool
	DeviceName    string
	Timezone      string
	EnableCloud   bool
	DisableCloud  bool
	Factory       *cmdutil.Factory

	// prompts, when its fields are non-nil, replaces the terminal prompts.
	prompts cmdutil.WiFiPasswordPrompts
}

// NewCommand creates the provision ble command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "ble <device-address>",
		Aliases: []string{"bluetooth"},
		Short:   "Provision a device via Bluetooth Low Energy",
		Long: `Provision a Shelly device using Bluetooth Low Energy (BLE).

This command allows you to configure WiFi credentials and other settings
on a new device without connecting to its AP network first.

BLE provisioning requires:
- Bluetooth hardware support on your computer
- The device to be in BLE advertising mode (typically when unconfigured)
- The device's BLE address (usually shown as ShellyXXX-YYYYYYYY)

Gen2+ devices support BLE provisioning. Gen1 devices do not have BLE capability.

Without --password, the network's passphrase is read from this host's stored
WiFi credentials, or asked for. A network is joined with no password only with
--open, or when an empty answer at the prompt is confirmed.`,
		Example: `  # Provision WiFi via BLE
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "MyNetwork" --password "secret"

  # Set device name during provisioning
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "MyNetwork" --password "secret" --name "Living Room Switch"

  # Configure with timezone
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "MyNetwork" --password "secret" --timezone "America/New_York"

  # Disable cloud during provisioning
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "MyNetwork" --password "secret" --no-cloud

  # Join a network that has no password
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "GuestWiFi" --open`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.DeviceAddress = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "WiFi network name (required)")
	cmdutil.AddWiFiPasswordFlag(cmd, &opts.Password, &opts.PasswordStdin)
	cmd.Flags().StringVar(&opts.DeviceName, "name", "", "Device name to set")
	cmd.Flags().StringVar(&opts.Timezone, "timezone", "", "Timezone (e.g., America/New_York)")
	cmd.Flags().BoolVar(&opts.EnableCloud, "cloud", false, "Enable Shelly Cloud")
	cmd.Flags().BoolVar(&opts.DisableCloud, "no-cloud", false, "Disable Shelly Cloud")

	cmdutil.AddOpenFlag(cmd, &opts.Open)
	cmd.MarkFlagsMutuallyExclusive("cloud", "no-cloud")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	if err := cmdutil.ReadWiFiPasswordStdin(ios, &opts.Password, opts.PasswordStdin); err != nil {
		return err
	}

	// Validate required options
	if opts.SSID == "" {
		ssid, err := ios.Input("WiFi network name (SSID):", "")
		if err != nil {
			return fmt.Errorf("failed to get SSID: %w", err)
		}
		opts.SSID = ssid
	}

	if opts.SSID == "" {
		return fmt.Errorf("SSID is required for BLE provisioning")
	}

	// BLE sends the password as given, so an empty one would join as an
	// open network; it must be found or confirmed first.
	if opts.Password == "" && !opts.Open {
		password, open, err := cmdutil.ResolveWiFiPassword(ctx, ios, opts.Factory.ShellyService().HostWiFiPassword, opts.SSID, opts.prompts)
		if err != nil {
			return err
		}
		opts.Password, opts.Open = password, open
	}

	// Create BLE transmitter
	ios.Info("Initializing Bluetooth...")
	transmitter, err := provisioning.NewTinyGoBLETransmitter()
	if err != nil {
		return fmt.Errorf("failed to initialize Bluetooth: %w", err)
	}

	// Create BLE provisioner with transmitter
	bleProvisioner := provisioning.NewBLEProvisioner()
	bleProvisioner.Transmitter = transmitter

	// Register the device for provisioning
	model, _ := provisioning.ParseBLEDeviceName(opts.DeviceAddress)
	bleProvisioner.AddDiscoveredDevice(&provisioning.BLEDevice{
		Name:     opts.DeviceAddress,
		Address:  opts.DeviceAddress,
		Model:    model,
		IsShelly: provisioning.IsShellyDevice(opts.DeviceAddress),
	})

	// Build provisioning configuration
	config := &provisioning.BLEProvisionConfig{
		WiFi: &provisioning.WiFiConfig{
			SSID:     opts.SSID,
			Password: opts.Password,
		},
		DeviceName: opts.DeviceName,
		Timezone:   opts.Timezone,
	}

	// Handle cloud option
	if opts.EnableCloud {
		enable := true
		config.EnableCloud = &enable
	} else if opts.DisableCloud {
		disable := false
		config.EnableCloud = &disable
	}

	// Perform provisioning
	ios.Info("Connecting to %s via BLE...", opts.DeviceAddress)
	result, err := bleProvisioner.ProvisionViaBLE(ctx, opts.DeviceAddress, config)
	if err != nil {
		return fmt.Errorf("BLE provisioning failed: %w", err)
	}

	// Display result
	term.DisplayBLEProvisionResult(ios, result, opts.SSID)

	if !result.Success {
		return fmt.Errorf("provisioning completed with errors")
	}

	return nil
}
