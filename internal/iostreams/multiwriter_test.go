package iostreams_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status iostreams.Status
		want   string
	}{
		{iostreams.StatusPending, "pending"},
		{iostreams.StatusRunning, "running"},
		{iostreams.StatusSuccess, "success"},
		{iostreams.StatusError, "error"},
		{iostreams.StatusSkipped, "skipped"},
		{iostreams.Status(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewMultiWriter(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, false)

	if mw == nil {
		t.Fatal("NewMultiWriter() returned nil")
	}
	if mw.LineCount() != 0 {
		t.Errorf("LineCount() = %d, want 0", mw.LineCount())
	}
}

func TestMultiWriter_AddLine(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, false) // non-TTY

	mw.AddLine("device1", "pending")
	mw.AddLine("device2", "pending")

	if mw.LineCount() != 2 {
		t.Errorf("LineCount() = %d, want 2", mw.LineCount())
	}

	// Check we can get lines
	line, ok := mw.GetLine("device1")
	if !ok {
		t.Error("GetLine(device1) should return true")
	}
	if line.ID != "device1" {
		t.Errorf("line.ID = %q, want %q", line.ID, "device1")
	}
	if line.Status != iostreams.StatusPending {
		t.Errorf("line.Status = %v, want %v", line.Status, iostreams.StatusPending)
	}
}

func TestMultiWriter_AddLine_TTY(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, true) // TTY mode

	mw.AddLine("device1", "pending")

	// Finalize stops the ticker goroutine so we can safely read the buffer.
	mw.Finalize()

	output := buf.String()
	if !strings.Contains(output, "device1") {
		t.Errorf("TTY AddLine should print immediately, got %q", output)
	}
}

func TestMultiWriter_UpdateLine(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, false)

	mw.AddLine("device1", "pending")
	mw.UpdateLine("device1", iostreams.StatusRunning, "working...")

	line, ok := mw.GetLine("device1")
	if !ok {
		t.Fatal("GetLine should return true")
	}
	if line.Status != iostreams.StatusRunning {
		t.Errorf("line.Status = %v, want %v", line.Status, iostreams.StatusRunning)
	}
	if line.Message != "working..." {
		t.Errorf("line.Message = %q, want %q", line.Message, "working...")
	}
}

func TestMultiWriter_UpdateLine_NonExistent(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, false)

	// Should not panic
	mw.UpdateLine("nonexistent", iostreams.StatusSuccess, "done")
}

func TestMultiWriter_GetLine_NonExistent(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, false)

	_, ok := mw.GetLine("nonexistent")
	if ok {
		t.Error("GetLine should return false for nonexistent line")
	}
}

func TestMultiWriter_Finalize_NonTTY(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, false) // non-TTY

	mw.AddLine("device1", "pending")
	mw.UpdateLine("device1", iostreams.StatusSuccess, "done")
	mw.AddLine("device2", "pending")
	mw.UpdateLine("device2", iostreams.StatusError, "failed")

	// Non-TTY prints on status transitions, so output accumulates during UpdateLine calls.
	// Finalize ensures any remaining unprinted lines are flushed.
	mw.Finalize()

	output := buf.String()
	if !strings.Contains(output, "device1") {
		t.Errorf("output should contain device1, got %q", output)
	}
	if !strings.Contains(output, "device2") {
		t.Errorf("output should contain device2, got %q", output)
	}
}

func TestMultiWriter_Finalize_TTY(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, true) // TTY

	mw.AddLine("device1", "pending")
	mw.UpdateLine("device1", iostreams.StatusSuccess, "done")

	mw.Finalize()

	// Finalize does a final render and emits ShowCursor. Verify it contains the
	// cursor-show escape sequence and that double-Finalize is safe.
	output := buf.String()
	if !strings.Contains(output, "\033[?25h") {
		t.Errorf("TTY Finalize should emit ShowCursor, got %q", output)
	}

	// Second Finalize should be a no-op (idempotent)
	lenAfterFirst := buf.Len()
	mw.Finalize()
	if buf.Len() != lenAfterFirst {
		t.Error("double Finalize should not add more output")
	}
}

func TestMultiWriter_Summary(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, false)

	mw.AddLine("d1", "pending")
	mw.UpdateLine("d1", iostreams.StatusSuccess, "done")
	mw.AddLine("d2", "pending")
	mw.UpdateLine("d2", iostreams.StatusSuccess, "done")
	mw.AddLine("d3", "pending")
	mw.UpdateLine("d3", iostreams.StatusError, "failed")
	mw.AddLine("d4", "pending")
	mw.UpdateLine("d4", iostreams.StatusSkipped, "skipped")
	mw.AddLine("d5", "pending")
	// d5 stays pending

	success, failed, skipped := mw.Summary()

	if success != 2 {
		t.Errorf("success = %d, want 2", success)
	}
	if failed != 1 {
		t.Errorf("failed = %d, want 1", failed)
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}
}

func TestMultiWriter_PrintSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(mw *iostreams.MultiWriter)
		contains []string
	}{
		{
			name: "all success",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				mw.UpdateLine("d1", iostreams.StatusSuccess, "done")
			},
			contains: []string{"1 succeeded"},
		},
		{
			name: "mixed results",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				mw.UpdateLine("d1", iostreams.StatusSuccess, "done")
				mw.AddLine("d2", "p")
				mw.UpdateLine("d2", iostreams.StatusError, "failed")
			},
			contains: []string{"1 succeeded", "1 failed"},
		},
		{
			name: "with skipped",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				mw.UpdateLine("d1", iostreams.StatusSkipped, "skipped")
			},
			contains: []string{"1 skipped"},
		},
		{
			name: "empty",
			setup: func(mw *iostreams.MultiWriter) {
				// No lines added
			},
			contains: []string{"Completed 0 operations"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			mw := iostreams.NewMultiWriter(buf, false)
			tt.setup(mw)

			buf.Reset()
			mw.PrintSummary()

			output := buf.String()
			for _, c := range tt.contains {
				if !strings.Contains(output, c) {
					t.Errorf("PrintSummary() should contain %q, got %q", c, output)
				}
			}
		})
	}
}

func TestMultiWriter_HasErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(mw *iostreams.MultiWriter)
		want  bool
	}{
		{
			name:  "empty",
			setup: func(mw *iostreams.MultiWriter) {},
			want:  false,
		},
		{
			name: "all success",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				mw.UpdateLine("d1", iostreams.StatusSuccess, "done")
			},
			want: false,
		},
		{
			name: "has error",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				mw.UpdateLine("d1", iostreams.StatusError, "failed")
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mw := iostreams.NewMultiWriter(&bytes.Buffer{}, false)
			tt.setup(mw)

			if got := mw.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiWriter_AllSucceeded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(mw *iostreams.MultiWriter)
		want  bool
	}{
		{
			name:  "empty",
			setup: func(mw *iostreams.MultiWriter) {},
			want:  true,
		},
		{
			name: "all success",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				mw.UpdateLine("d1", iostreams.StatusSuccess, "done")
			},
			want: true,
		},
		{
			name: "success and skipped",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				mw.UpdateLine("d1", iostreams.StatusSuccess, "done")
				mw.AddLine("d2", "p")
				mw.UpdateLine("d2", iostreams.StatusSkipped, "skip")
			},
			want: true,
		},
		{
			name: "has error",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				mw.UpdateLine("d1", iostreams.StatusError, "failed")
			},
			want: false,
		},
		{
			name: "has pending",
			setup: func(mw *iostreams.MultiWriter) {
				mw.AddLine("d1", "p")
				// stays pending
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mw := iostreams.NewMultiWriter(&bytes.Buffer{}, false)
			tt.setup(mw)

			if got := mw.AllSucceeded(); got != tt.want {
				t.Errorf("AllSucceeded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiWriter_Render_TTY(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := iostreams.NewMultiWriter(buf, true) // TTY mode

	mw.AddLine("device1", "pending")
	mw.AddLine("device2", "pending")

	// Update triggers render
	mw.UpdateLine("device1", iostreams.StatusRunning, "working")

	// Finalize stops the ticker goroutine so we can safely read the buffer.
	mw.Finalize()

	output := buf.String()
	// Should contain ANSI escape codes for cursor movement
	if !strings.Contains(output, "\033[") {
		t.Errorf("TTY render should contain ANSI escape codes, got %q", output)
	}
}
