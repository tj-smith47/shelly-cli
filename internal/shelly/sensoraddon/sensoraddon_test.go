// Package sensoraddon provides Sensor Add-on management for Shelly devices.
package sensoraddon

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

const (
	testSensorAddr = "28:FF:64:1E:2A:16:04:97"
)

// mockConnectionProvider is a test double for ConnectionProvider.
type mockConnectionProvider struct {
	withConnectionFn func(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

func (m *mockConnectionProvider) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := New(provider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != provider {
		t.Error("expected provider to be set")
	}
}

func TestPeripheralTypeConstants(t *testing.T) {
	t.Parallel()

	// Verify type constants are defined
	tests := []struct {
		name     string
		constant PeripheralType
	}{
		{"DS18B20", TypeDS18B20},
		{"DHT22", TypeDHT22},
		{"DigitalIn", TypeDigitalIn},
		{"AnalogIn", TypeAnalogIn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.constant == "" {
				t.Errorf("expected non-empty constant for %s", tt.name)
			}
		})
	}
}

func TestValidPeripheralTypes(t *testing.T) {
	t.Parallel()

	if len(ValidPeripheralTypes) == 0 {
		t.Error("expected non-empty ValidPeripheralTypes")
	}

	// Check all expected types are in the list
	expected := []string{
		string(TypeDS18B20),
		string(TypeDHT22),
		string(TypeDigitalIn),
		string(TypeAnalogIn),
	}

	for _, exp := range expected {
		found := false
		for _, vpt := range ValidPeripheralTypes {
			if vpt == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q in ValidPeripheralTypes", exp)
		}
	}
}

func TestPeripheral_Fields(t *testing.T) {
	t.Parallel()

	addr := testSensorAddr
	peripheral := Peripheral{
		Type:      TypeDS18B20,
		Component: "temperature:100",
		Addr:      &addr,
	}

	if peripheral.Type != TypeDS18B20 {
		t.Errorf("got Type=%q, want %q", peripheral.Type, TypeDS18B20)
	}
	if peripheral.Component != "temperature:100" {
		t.Errorf("got Component=%q, want %q", peripheral.Component, "temperature:100")
	}
	if peripheral.Addr == nil {
		t.Fatal("expected Addr to be non-nil")
	}
	if *peripheral.Addr != addr {
		t.Errorf("got Addr=%q, want %q", *peripheral.Addr, addr)
	}
}

func TestPeripheral_NilAddr(t *testing.T) {
	t.Parallel()

	peripheral := Peripheral{
		Type:      TypeDigitalIn,
		Component: "input:100",
		Addr:      nil,
	}

	if peripheral.Type != TypeDigitalIn {
		t.Errorf("got Type=%q, want %q", peripheral.Type, TypeDigitalIn)
	}
	if peripheral.Addr != nil {
		t.Error("expected Addr to be nil")
	}
}

func TestOneWireDevice_Fields(t *testing.T) {
	t.Parallel()

	component := "temperature:100"
	device := OneWireDevice{
		Type:      "DS18B20",
		Addr:      "28:FF:64:1E:2A:16:04:97",
		Component: &component,
	}

	if device.Type != "DS18B20" {
		t.Errorf("got Type=%q, want %q", device.Type, "DS18B20")
	}
	if device.Addr != "28:FF:64:1E:2A:16:04:97" {
		t.Errorf("got Addr=%q, want %q", device.Addr, "28:FF:64:1E:2A:16:04:97")
	}
	if device.Component == nil {
		t.Fatal("expected Component to be non-nil")
	}
	if *device.Component != component {
		t.Errorf("got Component=%q, want %q", *device.Component, component)
	}
}

func TestOneWireDevice_NilComponent(t *testing.T) {
	t.Parallel()

	device := OneWireDevice{
		Type:      "DS18B20",
		Addr:      "28:FF:64:1E:2A:16:04:97",
		Component: nil,
	}

	if device.Type != "DS18B20" {
		t.Errorf("got Type=%q, want %q", device.Type, "DS18B20")
	}
	if device.Component != nil {
		t.Error("expected Component to be nil")
	}
}

func TestAddOptions_Fields(t *testing.T) {
	t.Parallel()

	cid := 100
	addr := "28:FF:64:1E:2A:16:04:97"

	opts := AddOptions{
		CID:  &cid,
		Addr: &addr,
	}

	if opts.CID == nil {
		t.Fatal("expected CID to be non-nil")
	}
	if *opts.CID != cid {
		t.Errorf("got CID=%d, want %d", *opts.CID, cid)
	}
	if opts.Addr == nil {
		t.Fatal("expected Addr to be non-nil")
	}
	if *opts.Addr != addr {
		t.Errorf("got Addr=%q, want %q", *opts.Addr, addr)
	}
}

func TestAddOptions_NilFields(t *testing.T) {
	t.Parallel()

	opts := AddOptions{
		CID:  nil,
		Addr: nil,
	}

	if opts.CID != nil {
		t.Error("expected CID to be nil")
	}
	if opts.Addr != nil {
		t.Error("expected Addr to be nil")
	}
}

// ----- Service Connection Error Tests -----

func TestListPeripherals_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.ListPeripherals(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestAddPeripheral_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.AddPeripheral(context.Background(), "test-device", TypeDS18B20, nil)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestRemovePeripheral_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	err := svc.RemovePeripheral(context.Background(), "test-device", "temperature:100")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestScanOneWire_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.ScanOneWire(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

// ----- Identifier Passthrough Tests -----

func TestListPeripherals_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(provider)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_, _ = svc.ListPeripherals(context.Background(), "my-sensor-device")

	if capturedIdentifier != "my-sensor-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "my-sensor-device")
	}
}

func TestAddPeripheral_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(provider)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_, _ = svc.AddPeripheral(context.Background(), "add-device", TypeDHT22, nil)

	if capturedIdentifier != "add-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "add-device")
	}
}

func TestRemovePeripheral_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(provider)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_ = svc.RemovePeripheral(context.Background(), "remove-device", "temp:100")

	if capturedIdentifier != "remove-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "remove-device")
	}
}

func TestScanOneWire_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(provider)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_, _ = svc.ScanOneWire(context.Background(), "scan-device")

	if capturedIdentifier != "scan-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "scan-device")
	}
}

// ----- Nil Provider Tests -----

func TestNewWithNilProvider(t *testing.T) {
	t.Parallel()

	svc := New(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != nil {
		t.Error("expected provider to be nil")
	}
}

// ----- Peripheral Type Comparison Tests -----

func TestPeripheralTypes_AreDistinct(t *testing.T) {
	t.Parallel()

	types := []PeripheralType{
		TypeDS18B20,
		TypeDHT22,
		TypeDigitalIn,
		TypeAnalogIn,
	}

	// Check all types are distinct
	seen := make(map[PeripheralType]bool)
	for _, pt := range types {
		if seen[pt] {
			t.Errorf("duplicate peripheral type: %q", pt)
		}
		seen[pt] = true
	}
}

func TestValidPeripheralTypes_MatchesConstants(t *testing.T) {
	t.Parallel()

	// Verify count matches
	expectedCount := 4
	if len(ValidPeripheralTypes) != expectedCount {
		t.Errorf("got %d valid types, want %d", len(ValidPeripheralTypes), expectedCount)
	}

	// Verify all are non-empty
	for _, vpt := range ValidPeripheralTypes {
		if vpt == "" {
			t.Error("found empty string in ValidPeripheralTypes")
		}
	}
}

// ----- JSON Tag Tests -----

func TestPeripheral_ZeroValue(t *testing.T) {
	t.Parallel()

	var p Peripheral

	if p.Type != "" {
		t.Errorf("got Type=%q, want empty", p.Type)
	}
	if p.Component != "" {
		t.Errorf("got Component=%q, want empty", p.Component)
	}
	if p.Addr != nil {
		t.Error("expected Addr to be nil")
	}
}

func TestOneWireDevice_ZeroValue(t *testing.T) {
	t.Parallel()

	var d OneWireDevice

	if d.Type != "" {
		t.Errorf("got Type=%q, want empty", d.Type)
	}
	if d.Addr != "" {
		t.Errorf("got Addr=%q, want empty", d.Addr)
	}
	if d.Component != nil {
		t.Error("expected Component to be nil")
	}
}

func TestAddOptions_ZeroValue(t *testing.T) {
	t.Parallel()

	var opts AddOptions

	if opts.CID != nil {
		t.Error("expected CID to be nil")
	}
	if opts.Addr != nil {
		t.Error("expected Addr to be nil")
	}
}

// ----- Interface Compliance Test -----

func TestConnectionProvider_Interface(t *testing.T) {
	t.Parallel()

	// This test verifies that mockConnectionProvider satisfies the ConnectionProvider interface
	var _ ConnectionProvider = (*mockConnectionProvider)(nil)
}
