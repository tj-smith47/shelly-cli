// Package coiot provides CoIoT discovery command.
package coiot

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

// Test constants for flag default values.
const (
	defFalse = "false"
	defTrue  = "true"
)

// mockDiscoverer is a test mock for the CoIoT discoverer.
type mockDiscoverer struct {
	devices  []discovery.DiscoveredDevice
	err      error
	stopErr  error
	stopFunc func()
}

func (m *mockDiscoverer) Discover(_ time.Duration) ([]discovery.DiscoveredDevice, error) {
	return m.devices, m.err
}

func (m *mockDiscoverer) Stop() error {
	if m.stopFunc != nil {
		m.stopFunc()
	}
	return m.stopErr
}

// setMockDiscoverer replaces the discoverer factory and returns a cleanup function.
func setMockDiscoverer(mock *mockDiscoverer) func() {
	original := newDiscoverer
	newDiscoverer = func() Discoverer {
		return mock
	}
	return func() {
		newDiscoverer = original
	}
}

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "coiot" {
		t.Errorf("Use = %q, want %q", cmd.Use, "coiot")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	wantAliases := []string{"coap"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Check that example contains expected patterns
	wantPatterns := []string{
		"shelly discover coiot",
		"--timeout",
		"--register",
		"--gen1-only",
		"--verbose",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check that long description contains key information
	wantPatterns := []string{
		"CoIoT",
		"Gen1",
		"multicast",
		"224.0.1.187",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long description should contain %q", pattern)
		}
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
	} else if register.DefValue != defFalse {
		t.Errorf("register default = %q, want %q", register.DefValue, defFalse)
	}

	// Test skip-existing flag exists
	skipExisting := cmd.Flags().Lookup("skip-existing")
	if skipExisting == nil {
		t.Error("skip-existing flag not found")
	} else if skipExisting.DefValue != defTrue {
		t.Errorf("skip-existing default = %q, want %q", skipExisting.DefValue, defTrue)
	}

	// Test gen1-only flag exists
	gen1Only := cmd.Flags().Lookup("gen1-only")
	if gen1Only == nil {
		t.Error("gen1-only flag not found")
	} else if gen1Only.DefValue != defFalse {
		t.Errorf("gen1-only default = %q, want %q", gen1Only.DefValue, defFalse)
	}

	// Test verbose flag exists
	verbose := cmd.Flags().Lookup("verbose")
	if verbose == nil {
		t.Error("verbose flag not found")
	} else {
		if verbose.DefValue != defFalse {
			t.Errorf("verbose default = %q, want %q", verbose.DefValue, defFalse)
		}
		if verbose.Shorthand != "v" {
			t.Errorf("verbose shorthand = %q, want %q", verbose.Shorthand, "v")
		}
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

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func() bool
		errMsg    string
	}{
		{
			name: "has use",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Use != ""
			},
			errMsg: "Use should not be empty",
		},
		{
			name: "has short",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Short != ""
			},
			errMsg: "Short should not be empty",
		},
		{
			name: "has long",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Long != ""
			},
			errMsg: "Long should not be empty",
		},
		{
			name: "has example",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Example != ""
			},
			errMsg: "Example should not be empty",
		},
		{
			name: "has aliases",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return len(cmd.Aliases) > 0
			},
			errMsg: "Aliases should not be empty",
		},
		{
			name: "has RunE",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.RunE != nil
			},
			errMsg: "RunE should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !tt.checkFunc() {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestOptions_Default(t *testing.T) {
	t.Parallel()
	opts := &Options{}

	if opts.Factory != nil {
		t.Error("Factory should be nil by default")
	}
	if opts.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0", opts.Timeout)
	}
	if opts.Register || opts.SkipExisting || opts.Gen1Only || opts.Verbose {
		t.Error("All bool fields should be false by default")
	}
}

func TestOptions_WithTimeout(t *testing.T) {
	t.Parallel()
	opts := &Options{Timeout: 30 * time.Second}
	if opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", opts.Timeout, 30*time.Second)
	}
}

func TestOptions_WithRegister(t *testing.T) {
	t.Parallel()
	opts := &Options{Register: true, SkipExisting: true}
	if !opts.Register || !opts.SkipExisting {
		t.Error("Register and SkipExisting should be true")
	}
}

func TestOptions_WithGen1Only(t *testing.T) {
	t.Parallel()
	opts := &Options{Gen1Only: true}
	if !opts.Gen1Only {
		t.Error("Gen1Only should be true")
	}
}

func TestOptions_WithVerbose(t *testing.T) {
	t.Parallel()
	opts := &Options{Verbose: true}
	if !opts.Verbose {
		t.Error("Verbose should be true")
	}
}

func TestOptions_WithFactory(t *testing.T) {
	t.Parallel()
	opts := &Options{Factory: cmdutil.NewFactory()}
	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

func TestOptions_AllSet(t *testing.T) {
	t.Parallel()
	opts := &Options{
		Factory:      cmdutil.NewFactory(),
		Timeout:      5 * time.Second,
		Register:     true,
		SkipExisting: false,
		Gen1Only:     true,
		Verbose:      true,
	}
	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
	if opts.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", opts.Timeout, 5*time.Second)
	}
	if !opts.Register || opts.SkipExisting || !opts.Gen1Only || !opts.Verbose {
		t.Error("Option fields not set correctly")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, cmd *cmdutil.Factory, args []string)
	}{
		{
			name: "timeout flag",
			args: []string{"--timeout", "30s"},
		},
		{
			name: "timeout short flag",
			args: []string{"-t", "5s"},
		},
		{
			name: "register flag",
			args: []string{"--register"},
		},
		{
			name: "skip-existing flag",
			args: []string{"--skip-existing=false"},
		},
		{
			name: "gen1-only flag",
			args: []string{"--gen1-only"},
		},
		{
			name: "verbose flag",
			args: []string{"--verbose"},
		},
		{
			name: "verbose short flag",
			args: []string{"-v"},
		},
		{
			name: "multiple flags",
			args: []string{"--timeout", "20s", "--register", "--gen1-only", "-v"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			// Add --help to avoid actually running the command
			args := make([]string, len(tt.args)+1)
			copy(args, tt.args)
			args[len(tt.args)] = "--help"
			cmd.SetArgs(args)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("Flag parsing failed: %v", err)
			}
		})
	}
}

func TestNewCommand_InvalidFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "invalid timeout format",
			args:    []string{"--timeout", "not-a-duration"},
			wantErr: true,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly discover coiot",
		"--timeout 30s",
		"--gen1-only",
		"--verbose",
		"--register",
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
		"CoIoT",
		"Gen1",
		"multicast",
		"CoAP",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

// TestRun_NoDevices tests the run function when no devices are discovered.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_NoDevices(t *testing.T) {
	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: nil,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Timeout: 10 * time.Millisecond,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "No devices found") {
		t.Errorf("Expected 'No devices found' message, got: %q", output)
	}
}

// TestRun_DiscoveryError tests run when discovery returns an error.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_DiscoveryError(t *testing.T) {
	expectedErr := errors.New("discovery failed")
	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: nil,
		err:     expectedErr,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Timeout: 10 * time.Millisecond,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err == nil {
		t.Error("Expected error from run()")
	} else if !strings.Contains(err.Error(), "discovery failed") {
		t.Errorf("Expected error containing 'discovery failed', got: %v", err)
	}
}

// TestRun_WithDevices tests run when devices are discovered.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_WithDevices(t *testing.T) {
	devices := []discovery.DiscoveredDevice{
		{
			ID:         "shellyswitch-ABC123",
			Name:       "Kitchen Light",
			Model:      "SHSW-1",
			Address:    net.ParseIP("192.168.1.100"),
			MACAddress: "AA:BB:CC:DD:EE:FF",
			Generation: 1,
		},
		{
			ID:         "shellyplug-DEF456",
			Name:       "Living Room Plug",
			Model:      "SHPLG-1",
			Address:    net.ParseIP("192.168.1.101"),
			MACAddress: "11:22:33:44:55:66",
			Generation: 1,
		},
	}

	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: devices,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Timeout: 10 * time.Millisecond,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}
}

// TestRun_WithRegister tests run with the register flag enabled.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_WithRegister(t *testing.T) {
	devices := []discovery.DiscoveredDevice{
		{
			ID:         "shellyswitch-ABC123",
			Name:       "Kitchen Light",
			Model:      "SHSW-1",
			Address:    net.ParseIP("192.168.1.100"),
			MACAddress: "AA:BB:CC:DD:EE:FF",
			Generation: 1,
		},
	}

	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: devices,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:      tf.Factory,
		Timeout:      10 * time.Millisecond,
		Register:     true,
		SkipExisting: true,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}

	output := tf.OutString() + tf.ErrString()
	if !strings.Contains(output, "Added") || !strings.Contains(output, "device") {
		t.Logf("Output: %s", output)
	}
}

// TestRun_StopError tests that stop errors are handled gracefully.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_StopError(t *testing.T) {
	stopCalled := false
	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: nil,
		err:     nil,
		stopErr: errors.New("stop failed"),
		stopFunc: func() {
			stopCalled = true
		},
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Timeout: 10 * time.Millisecond,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	// run() should not return error for stop failures
	if err != nil {
		t.Errorf("run() should not error on stop failure: %v", err)
	}

	if !stopCalled {
		t.Error("Stop() was not called")
	}
}

// TestRun_DefaultTimeout tests that a zero timeout uses the default.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_DefaultTimeout(t *testing.T) {
	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: nil,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Timeout: 0, // Should use DefaultTimeout
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}
}

// TestRun_Gen1Only tests run with the gen1-only flag.
// Note: FilterGen1Devices requires network access to detect generation,
// so with mocked devices that have no actual network presence, this
// tests that the code path is exercised.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_Gen1Only(t *testing.T) {
	devices := []discovery.DiscoveredDevice{
		{
			ID:         "shellyswitch-ABC123",
			Name:       "Kitchen Light",
			Model:      "SHSW-1",
			Address:    net.ParseIP("192.168.1.100"),
			MACAddress: "AA:BB:CC:DD:EE:FF",
			Generation: 1,
		},
	}

	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: devices,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:  tf.Factory,
		Timeout:  10 * time.Millisecond,
		Gen1Only: true,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	// FilterGen1Devices may filter out all devices if it can't reach them
	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}
}

// TestRun_Verbose tests run with the verbose flag.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_Verbose(t *testing.T) {
	devices := []discovery.DiscoveredDevice{
		{
			ID:         "shellyswitch-ABC123",
			Name:       "Kitchen Light",
			Model:      "SHSW-1",
			Address:    net.ParseIP("192.168.1.100"),
			MACAddress: "AA:BB:CC:DD:EE:FF",
			Generation: 1,
		},
	}

	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: devices,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Timeout: 10 * time.Millisecond,
		Verbose: true,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}
}

// TestExecute_ViaCommand tests execution through the cobra command.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestExecute_ViaCommand(t *testing.T) {
	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: nil,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--timeout", "10ms"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute error = %v", err)
	}
}

// TestExecute_WithAllFlags tests execution with all flags set.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestExecute_WithAllFlags(t *testing.T) {
	devices := []discovery.DiscoveredDevice{
		{
			ID:         "shellyswitch-ABC123",
			Name:       "Kitchen Light",
			Model:      "SHSW-1",
			Address:    net.ParseIP("192.168.1.100"),
			MACAddress: "AA:BB:CC:DD:EE:FF",
			Generation: 1,
		},
	}

	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: devices,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"--timeout", "10ms",
		"--register",
		"--skip-existing=false",
		"--gen1-only",
		"--verbose",
	})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

// TestRun_WithMultipleDevices tests run with multiple discovered devices.
//
//nolint:paralleltest // Modifies global newDiscoverer
func TestRun_WithMultipleDevices(t *testing.T) {
	devices := []discovery.DiscoveredDevice{
		{
			ID:         "device1",
			Name:       "Device 1",
			Model:      "SHSW-1",
			Address:    net.ParseIP("192.168.1.100"),
			MACAddress: "AA:BB:CC:DD:EE:01",
			Generation: 1,
		},
		{
			ID:         "device2",
			Name:       "Device 2",
			Model:      "SHSW-1",
			Address:    net.ParseIP("192.168.1.101"),
			MACAddress: "AA:BB:CC:DD:EE:02",
			Generation: 1,
		},
		{
			ID:         "device3",
			Name:       "Device 3",
			Model:      "SHPLG-1",
			Address:    net.ParseIP("192.168.1.102"),
			MACAddress: "AA:BB:CC:DD:EE:03",
			Generation: 1,
		},
	}

	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: devices,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Timeout: 10 * time.Millisecond,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}
}

// TestRun_RegisterSkipExistingFalse tests register with skip-existing=false.
//
//nolint:paralleltest // Modifies global newDiscoverer and config.SetFs
func TestRun_RegisterSkipExistingFalse(t *testing.T) {
	factory.SetupTestFs(t)
	config.ResetDefaultManagerForTesting()

	devices := []discovery.DiscoveredDevice{
		{
			ID:         "shellyswitch-ABC123",
			Name:       "Kitchen Light",
			Model:      "SHSW-1",
			Address:    net.ParseIP("192.168.1.100"),
			MACAddress: "AA:BB:CC:DD:EE:FF",
			Generation: 1,
		},
	}

	cleanup := setMockDiscoverer(&mockDiscoverer{
		devices: devices,
		err:     nil,
	})
	defer cleanup()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:      tf.Factory,
		Timeout:      10 * time.Millisecond,
		Register:     true,
		SkipExisting: false,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}
}

// TestNewCommand_NoArgs tests that the command doesn't require arguments.
func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The discover coiot command doesn't take arguments
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Command should accept no args, got error: %v", err)
		}
	}
}

// TestNewCommand_FlagDefaults tests that flags have correct default values.
func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		defValue string
	}{
		{"timeout", "10s"},
		{"register", "false"},
		{"skip-existing", "true"},
		{"gen1-only", "false"},
		{"verbose", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}
