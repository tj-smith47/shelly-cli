// Package ble provides BLE discovery command.
package ble

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DefaultTimeout is the default BLE discovery timeout.
const DefaultTimeout = 15 * time.Second

// NewCommand creates the BLE discovery command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		timeout       time.Duration
		includeBTHome bool
		filterPrefix  string
	)

	cmd := &cobra.Command{
		Use:     "ble",
		Aliases: []string{"bluetooth", "bt"},
		Short:   "Discover devices using Bluetooth Low Energy",
		Long: `Discover Shelly devices using Bluetooth Low Energy (BLE).

This method finds Shelly devices that are in provisioning mode or
BLU devices broadcasting BTHome sensor data.

Requirements:
  - Bluetooth adapter on the host machine
  - Bluetooth must be enabled
  - May require elevated privileges on some systems`,
		Example: `  # Basic BLE discovery
  shelly discover ble

  # With longer timeout
  shelly discover ble --timeout 30s

  # Include BTHome sensor data
  shelly discover ble --bthome

  # Filter by device name prefix
  shelly discover ble --filter "Shelly"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, timeout, includeBTHome, filterPrefix)
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", DefaultTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&includeBTHome, "bthome", false, "Include BTHome sensor broadcasts")
	cmd.Flags().StringVarP(&filterPrefix, "filter", "f", "", "Filter by device name prefix")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, timeout time.Duration, includeBTHome bool, filterPrefix string) error {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ios := f.IOStreams()
	ios.StartProgress("Discovering devices via BLE...")

	bleDiscoverer, err := discovery.NewBLEDiscoverer()
	if err != nil {
		ios.StopProgress()
		if isBLENotSupportedError(err) {
			ios.Error("BLE discovery is not available on this system")
			ios.Hint("Ensure you have a Bluetooth adapter and it is enabled")
			ios.Hint("On Linux, you may need to run with elevated privileges")
			return nil
		}
		return fmt.Errorf("failed to initialize BLE: %w", err)
	}
	defer func() {
		if err := bleDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping BLE discoverer", err)
		}
	}()

	// Configure discoverer
	bleDiscoverer.IncludeBTHome = includeBTHome
	if filterPrefix != "" {
		bleDiscoverer.FilterPrefix = filterPrefix
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	devices, err := bleDiscoverer.DiscoverWithContext(ctx)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("BLE discovery failed: %w", err)
	}

	if len(devices) == 0 {
		ios.NoResults("BLE devices",
			"Put devices in provisioning mode (AP mode) to discover via BLE",
			"Ensure Bluetooth is enabled on this machine")
		return nil
	}

	// Get detailed BLE information
	bleDevices := bleDiscoverer.GetDiscoveredDevices()
	displayBLEDevices(ios, bleDevices)

	return nil
}

// isBLENotSupportedError checks if the error is due to BLE not being supported.
func isBLENotSupportedError(err error) bool {
	if err == nil {
		return false
	}
	// Check for the BLE not supported error
	var bleErr *discovery.BLEError
	if errors.As(err, &bleErr) {
		// Check if this is the sentinel error or has similar message
		if bleErr == discovery.ErrBLENotSupported {
			return true
		}
		// Check if the error message indicates BLE is not supported
		if strings.Contains(bleErr.Message, "not supported") ||
			strings.Contains(bleErr.Message, "BLE not supported") {
			return true
		}
		// Check if wrapped error is ErrBLENotSupported
		return errors.Is(bleErr.Err, discovery.ErrBLENotSupported)
	}
	return errors.Is(err, discovery.ErrBLENotSupported)
}

// displayBLEDevices prints a table of BLE discovered devices.
func displayBLEDevices(ios *iostreams.IOStreams, devices []discovery.BLEDiscoveredDevice) {
	if len(devices) == 0 {
		return
	}

	table := output.NewTable("Name", "Address", "Model", "RSSI", "Connectable", "BTHome")

	for _, d := range devices {
		name := d.LocalName
		if name == "" {
			name = d.ID
		}

		// Theme RSSI value (stronger is better)
		rssiStr := fmt.Sprintf("%d dBm", d.RSSI)
		switch {
		case d.RSSI > -50:
			rssiStr = theme.StatusOK().Render(rssiStr)
		case d.RSSI > -70:
			rssiStr = theme.StatusWarn().Render(rssiStr)
		default:
			rssiStr = theme.StatusError().Render(rssiStr)
		}

		// Connectable status
		connStr := theme.StatusError().Render("No")
		if d.Connectable {
			connStr = theme.StatusOK().Render("Yes")
		}

		// BTHome data indicator
		btHomeStr := "-"
		if d.BTHomeData != nil {
			btHomeStr = theme.StatusInfo().Render("Yes")
		}

		table.AddRow(
			name,
			d.Address.String(),
			d.Model,
			rssiStr,
			connStr,
			btHomeStr,
		)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print BLE devices table", err)
	}
	ios.Println("")
	ios.Count("BLE device", len(devices))
}
