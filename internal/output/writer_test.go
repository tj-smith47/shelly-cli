package output

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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
