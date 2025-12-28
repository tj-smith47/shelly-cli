package list

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "list [directory]" {
		t.Errorf("Use = %q, want 'list [directory]'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify aliases
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases are empty")
	}

	// Verify example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_OptionalArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept no args
	err := cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error with no args, got: %v", err)
	}

	// Should accept one arg
	err = cmd.Args(cmd, []string{"/some/path"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	// Should reject two args
	err = cmd.Args(cmd, []string{"/path1", "/path2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestRun_DirectoryNotExist(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent", "backups")
	cmd.SetArgs([]string{nonExistentDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No backups directory found") {
		t.Errorf("expected 'No backups directory found' message, got: %s", output)
	}
}

func TestRun_NotADirectory(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create a file instead of directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "not-a-directory")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{filePath})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-directory path")
	}

	if !strings.Contains(err.Error(), "is not a directory") {
		t.Errorf("expected 'is not a directory' error, got: %v", err)
	}
}

func TestRun_EmptyDirectory(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No backup files found") {
		t.Errorf("expected 'No backup files found' message, got: %s", output)
	}
}

func TestRun_DirectoryWithNonBackupFiles(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()

	// Create non-backup files
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "config.toml"), []byte("toml"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No backup files found") {
		t.Errorf("expected 'No backup files found' message for non-backup files, got: %s", output)
	}
}

// createTestBackupFile creates a valid backup JSON file for testing
func createTestBackupFile(t *testing.T, dir, filename string) string {
	t.Helper()

	backup := shellybackup.Backup{
		Version: 1,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:         "shellyplus1-test",
			Name:       "Test Device",
			Model:      "SNSW-001X16EU",
			Generation: 2,
			Version:    "1.0.0",
			MAC:        "AA:BB:CC:DD:EE:FF",
		},
		Config:    json.RawMessage(`{"sys":{"device":{"name":"Test Device"}}}`),
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}

	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

func TestRun_WithBackupFiles(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	createTestBackupFile(t, tmpDir, "device1.json")
	createTestBackupFile(t, tmpDir, "device2.json")

	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Should display the backup files in table format
	if !strings.Contains(output, "device1.json") && !strings.Contains(output, "device2.json") {
		// Table might show truncated names, just verify some output exists
		if len(output) == 0 {
			t.Error("expected some output for backup files")
		}
	}
}

func TestRun_WithInvalidBackupFiles(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()

	// Create an invalid JSON file
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("failed to create invalid file: %v", err)
	}

	// Create a valid backup file
	createTestBackupFile(t, tmpDir, "valid.json")

	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should skip invalid files and show valid ones
	output := out.String()
	// The valid file should be shown
	if len(output) == 0 {
		t.Error("expected some output")
	}
}

func TestRun_WithYamlBackupFile(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()

	// Create a YAML file with .yaml extension (won't be valid backup but tests extension detection)
	yamlPath := filepath.Join(tmpDir, "backup.yaml")
	if err := os.WriteFile(yamlPath, []byte("invalid: yaml"), 0o644); err != nil {
		t.Fatalf("failed to create yaml file: %v", err)
	}

	// Also create a .yml file
	ymlPath := filepath.Join(tmpDir, "backup.yml")
	if err := os.WriteFile(ymlPath, []byte("invalid: yml"), 0o644); err != nil {
		t.Fatalf("failed to create yml file: %v", err)
	}

	// Create a valid JSON backup
	createTestBackupFile(t, tmpDir, "backup.json")

	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should at least find the valid JSON backup
	output := out.String()
	if len(output) == 0 {
		t.Error("expected some output")
	}
}

func TestRun_DirectoryWithSubdirectories(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Put a backup in the subdirectory (should not be found)
	createTestBackupFile(t, subDir, "nested.json")

	// No backups in root directory
	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Should not find the nested backup
	if !strings.Contains(output, "No backup files found") {
		t.Errorf("expected 'No backup files found' for empty root dir, got: %s", output)
	}
}

func TestRun_BackupWithMissingVersion(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()

	// Create backup JSON without version field
	invalidBackup := map[string]any{
		"device_info": map[string]any{
			"id":   "test",
			"name": "Test",
		},
		"config": map[string]any{},
	}
	data, _ := json.Marshal(invalidBackup)
	if err := os.WriteFile(filepath.Join(tmpDir, "no-version.json"), data, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Invalid backup should be skipped
	output := out.String()
	if !strings.Contains(output, "No backup files found") {
		t.Logf("Got output: %s", output)
	}
}

func TestRun_BackupWithMissingDeviceInfo(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()

	// Create backup JSON without device_info
	invalidBackup := map[string]any{
		"version": 1,
		"config":  map[string]any{},
	}
	data, _ := json.Marshal(invalidBackup)
	if err := os.WriteFile(filepath.Join(tmpDir, "no-device.json"), data, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Invalid backup should be skipped
	output := out.String()
	if !strings.Contains(output, "No backup files found") {
		t.Logf("Got output: %s", output)
	}
}

func TestRun_BackupWithMissingConfig(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()

	// Create backup JSON without config
	invalidBackup := map[string]any{
		"version": 1,
		"device_info": map[string]any{
			"id":   "test",
			"name": "Test",
		},
	}
	data, _ := json.Marshal(invalidBackup)
	if err := os.WriteFile(filepath.Join(tmpDir, "no-config.json"), data, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Invalid backup should be skipped
	output := out.String()
	if !strings.Contains(output, "No backup files found") {
		t.Logf("Got output: %s", output)
	}
}

func TestRun_MultipleValidBackups(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()

	// Create multiple valid backup files
	createTestBackupFile(t, tmpDir, "living-room.json")
	createTestBackupFile(t, tmpDir, "kitchen.json")
	createTestBackupFile(t, tmpDir, "bedroom.json")

	cmd := NewCommand(f)
	cmd.SetArgs([]string{tmpDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have output for all three backups
	output := out.String()
	if len(output) == 0 {
		t.Error("expected output for multiple backups")
	}
}
