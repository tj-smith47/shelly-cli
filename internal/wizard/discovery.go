package wizard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

const defaultDiscoveryTimeout = 15 * time.Second

// selectDiscoveryMethods selects which discovery methods to use.
func selectDiscoveryMethods(ios *iostreams.IOStreams, opts *Options) []string {
	// Non-interactive: parse --discover-modes flag
	if opts.IsNonInteractive() {
		return parseDiscoverModes(opts.DiscoverModes)
	}

	// Interactive: multi-select
	options := []string{
		"HTTP (recommended, works everywhere)",
		"mDNS (fast, Gen2+ devices)",
		"CoIoT (Gen1 devices)",
		"BLE (Bluetooth, provisioning mode)",
	}
	defaults := []string{options[0]} // Default to HTTP

	selected, err := ios.MultiSelect("Select discovery methods:", options, defaults)
	if err != nil {
		ios.DebugErr("multi-select", err)
		return []string{"http"} // Fallback to http
	}

	var methods []string
	for _, s := range selected {
		switch {
		case strings.HasPrefix(s, "HTTP"):
			methods = append(methods, "http")
		case strings.HasPrefix(s, "mDNS"):
			methods = append(methods, "mdns")
		case strings.HasPrefix(s, "CoIoT"):
			methods = append(methods, "coiot")
		case strings.HasPrefix(s, "BLE"):
			methods = append(methods, "ble")
		}
	}

	if len(methods) == 0 {
		methods = []string{"http"} // Fallback
	}
	return methods
}

// parseDiscoverModes parses the --discover-modes flag value.
func parseDiscoverModes(modes string) []string {
	if modes == "" || modes == "all" {
		return []string{"http", "mdns", "coiot"}
	}

	var result []string
	for _, m := range strings.Split(modes, ",") {
		m = strings.TrimSpace(strings.ToLower(m))
		switch m {
		case "http", "scan":
			result = append(result, "http")
		case "mdns", "zeroconf", "bonjour":
			result = append(result, "mdns")
		case "coiot", "coap":
			result = append(result, "coiot")
		case "ble", "bluetooth":
			result = append(result, "ble")
		}
	}

	if len(result) == 0 {
		result = []string{"http"} // Default fallback
	}
	return result
}

// runDiscoveryMethod runs a single discovery method.
func runDiscoveryMethod(ctx context.Context, ios *iostreams.IOStreams, method string, timeout time.Duration, subnet string) ([]discovery.DiscoveredDevice, error) {
	switch method {
	case "http":
		return runHTTPDiscovery(ctx, ios, timeout, subnet)
	case "mdns":
		return runMDNSDiscovery(ios, timeout)
	case "coiot":
		return runCoIoTDiscovery(ios, timeout)
	case "ble":
		return runBLEDiscovery(ctx, ios, timeout)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func runHTTPDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration, subnet string) ([]discovery.DiscoveredDevice, error) {
	if subnet == "" {
		var err error
		subnet, err = utils.DetectSubnet()
		if err != nil {
			return nil, fmt.Errorf("failed to detect subnet: %w", err)
		}
	}

	ios.StartProgress(fmt.Sprintf("Scanning %s (timeout: %s)...", subnet, timeout))
	defer ios.StopProgress()

	addresses := discovery.GenerateSubnetAddresses(subnet)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no addresses in subnet %s", subnet)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	devices := discovery.ProbeAddressesWithProgress(ctx, addresses, func(_ discovery.ProbeProgress) bool {
		return ctx.Err() == nil
	})

	return devices, nil
}

func runMDNSDiscovery(ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios.StartProgress(fmt.Sprintf("Discovering via mDNS (timeout: %s)...", timeout))
	defer ios.StopProgress()

	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping mDNS discoverer", err)
		}
	}()

	return mdnsDiscoverer.Discover(timeout)
}

func runCoIoTDiscovery(ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios.StartProgress(fmt.Sprintf("Discovering via CoIoT (timeout: %s)...", timeout))
	defer ios.StopProgress()

	coiotDiscoverer := discovery.NewCoIoTDiscoverer()
	defer func() {
		if err := coiotDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping CoIoT discoverer", err)
		}
	}()

	return coiotDiscoverer.Discover(timeout)
}

func runBLEDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios.StartProgress(fmt.Sprintf("Discovering via BLE (timeout: %s)...", timeout))
	defer ios.StopProgress()

	bleDiscoverer, err := discovery.NewBLEDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("BLE not available: %w", err)
	}
	defer func() {
		if stopErr := bleDiscoverer.Stop(); stopErr != nil {
			ios.DebugErr("stopping BLE discoverer", stopErr)
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return bleDiscoverer.DiscoverWithContext(ctx)
}
