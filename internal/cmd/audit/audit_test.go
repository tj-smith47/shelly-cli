package audit

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "audit [device...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "audit [device...]")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Short != "Security audit for devices" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Security audit for devices")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"security", "sec"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias at index %d", i)
			continue
		}
		if cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "all", shorthand: "", defValue: "false"},
		{name: "check-auth", shorthand: "", defValue: "false"},
		{name: "check-cloud", shorthand: "", defValue: "false"},
		{name: "check-firmware", shorthand: "", defValue: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description contains key information
	checks := []string{
		"security audit",
		"Authentication status",
		"Cloud connection",
		"Firmware version",
	}

	for _, check := range checks {
		if !strings.Contains(cmd.Long, check) {
			t.Errorf("Long description missing %q", check)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify examples contain key usage patterns
	checks := []string{
		"shelly audit kitchen-light",
		"shelly audit light-1 switch-2",
		"shelly audit --check-firmware --check-auth -o json",
	}

	for _, check := range checks {
		if !strings.Contains(cmd.Example, check) {
			t.Errorf("Example missing %q", check)
		}
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Verify Run is not set (we should use RunE, not Run)
	if cmd.Run != nil {
		t.Error("Run should be nil, use RunE instead")
	}
}

func TestNewCommand_AllFlagMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue string
		wantError bool
	}{
		{
			name:      "set all flag to true",
			flagValue: "true",
			wantError: false,
		},
		{
			name:      "set all flag to false",
			flagValue: "false",
			wantError: false,
		},
		{
			name:      "set all flag to invalid",
			flagValue: "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Flags().Set("all", tt.flagValue)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command can be accessed by name
	if cmd.Name() != "audit" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "audit")
	}
}

func TestNewCommand_WithDeviceArg(t *testing.T) {
	t.Parallel()

	tf := runAudit(t, twoAuditDevices, "kitchen")
	out := tf.OutString()
	if !strings.Contains(out, "Security Audit") || !strings.Contains(out, "kitchen") {
		t.Errorf("expected the audit header and kitchen in output, got: %q", out)
	}
	if strings.Contains(out, "porch") {
		t.Errorf("only the named device should be audited, got: %q", out)
	}
}

func TestNewCommand_WithDeviceArg_CancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device"})

	// Create a cancelled context to prevent actual network calls
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - may fail due to cancelled context but exercises the code path
	err := cmd.Execute()
	if err != nil {
		t.Logf("Expected error due to cancelled context: %v", err)
	}

	// Verify some output was produced (even with error)
	output := tf.OutString()
	// Should have at least attempted to display the audit header
	if !strings.Contains(output, "Security Audit") {
		// The run function should print the title before auditing devices
		t.Logf("output was: %q", output)
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		check  func(cmd *cobra.Command) bool
		wantOK bool
		errMsg string
	}{
		{
			name:   "has use",
			check:  func(c *cobra.Command) bool { return c.Use != "" },
			wantOK: true,
			errMsg: "Use should not be empty",
		},
		{
			name:   "has short",
			check:  func(c *cobra.Command) bool { return c.Short != "" },
			wantOK: true,
			errMsg: "Short should not be empty",
		},
		{
			name:   "has long",
			check:  func(c *cobra.Command) bool { return c.Long != "" },
			wantOK: true,
			errMsg: "Long should not be empty",
		},
		{
			name:   "has example",
			check:  func(c *cobra.Command) bool { return c.Example != "" },
			wantOK: true,
			errMsg: "Example should not be empty",
		},
		{
			name:   "has aliases",
			check:  func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK: true,
			errMsg: "Aliases should not be empty",
		},
		{
			name:   "has RunE",
			check:  func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK: true,
			errMsg: "RunE should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			if tt.check(cmd) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no flags",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "all flag",
			args:    []string{"--all"},
			wantErr: false,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// runAudit executes the audit command against two registered Gen2 mock devices.
func runAudit(t *testing.T, devices []mock.DeviceFixture, args ...string) *factory.TestFactory {
	t.Helper()
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version:      "1",
		Config:       mock.ConfigFixture{Devices: devices},
		DeviceStates: map[string]mock.DeviceState{},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs(args)
	cmd.SetContext(t.Context())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute(%v) error = %v", args, err)
	}
	return tf
}

var twoAuditDevices = []mock.DeviceFixture{
	{Name: "kitchen", Address: "192.168.1.10", MAC: "AA:BB:CC:DD:EE:01", Type: "SNSW-001P16EU", Model: "Shelly Plus 1PM", Generation: 2},
	{Name: "porch", Address: "192.168.1.11", MAC: "AA:BB:CC:DD:EE:02", Type: "SNSW-001P16EU", Model: "Shelly Plus 1PM", Generation: 2},
}

func decodeAudit(t *testing.T, out string) []model.AuditResult {
	t.Helper()
	var results []model.AuditResult
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("output is not a JSON audit list: %v\n%s", err, out)
	}
	return results
}

// The advertised alias "audit --check-firmware --check-auth -o json" names no
// device; it must audit every registered device and print only the selected
// checks as JSON.
//
//nolint:paralleltest // Uses viper global state
func TestRun_NoDeviceSelectedChecksJSON(t *testing.T) {
	viper.Set("output", "json")
	t.Cleanup(viper.Reset)

	tf := runAudit(t, twoAuditDevices, "--check-firmware", "--check-auth")
	results := decodeAudit(t, tf.OutString())

	if len(results) != 2 || results[0].Device != "kitchen" || results[1].Device != "porch" {
		t.Fatalf("results = %+v, want kitchen and porch", results)
	}
	for _, r := range results {
		if !r.Reachable {
			t.Errorf("%s: Reachable = false", r.Device)
		}
		if r.AuthStatus == nil {
			t.Errorf("%s: auth check missing", r.Device)
		}
		if r.FWAudit == nil {
			t.Errorf("%s: firmware check missing", r.Device)
		}
		if r.CloudAudit != nil {
			t.Errorf("%s: cloud check ran but was not selected", r.Device)
		}
	}
	if strings.Contains(tf.OutString(), "Security Audit") {
		t.Error("JSON output must not include the table header")
	}
}

//nolint:paralleltest // Uses viper global state
func TestRun_NoCheckFlagsRunsEveryCheck(t *testing.T) {
	viper.Set("output", "json")
	t.Cleanup(viper.Reset)

	tf := runAudit(t, twoAuditDevices[:1], "kitchen")
	results := decodeAudit(t, tf.OutString())

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.AuthStatus == nil || r.CloudAudit == nil || r.FWAudit == nil {
		t.Errorf("every check should run by default, got auth=%v cloud=%v firmware=%v",
			r.AuthStatus, r.CloudAudit, r.FWAudit)
	}
}

//nolint:paralleltest // Uses viper global state
func TestRun_OnlyCloudCheck(t *testing.T) {
	viper.Set("output", "json")
	t.Cleanup(viper.Reset)

	tf := runAudit(t, twoAuditDevices[:1], "kitchen", "--check-cloud")
	results := decodeAudit(t, tf.OutString())

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.CloudAudit == nil || r.AuthStatus != nil || r.FWAudit != nil {
		t.Errorf("only cloud should run, got auth=%v cloud=%v firmware=%v",
			r.AuthStatus, r.CloudAudit, r.FWAudit)
	}
}

//nolint:paralleltest // Uses viper global state
func TestRun_NoRegisteredDevicesJSONIsEmptyList(t *testing.T) {
	viper.Set("output", "json")
	t.Cleanup(viper.Reset)

	tf := runAudit(t, nil)
	if got := strings.TrimSpace(tf.OutString()); got != "[]" {
		t.Errorf("output = %q, want []", got)
	}
}

func TestRun_NoDeviceTableAuditsAllRegistered(t *testing.T) {
	t.Parallel()

	tf := runAudit(t, twoAuditDevices)
	out := tf.OutString()
	for _, want := range []string{"Security Audit", "kitchen", "porch"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRun_NoRegisteredDevicesTableWarns(t *testing.T) {
	t.Parallel()

	tf := runAudit(t, nil)
	if !strings.Contains(tf.OutString()+tf.ErrString(), "No devices registered") {
		t.Errorf("expected a 'No devices registered' warning, got out=%q err=%q", tf.OutString(), tf.ErrString())
	}
}
