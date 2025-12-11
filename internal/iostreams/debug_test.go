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
	viper.Set("verbose", true)

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
	viper.Set("verbose", false)

	output := captureStderr(t, func() {
		iostreams.Debug("test message %s", "arg")
	})

	if output != "" {
		t.Errorf("Debug() should not output when verbose is false, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestDebugErr_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

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
	viper.Set("verbose", true)

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
	viper.Set("verbose", false)

	output := captureStderr(t, func() {
		iostreams.DebugErr("operation", errors.New("test error"))
	})

	if output != "" {
		t.Errorf("DebugErr() should not output when verbose is false, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stderr)
func TestCloseWithDebug_Success(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

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
	viper.Set("verbose", true)

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
	viper.Set("verbose", true)

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
	viper.Set("verbose", false)

	closer := &mockCloser{err: errors.New("close failed")}

	output := captureStderr(t, func() {
		iostreams.CloseWithDebug("test close", closer)
	})

	if !closer.closed {
		t.Error("Close() was not called")
	}
	if output != "" {
		t.Errorf("CloseWithDebug() should not output when verbose is false, got %q", output)
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
	viper.Set("verbose", true)

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
	viper.Set("verbose", false)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	ios.Debug("test message")

	if errOut.String() != "" {
		t.Errorf("Debug() should produce no output when verbose is false, got %q", errOut.String())
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIOStreams_DebugErr_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

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
	viper.Set("verbose", true)

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
	viper.Set("verbose", false)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	testErr := errors.New("test error")
	ios.DebugErr("operation", testErr)

	if errOut.String() != "" {
		t.Errorf("DebugErr() should produce no output when verbose is false, got %q", errOut.String())
	}
}
