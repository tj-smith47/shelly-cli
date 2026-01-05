// Package automation provides script, schedule, and event automation for Shelly devices.
package automation

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// mockConnectionProvider is a test double for ConnectionProvider.
type mockConnectionProvider struct {
	withConnectionFn        func(ctx context.Context, identifier string, fn func(*client.Client) error) error
	resolveWithGenerationFn func(ctx context.Context, identifier string) (model.Device, error)
	getGen1StatusJSONFn     func(ctx context.Context, identifier string) (json.RawMessage, error)
}

func (m *mockConnectionProvider) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionProvider) ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error) {
	if m.resolveWithGenerationFn != nil {
		return m.resolveWithGenerationFn(ctx, identifier)
	}
	return model.Device{}, nil
}

func (m *mockConnectionProvider) GetGen1StatusJSON(ctx context.Context, identifier string) (json.RawMessage, error) {
	if m.getGen1StatusJSONFn != nil {
		return m.getGen1StatusJSONFn(ctx, identifier)
	}
	return nil, nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := New(provider, nil, nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.parent != provider {
		t.Error("expected parent to be set")
	}
	if svc.cache != nil {
		t.Error("expected cache to be nil when not provided")
	}
}

func TestParseScheduleCalls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		callsJSON string
		wantCalls int
		wantErr   bool
	}{
		{
			name:      "single call",
			callsJSON: `[{"method": "Switch.Set", "params": {"id": 0, "on": true}}]`,
			wantCalls: 1,
		},
		{
			name:      "multiple calls",
			callsJSON: `[{"method": "Switch.Set", "params": {"id": 0, "on": true}}, {"method": "Light.Set", "params": {"id": 0, "on": false}}]`,
			wantCalls: 2,
		},
		{
			name:      "empty array",
			callsJSON: `[]`,
			wantCalls: 0,
		},
		{
			name:      "call without params",
			callsJSON: `[{"method": "Shelly.Reboot"}]`,
			wantCalls: 1,
		},
		{
			name:      "invalid JSON",
			callsJSON: `not valid json`,
			wantErr:   true,
		},
		{
			name:      "not an array",
			callsJSON: `{"method": "Switch.Set"}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			calls, err := ParseScheduleCalls(tt.callsJSON)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(calls) != tt.wantCalls {
				t.Errorf("got %d calls, want %d", len(calls), tt.wantCalls)
			}
		})
	}
}

func TestScheduleJob_Fields(t *testing.T) {
	t.Parallel()

	job := ScheduleJob{
		ID:       1,
		Enable:   true,
		Timespec: "0 0 8 * * *",
		Calls: []ScheduleCall{
			{
				Method: "Switch.Set",
				Params: map[string]any{"id": 0, "on": true},
			},
		},
	}

	if job.ID != 1 {
		t.Errorf("got ID=%d, want 1", job.ID)
	}
	if !job.Enable {
		t.Error("expected Enable to be true")
	}
	if job.Timespec != "0 0 8 * * *" {
		t.Errorf("got Timespec=%q, want %q", job.Timespec, "0 0 8 * * *")
	}
	if len(job.Calls) != 1 {
		t.Errorf("got %d calls, want 1", len(job.Calls))
	}
}

func TestScheduleCall_Fields(t *testing.T) {
	t.Parallel()

	call := ScheduleCall{
		Method: "Switch.Set",
		Params: map[string]any{
			"id": 0,
			"on": true,
		},
	}

	if call.Method != "Switch.Set" {
		t.Errorf("got Method=%q, want %q", call.Method, "Switch.Set")
	}
	if call.Params == nil {
		t.Error("expected non-nil Params")
	}
	if id, ok := call.Params["id"].(int); !ok || id != 0 {
		t.Errorf("got id=%v, want 0", call.Params["id"])
	}
}

func TestScriptInfo_Fields(t *testing.T) {
	t.Parallel()

	info := ScriptInfo{
		ID:      1,
		Name:    "test-script",
		Enable:  true,
		Running: true,
	}

	if info.ID != 1 {
		t.Errorf("got ID=%d, want 1", info.ID)
	}
	if info.Name != "test-script" {
		t.Errorf("got Name=%q, want %q", info.Name, "test-script")
	}
	if !info.Enable {
		t.Error("expected Enable to be true")
	}
	if !info.Running {
		t.Error("expected Running to be true")
	}
}

func TestScriptStatus_Fields(t *testing.T) {
	t.Parallel()

	status := ScriptStatus{
		ID:       1,
		Running:  true,
		MemUsage: 1024,
		MemPeak:  2048,
		MemFree:  4096,
		Errors:   []string{"error1", "error2"},
	}

	if status.ID != 1 {
		t.Errorf("got ID=%d, want 1", status.ID)
	}
	if !status.Running {
		t.Error("expected Running to be true")
	}
	if status.MemUsage != 1024 {
		t.Errorf("got MemUsage=%d, want 1024", status.MemUsage)
	}
	if status.MemPeak != 2048 {
		t.Errorf("got MemPeak=%d, want 2048", status.MemPeak)
	}
	if status.MemFree != 4096 {
		t.Errorf("got MemFree=%d, want 4096", status.MemFree)
	}
	if len(status.Errors) != 2 {
		t.Errorf("got %d errors, want 2", len(status.Errors))
	}
}

func TestScriptConfig_Fields(t *testing.T) {
	t.Parallel()

	cfg := ScriptConfig{
		ID:     1,
		Name:   "config-script",
		Enable: true,
	}

	if cfg.ID != 1 {
		t.Errorf("got ID=%d, want 1", cfg.ID)
	}
	if cfg.Name != "config-script" {
		t.Errorf("got Name=%q, want %q", cfg.Name, "config-script")
	}
	if !cfg.Enable {
		t.Error("expected Enable to be true")
	}
}

func TestInstallScriptResult_Fields(t *testing.T) {
	t.Parallel()

	result := InstallScriptResult{
		ID:      5,
		Name:    "installed-script",
		Enabled: true,
	}

	if result.ID != 5 {
		t.Errorf("got ID=%d, want 5", result.ID)
	}
	if result.Name != "installed-script" {
		t.Errorf("got Name=%q, want %q", result.Name, "installed-script")
	}
	if !result.Enabled {
		t.Error("expected Enabled to be true")
	}
}

func TestNewEventStream(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	es := NewEventStream(provider)

	if es == nil {
		t.Fatal("expected non-nil EventStream")
	}
	if es.svc != provider {
		t.Error("expected svc to be set")
	}
	if es.bus == nil {
		t.Error("expected bus to be non-nil")
	}
	if es.connections == nil {
		t.Error("expected connections map to be initialized")
	}
	if es.ctx == nil {
		t.Error("expected ctx to be non-nil")
	}
	if es.cancel == nil {
		t.Error("expected cancel to be non-nil")
	}
}

func TestEventStream_IsConnected(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	es := NewEventStream(provider)

	// Initially not connected
	if es.IsConnected("test-device") {
		t.Error("expected device to not be connected initially")
	}

	// Manually add a connection
	es.mu.Lock()
	es.connections["test-device"] = &deviceConnection{
		name:    "test-device",
		address: "192.168.1.100",
	}
	es.mu.Unlock()

	if !es.IsConnected("test-device") {
		t.Error("expected device to be connected after adding")
	}
}

func TestEventStream_ConnectedDevices(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	es := NewEventStream(provider)

	// Initially empty
	devices := es.ConnectedDevices()
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}

	// Add some connections
	es.mu.Lock()
	es.connections["device1"] = &deviceConnection{name: "device1"}
	es.connections["device2"] = &deviceConnection{name: "device2"}
	es.mu.Unlock()

	devices = es.ConnectedDevices()
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestEventStream_Stop(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	es := NewEventStream(provider)

	// Add a connection
	es.mu.Lock()
	cancelCalled := false
	es.connections["test-device"] = &deviceConnection{
		name:   "test-device",
		cancel: func() { cancelCalled = true },
	}
	es.mu.Unlock()

	es.Stop()

	// Check that connections are cleared
	es.mu.RLock()
	connCount := len(es.connections)
	es.mu.RUnlock()

	if connCount != 0 {
		t.Errorf("expected 0 connections after Stop, got %d", connCount)
	}
	if !cancelCalled {
		t.Error("expected device cancel to be called")
	}
}

func TestEventStream_RemoveDevice(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	es := NewEventStream(provider)

	// Add a connection
	cancelCalled := false
	es.mu.Lock()
	es.connections["test-device"] = &deviceConnection{
		name:   "test-device",
		cancel: func() { cancelCalled = true },
	}
	es.mu.Unlock()

	es.RemoveDevice("test-device")

	if es.IsConnected("test-device") {
		t.Error("expected device to be removed")
	}
	if !cancelCalled {
		t.Error("expected cancel to be called")
	}
}

func TestEventStream_RemoveDevice_NotExists(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	es := NewEventStream(provider)

	// Should not panic when removing non-existent device
	es.RemoveDevice("non-existent")
}
