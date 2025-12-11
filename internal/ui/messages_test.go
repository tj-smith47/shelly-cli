package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// captureStdout captures stdout during test execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Logf("warning: failed to close pipe reader: %v", err)
	}

	return buf.String()
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestInfo_NotQuiet(t *testing.T) {
	resetViper()
	viper.Set("quiet", false)

	output := captureStdout(t, func() {
		Info("test info message")
	})

	if !strings.Contains(output, "test info message") {
		t.Errorf("Info() output should contain message, got %q", output)
	}
	if !strings.Contains(output, "ℹ") {
		t.Errorf("Info() output should contain info icon, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestInfo_Quiet(t *testing.T) {
	resetViper()
	viper.Set("quiet", true)

	output := captureStdout(t, func() {
		Info("test info message")
	})

	if output != "" {
		t.Errorf("Info() should not output when quiet, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestInfo_WithFormatting(t *testing.T) {
	resetViper()
	viper.Set("quiet", false)

	output := captureStdout(t, func() {
		Info("found %d devices at %s", 5, "192.168.1.0/24")
	})

	if !strings.Contains(output, "found 5 devices at 192.168.1.0/24") {
		t.Errorf("Info() should format message, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestSuccess(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Success("operation completed")
	})

	if !strings.Contains(output, "operation completed") {
		t.Errorf("Success() output should contain message, got %q", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("Success() output should contain check mark, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestSuccess_WithFormatting(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Success("added %d devices", 3)
	})

	if !strings.Contains(output, "added 3 devices") {
		t.Errorf("Success() should format message, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestWarning(t *testing.T) {
	resetViper()

	output := captureStderr(t, func() {
		Warning("something is wrong")
	})

	if !strings.Contains(output, "something is wrong") {
		t.Errorf("Warning() output should contain message, got %q", output)
	}
	if !strings.Contains(output, "⚠") {
		t.Errorf("Warning() output should contain warning icon, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestError_Output(t *testing.T) {
	resetViper()

	output := captureStderr(t, func() {
		Error("operation failed")
	})

	if !strings.Contains(output, "operation failed") {
		t.Errorf("Error() output should contain message, got %q", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("Error() output should contain X mark, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestError_WithFormatting(t *testing.T) {
	resetViper()

	output := captureStderr(t, func() {
		Error("failed to connect to %s: %s", "192.168.1.100", "timeout")
	})

	if !strings.Contains(output, "failed to connect to 192.168.1.100: timeout") {
		t.Errorf("Error() should format message, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestTitle(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Title("Device List")
	})

	if !strings.Contains(output, "Device List") {
		t.Errorf("Title() output should contain message, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestHint_NotQuiet(t *testing.T) {
	resetViper()
	viper.Set("quiet", false)

	output := captureStdout(t, func() {
		Hint("try --help for more options")
	})

	if !strings.Contains(output, "try --help for more options") {
		t.Errorf("Hint() output should contain message, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestHint_Quiet(t *testing.T) {
	resetViper()
	viper.Set("quiet", true)

	output := captureStdout(t, func() {
		Hint("try --help for more options")
	})

	if output != "" {
		t.Errorf("Hint() should not output when quiet, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestCount_Singular(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Count("device", 1)
	})

	if !strings.Contains(output, "Found 1 device\n") {
		t.Errorf("Count() should use singular form, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestCount_Plural(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Count("device", 5)
	})

	if !strings.Contains(output, "Found 5 devices\n") {
		t.Errorf("Count() should use plural form, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestCount_Zero(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Count("device", 0)
	})

	if !strings.Contains(output, "Found 0 devices\n") {
		t.Errorf("Count() should use plural form for zero, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestNoResults_Basic(t *testing.T) {
	resetViper()
	viper.Set("quiet", false)

	output := captureStdout(t, func() {
		NoResults("devices")
	})

	if !strings.Contains(output, "No devices found") {
		t.Errorf("NoResults() output should contain message, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestNoResults_WithHints(t *testing.T) {
	resetViper()
	viper.Set("quiet", false)

	output := captureStdout(t, func() {
		NoResults("devices", "Check your network", "Try --verbose for more info")
	})

	if !strings.Contains(output, "No devices found") {
		t.Errorf("NoResults() should show no results message, got %q", output)
	}
	if !strings.Contains(output, "Check your network") {
		t.Errorf("NoResults() should show first hint, got %q", output)
	}
	if !strings.Contains(output, "Try --verbose for more info") {
		t.Errorf("NoResults() should show second hint, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestAdded_WithItems(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Added("device", 3)
	})

	if !strings.Contains(output, "Added 3 devices") {
		t.Errorf("Added() should show count message, got %q", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("Added() should show success icon, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestAdded_Singular(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Added("device", 1)
	})

	if !strings.Contains(output, "Added 1 device") {
		t.Errorf("Added() should use singular form, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestAdded_Zero(t *testing.T) {
	resetViper()

	output := captureStdout(t, func() {
		Added("device", 0)
	})

	if output != "" {
		t.Errorf("Added() should not output for zero count, got %q", output)
	}
}
