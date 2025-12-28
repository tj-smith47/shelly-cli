package restore

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const testFalseValue = "false"

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "restore <device> <file>" {
		t.Errorf("Use = %q, want 'restore <device> <file>'", cmd.Use)
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

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_Aliases(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	aliases := cmd.Aliases
	if len(aliases) < 2 {
		t.Errorf("expected at least 2 aliases, got %d", len(aliases))
	}

	// Check for expected aliases
	hasApply := false
	hasLoad := false
	for _, a := range aliases {
		if a == "apply" {
			hasApply = true
		}
		if a == "load" {
			hasLoad = true
		}
	}
	if !hasApply {
		t.Error("expected 'apply' alias")
	}
	if !hasLoad {
		t.Error("expected 'load' alias")
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_RequiresTwoArgs(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 2 arguments
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err == nil {
		t.Error("Expected error with only one arg")
	}

	err = cmd.Args(cmd, []string{"device1", "backup.json"})
	if err != nil {
		t.Errorf("Expected no error with two args, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "backup.json", "extra"})
	if err == nil {
		t.Error("Expected error with three args")
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_Flags(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	// Check dry-run flag
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("dry-run flag not found")
	}
	if dryRunFlag.DefValue != testFalseValue {
		t.Errorf("dry-run flag default = %q, want %q", dryRunFlag.DefValue, testFalseValue)
	}

	// Check skip-network flag (defaults to true)
	skipNetworkFlag := cmd.Flags().Lookup("skip-network")
	if skipNetworkFlag == nil {
		t.Fatal("skip-network flag not found")
	}
	if skipNetworkFlag.DefValue != "true" {
		t.Errorf("skip-network flag default = %q, want 'true'", skipNetworkFlag.DefValue)
	}

	// Check skip-scripts flag
	skipScriptsFlag := cmd.Flags().Lookup("skip-scripts")
	if skipScriptsFlag == nil {
		t.Fatal("skip-scripts flag not found")
	}
	if skipScriptsFlag.DefValue != testFalseValue {
		t.Errorf("skip-scripts flag default = %q, want %q", skipScriptsFlag.DefValue, testFalseValue)
	}

	// Check skip-schedules flag
	skipSchedulesFlag := cmd.Flags().Lookup("skip-schedules")
	if skipSchedulesFlag == nil {
		t.Fatal("skip-schedules flag not found")
	}
	if skipSchedulesFlag.DefValue != testFalseValue {
		t.Errorf("skip-schedules flag default = %q, want %q", skipSchedulesFlag.DefValue, testFalseValue)
	}

	// Check skip-webhooks flag
	skipWebhooksFlag := cmd.Flags().Lookup("skip-webhooks")
	if skipWebhooksFlag == nil {
		t.Fatal("skip-webhooks flag not found")
	}
	if skipWebhooksFlag.DefValue != testFalseValue {
		t.Errorf("skip-webhooks flag default = %q, want %q", skipWebhooksFlag.DefValue, testFalseValue)
	}

	// Check decrypt flag
	decryptFlag := cmd.Flags().Lookup("decrypt")
	if decryptFlag == nil {
		t.Fatal("decrypt flag not found")
	}
	if decryptFlag.Shorthand != "d" {
		t.Errorf("decrypt flag shorthand = %q, want 'd'", decryptFlag.Shorthand)
	}
}

// createTestBackupFile creates a valid backup JSON file for testing.
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
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	return filePath
}

//nolint:paralleltest // Uses package-level flag variables
func TestRun_FileNotFound(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device1", "/nonexistent/path/backup.json"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "failed to read backup file") {
		t.Errorf("expected 'failed to read backup file' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestRun_InvalidBackupFile(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("not valid json"), 0o600); err != nil {
		t.Fatalf("failed to create invalid file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device1", invalidFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid backup file")
	}

	if !strings.Contains(err.Error(), "invalid backup file") {
		t.Errorf("expected 'invalid backup file' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestRun_BackupMissingVersion(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	invalidBackup := map[string]any{
		"device_info": map[string]any{
			"id":   "test",
			"name": "Test",
		},
		"config": map[string]any{},
	}
	data, err := json.Marshal(invalidBackup)
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}
	invalidFile := filepath.Join(tmpDir, "no-version.json")
	if err := os.WriteFile(invalidFile, data, 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device1", invalidFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for backup missing version")
	}

	if !strings.Contains(err.Error(), "invalid backup file") {
		t.Errorf("expected 'invalid backup file' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestRun_BackupMissingDeviceInfo(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	invalidBackup := map[string]any{
		"version": 1,
		"config":  map[string]any{},
	}
	data, err := json.Marshal(invalidBackup)
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}
	invalidFile := filepath.Join(tmpDir, "no-device-info.json")
	if err := os.WriteFile(invalidFile, data, 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device1", invalidFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for backup missing device info")
	}

	if !strings.Contains(err.Error(), "invalid backup file") {
		t.Errorf("expected 'invalid backup file' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestRun_BackupMissingConfig(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	invalidBackup := map[string]any{
		"version": 1,
		"device_info": map[string]any{
			"id":   "test",
			"name": "Test",
		},
	}
	data, err := json.Marshal(invalidBackup)
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}
	invalidFile := filepath.Join(tmpDir, "no-config.json")
	if err := os.WriteFile(invalidFile, data, 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device1", invalidFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for backup missing config")
	}

	if !strings.Contains(err.Error(), "invalid backup file") {
		t.Errorf("expected 'invalid backup file' error, got: %v", err)
	}
}

//nolint:paralleltest // Modifies package-level flag variables
func TestRun_DryRun_ValidBackup(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	backupFile := createTestBackupFile(t, tmpDir, "backup.json")

	// Create command first - this binds to the global flag variables
	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--dry-run", "device1", backupFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Dry run should show preview
	if !strings.Contains(output, "Dry run") {
		t.Errorf("expected 'Dry run' in output, got: %s", output)
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_FlagParsing(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	testCases := []struct {
		name string
		args []string
	}{
		{"dry-run", []string{"--dry-run", "device", "backup.json"}},
		{"skip-network false", []string{"--skip-network=false", "device", "backup.json"}},
		{"skip-scripts", []string{"--skip-scripts", "device", "backup.json"}},
		{"skip-schedules", []string{"--skip-schedules", "device", "backup.json"}},
		{"skip-webhooks", []string{"--skip-webhooks", "device", "backup.json"}},
		{"decrypt", []string{"--decrypt", "mypassword", "device", "backup.json"}},
		{"decrypt short", []string{"-d", "mypassword", "device", "backup.json"}},
		{"all skip flags", []string{"--skip-scripts", "--skip-schedules", "--skip-webhooks", "device", "backup.json"}},
		{"combined flags", []string{"--dry-run", "-d", "pass", "--skip-scripts", "device", "backup.json"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags
			dryRunFlag = false
			skipNetworkFlag = true
			skipScriptsFlag = false
			skipSchedulesFlag = false
			skipWebhooksFlag = false
			decryptFlag = ""

			cmd := NewCommand(f)
			cmd.SetOut(out)
			cmd.SetErr(errOut)
			cmd.SetArgs(tc.args)

			err := cmd.ParseFlags(tc.args)
			if err != nil {
				t.Errorf("ParseFlags failed: %v", err)
			}
		})
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_Example_Content(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example

	// Check that example contains expected usage patterns
	examples := []string{
		"backup restore",
		"backup.json",
		"--dry-run",
		"--skip-network",
		"--decrypt",
	}

	for _, e := range examples {
		if !strings.Contains(example, e) {
			t.Errorf("expected example to contain %q", e)
		}
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestNewCommand_Long_Description(t *testing.T) {
	cmd := NewCommand(cmdutil.NewFactory())

	long := cmd.Long

	// Check that long description contains key information
	keywords := []string{
		"Restore",
		"backup",
		"--skip-",
		"network",
	}

	for _, kw := range keywords {
		if !strings.Contains(long, kw) {
			t.Errorf("expected long description to contain %q", kw)
		}
	}
}

//nolint:paralleltest // Uses package-level flag variables
func TestRun_DirectoryAsFile(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0o750); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device1", subDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when file is a directory")
	}

	// Should fail with some file-related error
	if err != nil {
		// Any error is acceptable since directories can't be read as backup files
		t.Logf("Got expected error: %v", err)
	}
}
