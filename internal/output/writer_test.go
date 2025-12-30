package output

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestWriteFile(t *testing.T) {
	t.Parallel()

	t.Run("write to new file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "test.txt")
		data := []byte("hello world")

		if err := WriteFile(file, data); err != nil {
			t.Fatalf("WriteFile() error: %v", err)
		}

		got, err := os.ReadFile(file) //nolint:gosec // test uses t.TempDir()
		if err != nil {
			t.Fatalf("ReadFile() error: %v", err)
		}
		if !bytes.Equal(got, data) {
			t.Errorf("got %q, want %q", got, data)
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "a", "b", "c", "test.txt")
		data := []byte("nested")

		if err := WriteFile(file, data); err != nil {
			t.Fatalf("WriteFile() error: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Error("file should exist")
		}
	})

	t.Run("file has secure permissions", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "secret.txt")
		data := []byte("secret data")

		if err := WriteFile(file, data); err != nil {
			t.Fatalf("WriteFile() error: %v", err)
		}

		info, err := os.Stat(file)
		if err != nil {
			t.Fatalf("Stat() error: %v", err)
		}

		perm := info.Mode().Perm()
		// Should be 0600 (owner read/write only)
		if perm != 0o600 {
			t.Errorf("file permissions = %o, want 0600", perm)
		}
	})
}

func TestWriteFile_CurrentDir(t *testing.T) {
	t.Parallel()

	// Save current dir and change to temp
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("warning: failed to restore working directory: %v", err)
		}
	}()

	// Write to current directory (no parent path)
	file := "local.txt"
	data := []byte("local content")

	if err := WriteFile(file, data); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestGetWriter(t *testing.T) {
	t.Parallel()

	t.Run("empty path returns stdout", func(t *testing.T) {
		t.Parallel()
		ios := newTestIOStreams()
		writer, closer, err := GetWriter(ios, "")
		if err != nil {
			t.Fatalf("GetWriter() error: %v", err)
		}
		if writer == nil {
			t.Fatal("expected non-nil writer")
		}
		if closer == nil {
			t.Fatal("expected non-nil closer")
		}
		// Call closer (no-op for stdout)
		closer()
	})

	t.Run("file path creates file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "output.txt")

		ios := newTestIOStreams()
		writer, closer, err := GetWriter(ios, file)
		if err != nil {
			t.Fatalf("GetWriter() error: %v", err)
		}
		defer closer()

		if writer == nil {
			t.Fatal("expected non-nil writer")
		}

		// Write something
		_, err = writer.Write([]byte("test content"))
		if err != nil {
			t.Fatalf("Write() error: %v", err)
		}
	})
}

func TestExportToFile(t *testing.T) {
	t.Parallel()

	t.Run("export to stdout", func(t *testing.T) {
		t.Parallel()
		ios := newTestIOStreams()
		data := map[string]string{"key": "value"}
		err := ExportToFile(ios, data, "", FormatJSON, "JSON")
		if err != nil {
			t.Fatalf("ExportToFile() error: %v", err)
		}
	})

	t.Run("export to file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "export.json")

		ios := newTestIOStreams()
		data := map[string]string{"key": "value"}
		err := ExportToFile(ios, data, file, FormatJSON, "JSON")
		if err != nil {
			t.Fatalf("ExportToFile() error: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Error("file should exist")
		}
	})
}

func TestExportCSV(t *testing.T) {
	t.Parallel()

	t.Run("export to stdout", func(t *testing.T) {
		t.Parallel()
		ios := newTestIOStreams()
		formatter := func() ([]byte, error) {
			return []byte("col1,col2\nval1,val2\n"), nil
		}
		err := ExportCSV(ios, "", formatter)
		if err != nil {
			t.Fatalf("ExportCSV() error: %v", err)
		}
	})

	t.Run("export to file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		file := filepath.Join(tmpDir, "export.csv")

		ios := newTestIOStreams()
		formatter := func() ([]byte, error) {
			return []byte("col1,col2\nval1,val2\n"), nil
		}
		err := ExportCSV(ios, file, formatter)
		if err != nil {
			t.Fatalf("ExportCSV() error: %v", err)
		}

		// Verify file was created
		content, err := os.ReadFile(file) //nolint:gosec // test uses t.TempDir()
		if err != nil {
			t.Fatalf("ReadFile() error: %v", err)
		}
		if string(content) != "col1,col2\nval1,val2\n" {
			t.Errorf("unexpected content: %q", content)
		}
	})

	t.Run("formatter error", func(t *testing.T) {
		t.Parallel()
		ios := newTestIOStreams()
		formatter := func() ([]byte, error) {
			return nil, os.ErrInvalid
		}
		err := ExportCSV(ios, "", formatter)
		if err == nil {
			t.Error("expected error from formatter")
		}
	})
}

// newTestIOStreams creates a test IOStreams with buffers.
func newTestIOStreams() *iostreams.IOStreams {
	var stdin, stdout, stderr bytes.Buffer
	return iostreams.Test(&stdin, &stdout, &stderr)
}

func TestExportToFile_Errors(t *testing.T) {
	t.Parallel()

	t.Run("marshal error returns error", func(t *testing.T) {
		t.Parallel()
		ios := newTestIOStreams()
		// Use a channel which can't be marshaled
		data := make(chan int)
		err := ExportToFile(ios, data, "", FormatJSON, "JSON")
		if err == nil {
			t.Error("expected error for unmarshalable type")
		}
	})
}
