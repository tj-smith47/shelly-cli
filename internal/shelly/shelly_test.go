// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// mockResolver is a mock device resolver for testing.
type mockResolver struct {
	device model.Device
	err    error
}

func (m *mockResolver) Resolve(_ string) (model.Device, error) {
	return m.device, m.err
}

func TestNew(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	if service == nil {
		t.Fatal("New() returned nil")
	}

	if service.resolver != resolver {
		t.Error("Service resolver not set correctly")
	}
}

func TestNewService(t *testing.T) {
	t.Parallel()
	service := NewService()

	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	if service.resolver == nil {
		t.Error("Service resolver is nil")
	}
}

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()
	expected := 10 * time.Second
	if DefaultTimeout != expected {
		t.Errorf("DefaultTimeout = %v, want %v", DefaultTimeout, expected)
	}
}

func TestSwitchInfo_Zero(t *testing.T) {
	t.Parallel()
	var info SwitchInfo
	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("Name = %q, want empty", info.Name)
	}
	if info.Output {
		t.Error("Output = true, want false")
	}
	if info.Power != 0 {
		t.Errorf("Power = %f, want 0", info.Power)
	}
}
