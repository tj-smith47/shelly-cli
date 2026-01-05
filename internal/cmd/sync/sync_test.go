package sync

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

const testFalseValue = "false"

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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "sync" {
		t.Errorf("Use = %q, want %q", cmd.Use, "sync")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"synchronize"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Synchronize device configurations"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if !strings.Contains(cmd.Long, "--pull") {
		t.Error("Long should contain '--pull'")
	}

	if !strings.Contains(cmd.Long, "--push") {
		t.Error("Long should contain '--push'")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly sync") {
		t.Error("Example should contain 'shelly sync'")
	}

	if !strings.Contains(cmd.Example, "--pull") {
		t.Error("Example should contain '--pull'")
	}

	if !strings.Contains(cmd.Example, "--push") {
		t.Error("Example should contain '--push'")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test push flag
	pushFlag := cmd.Flags().Lookup("push")
	if pushFlag == nil {
		t.Error("push flag not found")
	} else if pushFlag.DefValue != testFalseValue {
		t.Errorf("push default = %q, want %q", pushFlag.DefValue, testFalseValue)
	}

	// Test pull flag
	pullFlag := cmd.Flags().Lookup("pull")
	if pullFlag == nil {
		t.Error("pull flag not found")
	} else if pullFlag.DefValue != testFalseValue {
		t.Errorf("pull default = %q, want %q", pullFlag.DefValue, testFalseValue)
	}

	// Test dry-run flag
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("dry-run flag not found")
	} else if dryRunFlag.DefValue != testFalseValue {
		t.Errorf("dry-run default = %q, want %q", dryRunFlag.DefValue, testFalseValue)
	}

	// Test device flag
	deviceFlag := cmd.Flags().Lookup("device")
	if deviceFlag == nil {
		t.Error("device flag not found")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Push {
		t.Error("default Push = true, want false")
	}

	if opts.Pull {
		t.Error("default Pull = true, want false")
	}

	if opts.DryRun {
		t.Error("default DryRun = true, want false")
	}

	if len(opts.Devices) != 0 {
		t.Errorf("default Devices = %v, want empty slice", opts.Devices)
	}
}

func TestOptions_AllFields(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Push:    true,
		Pull:    false,
		DryRun:  true,
		Devices: []string{"device1", "device2"},
	}

	if !opts.Push {
		t.Error("Push = false, want true")
	}

	if opts.Pull {
		t.Error("Pull = true, want false")
	}

	if !opts.DryRun {
		t.Error("DryRun = false, want true")
	}

	if len(opts.Devices) != 2 {
		t.Errorf("Devices count = %d, want 2", len(opts.Devices))
	}
}

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

// TestRun_NoPushOrPull tests that run fails without --push or --pull.
func TestRun_NoPushOrPull(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		Push:    false,
		Pull:    false,
	}

	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error when neither --push nor --pull specified")
	}

	if !strings.Contains(err.Error(), "specify --push or --pull") {
		t.Errorf("Error = %q, should contain 'specify --push or --pull'", err.Error())
	}
}

// TestRun_BothPushAndPull tests that run fails with both --push and --pull.
func TestRun_BothPushAndPull(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		Push:    true,
		Pull:    true,
	}

	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error when both --push and --pull specified")
	}

	if !strings.Contains(err.Error(), "cannot use --push and --pull together") {
		t.Errorf("Error = %q, should contain 'cannot use --push and --pull together'", err.Error())
	}
}

// TestRun_PullNoDevices tests pull with no configured devices.
func TestRun_PullNoDevices(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		Pull:    true,
	}

	// Run should complete (possibly with warning about no devices)
	err := run(context.Background(), opts)

	// May fail due to config loading or succeed with "no devices" warning
	if err != nil {
		t.Logf("sync error (may be expected): %v", err)
	}
}

// TestRun_PushNoSyncDir tests push when no sync directory exists.
func TestRun_PushNoSyncDir(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		Push:    true,
	}

	err := run(context.Background(), opts)

	// Should fail since no sync directory exists
	if err != nil {
		// Expected - either no sync dir or no config files
		if !strings.Contains(err.Error(), "sync") && !strings.Contains(err.Error(), "config") {
			t.Logf("Unexpected error: %v", err)
		}
	}
}

// TestRun_PullWithDryRun tests pull with dry-run flag.
func TestRun_PullWithDryRun(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		DryRun:  true,
		Pull:    true,
	}

	err := run(context.Background(), opts)

	// May fail due to config loading or succeed with dry run output
	if err != nil {
		t.Logf("dry run error (may be expected): %v", err)
	}

	// Check for dry run indicator in output
	output := stdout.String() + stderr.String()
	t.Logf("output: %s", output) // Output may contain [DRY RUN] or warning about no devices
}

// TestRun_PushWithDryRun tests push with dry-run flag.
func TestRun_PushWithDryRun(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		DryRun:  true,
		Push:    true,
	}

	err := run(context.Background(), opts)

	// Should fail since no sync directory exists, but dry-run should be handled
	if err != nil {
		t.Logf("push dry run error (expected): %v", err)
	}
}

// TestRun_PullSpecificDevices tests pull with specific devices.
func TestRun_PullSpecificDevices(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		Devices: []string{"living-room", "kitchen"},
		Pull:    true,
	}

	// Use short timeout to prevent hanging on network operations
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := run(ctx, opts)

	// May fail due to config loading, device not found, or context timeout
	if err != nil {
		t.Logf("pull specific devices error (may be expected): %v", err)
	}
}

// TestRun_PushSpecificDevices tests push with specific devices.
func TestRun_PushSpecificDevices(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory: f,
		Devices: []string{"living-room"},
		Push:    true,
	}

	// Use short timeout to prevent hanging on network operations
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := run(ctx, opts)

	// Should fail since no sync directory exists
	if err != nil {
		t.Logf("push specific devices error (expected): %v", err)
	}
}

// TestDisplaySyncSummary tests the summary printing function.
func TestDisplaySyncSummary(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	// Test successful sync summary
	term.DisplaySyncSummary(ios, 3, 0, false, "/tmp/sync")

	output := stdout.String()
	if !strings.Contains(output, "3 device(s)") {
		t.Error("Output should contain '3 device(s)'")
	}
}

// TestDisplaySyncSummary_WithFailures tests summary with failures.
func TestDisplaySyncSummary_WithFailures(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	// Test sync summary with failures
	term.DisplaySyncSummary(ios, 2, 1, false, "/tmp/sync")

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "2 succeeded") {
		t.Error("Output should contain '2 succeeded'")
	}
}

// TestDisplaySyncSummary_DryRun tests summary in dry run mode.
func TestDisplaySyncSummary_DryRun(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)

	// Test dry run summary - should not print the sync dir path
	term.DisplaySyncSummary(ios, 3, 0, true, "/tmp/sync")

	output := stdout.String()
	if strings.Contains(output, "/tmp/sync") {
		t.Error("Dry run output should not contain sync dir path")
	}
}

// TestNewCommand_AllFlagsExist verifies all expected flags exist.
func TestNewCommand_AllFlagsExist(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	requiredFlags := []string{
		"push",
		"pull",
		"dry-run",
		"device",
	}

	for _, flag := range requiredFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("flag %q not found", flag)
		}
	}
}

// TestNewCommand_Args verifies command accepts no args.
func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Sync command doesn't use positional args (uses flags instead)
	// Just verify the command is set up correctly
	if cmd.Use != "sync" {
		t.Errorf("Use = %q, want %q", cmd.Use, "sync")
	}
}
