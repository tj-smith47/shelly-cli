package validate

import (
	"bytes"
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

	if cmd.Use != "validate <backup-file>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "validate <backup-file>")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"check", "verify"}
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
		{"one arg", []string{"backup.json"}, false},
		{"two args", []string{"backup.json", "extra"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
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
func createValidBackupFile(t *testing.T, dir, name string) string {
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

// createBackupWithScripts creates a backup file with scripts.
func createBackupWithScripts(t *testing.T, dir, name string) string {
	t.Helper()

	backup := shellybackup.Backup{
		Version:   1,
		CreatedAt: time.Now(),
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:         "script-device",
			Name:       "script-device",
			Model:      "SNSW-001X16EU",
			Version:    "1.0.0",
			Generation: 2,
		},
		Config: json.RawMessage(`{"sys":{}}`),
		Scripts: []*shellybackup.Script{
			{ID: 1, Name: "test-script", Enable: true, Code: "// test"},
			{ID: 2, Name: "another-script", Enable: false, Code: "// another"},
		},
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
func createInvalidBackupFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

func TestRun_ValidBackup(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create valid backup file
	validFile := createValidBackupFile(t, tmpDir, "valid.json")

	err := run(f, validFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String() + stderr.String()

	// Should mention valid
	if !bytes.Contains([]byte(output), []byte("valid")) {
		t.Errorf("output should mention 'valid', got: %s", output)
	}

	// Should show device ID
	if !bytes.Contains([]byte(output), []byte("test-device-id")) {
		t.Errorf("output should contain device ID, got: %s", output)
	}

	// Should show model
	if !bytes.Contains([]byte(output), []byte("SNSW-001X16EU")) {
		t.Errorf("output should contain model, got: %s", output)
	}
}

func TestRun_ValidBackupWithScripts(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create backup with scripts
	scriptsFile := createBackupWithScripts(t, tmpDir, "scripts.json")

	err := run(f, scriptsFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String() + stderr.String()

	// Should show script count
	if !bytes.Contains([]byte(output), []byte("2")) {
		t.Errorf("output should contain script count '2', got: %s", output)
	}
}

func TestRun_NonExistentFile(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f, "/nonexistent/backup.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Error should mention file reading failure
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("failed to read file")) {
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

	err := run(f, invalidFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
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

	err := run(f, missingVersionFile)
	if err == nil {
		t.Error("expected error for missing version")
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

	err := run(f, missingInfoFile)
	if err == nil {
		t.Error("expected error for missing device info")
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

	err := run(f, missingConfigFile)
	if err == nil {
		t.Error("expected error for missing config")
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

	err := run(f, emptyFile)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestRun_FileIsDirectory(t *testing.T) {
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

	err := run(f, dirPath)
	if err == nil {
		t.Error("expected error when path is a directory")
	}
}

func TestRun_OutputShowsConfigKeys(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create backup with config
	backup := shellybackup.Backup{
		Version:   1,
		CreatedAt: time.Now(),
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:    "config-device",
			Model: "SNSW-001",
		},
		Config: json.RawMessage(`{"sys":{},"switch":{},"wifi":{}}`),
	}

	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}

	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	err = run(f, configFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should show config key count (3)
	if !bytes.Contains([]byte(output), []byte("3")) {
		t.Errorf("output should contain config key count '3', got: %s", output)
	}
}

func TestRun_OutputShowsVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create valid backup
	backup := shellybackup.Backup{
		Version:   1,
		CreatedAt: time.Now(),
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:    "version-device",
			Model: "SNSW-001",
		},
		Config: json.RawMessage(`{"sys":{}}`),
	}

	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}

	versionFile := filepath.Join(tmpDir, "version.json")
	if err := os.WriteFile(versionFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	err = run(f, versionFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should show version info
	if !bytes.Contains([]byte(output), []byte("Version")) {
		t.Errorf("output should contain 'Version', got: %s", output)
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

	err := run(f, noReadFile)
	if err == nil {
		t.Error("expected error for permission denied")
	}
}

func TestRun_OutputShowsFirmware(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create backup with firmware version
	backup := shellybackup.Backup{
		Version:   1,
		CreatedAt: time.Now(),
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:      "fw-device",
			Model:   "SNSW-001",
			Version: "2.5.0",
		},
		Config: json.RawMessage(`{"sys":{}}`),
	}

	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}

	fwFile := filepath.Join(tmpDir, "firmware.json")
	if err := os.WriteFile(fwFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	err = run(f, fwFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should show firmware version
	if !bytes.Contains([]byte(output), []byte("2.5.0")) {
		t.Errorf("output should contain firmware version '2.5.0', got: %s", output)
	}
}
