// Package utils provides shared utility functions for the CLI.
package utils

import (
	"fmt"
	"net"
	"strings"
)

// shellyAPSubnet is the subnet used by Shelly devices in AP mode.
// DetectSubnet skips this to avoid picking up stale static IPs
// left after WiFi provisioning via nl80211.
const shellyAPSubnet = "192.168.33."

// DetectSubnet attempts to detect the local network subnet.
// It returns a CIDR notation string like "192.168.1.0/24".
//
// It iterates all network interfaces and picks the best candidate:
//   - Skips loopback, link-local, and Shelly AP (192.168.33.0/24) addresses
//   - Prefers interfaces that are up and not virtual (docker, veth, br-, etc.)
//   - Falls back to any non-loopback IPv4 address if no ideal match is found
func DetectSubnet() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var fallback string

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

			// Skip Shelly AP subnet.
			if strings.HasPrefix(ipNet.IP.String(), shellyAPSubnet) {
				if fallback == "" {
					fallback = cidr
				}
				continue
			}

			return cidr, nil
		}
	}

	// If all non-loopback addresses were Shelly AP, return that as last resort.
	if fallback != "" {
		return fallback, nil
	}

	return "", fmt.Errorf("no suitable network interface found")
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
