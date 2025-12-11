package iostreams_test

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestSystem(t *testing.T) {
	t.Parallel()

	ios := iostreams.System()

	if ios == nil {
		t.Fatal("System() returned nil")
	}

	if ios.In == nil {
		t.Error("In is nil")
	}
	if ios.Out == nil {
		t.Error("Out is nil")
	}
	if ios.ErrOut == nil {
		t.Error("ErrOut is nil")
	}
}

func TestTest(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	ios := iostreams.Test(in, out, errOut)

	if ios.In != in {
		t.Error("In not set correctly")
	}
	if ios.Out != out {
		t.Error("Out not set correctly")
	}
	if ios.ErrOut != errOut {
		t.Error("ErrOut not set correctly")
	}

	// Test streams should not be TTY
	if ios.IsStdinTTY() {
		t.Error("Test stdin should not be TTY")
	}
	if ios.IsStdoutTTY() {
		t.Error("Test stdout should not be TTY")
	}
	if ios.IsStderrTTY() {
		t.Error("Test stderr should not be TTY")
	}

	// Color should be disabled in tests
	if ios.ColorEnabled() {
		t.Error("Color should be disabled in tests")
	}
}

func TestIOStreams_TTYSetters(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// Test SetStdinTTY
	ios.SetStdinTTY(true)
	if !ios.IsStdinTTY() {
		t.Error("SetStdinTTY(true) failed")
	}
	ios.SetStdinTTY(false)
	if ios.IsStdinTTY() {
		t.Error("SetStdinTTY(false) failed")
	}

	// Test SetStdoutTTY
	ios.SetStdoutTTY(true)
	if !ios.IsStdoutTTY() {
		t.Error("SetStdoutTTY(true) failed")
	}
	ios.SetStdoutTTY(false)
	if ios.IsStdoutTTY() {
		t.Error("SetStdoutTTY(false) failed")
	}

	// Test SetStderrTTY
	ios.SetStderrTTY(true)
	if !ios.IsStderrTTY() {
		t.Error("SetStderrTTY(true) failed")
	}
	ios.SetStderrTTY(false)
	if ios.IsStderrTTY() {
		t.Error("SetStderrTTY(false) failed")
	}
}

func TestIOStreams_ColorEnabled(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// Initially disabled
	if ios.ColorEnabled() {
		t.Error("Color should be disabled by default in tests")
	}

	// Enable color
	ios.SetColorEnabled(true)
	if !ios.ColorEnabled() {
		t.Error("SetColorEnabled(true) failed")
	}

	// Disable color
	ios.SetColorEnabled(false)
	if ios.ColorEnabled() {
		t.Error("SetColorEnabled(false) failed")
	}
}

func TestIOStreams_Quiet(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// Initially not quiet
	if ios.IsQuiet() {
		t.Error("Should not be quiet by default")
	}

	// Enable quiet
	ios.SetQuiet(true)
	if !ios.IsQuiet() {
		t.Error("SetQuiet(true) failed")
	}

	// Disable quiet
	ios.SetQuiet(false)
	if ios.IsQuiet() {
		t.Error("SetQuiet(false) failed")
	}
}

func TestIOStreams_CanPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stdinTTY  bool
		stdoutTTY bool
		want      bool
	}{
		{
			name:      "both TTY",
			stdinTTY:  true,
			stdoutTTY: true,
			want:      true,
		},
		{
			name:      "stdin not TTY",
			stdinTTY:  false,
			stdoutTTY: true,
			want:      false,
		},
		{
			name:      "stdout not TTY",
			stdinTTY:  true,
			stdoutTTY: false,
			want:      false,
		},
		{
			name:      "neither TTY",
			stdinTTY:  false,
			stdoutTTY: false,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ios := iostreams.Test(nil, nil, nil)
			ios.SetStdinTTY(tt.stdinTTY)
			ios.SetStdoutTTY(tt.stdoutTTY)

			if got := ios.CanPrompt(); got != tt.want {
				t.Errorf("CanPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIOStreams_Printf(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Printf("Hello %s", "World")

	if got := out.String(); got != "Hello World" {
		t.Errorf("Printf() output = %q, want %q", got, "Hello World")
	}
}

func TestIOStreams_Println(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Println("Hello", "World")

	if got := out.String(); got != "Hello World\n" {
		t.Errorf("Println() output = %q, want %q", got, "Hello World\n")
	}
}

func TestIOStreams_Errorf(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	ios.Errorf("Error: %s", "test")

	if got := errOut.String(); got != "Error: test" {
		t.Errorf("Errorf() output = %q, want %q", got, "Error: test")
	}
}

func TestIOStreams_Errorln(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	ios.Errorln("Error", "message")

	if got := errOut.String(); got != "Error message\n" {
		t.Errorf("Errorln() output = %q, want %q", got, "Error message\n")
	}
}

func TestIOStreams_StartProgress_NonTTY(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)
	ios.SetStderrTTY(false)

	ios.StartProgress("Loading...")

	// For non-TTY, should print message directly
	if got := errOut.String(); got != "Loading...\n" {
		t.Errorf("StartProgress() output = %q, want %q", got, "Loading...\n")
	}

	// Stop should be safe to call
	ios.StopProgress()
}

func TestIOStreams_StartProgress_Quiet(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)
	ios.SetQuiet(true)

	ios.StartProgress("Loading...")

	// In quiet mode, should not print anything
	if got := errOut.String(); got != "" {
		t.Errorf("StartProgress() in quiet mode should not output, got %q", got)
	}
}

func TestIOStreams_Progress_TTY(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)
	ios.SetStderrTTY(true)

	// Start progress - this starts a spinner goroutine
	ios.StartProgress("Loading...")

	// Update progress
	ios.UpdateProgress("Still loading...")

	// Stop with success - verifies no panic and spinner state handling
	ios.StopProgressWithSuccess("Done!")

	// For TTY mode with real spinner, output depends on timing
	// The important thing is these calls don't panic
}

func TestIOStreams_StopProgressWithError(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)
	ios.SetStderrTTY(true)

	ios.StartProgress("Loading...")
	ios.StopProgressWithError("Failed!")

	// For TTY mode with real spinner, output depends on timing
	// The important thing is these calls don't panic
}

func TestIOStreams_StopProgress_NilIndicator(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// These should not panic when called without StartProgress
	ios.StopProgress()
	ios.StopProgressWithSuccess("Done")
	ios.StopProgressWithError("Failed")
	ios.UpdateProgress("Update")
}
