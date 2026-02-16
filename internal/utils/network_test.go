// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"strings"
	"testing"
)

func TestDetectSubnet_Format(t *testing.T) {
	t.Parallel()

	subnet, err := DetectSubnet()
	if err != nil {
		// In CI or containers, there may be no suitable interface
		t.Skipf("DetectSubnet() returned error (may be expected in CI): %v", err)
	}

	// Verify CIDR format
	if !strings.Contains(subnet, "/") {
		t.Errorf("DetectSubnet() = %q, expected CIDR format with /", subnet)
	}

	// Should be IPv4
	parts := strings.Split(subnet, "/")
	if len(parts) != 2 {
		t.Errorf("DetectSubnet() = %q, expected format IP/mask", subnet)
	}

	// Should NOT return the Shelly AP subnet
	if strings.HasPrefix(parts[0], shellyAPSubnet) {
		t.Errorf("DetectSubnet() = %q, should not return Shelly AP subnet", subnet)
	}
}

func TestDetectSubnet_ReturnsValidCIDR(t *testing.T) {
	t.Parallel()

	subnet, err := DetectSubnet()
	if err != nil {
		t.Skipf("DetectSubnet() returned error: %v", err)
	}

	// Check the format looks like x.x.x.x/n
	parts := strings.Split(subnet, "/")
	if len(parts) != 2 {
		t.Fatalf("Invalid CIDR format: %s", subnet)
	}

	// Check IP part has 4 octets
	ipParts := strings.Split(parts[0], ".")
	if len(ipParts) != 4 {
		t.Errorf("IP part should have 4 octets, got: %s", parts[0])
	}
}

func TestIsVirtualInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		iface    string
		expected bool
	}{
		{name: "docker0", iface: "docker0", expected: true},
		{name: "veth123", iface: "veth123abc", expected: true},
		{name: "br-network", iface: "br-abcdef", expected: true},
		{name: "virbr0", iface: "virbr0", expected: true},
		{name: "eth0", iface: "eth0", expected: false},
		{name: "wlan0", iface: "wlan0", expected: false},
		{name: "enp0s3", iface: "enp0s3", expected: false},
		{name: "wlp2s0", iface: "wlp2s0", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isVirtualInterface(tt.iface); got != tt.expected {
				t.Errorf("isVirtualInterface(%q) = %v, want %v", tt.iface, got, tt.expected)
			}
		})
	}
}
