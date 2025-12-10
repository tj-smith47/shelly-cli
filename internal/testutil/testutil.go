// Package testutil provides testing utilities.
package testutil

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// TempDir creates a temporary directory for testing and returns a cleanup function.
func TempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

// TempFile creates a temporary file with the given content.
func TempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
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
	c.outW.Close()
	c.errW.Close()

	var outBuf, errBuf bytes.Buffer
	outBuf.ReadFrom(c.outR)
	errBuf.ReadFrom(c.errR)

	os.Stdout = c.oldStdout
	os.Stderr = c.oldStderr

	c.outR.Close()
	c.errR.Close()

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
