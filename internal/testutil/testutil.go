// Package testutil provides testing utilities.
package testutil

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// TempDir creates a temporary directory for testing and returns a cleanup function.
func TempDir(t *testing.T) (dir string, cleanup func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("warning: failed to remove temp dir %s: %v", dir, err)
		}
	}
}

// TempFile creates a temporary file with the given content.
func TempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

// CaptureOutput captures stdout/stderr during test execution.
type CaptureOutput struct {
	oldStdout *os.File
	oldStderr *os.File
	outR      *os.File
	outW      *os.File
	errR      *os.File
	errW      *os.File
}

// NewCaptureOutput starts capturing stdout and stderr.
func NewCaptureOutput(t *testing.T) *CaptureOutput {
	t.Helper()

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	c := &CaptureOutput{
		oldStdout: os.Stdout,
		oldStderr: os.Stderr,
		outR:      outR,
		outW:      outW,
		errR:      errR,
		errW:      errW,
	}

	os.Stdout = outW
	os.Stderr = errW

	return c
}

// Stop stops capturing and returns the captured output.
func (c *CaptureOutput) Stop() (stdout, stderr string) {
	if err := c.outW.Close(); err != nil {
		// Log but don't fail - we still want to restore stdout/stderr.
		//nolint:errcheck // Best-effort warning output
		os.Stderr.WriteString("warning: failed to close stdout writer: " + err.Error() + "\n")
	}
	if err := c.errW.Close(); err != nil {
		//nolint:errcheck // Best-effort warning output
		os.Stderr.WriteString("warning: failed to close stderr writer: " + err.Error() + "\n")
	}

	var outBuf, errBuf bytes.Buffer
	if _, err := io.Copy(&outBuf, c.outR); err != nil {
		//nolint:errcheck // Best-effort warning output
		os.Stderr.WriteString("warning: failed to read stdout: " + err.Error() + "\n")
	}
	if _, err := io.Copy(&errBuf, c.errR); err != nil {
		//nolint:errcheck // Best-effort warning output
		os.Stderr.WriteString("warning: failed to read stderr: " + err.Error() + "\n")
	}

	os.Stdout = c.oldStdout
	os.Stderr = c.oldStderr

	if err := c.outR.Close(); err != nil {
		os.Stderr.WriteString("warning: failed to close stdout reader: " + err.Error() + "\n") //nolint:errcheck // Best-effort warning
	}
	if err := c.errR.Close(); err != nil {
		os.Stderr.WriteString("warning: failed to close stderr reader: " + err.Error() + "\n") //nolint:errcheck // Best-effort warning
	}

	return outBuf.String(), errBuf.String()
}

// ResetViper resets viper to a clean state for testing.
func ResetViper() {
	viper.Reset()
}

// SetupTestConfig sets up a minimal viper config for testing.
func SetupTestConfig(t *testing.T) {
	t.Helper()
	ResetViper()
	viper.SetDefault("output", "table")
	viper.SetDefault("color", true)
	viper.SetDefault("theme", "dracula")
	viper.SetDefault("api_mode", "local")
}

// AssertContains checks if a string contains a substring.
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !bytes.Contains([]byte(s), []byte(substr)) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

// AssertNotContains checks if a string does not contain a substring.
func AssertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if bytes.Contains([]byte(s), []byte(substr)) {
		t.Errorf("expected %q to not contain %q", s, substr)
	}
}

// AssertEqual checks if two values are equal.
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNotEqual checks if two values are not equal.
func AssertNotEqual[T comparable](t *testing.T, got, notWant T) {
	t.Helper()
	if got == notWant {
		t.Errorf("got %v, expected different value", got)
	}
}

// AssertNil checks that an error is nil.
func AssertNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// AssertError checks that an error is not nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// AssertErrorContains checks that an error contains a substring.
func AssertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error containing %q, got nil", substr)
		return
	}
	if !bytes.Contains([]byte(err.Error()), []byte(substr)) {
		t.Errorf("expected error containing %q, got %q", substr, err.Error())
	}
}

// AssertTrue checks that a condition is true.
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("expected true: %s", msg)
	}
}

// AssertFalse checks that a condition is false.
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("expected false: %s", msg)
	}
}

// Ptr returns a pointer to the given value. Useful for test data.
func Ptr[T any](v T) *T {
	return &v
}
