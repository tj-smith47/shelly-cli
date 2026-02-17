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

const (
	defaultDiscoveryTimeout = 15 * time.Second
	methodHTTP              = "http"
)

// selectDiscoveryMethods selects which discovery methods to use.
func selectDiscoveryMethods(ios *iostreams.IOStreams, opts *Options) []string {
	// Non-interactive: parse --discover-modes flag
	if opts.UseDefaults() || opts.Discover {
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
		return []string{methodHTTP} // Fallback to http
	}

	var methods []string
	for _, s := range selected {
		switch {
		case strings.HasPrefix(s, "HTTP"):
			methods = append(methods, methodHTTP)
		case strings.HasPrefix(s, "mDNS"):
			methods = append(methods, "mdns")
		case strings.HasPrefix(s, "CoIoT"):
			methods = append(methods, "coiot")
		case strings.HasPrefix(s, "BLE"):
			methods = append(methods, "ble")
		}
	}

	if len(methods) == 0 {
		methods = []string{methodHTTP} // Fallback
	}
	return methods
}

// parseDiscoverModes parses the --discover-modes flag value.
func parseDiscoverModes(modes string) []string {
	if modes == "" || modes == "all" {
		return []string{methodHTTP, "mdns", "coiot"}
	}

	var result []string
	for _, m := range strings.Split(modes, ",") {
		m = strings.TrimSpace(strings.ToLower(m))
		switch m {
		case methodHTTP, "scan":
			result = append(result, methodHTTP)
		case "mdns", "zeroconf", "bonjour":
			result = append(result, "mdns")
		case "coiot", "coap":
			result = append(result, "coiot")
		case "ble", "bluetooth":
			result = append(result, "ble")
		}
	}

	if len(result) == 0 {
		result = []string{methodHTTP} // Default fallback
	}
	return result
}

// runDiscoveryMethod runs a single discovery method.
func runDiscoveryMethod(ctx context.Context, ios *iostreams.IOStreams, method string, timeout time.Duration, subnets []string) ([]discovery.DiscoveredDevice, error) {
	switch method {
	case methodHTTP:
		return runHTTPDiscovery(ctx, ios, timeout, subnets)
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

func runHTTPDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration, subnets []string) ([]discovery.DiscoveredDevice, error) {
	if len(subnets) == 0 {
		var err error
		subnets, err = utils.DetectSubnets()
		if err != nil {
			return nil, fmt.Errorf("failed to detect subnets: %w", err)
		}
	}

	// Generate addresses across all subnets
	var allAddresses []string
	for _, subnet := range subnets {
		allAddresses = append(allAddresses, discovery.GenerateSubnetAddresses(subnet)...)
	}
	if len(allAddresses) == 0 {
		return nil, fmt.Errorf("no addresses in %s", strings.Join(subnets, ", "))
	}

	subnetLabel := strings.Join(subnets, ", ")
	ios.StartProgress(fmt.Sprintf("Scanning %s (timeout: %s)...", subnetLabel, timeout))
	defer ios.StopProgress()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	devices := discovery.ProbeAddressesWithProgress(ctx, allAddresses, func(_ discovery.ProbeProgress) bool {
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
