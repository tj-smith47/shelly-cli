package iostreams_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// setupViper creates a clean viper instance for testing.
func setupViper(t *testing.T) {
	t.Helper()
	viper.Reset()
}

func TestIOStreams_Info(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Info("Test message %d", 123)

	output := out.String()
	if !strings.Contains(output, "Test message 123") {
		t.Errorf("Info() should contain message, got %q", output)
	}
	if !strings.Contains(output, "→") {
		t.Errorf("Info() should contain info arrow, got %q", output)
	}
}

func TestIOStreams_Info_Quiet(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetQuiet(true)

	ios.Info("Test message")

	if out.Len() != 0 {
		t.Errorf("Info() in quiet mode should not output, got %q", out.String())
	}
}

func TestIOStreams_Success(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Success("Operation %s", "completed")

	output := out.String()
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("Success() should contain message, got %q", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("Success() should contain checkmark, got %q", output)
	}
}

func TestIOStreams_Warning(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	ios.Warning("Danger %s", "ahead")

	output := errOut.String()
	if !strings.Contains(output, "Danger ahead") {
		t.Errorf("Warning() should contain message, got %q", output)
	}
	if !strings.Contains(output, "⚠") {
		t.Errorf("Warning() should contain warning icon, got %q", output)
	}
}

func TestIOStreams_Error(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	ios.Error("Failed: %s", "reason")

	output := errOut.String()
	if !strings.Contains(output, "Failed: reason") {
		t.Errorf("Error() should contain message, got %q", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("Error() should contain error icon, got %q", output)
	}
}

func TestIOStreams_Plain(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Plain("Plain text: %s", "value")

	output := out.String()
	if !strings.Contains(output, "Plain text: value") {
		t.Errorf("Plain() should contain message, got %q", output)
	}
}

func TestIOStreams_Hint(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Hint("Try using --help")

	output := out.String()
	if !strings.Contains(output, "Try using --help") {
		t.Errorf("Hint() should contain message, got %q", output)
	}
}

func TestIOStreams_Hint_Quiet(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetQuiet(true)

	ios.Hint("Try using --help")

	if out.Len() != 0 {
		t.Errorf("Hint() in quiet mode should not output, got %q", out.String())
	}
}

func TestIOStreams_Title(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Title("Section: %s", "Devices")

	output := out.String()
	if !strings.Contains(output, "Section: Devices") {
		t.Errorf("Title() should contain message, got %q", output)
	}
}

func TestIOStreams_Subtitle(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Subtitle("Sub: %s", "Info")

	output := out.String()
	if !strings.Contains(output, "Sub: Info") {
		t.Errorf("Subtitle() should contain message, got %q", output)
	}
}

func TestIOStreams_Count(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		noun  string
		count int
		want  string
	}{
		{"zero", "device", 0, "Found 0 devices"},
		{"one", "device", 1, "Found 1 device"},
		{"many", "device", 5, "Found 5 devices"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			ios := iostreams.Test(nil, out, nil)

			ios.Count(tt.noun, tt.count)

			if !strings.Contains(out.String(), tt.want) {
				t.Errorf("Count() = %q, want %q", out.String(), tt.want)
			}
		})
	}
}

func TestIOStreams_NoResults(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.NoResults("devices", "Use discover command")

	output := out.String()
	if !strings.Contains(output, "No devices found") {
		t.Errorf("NoResults() should contain 'No X found', got %q", output)
	}
	if !strings.Contains(output, "Use discover command") {
		t.Errorf("NoResults() should contain hint, got %q", output)
	}
}

func TestIOStreams_Added(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		noun  string
		count int
		want  string
	}{
		{"zero", "device", 0, ""},
		{"one", "device", 1, "Added 1 device"},
		{"many", "device", 3, "Added 3 devices"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			ios := iostreams.Test(nil, out, nil)

			ios.Added(tt.noun, tt.count)

			if tt.want == "" {
				if out.Len() != 0 {
					t.Errorf("Added(0) should not output, got %q", out.String())
				}
			} else if !strings.Contains(out.String(), tt.want) {
				t.Errorf("Added() = %q, want %q", out.String(), tt.want)
			}
		})
	}
}

// Test static To functions

func TestInfoTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.InfoTo(buf, "Test %d", 1)

	if !strings.Contains(buf.String(), "Test 1") {
		t.Errorf("InfoTo() should contain message, got %q", buf.String())
	}
}

func TestSuccessTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.SuccessTo(buf, "Done %s", "here")

	if !strings.Contains(buf.String(), "Done here") {
		t.Errorf("SuccessTo() should contain message, got %q", buf.String())
	}
}

func TestWarningTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.WarningTo(buf, "Warn %s", "msg")

	if !strings.Contains(buf.String(), "Warn msg") {
		t.Errorf("WarningTo() should contain message, got %q", buf.String())
	}
}

func TestErrorTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.ErrorTo(buf, "Err %s", "msg")

	if !strings.Contains(buf.String(), "Err msg") {
		t.Errorf("ErrorTo() should contain message, got %q", buf.String())
	}
}

func TestPlainTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.PlainTo(buf, "Plain %s", "text")

	if !strings.Contains(buf.String(), "Plain text") {
		t.Errorf("PlainTo() should contain message, got %q", buf.String())
	}
}

func TestHintTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.HintTo(buf, "Hint %s", "text")

	if !strings.Contains(buf.String(), "Hint text") {
		t.Errorf("HintTo() should contain message, got %q", buf.String())
	}
}

func TestTitleTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.TitleTo(buf, "Title %s", "text")

	if !strings.Contains(buf.String(), "Title text") {
		t.Errorf("TitleTo() should contain message, got %q", buf.String())
	}
}

func TestSubtitleTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.SubtitleTo(buf, "Sub %s", "text")

	if !strings.Contains(buf.String(), "Sub text") {
		t.Errorf("SubtitleTo() should contain message, got %q", buf.String())
	}
}

func TestCountTo(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	iostreams.CountTo(buf, "item", 5)

	if !strings.Contains(buf.String(), "Found 5 items") {
		t.Errorf("CountTo() should contain message, got %q", buf.String())
	}
}

// Package-level convenience function tests
// Note: These functions write to os.Stdout/os.Stderr so we can't easily capture output.
// We test that they don't panic and exercise the code paths.

// Tests for package-level functions that modify global state cannot run in parallel.

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Info(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Info("Test %s", "message")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Info_Quiet(t *testing.T) {
	setupViper(t)
	viper.Set("quiet", true)
	// Should not panic in quiet mode
	iostreams.Info("Test %s", "message")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Success(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Success("Done %s", "successfully")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Warning(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Warning("Caution %s", "required")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Error(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Error("Error %s", "occurred")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Plain(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Plain("Plain %s", "text")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Hint(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Hint("Hint %s", "text")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Hint_Quiet(t *testing.T) {
	setupViper(t)
	viper.Set("quiet", true)
	// Should not panic in quiet mode
	iostreams.Hint("Hint %s", "text")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Title(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Title("Title %s", "text")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Subtitle(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Subtitle("Subtitle %s", "text")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Count(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Count("item", 5)
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_NoResults(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.NoResults("devices", "Use discover command")
}

//nolint:paralleltest // Tests modify global state (viper)
func TestPackageLevel_Added(t *testing.T) {
	setupViper(t)
	// Should not panic
	iostreams.Added("device", 3)
	// Zero count should not panic
	iostreams.Added("device", 0)
}
