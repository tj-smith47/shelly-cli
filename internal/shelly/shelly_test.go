// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/config"
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
