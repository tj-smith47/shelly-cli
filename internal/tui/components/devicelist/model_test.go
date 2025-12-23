package devicelist

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func testIOStreams() *iostreams.IOStreams {
	return iostreams.Test(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
}

func TestDevicesLoadedMsg(t *testing.T) {
	t.Parallel()

	// Create a minimal model
	ctx := context.Background()
	svc := shelly.NewService()
	ios := testIOStreams()

	m := New(Deps{
		Ctx:             ctx,
		Svc:             svc,
		IOS:             ios,
		RefreshInterval: 5 * time.Second,
	})

	// Initially loading should be true
	if !m.loading {
		t.Error("expected loading=true initially")
	}

	// Create a DevicesLoadedMsg with some devices
	devices := []DeviceStatus{
		{
			Device: model.Device{Name: "test1", Address: "10.0.0.1"},
			Online: true,
		},
		{
			Device: model.Device{Name: "test2", Address: "10.0.0.2"},
			Online: false,
		},
	}

	msg := DevicesLoadedMsg{Devices: devices}

	// Send the message to Update
	m, _ = m.Update(msg)

	// Now loading should be false
	if m.loading {
		t.Error("expected loading=false after DevicesLoadedMsg")
	}

	// Devices should be populated
	if len(m.devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(m.devices))
	}

	// Filtered should also be populated (no filter)
	if len(m.filtered) != 2 {
		t.Errorf("expected 2 filtered devices, got %d", len(m.filtered))
	}

	// DeviceCount should work
	if m.DeviceCount() != 2 {
		t.Errorf("expected DeviceCount()=2, got %d", m.DeviceCount())
	}
}

func TestInitReturnsBatchCmd(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := shelly.NewService()
	ios := testIOStreams()

	m := New(Deps{
		Ctx:             ctx,
		Svc:             svc,
		IOS:             ios,
		RefreshInterval: 5 * time.Second,
	})

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a non-nil command")
	}
}

func TestSetFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := shelly.NewService()
	ios := testIOStreams()

	m := New(Deps{
		Ctx:             ctx,
		Svc:             svc,
		IOS:             ios,
		RefreshInterval: 5 * time.Second,
	})

	// Load some devices
	devices := []DeviceStatus{
		{Device: model.Device{Name: "kitchen-light", Address: "10.0.0.1"}, Online: true},
		{Device: model.Device{Name: "bedroom-fan", Address: "10.0.0.2"}, Online: true},
		{Device: model.Device{Name: "kitchen-outlet", Address: "10.0.0.3"}, Online: true},
	}

	m, _ = m.Update(DevicesLoadedMsg{Devices: devices})

	// No filter - all devices
	if m.DeviceCount() != 3 {
		t.Errorf("expected 3 devices, got %d", m.DeviceCount())
	}

	// Filter by kitchen
	m = m.SetFilter("kitchen")
	if m.DeviceCount() != 2 {
		t.Errorf("expected 2 devices with 'kitchen' filter, got %d", m.DeviceCount())
	}

	// Clear filter
	m = m.SetFilter("")
	if m.DeviceCount() != 3 {
		t.Errorf("expected 3 devices after clearing filter, got %d", m.DeviceCount())
	}
}
