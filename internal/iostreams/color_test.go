package iostreams_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

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

func TestIOStreams_Success_Quiet(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetQuiet(true)

	ios.Success("Operation completed")

	if out.Len() != 0 {
		t.Errorf("Success() in quiet mode should not output, got %q", out.String())
	}
}

func TestIOStreams_Plain_Quiet(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetQuiet(true)

	ios.Plain("Plain text")

	if out.Len() != 0 {
		t.Errorf("Plain() in quiet mode should not output, got %q", out.String())
	}
}

func TestIOStreams_Title_Quiet(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetQuiet(true)

	ios.Title("Section Title")

	if out.Len() != 0 {
		t.Errorf("Title() in quiet mode should not output, got %q", out.String())
	}
}

func TestIOStreams_Subtitle_Quiet(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetQuiet(true)

	ios.Subtitle("Subsection")

	if out.Len() != 0 {
		t.Errorf("Subtitle() in quiet mode should not output, got %q", out.String())
	}
}

func TestIOStreams_Count_Quiet(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetQuiet(true)

	ios.Count("device", 5)

	if out.Len() != 0 {
		t.Errorf("Count() in quiet mode should not output, got %q", out.String())
	}
}

func TestIOStreams_NoResults_Quiet(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetQuiet(true)

	ios.NoResults("devices", "Use discover command")

	if out.Len() != 0 {
		t.Errorf("NoResults() in quiet mode should not output, got %q", out.String())
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
