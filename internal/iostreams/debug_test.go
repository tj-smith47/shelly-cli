package iostreams_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Note: Tests in this file modify global state (viper)
// and cannot run in parallel.

// setupDebugViper creates a clean viper instance for testing.
func setupDebugViper(t *testing.T) {
	t.Helper()
	viper.Reset()
}

// captureStderr captures stderr during test execution.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stderr = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Logf("warning: failed to close pipe reader: %v", err)
	}

	return buf.String()
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebug_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2) // Debug requires verbosity >= 2 (-vv)

	output := captureStderr(t, func() {
		iostreams.Debug("test message %d", 42)
	})

	if output != "debug: test message 42\n" {
		t.Errorf("Debug() output = %q, want %q", output, "debug: test message 42\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebug_QuietMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 0)

	output := captureStderr(t, func() {
		iostreams.Debug("test message %s", "arg")
	})

	if output != "" {
		t.Errorf("Debug() should not output when verbosity is 0, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebug_SingleVerbose(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1) // -v only, Debug requires >= 2

	output := captureStderr(t, func() {
		iostreams.Debug("test message")
	})

	if output != "" {
		t.Errorf("Debug() should not output when verbosity is 1 (requires 2), got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugErr_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1) // DebugErr requires verbosity >= 1 (-v)

	output := captureStderr(t, func() {
		iostreams.DebugErr("operation", errors.New("test error"))
	})

	if output != "debug: operation: test error\n" {
		t.Errorf("DebugErr() output = %q, want %q", output, "debug: operation: test error\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugErr_NilError(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)

	output := captureStderr(t, func() {
		iostreams.DebugErr("operation", nil)
	})

	if output != "" {
		t.Errorf("DebugErr() should not output for nil error, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugErr_QuietMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 0)

	output := captureStderr(t, func() {
		iostreams.DebugErr("operation", errors.New("test error"))
	})

	if output != "" {
		t.Errorf("DebugErr() should not output when verbosity is 0, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestCloseWithDebug_Success(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)

	closer := &mockCloser{err: nil}

	output := captureStderr(t, func() {
		iostreams.CloseWithDebug("test close", closer)
	})

	if !closer.closed {
		t.Error("Close() was not called")
	}
	if output != "" {
		t.Errorf("CloseWithDebug() should not output on success, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestCloseWithDebug_Error(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)

	closer := &mockCloser{err: errors.New("close failed")}

	output := captureStderr(t, func() {
		iostreams.CloseWithDebug("test close", closer)
	})

	if !closer.closed {
		t.Error("Close() was not called")
	}
	if output != "debug: test close: close failed\n" {
		t.Errorf("CloseWithDebug() output = %q, want %q", output, "debug: test close: close failed\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestCloseWithDebug_NilCloser(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)

	output := captureStderr(t, func() {
		iostreams.CloseWithDebug("test close", nil)
	})

	if output != "" {
		t.Errorf("CloseWithDebug() should not output for nil closer, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestCloseWithDebug_ErrorVerboseDisabled(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 0)

	closer := &mockCloser{err: errors.New("close failed")}

	output := captureStderr(t, func() {
		iostreams.CloseWithDebug("test close", closer)
	})

	if !closer.closed {
		t.Error("Close() was not called")
	}
	if output != "" {
		t.Errorf("CloseWithDebug() should not output when verbosity is 0, got %q", output)
	}
}

// mockCloser is a test helper that implements io.Closer.
type mockCloser struct {
	closed bool
	err    error
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.err
}

// IOStreams method tests

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_Debug_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2) // Debug requires verbosity >= 2 (-vv)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	ios.Debug("test message %s", "arg")

	got := errOut.String()
	if !strings.Contains(got, "debug:") {
		t.Errorf("Debug() output = %q, should contain 'debug:'", got)
	}
	if !strings.Contains(got, "test message arg") {
		t.Errorf("Debug() output = %q, should contain 'test message arg'", got)
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_Debug_QuietMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 0)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	ios.Debug("test message")

	if errOut.String() != "" {
		t.Errorf("Debug() should produce no output when verbosity is 0, got %q", errOut.String())
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_DebugErr_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1) // DebugErr requires verbosity >= 1 (-v)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	testErr := errors.New("test error")
	ios.DebugErr("operation", testErr)

	got := errOut.String()
	if !strings.Contains(got, "debug:") {
		t.Errorf("DebugErr() output = %q, should contain 'debug:'", got)
	}
	if !strings.Contains(got, "operation") {
		t.Errorf("DebugErr() output = %q, should contain 'operation'", got)
	}
	if !strings.Contains(got, "test error") {
		t.Errorf("DebugErr() output = %q, should contain 'test error'", got)
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_DebugErr_NilError(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	ios.DebugErr("operation", nil)

	if errOut.String() != "" {
		t.Errorf("DebugErr() with nil error should produce no output, got %q", errOut.String())
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_DebugErr_QuietMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 0)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	testErr := errors.New("test error")
	ios.DebugErr("operation", testErr)

	if errOut.String() != "" {
		t.Errorf("DebugErr() should produce no output when verbosity is 0, got %q", errOut.String())
	}
}

// =============================================================================
// Trace Tests
// =============================================================================

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestTrace_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 3) // Trace requires verbosity >= 3 (-vvv)

	output := captureStderr(t, func() {
		iostreams.Trace("trace message %d", 123)
	})

	if output != "trace: trace message 123\n" {
		t.Errorf("Trace() output = %q, want %q", output, "trace: trace message 123\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestTrace_InsufficientVerbosity(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2) // -vv only, Trace requires >= 3

	output := captureStderr(t, func() {
		iostreams.Trace("trace message")
	})

	if output != "" {
		t.Errorf("Trace() should not output when verbosity is 2 (requires 3), got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_Trace_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 3)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	ios.Trace("trace message %s", "arg")

	got := errOut.String()
	if !strings.Contains(got, "trace:") {
		t.Errorf("Trace() output = %q, should contain 'trace:'", got)
	}
	if !strings.Contains(got, "trace message arg") {
		t.Errorf("Trace() output = %q, should contain 'trace message arg'", got)
	}
}

// =============================================================================
// Category-Aware Function Tests
// =============================================================================

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugCat_WithCategory(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2)

	output := captureStderr(t, func() {
		iostreams.DebugCat(iostreams.CategoryNetwork, "network message %d", 42)
	})

	if output != "debug:network: network message 42\n" {
		t.Errorf("DebugCat() output = %q, want %q", output, "debug:network: network message 42\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugCat_CategoryFiltered(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2)
	viper.Set("log.categories", "device,config") // Only device and config

	output := captureStderr(t, func() {
		iostreams.DebugCat(iostreams.CategoryNetwork, "network message")
	})

	if output != "" {
		t.Errorf("DebugCat() should not output when category is filtered, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugCat_CategoryAllowed(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2)
	viper.Set("log.categories", "network,api")

	output := captureStderr(t, func() {
		iostreams.DebugCat(iostreams.CategoryNetwork, "network message")
	})

	if output != "debug:network: network message\n" {
		t.Errorf("DebugCat() output = %q, want %q", output, "debug:network: network message\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugErrCat_WithCategory(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)

	output := captureStderr(t, func() {
		iostreams.DebugErrCat(iostreams.CategoryDiscovery, "discovery failed", errors.New("timeout"))
	})

	if output != "debug:discovery: discovery failed: timeout\n" {
		t.Errorf("DebugErrCat() output = %q, want %q", output, "debug:discovery: discovery failed: timeout\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugErrCat_CategoryFiltered(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)
	viper.Set("log.categories", "network")

	output := captureStderr(t, func() {
		iostreams.DebugErrCat(iostreams.CategoryDiscovery, "discovery failed", errors.New("timeout"))
	})

	if output != "" {
		t.Errorf("DebugErrCat() should not output when category is filtered, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugErrCat_NilError(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)

	output := captureStderr(t, func() {
		iostreams.DebugErrCat(iostreams.CategoryNetwork, "operation", nil)
	})

	if output != "" {
		t.Errorf("DebugErrCat() should not output for nil error, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestTraceCat_WithCategory(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 3)

	output := captureStderr(t, func() {
		iostreams.TraceCat(iostreams.CategoryAPI, "api trace %s", "detail")
	})

	if output != "trace:api: api trace detail\n" {
		t.Errorf("TraceCat() output = %q, want %q", output, "trace:api: api trace detail\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestTraceCat_CategoryFiltered(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 3)
	viper.Set("log.categories", "device")

	output := captureStderr(t, func() {
		iostreams.TraceCat(iostreams.CategoryAPI, "api trace")
	})

	if output != "" {
		t.Errorf("TraceCat() should not output when category is filtered, got %q", output)
	}
}

// =============================================================================
// IOStreams Category-Aware Method Tests
// =============================================================================

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_DebugCat(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	ios.DebugCat(iostreams.CategoryDevice, "device message")

	got := errOut.String()
	if !strings.Contains(got, "debug:device:") {
		t.Errorf("DebugCat() output = %q, should contain 'debug:device:'", got)
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_DebugErrCat(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	ios.DebugErrCat(iostreams.CategoryFirmware, "firmware error", errors.New("update failed"))

	got := errOut.String()
	if !strings.Contains(got, "debug:firmware:") {
		t.Errorf("DebugErrCat() output = %q, should contain 'debug:firmware:'", got)
	}
	if !strings.Contains(got, "update failed") {
		t.Errorf("DebugErrCat() output = %q, should contain 'update failed'", got)
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_TraceCat(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 3)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	ios.TraceCat(iostreams.CategoryConfig, "config trace")

	got := errOut.String()
	if !strings.Contains(got, "trace:config:") {
		t.Errorf("TraceCat() output = %q, should contain 'trace:config:'", got)
	}
}

// =============================================================================
// JSON Output Tests
// =============================================================================

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebug_JSONOutput(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2)
	viper.Set("log.json", true)

	output := captureStderr(t, func() {
		iostreams.Debug("json test message")
	})

	if !strings.Contains(output, `"level":"debug"`) {
		t.Errorf("JSON output should contain level, got %q", output)
	}
	if !strings.Contains(output, `"message":"json test message"`) {
		t.Errorf("JSON output should contain message, got %q", output)
	}
	if !strings.Contains(output, `"time":`) {
		t.Errorf("JSON output should contain time, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugCat_JSONOutput(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2)
	viper.Set("log.json", true)

	output := captureStderr(t, func() {
		iostreams.DebugCat(iostreams.CategoryNetwork, "network json test")
	})

	if !strings.Contains(output, `"category":"network"`) {
		t.Errorf("JSON output should contain category, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugErr_JSONOutput(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 1)
	viper.Set("log.json", true)

	output := captureStderr(t, func() {
		iostreams.DebugErr("operation", errors.New("test error"))
	})

	if !strings.Contains(output, `"error":"test error"`) {
		t.Errorf("JSON output should contain error, got %q", output)
	}
}

// =============================================================================
// Category Filter Edge Cases
// =============================================================================

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugCat_EmptyFilterShowsAll(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2)
	viper.Set("log.categories", "") // Empty = show all

	output := captureStderr(t, func() {
		iostreams.DebugCat(iostreams.CategoryNetwork, "should show")
	})

	if output == "" {
		t.Error("DebugCat() with empty filter should show all categories")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugCat_FilterWithSpaces(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbosity", 2)
	viper.Set("log.categories", "device, network , api") // Spaces around commas

	output := captureStderr(t, func() {
		iostreams.DebugCat(iostreams.CategoryNetwork, "network message")
	})

	if output == "" {
		t.Error("DebugCat() should handle spaces in category filter")
	}
}
