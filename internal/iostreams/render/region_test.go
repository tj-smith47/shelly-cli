package render_test

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/tj-smith47/shelly-cli/internal/iostreams/render"
)

func TestSpinnerCharCycles(t *testing.T) {
	t.Parallel()

	// Verify spinner cycles through MiniDot frames
	seen := make(map[string]bool)
	for i := range 20 {
		r := &render.LineRegion{
			ID:           "test",
			Status:       render.StatusRunning,
			Message:      "working",
			SpinnerFrame: i,
		}
		line := render.FormatLine(r, render.NoColor())
		// Extract first character (spinner)
		firstRune, _ := utf8.DecodeRuneInString(line)
		char := string(firstRune)
		seen[char] = true
	}

	// MiniDot has 10 frames — we should see multiple unique characters
	if len(seen) < 5 {
		t.Errorf("expected at least 5 unique spinner chars, got %d: %v", len(seen), seen)
	}
}

func TestFormatLine_Pending(t *testing.T) {
	t.Parallel()

	r := &render.LineRegion{
		ID:      "device1",
		Status:  render.StatusPending,
		Message: "pending",
	}
	line := render.FormatLine(r, render.NoColor())

	if !strings.Contains(line, "○") {
		t.Errorf("pending line should contain ○, got %q", line)
	}
	if !strings.Contains(line, "device1") {
		t.Errorf("pending line should contain ID, got %q", line)
	}
}

func TestFormatLine_Running(t *testing.T) {
	t.Parallel()

	r := &render.LineRegion{
		ID:           "device1",
		Status:       render.StatusRunning,
		Message:      "working...",
		StartTime:    time.Now().Add(-3 * time.Second),
		SpinnerFrame: 0,
	}
	line := render.FormatLine(r, render.NoColor())

	if !strings.Contains(line, "device1") {
		t.Errorf("running line should contain ID, got %q", line)
	}
	if !strings.Contains(line, "working...") {
		t.Errorf("running line should contain message, got %q", line)
	}
	// Should contain elapsed time (e.g., "3.0s")
	if !strings.Contains(line, "s") {
		t.Errorf("running line should contain elapsed time, got %q", line)
	}
}

func TestFormatLine_Success(t *testing.T) {
	t.Parallel()

	r := &render.LineRegion{
		ID:       "device1",
		Status:   render.StatusSuccess,
		Message:  "done",
		Duration: "2s",
	}
	line := render.FormatLine(r, render.NoColor())

	if !strings.Contains(line, "\u2714") {
		t.Errorf("success line should contain ✔, got %q", line)
	}
	if !strings.Contains(line, "(2s)") {
		t.Errorf("success line should contain duration, got %q", line)
	}
}

func TestFormatLine_Error(t *testing.T) {
	t.Parallel()

	r := &render.LineRegion{
		ID:       "device1",
		Status:   render.StatusError,
		Message:  "connection refused",
		Duration: "5s",
	}
	line := render.FormatLine(r, render.NoColor())

	if !strings.Contains(line, "\u2718") {
		t.Errorf("error line should contain ✘, got %q", line)
	}
	if !strings.Contains(line, "connection refused") {
		t.Errorf("error line should contain error message, got %q", line)
	}
	if !strings.Contains(line, "(5s)") {
		t.Errorf("error line should contain duration, got %q", line)
	}
}

func TestFormatLine_Skipped(t *testing.T) {
	t.Parallel()

	r := &render.LineRegion{
		ID:      "device1",
		Status:  render.StatusSkipped,
		Message: "skipped",
	}
	line := render.FormatLine(r, render.NoColor())

	if !strings.Contains(line, "\u2298") {
		t.Errorf("skipped line should contain ⊘, got %q", line)
	}
}

func TestFormatElapsed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0s"},
		{5 * time.Second, "5s"},
		{59 * time.Second, "59s"},
		{60 * time.Second, "1m0s"},
		{90 * time.Second, "1m30s"},
		{125 * time.Second, "2m5s"},
	}

	for _, tt := range tests {
		if got := render.FormatElapsed(tt.d); got != tt.want {
			t.Errorf("FormatElapsed(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestNoColor(t *testing.T) {
	t.Parallel()

	cf := render.NoColor()
	input := "test"

	if cf.Primary(input) != input {
		t.Error("NoColor Primary should be identity")
	}
	if cf.Secondary(input) != input {
		t.Error("NoColor Secondary should be identity")
	}
	if cf.Error(input) != input {
		t.Error("NoColor Error should be identity")
	}
	if cf.Green(input) != input {
		t.Error("NoColor Green should be identity")
	}
}
