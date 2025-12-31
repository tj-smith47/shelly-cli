// Package mdns provides mDNS discovery command.
package mdns

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
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// mockDiscoverer is a mock implementation of the Discoverer interface.
type mockDiscoverer struct {
	devices     []discovery.DiscoveredDevice
	discoverErr error
	stopErr     error
	stopCalled  bool
}

func (m *mockDiscoverer) Discover(_ time.Duration) ([]discovery.DiscoveredDevice, error) {
	return m.devices, m.discoverErr
}

func (m *mockDiscoverer) Stop() error {
	m.stopCalled = true
	return m.stopErr
}

// setMockDiscoverer sets the discoverer factory to return a mock.
func setMockDiscoverer(m *mockDiscoverer) func() {
	original := discovererFactory
	discovererFactory = func() Discoverer { return m }
	return func() { discovererFactory = original }
}

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "mdns" {
		t.Errorf("Use = %q, want %q", cmd.Use, "mdns")
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
	wantAliases := []string{"zeroconf", "bonjour"}
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
	case timeout.DefValue != "10s":
		t.Errorf("timeout default = %q, want %q", timeout.DefValue, "10s")
	}

	// Test register flag exists
	register := cmd.Flags().Lookup("register")
	if register == nil {
		t.Error("register flag not found")
	} else if register.DefValue != "false" {
		t.Errorf("register default = %q, want %q", register.DefValue, "false")
	}

	// Test skip-existing flag exists
	skipExisting := cmd.Flags().Lookup("skip-existing")
	if skipExisting == nil {
		t.Error("skip-existing flag not found")
	} else if skipExisting.DefValue != "true" {
		t.Errorf("skip-existing default = %q, want %q", skipExisting.DefValue, "true")
	}
}

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()
	expected := 10 * time.Second
	if DefaultTimeout != expected {
		t.Errorf("DefaultTimeout = %v, want %v", DefaultTimeout, expected)
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
		"shelly discover mdns",
		"--timeout",
		"--register",
		"--skip-existing",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"mDNS",
		"Zeroconf",
		"_shelly._tcp.local",
		"Gen2+",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_NoDevices(t *testing.T) {
	mock := &mockDiscoverer{devices: nil}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "No devices found") {
		t.Errorf("output should contain 'No devices found', got: %q", output)
	}

	if !mock.stopCalled {
		t.Error("Stop() should have been called")
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_WithDevices(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "Kitchen Light",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "192.168.1.100") {
		t.Errorf("output should contain device address, got: %q", output)
	}

	if !mock.stopCalled {
		t.Error("Stop() should have been called")
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_DiscoveryError(t *testing.T) {
	expectedErr := errors.New("network error")
	mock := &mockDiscoverer{discoverErr: expectedErr}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should return error")
	}

	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("error should contain 'network error', got: %v", err)
	}

	if !mock.stopCalled {
		t.Error("Stop() should have been called even on error")
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_StopError(t *testing.T) {
	mock := &mockDiscoverer{
		devices: nil,
		stopErr: errors.New("stop failed"),
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms"})

	// Stop error should not cause command to fail
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil (stop error should be logged)", err)
	}

	if !mock.stopCalled {
		t.Error("Stop() should have been called")
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_WithRegister(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "Kitchen Light",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms", "--register"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Check that output mentions registration
	output := tf.OutString()
	if !strings.Contains(output, "device") {
		t.Errorf("output should mention device registration, got: %q", output)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_WithRegisterSkipExisting(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "Kitchen Light",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms", "--register", "--skip-existing"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

//nolint:paralleltest // modifies global discovererFactory and config.SetFs
func TestExecute_WithRegisterNoSkipExisting(t *testing.T) {
	factory.SetupTestFs(t)
	config.ResetDefaultManagerForTesting()

	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "Kitchen Light",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms", "--register", "--skip-existing=false"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_MultipleDevices(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "Kitchen Light",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:01",
				Generation: 2,
			},
			{
				ID:         "shellyplus2pm-def456",
				Name:       "Living Room",
				Model:      "SNSW-002P16EU",
				Address:    net.ParseIP("192.168.1.101"),
				MACAddress: "AA:BB:CC:DD:EE:02",
				Generation: 2,
			},
			{
				ID:         "shelly1-ghi789",
				Name:       "Bedroom",
				Model:      "SHSW-1",
				Address:    net.ParseIP("192.168.1.102"),
				MACAddress: "AA:BB:CC:DD:EE:03",
				Generation: 1,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	// Check that multiple devices are shown
	if !strings.Contains(output, "192.168.1.100") {
		t.Errorf("output should contain first device address, got: %q", output)
	}
	if !strings.Contains(output, "192.168.1.101") {
		t.Errorf("output should contain second device address, got: %q", output)
	}
	if !strings.Contains(output, "192.168.1.102") {
		t.Errorf("output should contain third device address, got: %q", output)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_CustomTimeout(t *testing.T) {
	mock := &mockDiscoverer{devices: nil}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "30s"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_ZeroTimeout(t *testing.T) {
	mock := &mockDiscoverer{devices: nil}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "0s"})

	// Zero timeout should default to DefaultTimeout
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_ContextCancelled(t *testing.T) {
	mock := &mockDiscoverer{
		discoverErr: context.Canceled,
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--timeout", "1ms"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should return error for cancelled context")
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_DeviceWithNoName(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "", // No name
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "192.168.1.100") {
		t.Errorf("output should contain device address, got: %q", output)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_DeviceWithSecureFlag(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "Secure Device",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_Gen1Device(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shelly1-abc123",
				Name:       "Gen1 Switch",
				Model:      "SHSW-1",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 1,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

func TestExecute_InvalidTimeoutFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error with invalid timeout")
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestRun_DirectCall(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "Test Device",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)

	err := run(context.Background(), tf.Factory, time.Millisecond, false, true)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestRun_WithRegistration(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-abc123",
				Name:       "Test Device",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.100"),
				MACAddress: "AA:BB:CC:DD:EE:FF",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)

	err := run(context.Background(), tf.Factory, time.Millisecond, true, true)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}
}

//nolint:paralleltest // modifies global discovererFactory
func TestRun_ZeroTimeoutDefault(t *testing.T) {
	mock := &mockDiscoverer{devices: nil}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)

	// Pass zero timeout - should use DefaultTimeout
	err := run(context.Background(), tf.Factory, 0, false, true)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}
}

func TestDiscovererInterface(t *testing.T) {
	t.Parallel()

	// Ensure mockDiscoverer implements Discoverer
	var _ Discoverer = (*mockDiscoverer)(nil)
}

//nolint:paralleltest // modifies global discovererFactory
func TestExecute_RegisterMultipleDevices(t *testing.T) {
	mock := &mockDiscoverer{
		devices: []discovery.DiscoveredDevice{
			{
				ID:         "shellyplus1pm-001",
				Name:       "Device 1",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.101"),
				MACAddress: "AA:BB:CC:DD:EE:01",
				Generation: 2,
			},
			{
				ID:         "shellyplus1pm-002",
				Name:       "Device 2",
				Model:      "SNSW-001P16EU",
				Address:    net.ParseIP("192.168.1.102"),
				MACAddress: "AA:BB:CC:DD:EE:02",
				Generation: 2,
			},
		},
	}
	cleanup := setMockDiscoverer(mock)
	defer cleanup()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "1ms", "--register"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

func TestAliasExecution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		alias string
	}{
		{"zeroconf alias", "zeroconf"},
		{"bonjour alias", "bonjour"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())

			found := false
			for _, alias := range cmd.Aliases {
				if alias == tt.alias {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("alias %q not found in command aliases", tt.alias)
			}
		})
	}
}
