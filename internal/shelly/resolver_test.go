// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestNewConfigResolver(t *testing.T) {
	t.Parallel()
	resolver := NewConfigResolver()
	if resolver == nil {
		t.Fatal("NewConfigResolver() returned nil")
	}
}

func TestDeviceResolver_Interface(t *testing.T) {
	t.Parallel()
	// Verify ConfigResolver implements DeviceResolver interface
	var _ DeviceResolver = (*ConfigResolver)(nil)
}

func TestDeviceResolver_MockImplementation(t *testing.T) {
	t.Parallel()
	// Verify mockResolver implements DeviceResolver interface
	var _ DeviceResolver = (*mockResolver)(nil)

	mock := &mockResolver{
		device: model.Device{
			Name:    "mock-device",
			Address: "192.168.1.50",
		},
	}

	device, err := mock.Resolve("any")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if device.Name != "mock-device" {
		t.Errorf("Name = %q, want %q", device.Name, "mock-device")
	}
}
