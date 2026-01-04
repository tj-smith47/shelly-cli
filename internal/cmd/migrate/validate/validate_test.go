package validate

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const (
	testBackupDir = "/test/backups"
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

	filePath := dir + "/" + name
	if err := afero.WriteFile(config.Fs(), filePath, data, 0o600); err != nil {
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

	filePath := dir + "/" + name
	if err := afero.WriteFile(config.Fs(), filePath, data, 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

// createInvalidBackupFile creates an invalid backup file for error testing.
func createInvalidBackupFile(t *testing.T, name, content string) string {
	t.Helper()

	filePath := testBackupDir + "/" + name
	if err := afero.WriteFile(config.Fs(), filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_ValidBackup(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create valid backup file
	validFile := createValidBackupFile(t, testBackupDir, "valid.json")

	opts := &Options{Factory: f, FilePath: validFile}
	err := run(opts)
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_ValidBackupWithScripts(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create backup with scripts
	scriptsFile := createBackupWithScripts(t, testBackupDir, "scripts.json")

	opts := &Options{Factory: f, FilePath: scriptsFile}
	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String() + stderr.String()

	// Should show script count
	if !bytes.Contains([]byte(output), []byte("2")) {
		t.Errorf("output should contain script count '2', got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_NonExistentFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, FilePath: "/nonexistent/backup.json"}
	err := run(opts)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Error should mention file reading failure
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("failed to read file")) {
		t.Errorf("error should mention file read failure, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_InvalidJSON(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create invalid JSON file
	invalidFile := createInvalidBackupFile(t, "invalid.json", "{ not valid json }")

	opts := &Options{Factory: f, FilePath: invalidFile}
	err := run(opts)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_MissingVersion(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without version field
	missingVersionFile := createInvalidBackupFile(t, "missing-version.json",
		`{"device_info": {"id": "test"}, "config": {}}`)

	opts := &Options{Factory: f, FilePath: missingVersionFile}
	err := run(opts)
	if err == nil {
		t.Error("expected error for missing version")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_MissingDeviceInfo(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without device_info
	missingInfoFile := createInvalidBackupFile(t, "missing-info.json",
		`{"version": 1, "config": {}}`)

	opts := &Options{Factory: f, FilePath: missingInfoFile}
	err := run(opts)
	if err == nil {
		t.Error("expected error for missing device info")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_MissingConfig(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create JSON without config
	missingConfigFile := createInvalidBackupFile(t, "missing-config.json",
		`{"version": 1, "device_info": {"id": "test"}}`)

	opts := &Options{Factory: f, FilePath: missingConfigFile}
	err := run(opts)
	if err == nil {
		t.Error("expected error for missing config")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_EmptyFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create empty file
	emptyFile := createInvalidBackupFile(t, "empty.json", "")

	opts := &Options{Factory: f, FilePath: emptyFile}
	err := run(opts)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_FileIsDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create a directory instead of a file
	dirPath := testBackupDir + "/not-a-file"
	if err := fs.MkdirAll(dirPath, 0o700); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	opts := &Options{Factory: f, FilePath: dirPath}
	err := run(opts)
	if err == nil {
		t.Error("expected error when path is a directory")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_OutputShowsConfigKeys(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

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

	configFile := testBackupDir + "/config.json"
	if err := afero.WriteFile(config.Fs(), configFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	opts := &Options{Factory: f, FilePath: configFile}
	err = run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should show config key count (3)
	if !bytes.Contains([]byte(output), []byte("3")) {
		t.Errorf("output should contain config key count '3', got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_OutputShowsVersion(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

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

	versionFile := testBackupDir + "/version.json"
	if err := afero.WriteFile(config.Fs(), versionFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	opts := &Options{Factory: f, FilePath: versionFile}
	err = run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should show version info
	if !bytes.Contains([]byte(output), []byte("Version")) {
		t.Errorf("output should contain 'Version', got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_PermissionDenied(t *testing.T) {
	// Skip on Windows where permissions work differently
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("skipping permission test on Windows")
	}

	// Use real filesystem for permission testing
	config.SetFs(afero.NewOsFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create file with no read permissions
	noReadFile := tmpDir + "/noaccess.json"
	if err := os.WriteFile(noReadFile, []byte("{}"), 0o000); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions for cleanup
		if err := os.Chmod(noReadFile, 0o600); err != nil {
			t.Logf("warning: failed to restore permissions: %v", err)
		}
	})

	opts := &Options{Factory: f, FilePath: noReadFile}
	err := run(opts)
	if err == nil {
		t.Error("expected error for permission denied")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestRun_OutputShowsFirmware(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

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

	fwFile := testBackupDir + "/firmware.json"
	if err := afero.WriteFile(config.Fs(), fwFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	opts := &Options{Factory: f, FilePath: fwFile}
	err = run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should show firmware version
	if !bytes.Contains([]byte(output), []byte("2.5.0")) {
		t.Errorf("output should contain firmware version '2.5.0', got: %s", output)
	}
}
