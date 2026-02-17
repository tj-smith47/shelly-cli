package shelly

import (
	"testing"
	"time"
)

const (
	testDeviceID = "shellyplus1pm-aabbcc"
	testDeviceIP = "192.168.1.100"
)

func TestDiscoveredDevice_Fields(t *testing.T) {
	t.Parallel()

	device := DiscoveredDevice{
		ID:         testDeviceID,
		Name:       "Kitchen Switch",
		Model:      "SNSW-001P16EU",
		Address:    testDeviceIP,
		Generation: 2,
		Firmware:   "1.2.0",
		AuthEn:     true,
		Added:      false,
	}

	if device.ID != testDeviceID {
		t.Errorf("expected ID %q, got %q", testDeviceID, device.ID)
	}
	if device.Name != "Kitchen Switch" {
		t.Errorf("expected Name 'Kitchen Switch', got %q", device.Name)
	}
	if device.Model != "SNSW-001P16EU" {
		t.Errorf("expected Model 'SNSW-001P16EU', got %q", device.Model)
	}
	if device.Address != testDeviceIP {
		t.Errorf("expected Address %q, got %q", testDeviceIP, device.Address)
	}
	if device.Generation != 2 {
		t.Errorf("expected Generation 2, got %d", device.Generation)
	}
	if device.Firmware != "1.2.0" {
		t.Errorf("expected Firmware '1.2.0', got %q", device.Firmware)
	}
	if !device.AuthEn {
		t.Error("expected AuthEn to be true")
	}
	if device.Added {
		t.Error("expected Added to be false")
	}
}

func TestDiscoveryOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := DiscoveryOptions{
		Method:     DiscoveryHTTP,
		Timeout:    15 * time.Second,
		Subnets:    []string{"192.168.1.0/24"},
		AutoDetect: true,
	}

	if opts.Method != DiscoveryHTTP {
		t.Error("expected Method DiscoveryHTTP")
	}
	if opts.Timeout != 15*time.Second {
		t.Errorf("expected Timeout 15s, got %v", opts.Timeout)
	}
	if len(opts.Subnets) != 1 || opts.Subnets[0] != "192.168.1.0/24" {
		t.Errorf("expected Subnets [192.168.1.0/24], got %v", opts.Subnets)
	}
	if !opts.AutoDetect {
		t.Error("expected AutoDetect to be true")
	}
}

func TestDiscoveryMethod_Constants(t *testing.T) {
	t.Parallel()

	if DiscoveryMDNS != 0 {
		t.Error("DiscoveryMDNS should be 0")
	}
	if DiscoveryHTTP != 1 {
		t.Error("DiscoveryHTTP should be 1")
	}
	if DiscoveryCoIoT != 2 {
		t.Error("DiscoveryCoIoT should be 2")
	}
	if DiscoveryBLE != 3 {
		t.Error("DiscoveryBLE should be 3")
	}
}

func TestDefaultDiscoveryTimeout(t *testing.T) {
	t.Parallel()

	if DefaultDiscoveryTimeout != 10*time.Second {
		t.Errorf("expected DefaultDiscoveryTimeout 10s, got %v", DefaultDiscoveryTimeout)
	}
}

func TestNormalizeMAC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"uppercase with colons", "AA:BB:CC:DD:EE:FF", "AABBCCDDEEFF"},
		{"lowercase with colons", "aa:bb:cc:dd:ee:ff", "AABBCCDDEEFF"},
		{"mixed case with colons", "Aa:Bb:Cc:Dd:Ee:Ff", "AABBCCDDEEFF"},
		{"uppercase with dashes", "AA-BB-CC-DD-EE-FF", "AABBCCDDEEFF"},
		{"no separators", "AABBCCDDEEFF", "AABBCCDDEEFF"},
		{"lowercase no separators", "aabbccddeeff", "AABBCCDDEEFF"},
		{"too short", "AA:BB:CC:DD:EE", ""},
		{"too long", "AA:BB:CC:DD:EE:FF:00", ""},
		{"invalid chars", "XX:YY:ZZ:AA:BB:CC", ""},
		{"with dots", "AA.BB.CC.DD.EE.FF", "AABBCCDDEEFF"},
		{"mixed separators", "AA:BB-CC.DD:EE:FF", "AABBCCDDEEFF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeMAC(tt.input)
			if got != tt.want {
				t.Errorf("normalizeMAC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
