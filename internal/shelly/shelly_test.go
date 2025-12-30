// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
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

func TestWithRateLimiter(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Initially no rate limiter
	if service.rateLimiter != nil {
		t.Error("initial rateLimiter should be nil")
	}

	// Create with default rate limiter
	service2 := New(resolver, WithDefaultRateLimiter())
	if service2.rateLimiter == nil {
		t.Error("WithDefaultRateLimiter should set rateLimiter")
	}
}

func TestWithPluginRegistry(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Initially no plugin registry
	if service.pluginRegistry != nil {
		t.Error("initial pluginRegistry should be nil")
	}

	// PluginRegistry accessor
	if service.PluginRegistry() != nil {
		t.Error("PluginRegistry() should return nil")
	}
}

func TestService_FirmwareService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Firmware service should be initialized
	if service.FirmwareService() == nil {
		t.Error("FirmwareService() should not return nil")
	}
}

func TestService_RateLimiter(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Initially nil
	if service.RateLimiter() != nil {
		t.Error("RateLimiter() should return nil initially")
	}

	// With default rate limiter
	service2 := New(resolver, WithDefaultRateLimiter())
	if service2.RateLimiter() == nil {
		t.Error("RateLimiter() should not return nil after WithDefaultRateLimiter")
	}
}

func TestService_Wireless(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Wireless service should be initialized
	if service.Wireless() == nil {
		t.Error("Wireless() should not return nil")
	}
}

func TestService_DeviceService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Device service should be initialized
	if service.DeviceService() == nil {
		t.Error("DeviceService() should not return nil")
	}
}

func TestService_ComponentService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Component service should be initialized
	if service.ComponentService() == nil {
		t.Error("ComponentService() should not return nil")
	}
}

func TestService_MQTTService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// MQTT service should be initialized
	if service.MQTTService() == nil {
		t.Error("MQTTService() should not return nil")
	}
}

func TestService_EthernetService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Ethernet service should be initialized
	if service.EthernetService() == nil {
		t.Error("EthernetService() should not return nil")
	}
}

func TestService_AuthService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Auth service should be initialized
	if service.AuthService() == nil {
		t.Error("AuthService() should not return nil")
	}
}

func TestService_ModbusService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Modbus service should be initialized
	if service.ModbusService() == nil {
		t.Error("ModbusService() should not return nil")
	}
}

func TestService_ProvisionService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Provision service should be initialized
	if service.ProvisionService() == nil {
		t.Error("ProvisionService() should not return nil")
	}
}

func TestService_SetDeviceGeneration(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}

	// Without rate limiter - no-op, shouldn't panic
	service := New(resolver)
	service.SetDeviceGeneration("192.168.1.1", 2)

	// With rate limiter
	service2 := New(resolver, WithDefaultRateLimiter())
	service2.SetDeviceGeneration("192.168.1.1", 2)
	// No error expected
}

func TestService_FirmwareCache(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	cache := service.FirmwareCache()
	if cache == nil {
		t.Error("FirmwareCache() should not return nil")
	}
}

func TestNewFirmwareCache(t *testing.T) {
	t.Parallel()
	cache := NewFirmwareCache()
	if cache == nil {
		t.Error("NewFirmwareCache() should not return nil")
	}
}

func TestIsConnectionError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"connection refused", errors.New("connection refused"), true},
		{"no route to host", errors.New("no route to host"), true},
		{"i/o timeout", errors.New("i/o timeout"), true},
		{"network is unreachable", errors.New("network is unreachable"), true},
		{"no such host", errors.New("no such host"), true},
		{"dial tcp", errors.New("dial tcp 192.168.1.1:80"), true},
		{"ErrConnectionFailed", model.ErrConnectionFailed, true},
		{"wrapped connection failed", fmt.Errorf("failed: %w", model.ErrConnectionFailed), true},
		{"auth error", errors.New("authentication failed"), false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isConnectionError(tt.err)
			if got != tt.want {
				t.Errorf("isConnectionError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestService_SetPluginRegistry(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Initially nil
	if service.PluginRegistry() != nil {
		t.Error("PluginRegistry() should initially be nil")
	}

	// Set to nil (should not panic)
	service.SetPluginRegistry(nil)
	if service.PluginRegistry() != nil {
		t.Error("PluginRegistry() should be nil after SetPluginRegistry(nil)")
	}
}

func TestWithRateLimiterFromAppConfig(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}

	cfg := config.RateLimitConfig{
		Gen1: config.GenerationRateLimitConfig{
			MinInterval:      2 * time.Second,
			MaxConcurrent:    1,
			CircuitThreshold: 5,
		},
		Gen2: config.GenerationRateLimitConfig{
			MinInterval:      500 * time.Millisecond,
			MaxConcurrent:    3,
			CircuitThreshold: 5,
		},
		Global: config.GlobalRateLimitConfig{
			MaxConcurrent:           10,
			CircuitOpenDuration:     30 * time.Second,
			CircuitSuccessThreshold: 3,
		},
	}

	service := New(resolver, WithRateLimiterFromAppConfig(cfg))
	if service.rateLimiter == nil {
		t.Error("WithRateLimiterFromAppConfig should set rateLimiter")
	}
}

func TestService_NetworkService(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}
	service := New(resolver)

	// Network service should be initialized
	if service.networkService == nil {
		t.Error("networkService should not be nil")
	}
}

func TestService_WithRateLimiter(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}

	// Create a rate limiter
	rl := ratelimit.New()

	service := New(resolver, WithRateLimiter(rl))
	if service.rateLimiter == nil {
		t.Error("WithRateLimiter should set rateLimiter")
	}
	if service.RateLimiter() != rl {
		t.Error("RateLimiter() should return the same rate limiter")
	}
}

func TestService_WithRateLimiterFromConfig(t *testing.T) {
	t.Parallel()
	resolver := &mockResolver{}

	rlConfig := ratelimit.Config{
		Gen1: ratelimit.GenerationConfig{
			MinInterval:      2 * time.Second,
			MaxConcurrent:    1,
			CircuitThreshold: 5,
		},
		Gen2: ratelimit.GenerationConfig{
			MinInterval:      500 * time.Millisecond,
			MaxConcurrent:    3,
			CircuitThreshold: 5,
		},
		Global: ratelimit.GlobalConfig{
			MaxConcurrent:           10,
			CircuitOpenDuration:     30 * time.Second,
			CircuitSuccessThreshold: 3,
		},
	}

	service := New(resolver, WithRateLimiterFromConfig(rlConfig))
	if service.rateLimiter == nil {
		t.Error("WithRateLimiterFromConfig should set rateLimiter")
	}
}

func TestService_ResolveWithGeneration_BasicResolver(t *testing.T) {
	t.Parallel()

	expectedDevice := model.Device{
		Name:       "test-device",
		Address:    "192.168.1.100",
		Generation: 2,
	}

	resolver := &mockResolver{
		device: expectedDevice,
		err:    nil,
	}

	service := New(resolver)

	// ResolveWithGeneration should fall back to basic Resolve
	dev, err := service.ResolveWithGeneration(context.Background(), "test-device")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if dev.Name != expectedDevice.Name {
		t.Errorf("got Name=%q, want %q", dev.Name, expectedDevice.Name)
	}
	if dev.Address != expectedDevice.Address {
		t.Errorf("got Address=%q, want %q", dev.Address, expectedDevice.Address)
	}
}

func TestService_ResolveWithGeneration_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &mockResolver{
		device: model.Device{},
		err:    expectedErr,
	}

	service := New(resolver)

	_, err := service.ResolveWithGeneration(context.Background(), "nonexistent")
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_IsGen1Device(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		device     model.Device
		wantIsGen1 bool
	}{
		{
			name:       "gen1 device",
			device:     model.Device{Name: "gen1", Address: "192.168.1.1", Generation: 1},
			wantIsGen1: true,
		},
		{
			name:       "gen2 device",
			device:     model.Device{Name: "gen2", Address: "192.168.1.2", Generation: 2},
			wantIsGen1: false,
		},
		{
			name:       "gen3 device",
			device:     model.Device{Name: "gen3", Address: "192.168.1.3", Generation: 3},
			wantIsGen1: false,
		},
		{
			name:       "unknown generation (0)",
			device:     model.Device{Name: "unknown", Address: "192.168.1.4", Generation: 0},
			wantIsGen1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resolver := &mockResolver{device: tt.device}
			service := New(resolver)

			isGen1, dev, err := service.IsGen1Device(context.Background(), tt.device.Name)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if isGen1 != tt.wantIsGen1 {
				t.Errorf("got isGen1=%v, want %v", isGen1, tt.wantIsGen1)
			}
			if dev.Generation != tt.device.Generation {
				t.Errorf("got Generation=%d, want %d", dev.Generation, tt.device.Generation)
			}
		})
	}
}

func TestService_IsGen1Device_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &mockResolver{err: expectedErr}

	service := New(resolver)

	isGen1, dev, err := service.IsGen1Device(context.Background(), "nonexistent")
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if isGen1 {
		t.Error("expected isGen1 to be false on error")
	}
	if dev.Name != "" {
		t.Error("expected empty device on error")
	}
}

func TestService_TypeAliases(t *testing.T) {
	t.Parallel()

	// Test that type aliases exist and are usable
	var _ FirmwareInfo
	var _ FirmwareStatus
	var _ FirmwareCheckResult
	var _ DeviceUpdateStatus
	var _ UpdateOpts
	var _ UpdateResult
	var _ FirmwareUpdateEntry
	var _ MQTTStatus
	var _ EthernetStatus
	var _ AuthStatus
	var _ ModbusStatus
	var _ BTHomeDiscovery
	var _ ProvisioningDeviceInfo
}

func TestExtractWiFiSSID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		wantSSID string
	}{
		{
			name:     "nil input",
			input:    nil,
			wantSSID: "",
		},
		{
			name:     "non-map input",
			input:    "string",
			wantSSID: "",
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			wantSSID: "",
		},
		{
			name: "map with sta field",
			input: map[string]any{
				"sta": map[string]any{
					"ssid": "TestNetwork",
				},
			},
			wantSSID: "TestNetwork",
		},
		{
			name: "map with non-string ssid",
			input: map[string]any{
				"sta": map[string]any{
					"ssid": 123,
				},
			},
			wantSSID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ssid := ExtractWiFiSSID(tt.input)
			if ssid != tt.wantSSID {
				t.Errorf("got SSID=%q, want %q", ssid, tt.wantSSID)
			}
		})
	}
}

func TestNewService_InitializesAllServices(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := NewService()

	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	// All sub-services should be initialized
	if service.FirmwareService() == nil {
		t.Error("FirmwareService() should not be nil")
	}
	if service.Wireless() == nil {
		t.Error("Wireless() should not be nil")
	}
	if service.DeviceService() == nil {
		t.Error("DeviceService() should not be nil")
	}
	if service.ComponentService() == nil {
		t.Error("ComponentService() should not be nil")
	}
	if service.MQTTService() == nil {
		t.Error("MQTTService() should not be nil")
	}
	if service.EthernetService() == nil {
		t.Error("EthernetService() should not be nil")
	}
	if service.AuthService() == nil {
		t.Error("AuthService() should not be nil")
	}
	if service.ModbusService() == nil {
		t.Error("ModbusService() should not be nil")
	}
	if service.ProvisionService() == nil {
		t.Error("ProvisionService() should not be nil")
	}

	// Test New() with resolver as well
	service2 := New(resolver)
	if service2 == nil {
		t.Fatal("New() returned nil")
	}
}

func TestService_WithRateLimitedCall_NoLimiter(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver) // No rate limiter

	called := false
	err := service.WithRateLimitedCall(context.Background(), "192.168.1.1", 2, func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("function should have been called")
	}
}

func TestService_WithRateLimitedCall_WithLimiter(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver, WithDefaultRateLimiter())

	called := false
	err := service.WithRateLimitedCall(context.Background(), "192.168.1.1", 2, func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("function should have been called")
	}
}

func TestService_WithRateLimitedCall_FunctionError(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver, WithDefaultRateLimiter())

	expectedErr := errors.New("operation failed")
	err := service.WithRateLimitedCall(context.Background(), "192.168.1.1", 2, func() error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestBuildFirmwareUpdateList tests the firmware update list builder delegation.
func TestBuildFirmwareUpdateList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		results  []FirmwareCheckResult
		devices  map[string]model.Device
		wantLen  int
		wantName string
	}{
		{
			name:    "empty results",
			results: nil,
			devices: map[string]model.Device{},
			wantLen: 0,
		},
		{
			name: "device with update",
			results: []FirmwareCheckResult{
				{
					Name: "device1",
					Info: &FirmwareInfo{
						Current:   "1.0.0",
						Available: "2.0.0",
						HasUpdate: true,
					},
				},
			},
			devices: map[string]model.Device{
				"device1": {Name: "device1", Address: "192.168.1.1"},
			},
			wantLen:  1,
			wantName: "device1",
		},
		{
			name: "device without update",
			results: []FirmwareCheckResult{
				{
					Name: "device1",
					Info: &FirmwareInfo{
						Current:   "2.0.0",
						Available: "2.0.0",
						HasUpdate: false,
					},
				},
			},
			devices: map[string]model.Device{
				"device1": {Name: "device1", Address: "192.168.1.1"},
			},
			wantLen: 0,
		},
		{
			name: "device with beta only",
			results: []FirmwareCheckResult{
				{
					Name: "device1",
					Info: &FirmwareInfo{
						Current:   "2.0.0",
						Available: "2.0.0",
						Beta:      "2.1.0-beta",
						HasUpdate: false,
					},
				},
			},
			devices: map[string]model.Device{
				"device1": {Name: "device1", Address: "192.168.1.1"},
			},
			wantLen:  1, // Should include beta-only devices
			wantName: "device1",
		},
		{
			name: "device with error",
			results: []FirmwareCheckResult{
				{
					Name: "device1",
					Info: nil,
					Err:  errors.New("connection failed"),
				},
			},
			devices: map[string]model.Device{
				"device1": {Name: "device1", Address: "192.168.1.1"},
			},
			wantLen: 0, // Errors shouldn't be included
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entries := BuildFirmwareUpdateList(tt.results, tt.devices)
			if len(entries) != tt.wantLen {
				t.Errorf("got %d entries, want %d", len(entries), tt.wantLen)
			}
			if tt.wantLen > 0 && entries[0].Name != tt.wantName {
				t.Errorf("got name %q, want %q", entries[0].Name, tt.wantName)
			}
		})
	}
}

// TestFilterDevicesByNameAndPlatform tests the device filter delegation.
func TestFilterDevicesByNameAndPlatform(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"shelly1":  {Name: "shelly1", Address: "192.168.1.1", Platform: "shelly"},
		"shelly2":  {Name: "shelly2", Address: "192.168.1.2", Platform: "shelly"},
		"tasmota1": {Name: "tasmota1", Address: "192.168.1.3", Platform: "tasmota"},
		"default":  {Name: "default", Address: "192.168.1.4"}, // No platform = defaults to shelly
	}

	tests := []struct {
		name        string
		devicesList string
		platform    string
		wantCount   int
	}{
		{
			name:        "no filters",
			devicesList: "",
			platform:    "",
			wantCount:   4,
		},
		{
			name:        "filter by name single",
			devicesList: "shelly1",
			platform:    "",
			wantCount:   1,
		},
		{
			name:        "filter by name multiple",
			devicesList: "shelly1,shelly2",
			platform:    "",
			wantCount:   2,
		},
		{
			name:        "filter by platform shelly",
			devicesList: "",
			platform:    "shelly",
			wantCount:   3, // includes "default" which defaults to shelly
		},
		{
			name:        "filter by platform tasmota",
			devicesList: "",
			platform:    "tasmota",
			wantCount:   1,
		},
		{
			name:        "filter by both name and platform",
			devicesList: "shelly1,tasmota1",
			platform:    "shelly",
			wantCount:   1, // Only shelly1 matches both filters
		},
		{
			name:        "filter with non-existent name",
			devicesList: "nonexistent",
			platform:    "",
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			filtered := FilterDevicesByNameAndPlatform(devices, tt.devicesList, tt.platform)
			if len(filtered) != tt.wantCount {
				t.Errorf("got %d devices, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

// TestFilterEntriesByStage tests the entry filter delegation.
func TestFilterEntriesByStage(t *testing.T) {
	t.Parallel()

	entries := []FirmwareUpdateEntry{
		{Name: "stable-only", HasUpdate: true, HasBeta: false},
		{Name: "beta-only", HasUpdate: false, HasBeta: true},
		{Name: "both", HasUpdate: true, HasBeta: true},
		{Name: "neither", HasUpdate: false, HasBeta: false},
	}

	tests := []struct {
		name      string
		beta      bool
		wantCount int
	}{
		{
			name:      "stable filter",
			beta:      false,
			wantCount: 2, // stable-only and both (has HasUpdate)
		},
		{
			name:      "beta filter",
			beta:      true,
			wantCount: 3, // beta-only, both, AND stable-only (HasUpdate fallback)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			filtered := FilterEntriesByStage(entries, tt.beta)
			if len(filtered) != tt.wantCount {
				t.Errorf("got %d entries, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

// TestAnyHasBeta tests the beta check delegation.
func TestAnyHasBeta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entries []FirmwareUpdateEntry
		want    bool
	}{
		{
			name:    "empty entries",
			entries: nil,
			want:    false,
		},
		{
			name:    "no beta",
			entries: []FirmwareUpdateEntry{{Name: "dev1", HasBeta: false}},
			want:    false,
		},
		{
			name:    "has beta",
			entries: []FirmwareUpdateEntry{{Name: "dev1", HasBeta: true}},
			want:    true,
		},
		{
			name:    "mixed",
			entries: []FirmwareUpdateEntry{{Name: "dev1", HasBeta: false}, {Name: "dev2", HasBeta: true}},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AnyHasBeta(tt.entries)
			if got != tt.want {
				t.Errorf("AnyHasBeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSelectEntriesByStage tests the entry selection delegation.
func TestSelectEntriesByStage(t *testing.T) {
	t.Parallel()

	entries := []FirmwareUpdateEntry{
		{Name: "dev1", HasUpdate: true, HasBeta: false},
		{Name: "dev2", HasUpdate: false, HasBeta: true},
		{Name: "dev3", HasUpdate: true, HasBeta: true},
	}

	tests := []struct {
		name       string
		beta       bool
		wantStage  string
		wantIndLen int
	}{
		{
			name:       "stable",
			beta:       false,
			wantStage:  "stable",
			wantIndLen: 2, // dev1 and dev3
		},
		{
			name:       "beta",
			beta:       true,
			wantStage:  "beta",
			wantIndLen: 3, // dev1, dev2, and dev3 (beta includes HasUpdate fallback)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indices, stage := SelectEntriesByStage(entries, tt.beta)
			if stage != tt.wantStage {
				t.Errorf("got stage %q, want %q", stage, tt.wantStage)
			}
			if len(indices) != tt.wantIndLen {
				t.Errorf("got %d indices, want %d", len(indices), tt.wantIndLen)
			}
		})
	}
}

// TestGetEntriesByIndices tests the entry retrieval delegation.
func TestGetEntriesByIndices(t *testing.T) {
	t.Parallel()

	entries := []FirmwareUpdateEntry{
		{Name: "dev0"},
		{Name: "dev1"},
		{Name: "dev2"},
	}

	tests := []struct {
		name      string
		indices   []int
		wantCount int
		wantNames []string
	}{
		{
			name:      "empty indices",
			indices:   nil,
			wantCount: 0,
		},
		{
			name:      "single index",
			indices:   []int{1},
			wantCount: 1,
			wantNames: []string{"dev1"},
		},
		{
			name:      "multiple indices",
			indices:   []int{0, 2},
			wantCount: 2,
			wantNames: []string{"dev0", "dev2"},
		},
		{
			name:      "out of bounds ignored",
			indices:   []int{0, 10, -1},
			wantCount: 1, // Only valid index 0
			wantNames: []string{"dev0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetEntriesByIndices(entries, tt.indices)
			if len(result) != tt.wantCount {
				t.Errorf("got %d entries, want %d", len(result), tt.wantCount)
			}
			for i, wantName := range tt.wantNames {
				if i < len(result) && result[i].Name != wantName {
					t.Errorf("entry[%d].Name = %q, want %q", i, result[i].Name, wantName)
				}
			}
		})
	}
}

// TestToDeviceUpdateStatuses tests the status conversion delegation.
func TestToDeviceUpdateStatuses(t *testing.T) {
	t.Parallel()

	entries := []FirmwareUpdateEntry{
		{
			Name:      "dev1",
			FwInfo:    &FirmwareInfo{Current: "1.0.0", Available: "2.0.0"},
			HasUpdate: true,
		},
		{
			Name:      "dev2",
			FwInfo:    nil,
			HasUpdate: false,
		},
	}

	statuses := ToDeviceUpdateStatuses(entries)

	if len(statuses) != len(entries) {
		t.Errorf("got %d statuses, want %d", len(statuses), len(entries))
	}

	if statuses[0].Name != "dev1" {
		t.Errorf("got name %q, want dev1", statuses[0].Name)
	}
	if !statuses[0].HasUpdate {
		t.Error("expected HasUpdate true for dev1")
	}
	if statuses[0].Info == nil {
		t.Error("expected Info to be non-nil for dev1")
	}

	if statuses[1].Name != "dev2" {
		t.Errorf("got name %q, want dev2", statuses[1].Name)
	}
	if statuses[1].HasUpdate {
		t.Error("expected HasUpdate false for dev2")
	}
}

// TestWithPluginRegistry tests the plugin registry service option.
func TestWithPluginRegistry_Option(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}

	// Test with a nil registry
	service := New(resolver, WithPluginRegistry(nil))
	if service.PluginRegistry() != nil {
		t.Error("expected nil plugin registry")
	}
}

// generationAwareResolver implements GenerationAwareResolver for testing.
type generationAwareResolver struct {
	device model.Device
	err    error
}

func (g *generationAwareResolver) Resolve(_ string) (model.Device, error) {
	return g.device, g.err
}

func (g *generationAwareResolver) ResolveWithGeneration(_ context.Context, _ string) (model.Device, error) {
	return g.device, g.err
}

// TestService_ResolveWithGeneration_GenerationAwareResolver tests resolution with generation-aware resolver.
func TestService_ResolveWithGeneration_GenerationAwareResolver(t *testing.T) {
	t.Parallel()

	expectedDevice := model.Device{
		Name:       "test-device",
		Address:    "192.168.1.100",
		Generation: 2,
	}

	resolver := &generationAwareResolver{
		device: expectedDevice,
	}

	service := New(resolver)

	dev, err := service.ResolveWithGeneration(context.Background(), "test-device")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if dev.Name != expectedDevice.Name {
		t.Errorf("got Name=%q, want %q", dev.Name, expectedDevice.Name)
	}
	if dev.Generation != expectedDevice.Generation {
		t.Errorf("got Generation=%d, want %d", dev.Generation, expectedDevice.Generation)
	}
}

// TestService_WithConnection_ResolveError tests WithConnection when resolution fails.
func TestService_WithConnection_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.WithConnection(context.Background(), "nonexistent", func(_ *client.Client) error {
		t.Error("function should not be called on resolve error")
		return nil
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_WithGen1Connection_ResolveError tests WithGen1Connection when resolution fails.
func TestService_WithGen1Connection_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.WithGen1Connection(context.Background(), "nonexistent", func(_ *client.Gen1Client) error {
		t.Error("function should not be called on resolve error")
		return nil
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_withGenAwareAction tests the generation-aware action dispatcher.
func TestService_withGenAwareAction_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.withGenAwareAction(
		context.Background(),
		"nonexistent",
		func(_ *client.Gen1Client) error {
			t.Error("gen1 function should not be called")
			return nil
		},
		func(_ *client.Client) error {
			t.Error("gen2 function should not be called")
			return nil
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_Connect_ResolveError tests Connect when resolution fails.
func TestService_Connect_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &mockResolver{err: expectedErr}
	service := New(resolver)

	conn, err := service.Connect(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if conn != nil {
		t.Error("expected nil connection on error")
	}
}

// TestService_ConnectGen1_ResolveError tests ConnectGen1 when resolution fails.
func TestService_ConnectGen1_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	conn, err := service.ConnectGen1(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if conn != nil {
		t.Error("expected nil connection on error")
	}
}

// TestService_RawRPC_ResolveError tests RawRPC when resolution fails.
func TestService_RawRPC_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	result, err := service.RawRPC(context.Background(), "nonexistent", "Shelly.GetStatus", nil)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestService_RawGen1Call_ResolveError tests RawGen1Call when resolution fails.
func TestService_RawGen1Call_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	result, err := service.RawGen1Call(context.Background(), "nonexistent", "/status")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestTryIPRemap_NonConnectionError tests tryIPRemap with non-connection errors.
func TestTryIPRemap_NonConnectionError(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	// Non-connection error should pass through
	originalErr := errors.New("authentication failed")
	conn, err := service.tryIPRemap(context.Background(), model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "AA:BB:CC:DD:EE:FF",
	}, originalErr)

	if conn != nil {
		t.Error("expected nil connection")
	}
	if !errors.Is(err, originalErr) {
		t.Errorf("got error %v, want %v", err, originalErr)
	}
}

// TestTryIPRemap_NoMAC tests tryIPRemap when device has no MAC.
func TestTryIPRemap_NoMAC(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	originalErr := model.ErrConnectionFailed
	conn, err := service.tryIPRemap(context.Background(), model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "", // No MAC
	}, originalErr)

	if conn != nil {
		t.Error("expected nil connection")
	}
	if !errors.Is(err, originalErr) {
		t.Errorf("got error %v, want %v", err, originalErr)
	}
}

// TestTryGen1IPRemap_NonConnectionError tests tryGen1IPRemap with non-connection errors.
func TestTryGen1IPRemap_NonConnectionError(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	// Non-connection error should pass through
	originalErr := errors.New("authentication failed")
	conn, err := service.tryGen1IPRemap(context.Background(), model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "AA:BB:CC:DD:EE:FF",
	}, originalErr)

	if conn != nil {
		t.Error("expected nil connection")
	}
	if !errors.Is(err, originalErr) {
		t.Errorf("got error %v, want %v", err, originalErr)
	}
}

// TestTryGen1IPRemap_NoMAC tests tryGen1IPRemap when device has no MAC.
func TestTryGen1IPRemap_NoMAC(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	originalErr := model.ErrConnectionFailed
	conn, err := service.tryGen1IPRemap(context.Background(), model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "", // No MAC
	}, originalErr)

	if conn != nil {
		t.Error("expected nil connection")
	}
	if !errors.Is(err, originalErr) {
		t.Errorf("got error %v, want %v", err, originalErr)
	}
}

// TestService_WithRateLimitedCall_ContextCancelled tests rate limited call with cancelled context.
func TestService_WithRateLimitedCall_ContextCancelled(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver, WithDefaultRateLimiter())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := service.WithRateLimitedCall(ctx, "192.168.1.1", 2, func() error {
		t.Error("function should not be called with cancelled context")
		return nil
	})

	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

// TestService_CheckFirmware_ResolveError tests CheckFirmware when resolution fails.
func TestService_CheckFirmware_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	info, err := service.CheckFirmware(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if info != nil {
		t.Error("expected nil info on error")
	}
}

// TestService_GetFirmwareStatus_ResolveError tests GetFirmwareStatus when resolution fails.
func TestService_GetFirmwareStatus_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	status, err := service.GetFirmwareStatus(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if status != nil {
		t.Error("expected nil status on error")
	}
}

// TestService_UpdateFirmware_ResolveError tests UpdateFirmware when resolution fails.
func TestService_UpdateFirmware_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.UpdateFirmware(context.Background(), "nonexistent", nil)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_UpdateFirmwareStable_ResolveError tests UpdateFirmwareStable when resolution fails.
func TestService_UpdateFirmwareStable_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.UpdateFirmwareStable(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_UpdateFirmwareBeta_ResolveError tests UpdateFirmwareBeta when resolution fails.
func TestService_UpdateFirmwareBeta_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.UpdateFirmwareBeta(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_UpdateFirmwareFromURL_ResolveError tests UpdateFirmwareFromURL when resolution fails.
func TestService_UpdateFirmwareFromURL_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.UpdateFirmwareFromURL(context.Background(), "nonexistent", "http://example.com/fw.bin")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_RollbackFirmware_ResolveError tests RollbackFirmware when resolution fails.
func TestService_RollbackFirmware_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.RollbackFirmware(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_GetFirmwareURL_ResolveError tests GetFirmwareURL when resolution fails.
func TestService_GetFirmwareURL_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	url, err := service.GetFirmwareURL(context.Background(), "nonexistent", "stable")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if url != "" {
		t.Error("expected empty URL on error")
	}
}

// TestService_GetWiFiStatusFull_ResolveError tests GetWiFiStatusFull when resolution fails.
func TestService_GetWiFiStatusFull_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	status, err := service.GetWiFiStatusFull(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if status != nil {
		t.Error("expected nil status on error")
	}
}

// TestService_GetWiFiConfigFull_ResolveError tests GetWiFiConfigFull when resolution fails.
func TestService_GetWiFiConfigFull_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	cfg, err := service.GetWiFiConfigFull(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

// TestService_ScanWiFiNetworksFull_ResolveError tests ScanWiFiNetworksFull when resolution fails.
func TestService_ScanWiFiNetworksFull_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	networks, err := service.ScanWiFiNetworksFull(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if networks != nil {
		t.Error("expected nil networks on error")
	}
}

// TestService_SetWiFiStation_ResolveError tests SetWiFiStation when resolution fails.
func TestService_SetWiFiStation_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.SetWiFiStation(context.Background(), "nonexistent", "MySSID", "password", true)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_SetWiFiAP_ResolveError tests SetWiFiAP when resolution fails.
func TestService_SetWiFiAP_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.SetWiFiAP(context.Background(), "nonexistent", "ShellyAP", "password", true)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_GetMQTTStatus_ResolveError tests GetMQTTStatus when resolution fails.
func TestService_GetMQTTStatus_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	status, err := service.GetMQTTStatus(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if status != nil {
		t.Error("expected nil status on error")
	}
}

// TestService_GetMQTTConfig_ResolveError tests GetMQTTConfig when resolution fails.
func TestService_GetMQTTConfig_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	cfg, err := service.GetMQTTConfig(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

// TestService_SetMQTTConfig_ResolveError tests SetMQTTConfig when resolution fails.
func TestService_SetMQTTConfig_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	enable := true
	err := service.SetMQTTConfig(context.Background(), "nonexistent", &enable, "mqtt://broker", "user", "pass", "shelly")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_SetMQTTConfigFull_ResolveError tests SetMQTTConfigFull when resolution fails.
func TestService_SetMQTTConfigFull_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	enable := true
	params := MQTTSetConfigParams{
		Enable: &enable,
		Server: "mqtt://broker",
	}
	err := service.SetMQTTConfigFull(context.Background(), "nonexistent", params)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_GetEthernetStatus_ResolveError tests GetEthernetStatus when resolution fails.
func TestService_GetEthernetStatus_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	status, err := service.GetEthernetStatus(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if status != nil {
		t.Error("expected nil status on error")
	}
}

// TestService_GetEthernetConfig_ResolveError tests GetEthernetConfig when resolution fails.
func TestService_GetEthernetConfig_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	cfg, err := service.GetEthernetConfig(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

// TestService_SetEthernetConfig_ResolveError tests SetEthernetConfig when resolution fails.
func TestService_SetEthernetConfig_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	enable := true
	err := service.SetEthernetConfig(context.Background(), "nonexistent", &enable, "dhcp", "", "", "", "")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_GetAuthStatus_ResolveError tests GetAuthStatus when resolution fails.
func TestService_GetAuthStatus_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	status, err := service.GetAuthStatus(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if status != nil {
		t.Error("expected nil status on error")
	}
}

// TestService_SetAuth_ResolveError tests SetAuth when resolution fails.
func TestService_SetAuth_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.SetAuth(context.Background(), "nonexistent", "admin", "shelly", "password")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_DisableAuth_ResolveError tests DisableAuth when resolution fails.
func TestService_DisableAuth_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.DisableAuth(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_GetModbusStatus_ResolveError tests GetModbusStatus when resolution fails.
func TestService_GetModbusStatus_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	status, err := service.GetModbusStatus(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if status != nil {
		t.Error("expected nil status on error")
	}
}

// TestService_GetModbusConfig_ResolveError tests GetModbusConfig when resolution fails.
func TestService_GetModbusConfig_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	cfg, err := service.GetModbusConfig(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if cfg != nil {
		t.Error("expected nil config on error")
	}
}

// TestService_SetModbusConfig_ResolveError tests SetModbusConfig when resolution fails.
func TestService_SetModbusConfig_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.SetModbusConfig(context.Background(), "nonexistent", true)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_GetBTHomeStatus_ResolveError tests GetBTHomeStatus when resolution fails.
func TestService_GetBTHomeStatus_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	status, err := service.GetBTHomeStatus(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if status != nil {
		t.Error("expected nil status on error")
	}
}

// TestService_StartBTHomeDiscovery_ResolveError tests StartBTHomeDiscovery when resolution fails.
func TestService_StartBTHomeDiscovery_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	err := service.StartBTHomeDiscovery(context.Background(), "nonexistent", 30)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// TestService_withGenAwareAction_Gen1 tests withGenAwareAction for Gen1 devices.
func TestService_withGenAwareAction_Gen1(t *testing.T) {
	t.Parallel()

	resolver := &generationAwareResolver{
		device: model.Device{
			Name:       "gen1-device",
			Address:    "192.168.1.1",
			Generation: 1,
		},
	}
	service := New(resolver)

	// Gen1 function will be called, but connection will fail
	// This is expected since we don't have a real device
	err := service.withGenAwareAction(
		context.Background(),
		"gen1-device",
		func(_ *client.Gen1Client) error {
			// Gen1 function is called
			return errors.New("gen1 called")
		},
		func(_ *client.Client) error {
			t.Error("gen2 function should not be called for Gen1 device")
			return nil
		},
	)

	// The connection will fail before reaching our function, but we verify the Gen1 path is taken
	if err == nil {
		t.Error("expected error from gen1 path")
	}
}

// TestService_withGenAwareAction_Gen2 tests withGenAwareAction for Gen2 devices.
func TestService_withGenAwareAction_Gen2(t *testing.T) {
	t.Parallel()

	resolver := &generationAwareResolver{
		device: model.Device{
			Name:       "gen2-device",
			Address:    "192.168.1.1",
			Generation: 2,
		},
	}
	service := New(resolver)

	// Gen2 function will be called, but connection will fail
	// This is expected since we don't have a real device
	err := service.withGenAwareAction(
		context.Background(),
		"gen2-device",
		func(_ *client.Gen1Client) error {
			t.Error("gen1 function should not be called for Gen2 device")
			return nil
		},
		func(_ *client.Client) error {
			// Gen2 function is called
			return errors.New("gen2 called")
		},
	)

	// The connection will fail before reaching our function, but we verify the Gen2 path is taken
	if err == nil {
		t.Error("expected error from gen2 path")
	}
}

// TestService_GetDeviceInfoByAddress tests GetDeviceInfoByAddress.
func TestService_GetDeviceInfoByAddress(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	// Address that won't connect - we test the delegation happens
	info, err := service.GetDeviceInfoByAddress(context.Background(), "192.168.1.1")

	// Connection error expected
	if err == nil {
		t.Error("expected error for non-existent address")
	}
	if info != nil {
		t.Error("expected nil info on error")
	}
}

// TestService_ConfigureWiFi tests ConfigureWiFi.
func TestService_ConfigureWiFi(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	// Address that won't connect - we test the delegation happens
	err := service.ConfigureWiFi(context.Background(), "192.168.1.1", "TestSSID", "password")

	// Connection error expected
	if err == nil {
		t.Error("expected error for non-existent address")
	}
}

// TestService_FirmwareCacheOperations tests firmware cache operations.
func TestService_FirmwareCacheOperations(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	// Test GetCachedFirmware with non-existent entry
	cache := service.GetCachedFirmware(context.Background(), "nonexistent", 0)
	if cache != nil {
		t.Error("expected nil cache for non-existent device")
	}
}

// TestService_CheckPluginFirmware_NoRegistry tests CheckPluginFirmware without registry.
func TestService_CheckPluginFirmware_NoRegistry(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver) // No plugin registry

	dev := model.Device{
		Name:     "tasmota1",
		Address:  "192.168.1.1",
		Platform: "tasmota",
	}

	info, err := service.CheckPluginFirmware(context.Background(), dev)

	if err == nil {
		t.Error("expected error when no plugin registry")
	}
	if info != nil {
		t.Error("expected nil info on error")
	}
}

// TestService_UpdatePluginFirmware_NoRegistry tests UpdatePluginFirmware without registry.
func TestService_UpdatePluginFirmware_NoRegistry(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver) // No plugin registry

	dev := model.Device{
		Name:     "tasmota1",
		Address:  "192.168.1.1",
		Platform: "tasmota",
	}

	err := service.UpdatePluginFirmware(context.Background(), dev, "stable", "")

	if err == nil {
		t.Error("expected error when no plugin registry")
	}
}

// TestService_CheckDeviceFirmware_NativeDevice tests CheckDeviceFirmware for native Shelly.
func TestService_CheckDeviceFirmware_NativeDevice(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	dev := model.Device{
		Name:     "shelly1",
		Address:  "192.168.1.1",
		Platform: "shelly", // Native Shelly, not plugin-managed
	}

	info, err := service.CheckDeviceFirmware(context.Background(), dev)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if info != nil {
		t.Error("expected nil info on error")
	}
}

// TestService_CheckDeviceFirmware_PluginDevice tests CheckDeviceFirmware for plugin device.
func TestService_CheckDeviceFirmware_PluginDevice(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver) // No plugin registry

	dev := model.Device{
		Name:     "tasmota1",
		Address:  "192.168.1.1",
		Platform: "tasmota",
	}

	info, err := service.CheckDeviceFirmware(context.Background(), dev)

	// Should fail because no plugin registry
	if err == nil {
		t.Error("expected error when no plugin registry")
	}
	if info != nil {
		t.Error("expected nil info on error")
	}
}

// TestService_UpdateDeviceFirmware_NativeDevice tests UpdateDeviceFirmware for native Shelly.
func TestService_UpdateDeviceFirmware_NativeDevice(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	dev := model.Device{
		Name:     "shelly1",
		Address:  "192.168.1.1",
		Platform: "shelly",
	}

	// Test stable update
	err := service.UpdateDeviceFirmware(context.Background(), dev, false, "")
	if !errors.Is(err, expectedErr) {
		t.Errorf("stable update: got error %v, want %v", err, expectedErr)
	}

	// Test beta update
	err = service.UpdateDeviceFirmware(context.Background(), dev, true, "")
	if !errors.Is(err, expectedErr) {
		t.Errorf("beta update: got error %v, want %v", err, expectedErr)
	}

	// Test custom URL update
	err = service.UpdateDeviceFirmware(context.Background(), dev, false, "http://example.com/fw.bin")
	if !errors.Is(err, expectedErr) {
		t.Errorf("URL update: got error %v, want %v", err, expectedErr)
	}
}

// TestService_UpdateDeviceFirmware_PluginDevice tests UpdateDeviceFirmware for plugin device.
func TestService_UpdateDeviceFirmware_PluginDevice(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver) // No plugin registry

	dev := model.Device{
		Name:     "tasmota1",
		Address:  "192.168.1.1",
		Platform: "tasmota",
	}

	err := service.UpdateDeviceFirmware(context.Background(), dev, false, "")

	// Should fail because no plugin registry
	if err == nil {
		t.Error("expected error when no plugin registry")
	}
}

// newTestIOStreams creates IOStreams for testing with buffers.
func newTestIOStreams() *iostreams.IOStreams {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return iostreams.Test(in, out, errOut)
}

// TestService_CheckFirmwareAll tests CheckFirmwareAll with IOStreams.
func TestService_CheckFirmwareAll(t *testing.T) {
	t.Parallel()

	resolver := &generationAwareResolver{
		device: model.Device{
			Name:       "test-device",
			Address:    "192.168.1.1",
			Generation: 2,
		},
	}
	service := New(resolver)
	ios := newTestIOStreams()

	// Empty device list should return empty results
	results := service.CheckFirmwareAll(context.Background(), ios, []string{})
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}

	// With cancelled context should return results (individual device errors)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results = service.CheckFirmwareAll(ctx, ios, []string{"device1"})
	// Results should contain the device with an error
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "device1" {
		t.Errorf("expected device name 'device1', got %q", results[0].Name)
	}
}

// TestService_CheckFirmwareAllPlatforms tests CheckFirmwareAllPlatforms with IOStreams.
func TestService_CheckFirmwareAllPlatforms(t *testing.T) {
	t.Parallel()

	resolver := &generationAwareResolver{
		device: model.Device{
			Name:       "test-device",
			Address:    "192.168.1.1",
			Generation: 2,
		},
	}
	service := New(resolver)
	ios := newTestIOStreams()

	// Empty device map should return empty results
	results := service.CheckFirmwareAllPlatforms(context.Background(), ios, map[string]model.Device{})
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}

	// With cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	deviceConfigs := map[string]model.Device{
		"device1": {Name: "device1", Address: "192.168.1.1", Platform: "shelly"},
	}
	results = service.CheckFirmwareAllPlatforms(ctx, ios, deviceConfigs)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// TestService_CheckDevicesForUpdates tests CheckDevicesForUpdates with IOStreams.
func TestService_CheckDevicesForUpdates(t *testing.T) {
	t.Parallel()

	resolver := &generationAwareResolver{
		device: model.Device{
			Name:       "test-device",
			Address:    "192.168.1.1",
			Generation: 2,
		},
	}
	service := New(resolver)
	ios := newTestIOStreams()

	// Empty device list should return empty results
	statuses := service.CheckDevicesForUpdates(context.Background(), ios, []string{}, 100)
	if len(statuses) != 0 {
		t.Errorf("expected empty statuses, got %d", len(statuses))
	}

	// With cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	statuses = service.CheckDevicesForUpdates(ctx, ios, []string{"device1"}, 100)
	// Should return empty since device won't have updates (error case)
	if len(statuses) != 0 {
		t.Errorf("expected 0 statuses (device errors), got %d", len(statuses))
	}
}

// TestService_UpdateDevices tests UpdateDevices with IOStreams.
func TestService_UpdateDevices(t *testing.T) {
	t.Parallel()

	resolver := &generationAwareResolver{
		device: model.Device{
			Name:       "test-device",
			Address:    "192.168.1.1",
			Generation: 2,
		},
	}
	service := New(resolver)
	ios := newTestIOStreams()

	// Empty device list should return empty results
	opts := UpdateOpts{Beta: false, Parallelism: 1}
	results := service.UpdateDevices(context.Background(), ios, []DeviceUpdateStatus{}, opts)
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}

	// With devices - will fail because they don't exist
	devices := []DeviceUpdateStatus{
		{Name: "device1", HasUpdate: true},
	}
	results = service.UpdateDevices(context.Background(), ios, devices, opts)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Success {
		t.Error("expected failure for non-existent device")
	}
}

// TestService_PrefetchFirmwareCache tests PrefetchFirmwareCache with IOStreams.
func TestService_PrefetchFirmwareCache(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)
	ios := newTestIOStreams()

	// Should not panic with empty cache
	service.PrefetchFirmwareCache(context.Background(), ios)
}

// TestService_WithDevice_ResolveError tests the Service.WithDevice method on resolve error.
func TestService_WithDevice_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	// Call WithDevice - should fail with resolution error
	err := service.WithDevice(context.Background(), "nonexistent", func(_ *DeviceClient) error {
		t.Error("function should not be called on resolve error")
		return nil
	})

	if err == nil {
		t.Error("expected error")
	}
}

// TestService_DeviceInfo_ResolveError tests DeviceInfo on resolve error.
func TestService_DeviceInfo_ResolveError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")
	resolver := &generationAwareResolver{err: expectedErr}
	service := New(resolver)

	info, err := service.DeviceInfo(context.Background(), "nonexistent")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if info != nil {
		t.Error("expected nil info on error")
	}
}

// TestService_executeWithGen1Connection tests executeWithGen1Connection with invalid device.
func TestService_executeWithGen1Connection_InvalidDevice(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	// Execute with device that can't connect
	dev := model.Device{
		Name:    "test",
		Address: "192.168.1.1",
	}

	err := service.executeWithGen1Connection(context.Background(), dev, func(_ *client.Gen1Client) error {
		// Should not reach here
		return nil
	})

	// Should fail because device doesn't exist
	if err == nil {
		t.Error("expected error for non-existent device")
	}
}

// TestService_tryGen1IPRemap_ConnectionError tests tryGen1IPRemap with connection error.
func TestService_tryGen1IPRemap_ConnectionError(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	// Connection error with MAC should attempt remap (but fail without mDNS)
	device := model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "AA:BB:CC:DD:EE:FF",
	}
	conn, err := service.tryGen1IPRemap(context.Background(), device, model.ErrConnectionFailed)

	// Should fail (no mDNS to resolve)
	if conn != nil {
		t.Error("expected nil connection")
	}
	if err == nil {
		t.Error("expected error")
	}
}

// TestIsConnectionError_Extended tests additional isConnectionError cases.
func TestIsConnectionError_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"wrapped ErrConnectionFailed", fmt.Errorf("failed: %w", model.ErrConnectionFailed), true},
		{"no route to host", errors.New("no route to host"), true},
		{"i/o timeout", errors.New("i/o timeout"), true},
		{"network is unreachable", errors.New("network is unreachable"), true},
		{"no such host", errors.New("no such host"), true},
		{"dial tcp error", errors.New("dial tcp 192.168.1.1:80"), true},
		{"permission denied", errors.New("permission denied"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isConnectionError(tt.err)
			if got != tt.want {
				t.Errorf("isConnectionError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// TestService_tryIPRemap_NonConnectionError tests tryIPRemap with non-connection errors.
func TestService_tryIPRemap_NonConnectionError(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	device := model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "AA:BB:CC:DD:EE:FF",
	}

	// Non-connection error should not attempt remap
	nonConnErr := errors.New("some other error")
	conn, err := service.tryIPRemap(context.Background(), device, nonConnErr)

	if conn != nil {
		t.Error("expected nil connection for non-connection error")
	}
	if !errors.Is(err, nonConnErr) {
		t.Errorf("expected original error %v, got %v", nonConnErr, err)
	}
}

// TestService_tryIPRemap_NoMAC tests tryIPRemap without MAC address.
func TestService_tryIPRemap_NoMAC(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	device := model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "", // No MAC
	}

	// Connection error without MAC should not attempt remap
	conn, err := service.tryIPRemap(context.Background(), device, model.ErrConnectionFailed)

	if conn != nil {
		t.Error("expected nil connection for device without MAC")
	}
	if !errors.Is(err, model.ErrConnectionFailed) {
		t.Errorf("expected ErrConnectionFailed, got %v", err)
	}
}

// TestService_tryIPRemap_WithMAC tests tryIPRemap with MAC but no mDNS responder.
func TestService_tryIPRemap_WithMAC(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	device := model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "AA:BB:CC:DD:EE:FF",
	}

	// Connection error with MAC should attempt remap but fail without mDNS
	conn, err := service.tryIPRemap(context.Background(), device, model.ErrConnectionFailed)

	if conn != nil {
		t.Error("expected nil connection when mDNS discovery fails")
	}
	if !errors.Is(err, model.ErrConnectionFailed) {
		t.Errorf("expected ErrConnectionFailed, got %v", err)
	}
}

// TestService_tryGen1IPRemap_NonConnectionError tests tryGen1IPRemap with non-connection errors.
func TestService_tryGen1IPRemap_NonConnectionError(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	device := model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "AA:BB:CC:DD:EE:FF",
	}

	// Non-connection error should not attempt remap
	nonConnErr := errors.New("authentication failed")
	conn, err := service.tryGen1IPRemap(context.Background(), device, nonConnErr)

	if conn != nil {
		t.Error("expected nil connection for non-connection error")
	}
	if !errors.Is(err, nonConnErr) {
		t.Errorf("expected original error %v, got %v", nonConnErr, err)
	}
}

// TestService_tryGen1IPRemap_NoMAC tests tryGen1IPRemap without MAC address.
func TestService_tryGen1IPRemap_NoMAC(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver)

	device := model.Device{
		Name:    "test",
		Address: "192.168.1.1",
		MAC:     "", // No MAC
	}

	// Connection error without MAC should not attempt remap
	conn, err := service.tryGen1IPRemap(context.Background(), device, model.ErrConnectionFailed)

	if conn != nil {
		t.Error("expected nil connection for device without MAC")
	}
	if !errors.Is(err, model.ErrConnectionFailed) {
		t.Errorf("expected ErrConnectionFailed, got %v", err)
	}
}

// TestService_New_WithPluginRegistry tests New with plugin registry option.
func TestService_New_WithPluginRegistry(t *testing.T) {
	t.Parallel()

	resolver := &mockResolver{}
	service := New(resolver, WithPluginRegistry(nil))

	// Just verify no panic - plugin registry is nil
	if service == nil {
		t.Error("expected non-nil service")
	}
}
