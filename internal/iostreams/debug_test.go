package iostreams_test

import (
	"bytes"
	"errors"
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

//nolint:paralleltest // Tests modify global state (viper)
func TestDebug_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

	// Debug prints to os.Stderr, so we can't easily capture it
	// without redirecting os.Stderr. The function doesn't panic,
	// which is the main thing we want to verify.
	iostreams.Debug("test message %s", "arg")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestDebug_QuietMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", false)

	// Should not panic when verbose is disabled
	iostreams.Debug("test message %s", "arg")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestDebugErr_VerboseMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

	testErr := errors.New("test error")
	// Should not panic
	iostreams.DebugErr("operation", testErr)
}

//nolint:paralleltest // Tests modify global state (viper)
func TestDebugErr_NilError(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

	// Should not panic with nil error
	iostreams.DebugErr("operation", nil)
}

//nolint:paralleltest // Tests modify global state (viper)
func TestDebugErr_QuietMode(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", false)

	testErr := errors.New("test error")
	// Should not panic when verbose is disabled
	iostreams.DebugErr("operation", testErr)
}

//nolint:paralleltest // Tests modify global state (viper)
func TestCloseWithDebug_Success(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

	closer := &mockCloser{err: nil}
	// Should not panic
	iostreams.CloseWithDebug("closing resource", closer)

	if !closer.closed {
		t.Error("CloseWithDebug should call Close()")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestCloseWithDebug_Error(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

	closer := &mockCloser{err: errors.New("close error")}
	// Should not panic even with error
	iostreams.CloseWithDebug("closing resource", closer)

	if !closer.closed {
		t.Error("CloseWithDebug should call Close()")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestCloseWithDebug_NilCloser(t *testing.T) {
	setupDebugViper(t)
	viper.Set("verbose", true)

	// Should not panic with nil closer
	iostreams.CloseWithDebug("closing resource", nil)
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
