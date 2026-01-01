package remove

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "remove <device> <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "remove <device> <id>")
	}

	// Test Aliases
	wantAliases := []string{"rm", "delete", "del"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"device"}, true},
		{"two args valid", []string{"device", "200"}, false},
		{"three args", []string{"device", "200", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test yes flag
	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("--yes flag not found")
	}
	if flag.Shorthand != "y" {
		t.Errorf("--yes shorthand = %q, want %q", flag.Shorthand, "y")
	}
	if flag.DefValue != "false" {
		t.Errorf("--yes default = %q, want %q", flag.DefValue, "false")
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

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly bthome remove",
		"--yes",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device: "test-device",
		ID:     200,
	}
	opts.Yes = true

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.ID != 200 {
		t.Errorf("ID = %d, want %d", opts.ID, 200)
	}
	if !opts.Yes {
		t.Error("Yes should be true")
	}
}

func TestNewCommand_InvalidID(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device", "not-a-number"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}

	if !strings.Contains(err.Error(), "invalid device ID") {
		t.Errorf("expected 'invalid device ID' error, got: %v", err)
	}
}

func TestRun_WithMock_Success(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{Name: "test-device", Address: "10.0.0.1", Generation: 2},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	// Use --yes to skip confirmation
	cmd.SetArgs([]string{"test-device", "200", "--yes"})

	err = cmd.Execute()
	// The command will likely fail because BTHomeRemoveDevice isn't mocked,
	// but this tests the confirmation bypass path
	// We expect a device resolution or RPC error, not a confirmation error
	if err != nil {
		// This is expected since we don't have a real BTHome device
		// Just verify it got past the confirmation step
		if strings.Contains(err.Error(), "confirm") {
			t.Errorf("Should have bypassed confirmation with --yes: %v", err)
		}
	}
}

func TestRun_Cancelled(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{Name: "test-device", Address: "10.0.0.1", Generation: 2},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	// Don't use --yes so confirmation is required
	// In non-TTY mode (test mode), confirmation returns false by default
	cmd.SetArgs([]string{"test-device", "200"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should show "Operation cancelled" - Info() goes to stdout
	output := tf.OutString()
	if !strings.Contains(output, "cancelled") {
		t.Errorf("Output = %q, want to contain 'cancelled'", output)
	}
}

func TestOptions_Factory(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		Device:  "gateway",
		ID:      100,
	}
	opts.Yes = true

	if opts.Factory != f {
		t.Error("Factory not set correctly")
	}
	if opts.Device != "gateway" {
		t.Errorf("Device = %q, want %q", opts.Device, "gateway")
	}
	if opts.ID != 100 {
		t.Errorf("ID = %d, want %d", opts.ID, 100)
	}
	if !opts.Yes {
		t.Error("Yes should be true")
	}
}
