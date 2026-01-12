package members

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "members <group>" {
		t.Errorf("Use = %q, want \"members <group>\"", cmd.Use)
	}
	aliases := []string{"show", "ls"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
	}
	for i, alias := range aliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test requires exactly 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"group"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"group", "extra"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}
	if outputFlag.Shorthand != "o" {
		t.Errorf("output shorthand = %q, want o", outputFlag.Shorthand)
	}
	if outputFlag.DefValue != "table" {
		t.Errorf("output default = %q, want table", outputFlag.DefValue)
	}
}

func TestExecute_Help(t *testing.T) {
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

func TestExecute_GroupNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"nonexistent-group"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent group")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestExecute_EmptyGroup(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"empty-group": {
				Devices: []string{},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"empty-group"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should show "no results" message
	output := out.String()
	if !strings.Contains(output, "No") || !strings.Contains(output, "member") {
		t.Errorf("expected 'no members' message, got: %q", output)
	}
}

func TestExecute_GroupWithMembers(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"living-room": {
				Devices: []string{"light-1", "light-2", "switch-1"},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"living-room"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	// Verify all devices are listed
	if !strings.Contains(output, "light-1") {
		t.Errorf("expected output to contain 'light-1', got: %q", output)
	}
	if !strings.Contains(output, "light-2") {
		t.Errorf("expected output to contain 'light-2', got: %q", output)
	}
	if !strings.Contains(output, "switch-1") {
		t.Errorf("expected output to contain 'switch-1', got: %q", output)
	}
	// Verify count message
	if !strings.Contains(output, "3") && !strings.Contains(output, "member") {
		t.Errorf("expected member count in output, got: %q", output)
	}
}

//nolint:paralleltest // Tests that modify viper global state cannot run in parallel
func TestExecute_JSONOutput(t *testing.T) {
	viper.Set("output", "json")
	defer viper.Set("output", "")

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"office": {
				Devices: []string{"desk-lamp", "ceiling-light"},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"office", "-o", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	// Verify JSON structure
	if !strings.Contains(output, `"group"`) {
		t.Errorf("expected JSON output to contain 'group' key, got: %q", output)
	}
	if !strings.Contains(output, `"members"`) {
		t.Errorf("expected JSON output to contain 'members' key, got: %q", output)
	}
	if !strings.Contains(output, `"count"`) {
		t.Errorf("expected JSON output to contain 'count' key, got: %q", output)
	}
	if !strings.Contains(output, "desk-lamp") {
		t.Errorf("expected JSON output to contain 'desk-lamp', got: %q", output)
	}
}

//nolint:paralleltest // Tests that modify viper global state cannot run in parallel
func TestExecute_YAMLOutput(t *testing.T) {
	viper.Set("output", "yaml")
	defer viper.Set("output", "")

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"bedroom": {
				Devices: []string{"nightstand-lamp"},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"bedroom", "-o", "yaml"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	// Verify YAML structure (keys without quotes)
	if !strings.Contains(output, "group:") {
		t.Errorf("expected YAML output to contain 'group:' key, got: %q", output)
	}
	if !strings.Contains(output, "members:") {
		t.Errorf("expected YAML output to contain 'members:' key, got: %q", output)
	}
	if !strings.Contains(output, "count:") {
		t.Errorf("expected YAML output to contain 'count:' key, got: %q", output)
	}
	if !strings.Contains(output, "nightstand-lamp") {
		t.Errorf("expected YAML output to contain 'nightstand-lamp', got: %q", output)
	}
}

func TestExecute_SingleMember(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"solo-group": {
				Devices: []string{"only-device"},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"solo-group"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "only-device") {
		t.Errorf("expected output to contain 'only-device', got: %q", output)
	}
	// Singular "member" or "1 member" message
	if !strings.Contains(output, "1") {
		t.Errorf("expected output to contain count '1', got: %q", output)
	}
}

func TestExecute_ManyMembers(t *testing.T) {
	t.Parallel()

	devices := []string{
		"device-1", "device-2", "device-3", "device-4", "device-5",
		"device-6", "device-7", "device-8", "device-9", "device-10",
	}

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"large-group": {
				Devices: devices,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"large-group"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	// Verify all devices are listed
	for _, device := range devices {
		if !strings.Contains(output, device) {
			t.Errorf("expected output to contain %q, got: %q", device, output)
		}
	}
	// Verify count
	if !strings.Contains(output, "10") {
		t.Errorf("expected output to contain count '10', got: %q", output)
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no group name provided")
	}
}

func TestExecute_TooManyArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"group1", "group2"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with too many args")
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		GroupName: "test-group",
	}

	if opts.GroupName != "test-group" {
		t.Errorf("GroupName = %q, want %q", opts.GroupName, "test-group")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly group members",
		"-o json",
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
		"device",
		"member",
		"group",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(strings.ToLower(cmd.Long), pattern) {
			t.Errorf("expected Long to contain %q (case-insensitive)", pattern)
		}
	}
}

func TestExecute_GroupWithSpecialChars(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"living-room-2nd-floor": {
				Devices: []string{"smart-plug-1"},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"living-room-2nd-floor"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "smart-plug-1") {
		t.Errorf("expected output to contain 'smart-plug-1', got: %q", output)
	}
}

func TestRun_TableOutput(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"kitchen": {
				Devices: []string{"oven-outlet", "fridge-monitor"},
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)
	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"kitchen", "-o", "table"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	// Table output should contain header (may be uppercase DEVICE) and devices
	if !strings.Contains(strings.ToUpper(output), "DEVICE") {
		t.Errorf("expected table header 'DEVICE', got: %q", output)
	}
	if !strings.Contains(output, "oven-outlet") {
		t.Errorf("expected output to contain 'oven-outlet', got: %q", output)
	}
	if !strings.Contains(output, "fridge-monitor") {
		t.Errorf("expected output to contain 'fridge-monitor', got: %q", output)
	}
}
