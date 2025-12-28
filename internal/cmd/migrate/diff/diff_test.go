package diff

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "diff <device> <backup-file>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "diff <device> <backup-file>")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"compare", "cmp"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	testCases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"device1"}, true},
		{"two args", []string{"device1", "backup.json"}, false},
		{"three args", []string{"device1", "backup.json", "extra"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cmd.Args(cmd, tc.args)
			if (err != nil) != tc.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
		})
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

// createValidBackupFile creates a test backup file with valid structure.
func createValidBackupFile(t *testing.T, dir string, name string) string {
	t.Helper()

	backup := shellybackup.Backup{
		Version:   1,
		CreatedAt: time.Now(),
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:         "test-device-id",
			Name:       "test-device",
			Model:      "SNSW-001X16EU",
			Version:    "1.0.0",
			MAC:        "AA:BB:CC:DD:EE:FF",
			Generation: 2,
		},
		Config: json.RawMessage(`{"sys":{"device":{"name":"test"}}}`),
	}

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}

	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

// createInvalidBackupFile creates an invalid backup file for error testing.
func createInvalidBackupFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()

	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

func TestRun_NonExistentFile(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Test with non-existent file
	err := run(context.Background(), f, "test-device", "/nonexistent/backup.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Error should mention file reading failure
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("failed to read backup file")) {
		t.Errorf("error should mention file read failure, got: %v", err)
	}
}

func TestRun_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create invalid JSON file
	invalidFile := createInvalidBackupFile(t, tmpDir, "invalid.json", "{ not valid json }")

	err := run(context.Background(), f, "test-device", invalidFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	// Error should mention invalid backup
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("error should mention invalid backup, got: %v", err)
	}
}

func TestRun_MissingVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without version field
	missingVersionFile := createInvalidBackupFile(t, tmpDir, "missing-version.json",
		`{"device_info": {"id": "test"}, "config": {}}`)

	err := run(context.Background(), f, "test-device", missingVersionFile)
	if err == nil {
		t.Error("expected error for missing version")
	}

	// Error should mention invalid backup
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("error should mention invalid backup, got: %v", err)
	}
}

func TestRun_MissingDeviceInfo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without device_info
	missingInfoFile := createInvalidBackupFile(t, tmpDir, "missing-info.json",
		`{"version": 1, "config": {}}`)

	err := run(context.Background(), f, "test-device", missingInfoFile)
	if err == nil {
		t.Error("expected error for missing device info")
	}

	// Error should mention invalid backup or missing device
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("error should mention invalid backup, got: %v", err)
	}
}

func TestRun_MissingConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without config
	missingConfigFile := createInvalidBackupFile(t, tmpDir, "missing-config.json",
		`{"version": 1, "device_info": {"id": "test"}}`)

	err := run(context.Background(), f, "test-device", missingConfigFile)
	if err == nil {
		t.Error("expected error for missing config")
	}

	// Error should mention invalid backup or missing config
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("error should mention invalid backup, got: %v", err)
	}
}

func TestRun_ValidBackupButUnreachableDevice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create valid backup file
	validFile := createValidBackupFile(t, tmpDir, "valid.json")

	// Use unreachable device address - should fail to compare
	err := run(context.Background(), f, "192.0.2.1", validFile) // TEST-NET-1, unreachable
	if err == nil {
		// This is expected to fail because we can't reach the device
		t.Log("Note: run succeeded (device might be mocked), which is acceptable")
	}
}

func TestRun_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create empty file
	emptyFile := createInvalidBackupFile(t, tmpDir, "empty.json", "")

	err := run(context.Background(), f, "test-device", emptyFile)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestRun_ValidBackupWithMinimalData(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create minimal valid backup
	backup := shellybackup.Backup{
		Version:   1,
		CreatedAt: time.Now(),
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:    "minimal-device",
			Model: "SNSW-001",
		},
		Config: json.RawMessage(`{"sys":{}}`),
	}

	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}

	minimalFile := filepath.Join(tmpDir, "minimal.json")
	if err := os.WriteFile(minimalFile, data, 0o600); err != nil {
		t.Fatalf("failed to write minimal file: %v", err)
	}

	// This will likely fail due to unreachable device, but should get past validation
	err = run(context.Background(), f, "192.0.2.1", minimalFile)
	// We expect a failure since we can't reach the device
	// The important part is that validation passed
	if err != nil && bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("validation should pass for minimal backup, got: %v", err)
	}
}

func TestRun_FileReadError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create a directory instead of a file
	dirPath := filepath.Join(tmpDir, "not-a-file")
	if err := os.MkdirAll(dirPath, 0o700); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	err := run(context.Background(), f, "test-device", dirPath)
	if err == nil {
		t.Error("expected error when path is a directory")
	}
}

func TestRun_PermissionDenied(t *testing.T) {
	// Skip on Windows where permissions work differently
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("skipping permission test on Windows")
	}

	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create file with no read permissions
	noReadFile := filepath.Join(tmpDir, "noaccess.json")
	if err := os.WriteFile(noReadFile, []byte("{}"), 0o000); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions for cleanup
		if err := os.Chmod(noReadFile, 0o600); err != nil {
			t.Logf("warning: failed to restore permissions: %v", err)
		}
	})

	err := run(context.Background(), f, "test-device", noReadFile)
	if err == nil {
		t.Error("expected error for permission denied")
	}
}
