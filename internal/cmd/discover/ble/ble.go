// Package ble provides BLE discovery command.
package ble

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
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
		if shelly.IsBLENotSupportedError(err) {
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
	term.DisplayBLEDevices(ios, bleDevices)

	return nil
}
