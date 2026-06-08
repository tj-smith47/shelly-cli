package wizard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// Discovery method identifiers and the flag aliases that map onto them.
const (
	methodHTTP  = "http"
	methodMDNS  = "mdns"
	methodCoIoT = "coiot"
	methodBLE   = "ble"

	modeAll        = "all"
	aliasScan      = "scan"
	aliasBonjour   = "bonjour"
	aliasZeroconf  = "zeroconf"
	aliasCoAP      = "coap"
	aliasBluetooth = "bluetooth"
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
			methods = append(methods, methodMDNS)
		case strings.HasPrefix(s, "CoIoT"):
			methods = append(methods, methodCoIoT)
		case strings.HasPrefix(s, "BLE"):
			methods = append(methods, methodBLE)
		}
	}

	if len(methods) == 0 {
		methods = []string{methodHTTP} // Fallback
	}
	return methods
}

// parseDiscoverModes parses the --discover-modes flag value.
func parseDiscoverModes(modes string) []string {
	if modes == "" || modes == modeAll {
		return []string{methodHTTP, methodMDNS, methodCoIoT}
	}

	var result []string
	for _, m := range strings.Split(modes, ",") {
		m = strings.TrimSpace(strings.ToLower(m))
		switch m {
		case methodHTTP, aliasScan:
			result = append(result, methodHTTP)
		case methodMDNS, aliasZeroconf, aliasBonjour:
			result = append(result, methodMDNS)
		case methodCoIoT, aliasCoAP:
			result = append(result, methodCoIoT)
		case methodBLE, aliasBluetooth:
			result = append(result, methodBLE)
		}
	}

	if len(result) == 0 {
		result = []string{methodHTTP} // Default fallback
	}
	return result
}

// runDiscoveryMethod runs a single discovery method, delegating to the
// canonical cmdutil discovery helpers so the wizard, discover, and onboard
// paths share one implementation (subnet/address generation, timeouts,
// ctx-driven cancellation, and progress rendering).
func runDiscoveryMethod(ctx context.Context, ios *iostreams.IOStreams, method string, timeout time.Duration, subnets []string) ([]discovery.DiscoveredDevice, error) {
	switch method {
	case methodHTTP:
		return runHTTPDiscovery(ctx, ios, timeout, subnets)
	case methodMDNS:
		// mDNS is a fast multicast sweep, not a per-host scan, so it uses the
		// helper's short default rather than the 2-minute HTTP scan timeout.
		return cmdutil.RunMDNSDiscovery(ctx, ios, timeout)
	case methodCoIoT:
		return cmdutil.RunCoIoTDiscovery(ctx, ios, timeout)
	case methodBLE:
		return cmdutil.RunBLEDiscovery(ctx, ios, timeout)
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

	// Delegate to the single canonical HTTP scan. It applies DefaultScanTimeout
	// when timeout==0, generates probe addresses (with per-subnet CIDR
	// validation), and honors ctx cancellation — keeping the wizard's HTTP path
	// from drifting away from `shelly discover`.
	return cmdutil.RunHTTPDiscovery(ctx, ios, timeout, subnets)
}
