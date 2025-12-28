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

	// Verify it's a network address (ends with .0 for /24 or similar)
	// Note: This may not always be true for all subnets
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
