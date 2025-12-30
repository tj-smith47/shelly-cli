// Package ble provides BLE discovery command.
package ble

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly/wireless"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "ble" {
		t.Errorf("Use = %q, want %q", cmd.Use, "ble")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Aliases
	wantAliases := []string{"bluetooth", "bt"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test timeout flag exists
	timeout := cmd.Flags().Lookup("timeout")
	switch {
	case timeout == nil:
		t.Error("timeout flag not found")
	case timeout.Shorthand != "t":
		t.Errorf("timeout shorthand = %q, want %q", timeout.Shorthand, "t")
	case timeout.DefValue != "15s":
		t.Errorf("timeout default = %q, want %q", timeout.DefValue, "15s")
	}

	// Test bthome flag exists
	bthome := cmd.Flags().Lookup("bthome")
	if bthome == nil {
		t.Error("bthome flag not found")
	} else if bthome.DefValue != "false" {
		t.Errorf("bthome default = %q, want %q", bthome.DefValue, "false")
	}

	// Test filter flag exists
	filter := cmd.Flags().Lookup("filter")
	if filter == nil {
		t.Error("filter flag not found")
	} else if filter.Shorthand != "f" {
		t.Errorf("filter shorthand = %q, want %q", filter.Shorthand, "f")
	}
}

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()
	expected := 15 * time.Second
	if DefaultTimeout != expected {
		t.Errorf("DefaultTimeout = %v, want %v", DefaultTimeout, expected)
	}
}

func TestIsBLENotSupportedError_NilError(t *testing.T) {
	t.Parallel()
	if wireless.IsBLENotSupportedError(nil) {
		t.Error("IsBLENotSupportedError(nil) = true, want false")
	}
}

func TestIsBLENotSupportedError_GenericError(t *testing.T) {
	t.Parallel()
	err := errors.New("some other error")
	if wireless.IsBLENotSupportedError(err) {
		t.Error("IsBLENotSupportedError(generic) = true, want false")
	}
}

func TestIsBLENotSupportedError_NotSupportedError(t *testing.T) {
	t.Parallel()
	if !wireless.IsBLENotSupportedError(discovery.ErrBLENotSupported) {
		t.Error("IsBLENotSupportedError(ErrBLENotSupported) = false, want true")
	}
}

func TestIsBLENotSupportedError_WrappedError(t *testing.T) {
	t.Parallel()
	wrappedErr := &discovery.BLEError{
		Message: "BLE not supported",
		Err:     discovery.ErrBLENotSupported,
	}
	if !wireless.IsBLENotSupportedError(wrappedErr) {
		t.Error("IsBLENotSupportedError(wrapped) = false, want true")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly discover ble",
		"--timeout",
		"--bthome",
		"--filter",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

// mockBLEDiscoverer is a mock implementation of BLEDiscoverer for testing.
type mockBLEDiscoverer struct {
	discoverDevices    []discovery.DiscoveredDevice
	bleDevices         []discovery.BLEDiscoveredDevice
	discoverErr        error
	stopErr            error
	stopCalled         bool
	includeBTHome      bool
	filterPrefix       string
	discoverCalledWith context.Context
}

func (m *mockBLEDiscoverer) DiscoverWithContext(ctx context.Context) ([]discovery.DiscoveredDevice, error) {
	m.discoverCalledWith = ctx
	return m.discoverDevices, m.discoverErr
}

func (m *mockBLEDiscoverer) GetDiscoveredDevices() []discovery.BLEDiscoveredDevice {
	return m.bleDevices
}

func (m *mockBLEDiscoverer) Stop() error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockBLEDiscoverer) SetIncludeBTHome(include bool) {
	m.includeBTHome = include
}

func (m *mockBLEDiscoverer) SetFilterPrefix(prefix string) {
	m.filterPrefix = prefix
}

// Verify mockBLEDiscoverer implements Discoverer.
var _ Discoverer = (*mockBLEDiscoverer)(nil)

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_BLENotSupported(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// Replace the BLE discoverer factory to return BLE not supported error
	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return nil, discovery.ErrBLENotSupported
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error for BLE not supported, got: %v", err)
	}

	// Check stderr for error and hints
	errOutput := tf.ErrString()
	if !strings.Contains(errOutput, "not available") {
		t.Errorf("expected 'not available' message, got: %s", errOutput)
	}

	// Check stdout for hints
	outOutput := tf.OutString()
	combined := errOutput + outOutput
	if !strings.Contains(combined, "Bluetooth adapter") {
		t.Errorf("expected hint about Bluetooth adapter in output, got stderr: %s, stdout: %s", errOutput, outOutput)
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_InitError(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// Replace the BLE discoverer factory to return a non-BLE-not-supported error
	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return nil, errors.New("failed to initialize adapter")
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for init failure")
	}
	if !strings.Contains(err.Error(), "failed to initialize BLE") {
		t.Errorf("expected 'failed to initialize BLE' error, got: %v", err)
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_NoDevicesFound(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{},
		bleDevices:      []discovery.BLEDiscoveredDevice{},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "BLE devices") {
		t.Errorf("expected 'BLE devices' in output, got: %s", output)
	}

	if !mock.stopCalled {
		t.Error("expected Stop() to be called")
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_DiscoverError(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverErr: errors.New("discovery failed"),
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for discovery failure")
	}
	if !strings.Contains(err.Error(), "BLE discovery failed") {
		t.Errorf("expected 'BLE discovery failed' error, got: %v", err)
	}

	if !mock.stopCalled {
		t.Error("expected Stop() to be called even on error")
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_DevicesFound(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{
			{
				ID:      "shelly-blu-1",
				Name:    "Shelly BLU Button",
				Model:   "SBBT-002C",
				Address: net.ParseIP("0.0.0.0"),
			},
		},
		bleDevices: []discovery.BLEDiscoveredDevice{
			{
				DiscoveredDevice: discovery.DiscoveredDevice{
					ID:    "shelly-blu-1",
					Model: "SBBT-002C",
				},
				LocalName:   "Shelly BLU Button",
				RSSI:        -55,
				Connectable: true,
				BTHomeData:  nil,
			},
		},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly BLU Button") {
		t.Errorf("expected device name in output, got: %s", output)
	}
	if !strings.Contains(output, "dBm") {
		t.Errorf("expected RSSI in output, got: %s", output)
	}

	if !mock.stopCalled {
		t.Error("expected Stop() to be called")
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_WithBTHomeFlag(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{},
		bleDevices:      []discovery.BLEDiscoveredDevice{},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--bthome"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	if !mock.includeBTHome {
		t.Error("expected SetIncludeBTHome(true) to be called")
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_WithFilterFlag(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{},
		bleDevices:      []discovery.BLEDiscoveredDevice{},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--filter", "Shelly"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	if mock.filterPrefix != "Shelly" {
		t.Errorf("expected filterPrefix = 'Shelly', got: %q", mock.filterPrefix)
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_WithTimeoutFlag(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{},
		bleDevices:      []discovery.BLEDiscoveredDevice{},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--timeout", "30s"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	// Context with timeout was used for discovery
	if mock.discoverCalledWith == nil {
		t.Error("expected DiscoverWithContext to be called")
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_StopError(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{},
		bleDevices:      []discovery.BLEDiscoveredDevice{},
		stopErr:         errors.New("stop failed"),
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	// Stop error should be logged but not cause command to fail
	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error even with stop error, got: %v", err)
	}

	if !mock.stopCalled {
		t.Error("expected Stop() to be called")
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_MultipleDevicesWithBTHome(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{
			{ID: "device1", Name: "Sensor 1"},
			{ID: "device2", Name: "Sensor 2"},
		},
		bleDevices: []discovery.BLEDiscoveredDevice{
			{
				DiscoveredDevice: discovery.DiscoveredDevice{ID: "device1"},
				LocalName:        "Temperature Sensor",
				RSSI:             -40,
				Connectable:      true,
				BTHomeData:       &discovery.BTHomeData{},
			},
			{
				DiscoveredDevice: discovery.DiscoveredDevice{ID: "device2"},
				LocalName:        "Motion Sensor",
				RSSI:             -65,
				Connectable:      false,
				BTHomeData:       nil,
			},
		},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--bthome"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Temperature Sensor") {
		t.Errorf("expected 'Temperature Sensor' in output, got: %s", output)
	}
	if !strings.Contains(output, "Motion Sensor") {
		t.Errorf("expected 'Motion Sensor' in output, got: %s", output)
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_DefaultTimeoutUsedWhenZero(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{},
		bleDevices:      []discovery.BLEDiscoveredDevice{},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	// Test that the run function handles zero timeout by using default
	err := run(t.Context(), tf.Factory, 0, false, "")
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_EmptyFilterPrefixNotSet(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{},
		bleDevices:      []discovery.BLEDiscoveredDevice{},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	err := run(t.Context(), tf.Factory, DefaultTimeout, false, "")
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	// Empty filter should not be set
	if mock.filterPrefix != "" {
		t.Errorf("expected empty filterPrefix, got: %q", mock.filterPrefix)
	}
}

func TestBLEDiscovererAdapter(t *testing.T) {
	t.Parallel()

	// Verify the adapter implements Discoverer interface.
	var _ Discoverer = (*bleDiscovererAdapter)(nil)
}

func TestNewCommand_AllFlagsAvailable(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flags := cmd.Flags()

	// Verify all flags exist
	requiredFlags := []string{"timeout", "bthome", "filter"}
	for _, name := range requiredFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("flag %q not found", name)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the long description mentions key requirements
	wantPatterns := []string{
		"Bluetooth Low Energy",
		"BLE",
		"provisioning mode",
		"Bluetooth adapter",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_DeviceWithNoLocalName(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{
			{ID: "device-id-only"},
		},
		bleDevices: []discovery.BLEDiscoveredDevice{
			{
				DiscoveredDevice: discovery.DiscoveredDevice{ID: "device-id-only", Model: "Unknown"},
				LocalName:        "", // No local name
				RSSI:             -80,
				Connectable:      false,
			},
		},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	// Should fall back to using ID when LocalName is empty
	output := tf.OutString()
	if !strings.Contains(output, "device-id-only") {
		t.Errorf("expected device ID in output when no local name, got: %s", output)
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_WeakSignalDevice(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{
			{ID: "weak-signal"},
		},
		bleDevices: []discovery.BLEDiscoveredDevice{
			{
				DiscoveredDevice: discovery.DiscoveredDevice{ID: "weak-signal"},
				LocalName:        "Weak Device",
				RSSI:             -90, // Very weak signal
				Connectable:      true,
			},
		},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "-90") {
		t.Errorf("expected RSSI value in output, got: %s", output)
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_StrongSignalDevice(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{
			{ID: "strong-signal"},
		},
		bleDevices: []discovery.BLEDiscoveredDevice{
			{
				DiscoveredDevice: discovery.DiscoveredDevice{ID: "strong-signal"},
				LocalName:        "Strong Device",
				RSSI:             -30, // Very strong signal
				Connectable:      true,
			},
		},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "-30") {
		t.Errorf("expected RSSI value in output, got: %s", output)
	}
}

//nolint:paralleltest // Modifies global newBLEDiscoverer variable
func TestRun_CombinedFlags(t *testing.T) {
	tf := factory.NewTestFactory(t)
	mock := &mockBLEDiscoverer{
		discoverDevices: []discovery.DiscoveredDevice{},
		bleDevices:      []discovery.BLEDiscoveredDevice{},
	}

	oldFactory := newBLEDiscoverer
	newBLEDiscoverer = func() (Discoverer, error) {
		return mock, nil
	}
	t.Cleanup(func() { newBLEDiscoverer = oldFactory })

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--timeout", "20s", "--bthome", "--filter", "BTHome"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	if !mock.includeBTHome {
		t.Error("expected SetIncludeBTHome(true) to be called")
	}
	if mock.filterPrefix != "BTHome" {
		t.Errorf("expected filterPrefix = 'BTHome', got: %q", mock.filterPrefix)
	}
}
