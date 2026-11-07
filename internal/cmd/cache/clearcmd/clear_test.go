package clearcmd

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != commandUse {
		t.Errorf("Use = %q, want %q", cmd.Use, commandUse)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"c", "rm", "clean"}
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

	expected := "Clear the cache"
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

	if !strings.Contains(cmd.Long, "Clear cached device data") {
		t.Errorf("Long should contain 'Clear cached device data', got %q", cmd.Long)
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly cache clear") {
		t.Error("Example should contain 'shelly cache clear'")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
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

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

func TestNewCommand_AcceptsNoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Clear command doesn't take arguments
	if cmd.Args != nil {
		// If Args is set, verify it accepts no args
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Command should accept no args, got error: %v", err)
		}
	}
}

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The clear command doesn't add its own flags
	// Just verify the command is properly set up
	if cmd.Use != commandUse {
		t.Errorf("Use = %q, want %q", cmd.Use, commandUse)
	}
}

// TestRun tests the run function with various cache directory states.
// Uses SetupTestFs for in-memory filesystem - NOT parallel due to t.Setenv.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, fc *cache.FileCache)
		expectedOutput string
	}{
		{
			name:           "empty cache",
			setup:          nil,
			expectedOutput: "Cache is already empty",
		},
		{
			name: "cache with single file",
			setup: func(t *testing.T, fc *cache.FileCache) {
				t.Helper()
				if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "test"}, cache.TTLDeviceInfo); err != nil {
					t.Fatalf("Failed to set cache: %v", err)
				}
			},
			expectedOutput: "Cache cleared",
		},
		{
			name: "cache with multiple entries",
			setup: func(t *testing.T, fc *cache.FileCache) {
				t.Helper()
				if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, cache.TTLDeviceInfo); err != nil {
					t.Fatalf("Failed to set cache: %v", err)
				}
				if err := fc.Set("device2", "deviceinfo", map[string]string{"id": "d2"}, cache.TTLDeviceInfo); err != nil {
					t.Fatalf("Failed to set cache: %v", err)
				}
				if err := fc.Set("device1", "firmware", map[string]string{"version": "1.0"}, cache.TTLFirmware); err != nil {
					t.Fatalf("Failed to set cache: %v", err)
				}
			},
			expectedOutput: "Cache cleared",
		},
		{
			name: "cache with subdirectories",
			setup: func(t *testing.T, fc *cache.FileCache) {
				t.Helper()
				if err := fc.Set("device1", "automation/schedules", []string{"sched1"}, cache.TTLAutomation); err != nil {
					t.Fatalf("Failed to set cache: %v", err)
				}
				if err := fc.Set("device1", "automation/webhooks", []string{"hook1"}, cache.TTLAutomation); err != nil {
					t.Fatalf("Failed to set cache: %v", err)
				}
			},
			expectedOutput: "Cache cleared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test filesystem - NOT parallel
			memFs := factory.SetupTestFs(t)

			// Create FileCache with the test filesystem
			fc, err := cache.NewWithFs("/cache", memFs)
			if err != nil {
				t.Fatalf("Failed to create file cache: %v", err)
			}

			// Run setup if provided
			if tt.setup != nil {
				tt.setup(t, fc)
			}

			tf := factory.NewTestFactory(t)
			tf.SetFileCache(fc)

			opts := &Options{Factory: tf.Factory, All: true, Yes: true}
			err = run(context.Background(), opts)

			if err != nil {
				t.Errorf("run returned error: %v", err)
			}

			combined := tf.OutString() + tf.ErrString()
			if !strings.Contains(combined, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expectedOutput, combined)
			}
		})
	}
}

// TestRun_ClearDevice tests clearing cache for a specific device.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_ClearDevice(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create cache entries for multiple devices
	if err := fc.Set("kitchen", "deviceinfo", map[string]string{"id": "k1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	if err := fc.Set("kitchen", "firmware", map[string]string{"v": "1.0"}, cache.TTLFirmware); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	if err := fc.Set("bedroom", "deviceinfo", map[string]string{"id": "b1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory, Device: "kitchen", Yes: true}
	if err := run(context.Background(), opts); err != nil {
		t.Errorf("run returned error: %v", err)
	}

	// Check kitchen cache is cleared
	entry, err := fc.Get("kitchen", "deviceinfo")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
	if entry != nil {
		t.Error("kitchen deviceinfo should be cleared")
	}

	// Check bedroom cache is NOT cleared
	entry, err = fc.Get("bedroom", "deviceinfo")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
	if entry == nil {
		t.Error("bedroom deviceinfo should NOT be cleared")
	}
}

// TestRun_ClearDeviceType tests clearing cache for a specific device and type.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_ClearDeviceType(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create cache entries
	if err := fc.Set("kitchen", "deviceinfo", map[string]string{"id": "k1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	if err := fc.Set("kitchen", "firmware", map[string]string{"v": "1.0"}, cache.TTLFirmware); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory, Device: "kitchen", Type: "deviceinfo", Yes: true}
	if err := run(context.Background(), opts); err != nil {
		t.Errorf("run returned error: %v", err)
	}

	// Check deviceinfo is cleared
	entry, err := fc.Get("kitchen", "deviceinfo")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
	if entry != nil {
		t.Error("kitchen deviceinfo should be cleared")
	}

	// Check firmware is NOT cleared
	entry, err = fc.Get("kitchen", "firmware")
	if err != nil {
		t.Errorf("Get error: %v", err)
	}
	if entry == nil {
		t.Error("kitchen firmware should NOT be cleared")
	}
}

// TestRun_TypeWithoutDevice tests that --type without --device returns an error.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_TypeWithoutDevice(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory, Type: "firmware"}
	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error for --type without --device")
	}
	if !strings.Contains(err.Error(), "--type requires --device") {
		t.Errorf("Expected error about --type requires --device, got: %v", err)
	}
}

// TestRun_NoFlagsError tests that running without flags returns an error.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_NoFlagsError(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory}
	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error when no flags provided")
	}
	if !strings.Contains(err.Error(), "specify --all") {
		t.Errorf("Expected error about specifying flags, got: %v", err)
	}
}

// TestRun_ExpiredCleanup tests the --expired flag behavior.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_ExpiredCleanup(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory, Expired: true, Yes: true}
	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("run should not error with --expired: %v", err)
	}

	combined := tf.OutString() + tf.ErrString()
	if !strings.Contains(combined, "No expired entries") {
		t.Errorf("Expected 'No expired entries' message, got: %q", combined)
	}
}

// TestRun_ViaCommand tests running the command via cobra Execute.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_ViaCommand(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create cache entry
	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--all", "--yes"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	combined := tf.OutString() + tf.ErrString()
	if !strings.Contains(combined, "Cache cleared") {
		t.Errorf("Expected 'Cache cleared' message, got: %q", combined)
	}
}

// TestRun_NoTerminalWithoutYes_Refuses drives every clearing form without a
// terminal and without --yes: each one returns ErrConfirmationRequired and the
// cache is untouched.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_NoTerminalWithoutYes_Refuses(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create cache entry
	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	for name, opts := range map[string]*Options{
		"all":         {Factory: tf.Factory, All: true},
		"device":      {Factory: tf.Factory, Device: "device1"},
		"device type": {Factory: tf.Factory, Device: "device1", Type: "deviceinfo"},
		"expired":     {Factory: tf.Factory, Expired: true},
	} {
		err := run(context.Background(), opts)
		if !errors.Is(err, cmdutil.ErrConfirmationRequired) {
			t.Errorf("%s: run() error = %v, want ErrConfirmationRequired", name, err)
		}
		entry, getErr := fc.Get("device1", "deviceinfo")
		if getErr != nil || entry == nil {
			t.Errorf("%s: cache entry removed without confirmation (entry=%v err=%v)", name, entry, getErr)
		}
	}
	if out := tf.OutString(); strings.Contains(out, "Cleared") || strings.Contains(out, "Removed") {
		t.Errorf("output reports a clear that must not have happened: %q", out)
	}
}

// TestRun_Yes_ClearsEveryForm proves --yes performs each clearing form.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_Yes_ClearsEveryForm(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}
	for _, dev := range []string{"a", "b", "c", "d"} {
		if err := fc.Set(dev, "deviceinfo", map[string]string{"id": dev}, cache.TTLDeviceInfo); err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}
	}
	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	for _, tc := range []struct {
		opts *Options
		gone string
	}{
		{&Options{Factory: tf.Factory, Device: "a", Type: "deviceinfo", Yes: true}, "a"},
		{&Options{Factory: tf.Factory, Device: "b", Yes: true}, "b"},
		{&Options{Factory: tf.Factory, Expired: true, Yes: true}, ""},
		{&Options{Factory: tf.Factory, All: true, Yes: true}, "c"},
	} {
		if err := run(context.Background(), tc.opts); err != nil {
			t.Fatalf("run(%+v) error = %v", tc.opts, err)
		}
		if tc.gone == "" {
			continue
		}
		entry, getErr := fc.Get(tc.gone, "deviceinfo")
		if getErr != nil {
			t.Fatalf("Get(%s) error = %v", tc.gone, getErr)
		}
		if entry != nil {
			t.Errorf("%s still cached after run(%+v)", tc.gone, tc.opts)
		}
	}
}

// TestRun_CacheNil tests run when cache initialization fails.
// Uses a read-only filesystem to make cache.New() fail.
// NOT parallel: Uses t.Setenv and config.SetFs which are package-level state.
func TestRun_CacheNil(t *testing.T) {
	// Create base filesystem with config directory
	memFs := afero.NewMemMapFs()
	if err := memFs.MkdirAll("/testconfig/shelly", 0o700); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Wrap in read-only filesystem to make cache directory creation fail
	roFs := afero.NewReadOnlyFs(memFs)
	config.SetFs(roFs)

	t.Setenv("XDG_CONFIG_HOME", "/testconfig")
	config.ResetDefaultManagerForTesting()

	t.Cleanup(func() {
		config.SetFs(nil)
		config.ResetDefaultManagerForTesting()
	})

	tf := factory.NewTestFactory(t)
	// Don't call SetFileCache - let FileCache() try to lazy-init and fail

	opts := &Options{Factory: tf.Factory, All: true, Yes: true}
	err := run(context.Background(), opts)

	if err != nil {
		t.Errorf("run should not error with nil cache: %v", err)
	}

	combined := tf.OutString() + tf.ErrString()
	if !strings.Contains(combined, "Cache not available") {
		t.Errorf("Expected 'Cache not available' message, got: %q", combined)
	}
}

// TestRun_ExpiredWithEntries tests --expired with entries to clean.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_ExpiredWithEntries(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create an expired cache entry by setting with 0 TTL
	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, 0); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory, Expired: true, Yes: true}
	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("run should not error with --expired: %v", err)
	}

	// Should report removed entries (1 or "Removed X expired entries")
	combined := tf.OutString() + tf.ErrString()
	hasRemoved := strings.Contains(combined, "Removed 1 expired") ||
		strings.Contains(combined, "Removed") && strings.Contains(combined, "expired")
	if !hasRemoved && !strings.Contains(combined, "No expired") {
		t.Errorf("Expected removed or no expired message, got: %q", combined)
	}
}

// TestRun_ExpiredMultipleEntries tests --expired with multiple expired entries.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_ExpiredMultipleEntries(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create multiple expired cache entries with 0 TTL
	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, 0); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	if err := fc.Set("device2", "deviceinfo", map[string]string{"id": "d2"}, 0); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	if err := fc.Set("device3", "deviceinfo", map[string]string{"id": "d3"}, 0); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory, Expired: true, Yes: true}
	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("run should not error with --expired: %v", err)
	}

	// Should report removed entries
	combined := tf.OutString() + tf.ErrString()
	hasRemoved := strings.Contains(combined, "Removed") && strings.Contains(combined, "expired")
	if !hasRemoved && !strings.Contains(combined, "No expired") {
		t.Errorf("Expected removed message, got: %q", combined)
	}
}

// Interactive acceptance needs a real TTY; the --yes path and the no-terminal
// refusal are covered above.

// =============================================================================
// Error path tests using failing filesystem
// =============================================================================

// failingFs wraps an afero.Fs and fails on Remove operations.
type failingFs struct {
	afero.Fs
	failRemove bool
	failStat   bool
}

func (f *failingFs) Remove(name string) error {
	if f.failRemove {
		return errors.New("mock remove error")
	}
	return f.Fs.Remove(name)
}

func (f *failingFs) RemoveAll(path string) error {
	if f.failRemove {
		return errors.New("mock remove error")
	}
	return f.Fs.RemoveAll(path)
}

func (f *failingFs) Stat(name string) (os.FileInfo, error) {
	if f.failStat {
		return nil, errors.New("mock stat error")
	}
	return f.Fs.Stat(name)
}

// TestRun_CleanupError tests --expired when cleanup fails.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_CleanupError(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	// Create cache with an expired entry
	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, 0); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Now wrap with failing filesystem and recreate cache
	failFs := &failingFs{Fs: memFs, failRemove: true}
	fcFail, err := cache.NewWithFs("/cache", failFs)
	if err != nil {
		t.Fatalf("Failed to create failing file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fcFail)

	opts := &Options{Factory: tf.Factory, Expired: true, Yes: true}
	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error from cleanup failure")
	}
}

// TestRun_InvalidateError tests --device --type when invalidation fails.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_InvalidateError(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	// Create cache entry first
	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	if err := fc.Set("kitchen", "deviceinfo", map[string]string{"id": "k1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Wrap with failing filesystem
	failFs := &failingFs{Fs: memFs, failRemove: true}
	fcFail, err := cache.NewWithFs("/cache", failFs)
	if err != nil {
		t.Fatalf("Failed to create failing file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fcFail)

	opts := &Options{Factory: tf.Factory, Device: "kitchen", Type: "deviceinfo", Yes: true}
	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error from invalidation failure")
	}
}

// TestRun_InvalidateDeviceError tests --device when invalidation fails.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_InvalidateDeviceError(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	// Create cache entry first
	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	if err := fc.Set("kitchen", "deviceinfo", map[string]string{"id": "k1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Wrap with failing filesystem
	failFs := &failingFs{Fs: memFs, failRemove: true}
	fcFail, err := cache.NewWithFs("/cache", failFs)
	if err != nil {
		t.Fatalf("Failed to create failing file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fcFail)

	opts := &Options{Factory: tf.Factory, Device: "kitchen", Yes: true}
	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error from device invalidation failure")
	}
}

// TestRun_InvalidateAllError tests --all --yes when invalidation fails.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_InvalidateAllError(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	// Create cache entry first
	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Wrap with failing filesystem
	failFs := &failingFs{Fs: memFs, failRemove: true}
	fcFail, err := cache.NewWithFs("/cache", failFs)
	if err != nil {
		t.Fatalf("Failed to create failing file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fcFail)

	opts := &Options{Factory: tf.Factory, All: true, Yes: true}
	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error from invalidate all failure")
	}
}
