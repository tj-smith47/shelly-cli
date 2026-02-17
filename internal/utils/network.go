// Package utils provides shared utility functions for the CLI.
package utils

import (
	"fmt"
	"net"
	"strings"
)

// shellyAPSubnet is the subnet used by Shelly devices in AP mode.
// DetectSubnets skips this to avoid picking up stale static IPs
// left after WiFi provisioning via nl80211.
const shellyAPSubnet = "192.168.33."

// DetectSubnets returns all local network subnets in CIDR notation.
// It iterates all network interfaces and collects valid candidates:
//   - Skips loopback, link-local, and Shelly AP (192.168.33.0/24) addresses
//   - Skips virtual/container interfaces (docker, veth, br-, etc.)
//   - Skips down interfaces
//   - Deduplicates subnets (same interface may have multiple addresses in one subnet)
//
// Returns an error only if no suitable interface is found at all.
func DetectSubnets() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var subnets []string
	var fallbacks []string

	for _, iface := range ifaces {
		// Skip down interfaces.
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Skip loopback.
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip virtual/container interfaces.
		if isVirtualInterface(iface.Name) {
			continue
		}

		addrs, addrErr := iface.Addrs()
		if addrErr != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}

			// Skip link-local (169.254.x.x).
			if ipNet.IP.IsLinkLocalUnicast() {
				continue
			}

			network := ipNet.IP.Mask(ipNet.Mask)
			ones, _ := ipNet.Mask.Size()
			cidr := fmt.Sprintf("%s/%d", network, ones)

			if seen[cidr] {
				continue
			}
			seen[cidr] = true

			// Shelly AP subnet goes to fallback list.
			if strings.HasPrefix(ipNet.IP.String(), shellyAPSubnet) {
				fallbacks = append(fallbacks, cidr)
				continue
			}

			subnets = append(subnets, cidr)
		}
	}

	if len(subnets) > 0 {
		return subnets, nil
	}

	// If all non-loopback addresses were Shelly AP, return those as last resort.
	if len(fallbacks) > 0 {
		return fallbacks, nil
	}

	return nil, fmt.Errorf("no suitable network interface found")
}

// DetectSubnet returns the first detected local network subnet.
// Use DetectSubnets to get all subnets when scanning for devices.
func DetectSubnet() (string, error) {
	subnets, err := DetectSubnets()
	if err != nil {
		return "", err
	}
	return subnets[0], nil
}

// isVirtualInterface returns true for known virtual interface name prefixes.
func isVirtualInterface(name string) bool {
	virtualPrefixes := []string{"docker", "veth", "br-", "virbr", "lxc", "cni", "flannel", "calico"}
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
