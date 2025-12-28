package output

import (
	"strings"
	"testing"
)

func TestFormatREPLPrompt(t *testing.T) {
	t.Parallel()

	t.Run("no active device", func(t *testing.T) {
		t.Parallel()
		got := FormatREPLPrompt("")
		if got != "shelly> " {
			t.Errorf("FormatREPLPrompt(\"\") = %q, want %q", got, "shelly> ")
		}
	})

	t.Run("with active device", func(t *testing.T) {
		t.Parallel()
		got := FormatREPLPrompt("kitchen")
		// Should contain device name and brackets (possibly with ANSI codes)
		if !strings.Contains(got, "shelly") {
			t.Error("expected prompt to contain 'shelly'")
		}
		if !strings.Contains(got, "kitchen") {
			t.Error("expected prompt to contain 'kitchen'")
		}
	})
}

func TestFormatShellPrompt(t *testing.T) {
	t.Parallel()

	got := FormatShellPrompt("device1")
	// Should contain device name and end with "> " (possibly with ANSI codes)
	if !strings.Contains(got, "device1") {
		t.Error("expected prompt to contain device name")
	}
	if !strings.Contains(got, ">") {
		t.Error("expected prompt to contain '>'")
	}
}
