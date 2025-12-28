package cmdutil_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

const testLine3 = "line3"

func TestGetLogPath(t *testing.T) {
	t.Parallel()

	path, err := cmdutil.GetLogPath()

	if err != nil {
		t.Fatalf("GetLogPath() error = %v", err)
	}
	if path == "" {
		t.Error("GetLogPath() returned empty path")
	}
	if !strings.HasSuffix(path, "shelly.log") {
		t.Errorf("GetLogPath() = %q, want suffix %q", path, "shelly.log")
	}
}

//nolint:gocyclo // test function with many sub-tests
func TestReadLastLines(t *testing.T) {
	t.Parallel()

	t.Run("read all lines when less than N", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "test.log")
		content := "line1\nline2\nline3\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		lines, err := cmdutil.ReadLastLines(path, 10)
		if err != nil {
			t.Fatalf("ReadLastLines() error = %v", err)
		}

		if len(lines) != 3 {
			t.Errorf("got %d lines, want 3", len(lines))
		}
		if lines[0] != "line1" || lines[1] != "line2" || lines[2] != testLine3 {
			t.Errorf("lines = %v, want [line1 line2 %s]", lines, testLine3)
		}
	})

	t.Run("read last N lines", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "test.log")
		content := "line1\nline2\nline3\nline4\nline5\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		lines, err := cmdutil.ReadLastLines(path, 3)
		if err != nil {
			t.Fatalf("ReadLastLines() error = %v", err)
		}

		if len(lines) != 3 {
			t.Errorf("got %d lines, want 3", len(lines))
		}
		if lines[0] != testLine3 || lines[1] != "line4" || lines[2] != "line5" {
			t.Errorf("lines = %v, want [%s line4 line5]", lines, testLine3)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "test.log")
		if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		lines, err := cmdutil.ReadLastLines(path, 10)
		if err != nil {
			t.Fatalf("ReadLastLines() error = %v", err)
		}

		if len(lines) != 0 {
			t.Errorf("got %d lines, want 0", len(lines))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		_, err := cmdutil.ReadLastLines(filepath.Join(dir, "nonexistent.log"), 10)
		if err == nil {
			t.Error("ReadLastLines() should error for nonexistent file")
		}
	})

	t.Run("file without trailing newline", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "test.log")
		content := "line1\nline2\nline3" // No trailing newline
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		lines, err := cmdutil.ReadLastLines(path, 10)
		if err != nil {
			t.Fatalf("ReadLastLines() error = %v", err)
		}

		if len(lines) != 3 {
			t.Errorf("got %d lines, want 3", len(lines))
		}
		if lines[2] != testLine3 {
			t.Errorf("last line = %q, want %q", lines[2], testLine3)
		}
	})

	t.Run("exact number of lines", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "test.log")
		content := "a\nb\nc\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		lines, err := cmdutil.ReadLastLines(path, 3)
		if err != nil {
			t.Fatalf("ReadLastLines() error = %v", err)
		}

		if len(lines) != 3 {
			t.Errorf("got %d lines, want 3", len(lines))
		}
	})

	t.Run("request more lines than exist", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "test.log")
		content := "only\ntwo\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		lines, err := cmdutil.ReadLastLines(path, 100)
		if err != nil {
			t.Fatalf("ReadLastLines() error = %v", err)
		}

		if len(lines) != 2 {
			t.Errorf("got %d lines, want 2", len(lines))
		}
	})
}
