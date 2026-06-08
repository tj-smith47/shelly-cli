package cmdutil

import (
	"net"
	"testing"

	"github.com/tj-smith47/shelly-go/discovery"
)

func TestDedupDiscoveredDevices_AddressLessNotCollapsed(t *testing.T) {
	t.Parallel()

	// Two distinct BLE devices: nil Address, no shared ID/MAC. Keying on
	// Address.String() alone would collapse both onto "<nil>" and drop one.
	devices := []discovery.DiscoveredDevice{
		{Name: "shelly-ble-a", Address: nil},
		{Name: "shelly-ble-b", Address: nil},
	}

	got := DedupDiscoveredDevices(devices)
	if len(got) != 2 {
		t.Fatalf("address-less devices collapsed: got %d, want 2", len(got))
	}
}

func TestDedupDiscoveredDevices_PrecedenceIDMACAddress(t *testing.T) {
	t.Parallel()

	ip := net.ParseIP("192.168.1.10")
	devices := []discovery.DiscoveredDevice{
		// First sighting: ID present, so the ID is the dedup key.
		{ID: "shellyplug-aabb", MACAddress: "AA:BB:CC:DD:EE:FF", Address: ip},
		// Same ID discovered again via a second method: must collapse.
		{ID: "shellyplug-aabb"},
		// A device with no ID falls back to its MAC key — distinct identity,
		// so it must survive rather than merge with the ID-keyed entry above.
		{MACAddress: "AA:BB:CC:DD:EE:FF"},
	}

	got := DedupDiscoveredDevices(devices)
	if len(got) != 2 {
		t.Fatalf("expected ID dedup + MAC-keyed survivor, got %d, want 2", len(got))
	}
}

func TestDedupDiscoveredDevices_DistinctIDsKept(t *testing.T) {
	t.Parallel()

	devices := []discovery.DiscoveredDevice{
		{ID: "a", Address: net.ParseIP("10.0.0.1")},
		{ID: "b", Address: net.ParseIP("10.0.0.2")},
	}

	got := DedupDiscoveredDevices(devices)
	if len(got) != 2 {
		t.Fatalf("distinct devices wrongly merged: got %d, want 2", len(got))
	}
}
