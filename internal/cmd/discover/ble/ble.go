// Package ble provides BLE discovery command.
package ble

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly/wireless"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// DefaultTimeout is the default BLE discovery timeout.
const DefaultTimeout = 15 * time.Second

// Options holds the command options.
type Options struct {
	Factory       *cmdutil.Factory
	FilterPrefix  string
	IncludeBTHome bool
	Timeout       time.Duration
}

// Discoverer is the interface for BLE device discovery.
// This interface allows for dependency injection in tests.
type Discoverer interface {
	DiscoverWithContext(ctx context.Context) ([]discovery.DiscoveredDevice, error)
	GetDiscoveredDevices() []discovery.BLEDiscoveredDevice
	Stop() error
	SetIncludeBTHome(include bool)
	SetFilterPrefix(prefix string)
}

// bleDiscovererAdapter wraps discovery.BLEDiscoverer to implement Discoverer interface.
type bleDiscovererAdapter struct {
	*discovery.BLEDiscoverer
}

func (a *bleDiscovererAdapter) SetIncludeBTHome(include bool) {
	a.IncludeBTHome = include
}

func (a *bleDiscovererAdapter) SetFilterPrefix(prefix string) {
	a.FilterPrefix = prefix
}

// newBLEDiscoverer is the factory function for creating BLE discoverers.
// This can be replaced in tests.
var newBLEDiscoverer = func() (Discoverer, error) {
	d, err := discovery.NewBLEDiscoverer()
	if err != nil {
		return nil, err
	}
	return &bleDiscovererAdapter{BLEDiscoverer: d}, nil
}

// NewCommand creates the BLE discovery command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", DefaultTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&opts.IncludeBTHome, "bthome", false, "Include BTHome sensor broadcasts")
	cmd.Flags().StringVarP(&opts.FilterPrefix, "filter", "f", "", "Filter by device name prefix")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ios := opts.Factory.IOStreams()

	bleDiscoverer, err := newBLEDiscoverer()
	if err != nil {
		if wireless.IsBLENotSupportedError(err) {
			ios.Error("BLE discovery is not available on this system")
			ios.Hint("Ensure you have a Bluetooth adapter and it is enabled")
			ios.Hint("On Linux, you may need to run with elevated privileges")
			return nil
		}
		return fmt.Errorf("failed to initialize BLE: %w", err)
	}
	defer func() {
		if stopErr := bleDiscoverer.Stop(); stopErr != nil {
			ios.DebugErr("stopping BLE discoverer", stopErr)
		}
	}()

	// Configure discoverer
	bleDiscoverer.SetIncludeBTHome(opts.IncludeBTHome)
	if opts.FilterPrefix != "" {
		bleDiscoverer.SetFilterPrefix(opts.FilterPrefix)
	}

	var devices []discovery.DiscoveredDevice
	err = cmdutil.RunWithSpinner(ctx, ios, "Discovering devices via BLE...", func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		var discoverErr error
		devices, discoverErr = bleDiscoverer.DiscoverWithContext(ctx)
		return discoverErr
	})
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
