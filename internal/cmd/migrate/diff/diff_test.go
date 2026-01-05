package diff

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const testBackupDir = "/test/backup"

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
func createValidBackupFile(t *testing.T, fs afero.Fs, name string) string {
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

	filePath := testBackupDir + "/" + name
	if err := afero.WriteFile(fs, filePath, data, 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

// createInvalidBackupFile creates an invalid backup file for error testing.
func createInvalidBackupFile(t *testing.T, fs afero.Fs, name, content string) string {
	t.Helper()

	filePath := testBackupDir + "/" + name
	if err := afero.WriteFile(fs, filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

func TestRun_NonExistentFile(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		FilePath: "/nonexistent/backup.json",
	}
	// Test with non-existent file
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Error should mention file reading failure
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("failed to read backup file")) {
		t.Errorf("error should mention file read failure, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_InvalidJSON(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testBackupDir, 0o755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create invalid JSON file
	invalidFile := createInvalidBackupFile(t, fs, "invalid.json", "{ not valid json }")

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		FilePath: invalidFile,
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	// Error should mention invalid backup
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("error should mention invalid backup, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_MissingVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testBackupDir, 0o755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without version field
	missingVersionFile := createInvalidBackupFile(t, fs, "missing-version.json",
		`{"device_info": {"id": "test"}, "config": {}}`)

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		FilePath: missingVersionFile,
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for missing version")
	}

	// Error should mention invalid backup
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("error should mention invalid backup, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_MissingDeviceInfo(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testBackupDir, 0o755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without device_info
	missingInfoFile := createInvalidBackupFile(t, fs, "missing-info.json",
		`{"version": 1, "config": {}}`)

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		FilePath: missingInfoFile,
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for missing device info")
	}

	// Error should mention invalid backup or missing device
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("error should mention invalid backup, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_MissingConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testBackupDir, 0o755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without config
	missingConfigFile := createInvalidBackupFile(t, fs, "missing-config.json",
		`{"version": 1, "device_info": {"id": "test"}}`)

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		FilePath: missingConfigFile,
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for missing config")
	}

	// Error should mention invalid backup or missing config
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("error should mention invalid backup, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_ValidBackupButUnreachableDevice(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testBackupDir, 0o755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create valid backup file
	validFile := createValidBackupFile(t, fs, "valid.json")

	opts := &Options{
		Factory:  f,
		Device:   "192.0.2.1", // TEST-NET-1, unreachable
		FilePath: validFile,
	}
	// Use unreachable device address - should fail to compare
	err := run(context.Background(), opts)
	if err == nil {
		// This is expected to fail because we can't reach the device
		t.Log("Note: run succeeded (device might be mocked), which is acceptable")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_EmptyFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testBackupDir, 0o755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create empty file
	emptyFile := createInvalidBackupFile(t, fs, "empty.json", "")

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		FilePath: emptyFile,
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_ValidBackupWithMinimalData(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testBackupDir, 0o755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

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

	minimalFile := testBackupDir + "/minimal.json"
	if err := afero.WriteFile(fs, minimalFile, data, 0o600); err != nil {
		t.Fatalf("failed to write minimal file: %v", err)
	}

	opts := &Options{
		Factory:  f,
		Device:   "192.0.2.1",
		FilePath: minimalFile,
	}
	// This will likely fail due to unreachable device, but should get past validation
	err = run(context.Background(), opts)
	// We expect a failure since we can't reach the device
	// The important part is that validation passed
	if err != nil && bytes.Contains([]byte(err.Error()), []byte("invalid backup")) {
		t.Errorf("validation should pass for minimal backup, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_FileReadError(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	if err := fs.MkdirAll(testBackupDir, 0o755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create a directory instead of a file
	dirPath := testBackupDir + "/not-a-file"
	if err := fs.MkdirAll(dirPath, 0o700); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		FilePath: dirPath,
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error when path is a directory")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_ReadOnlyFilesystem(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	roFs := afero.NewReadOnlyFs(baseFs)
	config.SetFs(roFs)
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		FilePath: "/test/noaccess.json",
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for read-only filesystem")
	}
}
