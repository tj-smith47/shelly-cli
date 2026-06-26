package restore

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	clibackup "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testFalseValue = "false"

func TestNewCommand(t *testing.T) {
	t.Parallel()
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

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
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

func TestNewCommand_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check dry-run flag
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("dry-run flag not found")
	}
	if dryRunFlag.DefValue != testFalseValue {
		t.Errorf("dry-run flag default = %q, want %q", dryRunFlag.DefValue, testFalseValue)
	}

	// Check skip-auth flag (defaults to false)
	skipAuthFlag := cmd.Flags().Lookup("skip-auth")
	if skipAuthFlag == nil {
		t.Fatal("skip-auth flag not found")
	}
	if skipAuthFlag.DefValue != testFalseValue {
		t.Errorf("skip-auth flag default = %q, want %q", skipAuthFlag.DefValue, testFalseValue)
	}

	// Check skip-network flag (defaults to false)
	skipNetworkFlag := cmd.Flags().Lookup("skip-network")
	if skipNetworkFlag == nil {
		t.Fatal("skip-network flag not found")
	}
	if skipNetworkFlag.DefValue != testFalseValue {
		t.Errorf("skip-network flag default = %q, want %q", skipNetworkFlag.DefValue, testFalseValue)
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

func TestRun_FileNotFound(t *testing.T) {
	t.Parallel()
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

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_InvalidBackupFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	invalidFile := "/test/invalid.json"
	if err := afero.WriteFile(config.Fs(), invalidFile, []byte("not valid json"), 0o600); err != nil {
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

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_BackupMissingVersion(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

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
	invalidFile := "/test/no-version.json"
	if err := afero.WriteFile(config.Fs(), invalidFile, data, 0o600); err != nil {
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

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_BackupMissingDeviceInfo(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	invalidBackup := map[string]any{
		"version": 1,
		"config":  map[string]any{},
	}
	data, err := json.Marshal(invalidBackup)
	if err != nil {
		t.Fatalf("failed to marshal backup: %v", err)
	}
	invalidFile := "/test/no-device-info.json"
	if err := afero.WriteFile(config.Fs(), invalidFile, data, 0o600); err != nil {
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

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_BackupMissingConfig(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

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
	invalidFile := "/test/no-config.json"
	if err := afero.WriteFile(config.Fs(), invalidFile, data, 0o600); err != nil {
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

func encryptedBackupFile(t *testing.T, password, path string) {
	t.Helper()
	bkp := &clibackup.DeviceBackup{Backup: &shellybackup.Backup{
		Version: shellybackup.BackupVersion,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:         "shellyplus1-enc",
			Model:      "SNSW-001X16EU",
			Generation: 2,
		},
		Config:    json.RawMessage(`{"sys":{}}`),
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}}
	data, err := clibackup.Encrypt(bkp, password)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if err := afero.WriteFile(config.Fs(), path, data, 0o600); err != nil {
		t.Fatalf("write encrypted backup: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_EncryptedRequiresDecrypt(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	f := cmdutil.NewFactory().SetIOStreams(iostreams.Test(nil, out, errOut))

	encFile := "/test/enc.json"
	encryptedBackupFile(t, "s3cret", encFile)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device1", encFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error restoring an encrypted backup without --decrypt")
	}
	if !strings.Contains(err.Error(), "--decrypt") {
		t.Errorf("expected '--decrypt' hint, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_EncryptedDryRunWithDecrypt(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	f := cmdutil.NewFactory().SetIOStreams(iostreams.Test(nil, out, errOut))

	encFile := "/test/enc.json"
	encryptedBackupFile(t, "s3cret", encFile)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--dry-run", "--decrypt", "s3cret", "device1", encFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error decrypting+previewing: %v", err)
	}
	if !strings.Contains(out.String(), "shellyplus1-enc") {
		t.Errorf("expected decrypted device ID in preview, got: %s", out.String())
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_EncryptedWrongPassword(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	f := cmdutil.NewFactory().SetIOStreams(iostreams.Test(nil, out, errOut))

	encFile := "/test/enc.json"
	encryptedBackupFile(t, "s3cret", encFile)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--dry-run", "--decrypt", "wrong", "device1", encFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error restoring with the wrong decryption password")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_DryRun_ValidBackup(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

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
	backupFile := "/test/backup.json"
	if err := afero.WriteFile(config.Fs(), backupFile, data, 0o600); err != nil {
		t.Fatalf("failed to write backup file: %v", err)
	}

	// Create command first - this binds to the global flag variables
	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--dry-run", "device1", backupFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Dry run should show preview
	if !strings.Contains(output, "Dry run") {
		t.Errorf("expected 'Dry run' in output, got: %s", output)
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

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

func TestNewCommand_Example_Content(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example

	// Check that example contains expected usage patterns
	examples := []string{
		"backup restore",
		"backup.json",
		"--dry-run",
		"--skip-network",
		"--skip-auth",
		"--decrypt",
	}

	for _, e := range examples {
		if !strings.Contains(example, e) {
			t.Errorf("expected example to contain %q", e)
		}
	}
}

func TestNewCommand_Long_Description(t *testing.T) {
	t.Parallel()
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

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_DirectoryAsFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	subDir := "/test/subdir"
	if err := config.Fs().MkdirAll(subDir, 0o750); err != nil {
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
	// Any error is acceptable since directories can't be read as backup files.
	t.Logf("Got expected error: %v", err)
}

func TestValidateFlags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		opts        Options
		wantErrSubs string
	}{
		{name: "no flags", opts: Options{}},
		{name: "to-ap with static-ip", opts: Options{ToAP: "ShellyBulbDuo-AABBCC", StaticIP: "10.0.0.5"}},
		{name: "static-ip with skip-network", opts: Options{StaticIP: "10.0.0.5", SkipNetwork: true}, wantErrSubs: "static-ip cannot be used with --skip-network"},
		{name: "to-ap with skip-network", opts: Options{ToAP: "ShellyBulbDuo-AABBCC", SkipNetwork: true}, wantErrSubs: "to-ap cannot be used with --skip-network"},
		{name: "to-ap with dry-run", opts: Options{ToAP: "ShellyBulbDuo-AABBCC", DryRun: true}, wantErrSubs: "to-ap cannot be combined with --dry-run"},
		{name: "ap-ip without to-ap", opts: Options{APIP: "192.168.33.140"}, wantErrSubs: "ap-ip only applies with --to-ap"},
		{name: "ap-ip with to-ap", opts: Options{ToAP: "ShellyBulbDuo-AABBCC", APIP: "192.168.33.140"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.opts.validateFlags()
			if tt.wantErrSubs == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErrSubs) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErrSubs)
			}
		})
	}
}

func TestToAPFlagRegistered(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())
	if cmd.Flags().Lookup("to-ap") == nil {
		t.Error("--to-ap flag not registered")
	}
}

func TestNewCommand_SkipStateAndMetersFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	for _, name := range []string{"skip-state", "skip-meters"} {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Fatalf("--%s flag not registered", name)
		}
		if f.DefValue != testFalseValue {
			t.Errorf("--%s default = %q, want %q", name, f.DefValue, testFalseValue)
		}
	}
}

func TestNewCommand_FirmwareAndTraceFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	downgrade := cmd.Flags().Lookup("allow-firmware-downgrade")
	if downgrade == nil {
		t.Fatal("--allow-firmware-downgrade flag not registered")
	}
	if downgrade.DefValue != testFalseValue {
		t.Errorf("--allow-firmware-downgrade default = %q, want %q", downgrade.DefValue, testFalseValue)
	}

	trace := cmd.Flags().Lookup("trace-file")
	if trace == nil {
		t.Fatal("--trace-file flag not registered")
	}
	// The trace file is a debug seam, hidden from normal help so the command's
	// surface stays clean.
	if !trace.Hidden {
		t.Error("--trace-file should be hidden")
	}
}

// writeValidBackup writes a minimal but valid Gen2 backup (no WiFi blob) to the
// in-memory FS and returns its path. A WiFi-less backup is what lets the --to-ap
// paths fail safely in resolveJoinNetwork before any host WiFi mutation.
func writeValidBackup(t *testing.T, path string) {
	t.Helper()
	bkp := shellybackup.Backup{
		Version: 1,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:         "shellybulbduo-test",
			Name:       "src",
			Model:      "SHBDUO-1",
			Generation: 2,
			Version:    "1.0.0",
			MAC:        "AA:BB:CC:DD:EE:FF",
		},
		Config:    json.RawMessage(`{"sys":{"device":{"name":"src"}}}`),
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	data, err := json.Marshal(bkp)
	if err != nil {
		t.Fatalf("marshal backup: %v", err)
	}
	if err := afero.WriteFile(config.Fs(), path, data, 0o600); err != nil {
		t.Fatalf("write backup: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_RestoreBackupFails(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	const backupFile = "/test/restore-fail.json"
	writeValidBackup(t, backupFile)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// TEST-NET-1 device: run reaches the RestoreBackup call (covering the full
	// non-dry-run, non-AP body) and fails dialing the unreachable device.
	opts := &Options{Factory: tf.Factory, Device: "192.0.2.30", FilePath: backupFile}
	err := run(ctx, opts)
	if err == nil {
		t.Fatal("expected restore to fail against an unreachable device")
	}
	if !strings.Contains(err.Error(), "failed to restore backup") {
		t.Errorf("got %v, want error containing %q", err, "failed to restore backup")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_ToAP_RestoreViaAPFails(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	const backupFile = "/test/to-ap.json"
	writeValidBackup(t, backupFile)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// --to-ap drives run through restoreViaAP -> RestoreToAP. The WiFi-less backup
	// resolves no station passphrase, so it fails in resolveJoinNetwork BEFORE any
	// host WiFi hop is attempted.
	opts := &Options{Factory: tf.Factory, Device: "dst", FilePath: backupFile, ToAP: "ShellyBulbDuo-AABBCC"}
	err := run(ctx, opts)
	if err == nil {
		t.Fatal("expected restore-via-AP to fail without a resolvable passphrase")
	}
	if !strings.Contains(err.Error(), "failed to restore via AP") {
		t.Errorf("got %v, want error containing %q", err, "failed to restore via AP")
	}
	// not covered: restoreViaAP's success tail (ios.Success / newAddr / display)
	// requires RestoreToAP to finish a real AP hop + device write, which mutates
	// host WiFi and contacts hardware — never exercised in a unit test.
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRestoreViaAP_DirectFailsBeforeHop(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	bkp := &clibackup.DeviceBackup{Backup: &shellybackup.Backup{
		Version:    1,
		DeviceInfo: &shellybackup.DeviceInfo{ID: "shellybulbduo-test", Generation: 2, Version: "1.0.0"},
		Config:     json.RawMessage(`{}`),
	}}
	opts := &Options{Factory: tf.Factory, Device: "dst", ToAP: "ShellyBulbDuo-AABBCC"}
	err := opts.restoreViaAP(ctx, tf.ShellyService(), bkp, clibackup.RestoreOptions{})
	if err == nil {
		t.Fatal("expected restoreViaAP to fail without a resolvable passphrase")
	}
	if !strings.Contains(err.Error(), "failed to restore via AP") {
		t.Errorf("got %v, want error containing %q", err, "failed to restore via AP")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_InvalidFlagCombo(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	const backupFile = "/test/bad-flags.json"
	writeValidBackup(t, backupFile)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// --static-ip with --skip-network is rejected by validateFlags inside run,
	// after the backup is read but before any device I/O.
	opts := &Options{
		Factory:     tf.Factory,
		Device:      "dst",
		FilePath:    backupFile,
		StaticIP:    "10.0.0.9",
		SkipNetwork: true,
	}
	err := run(ctx, opts)
	if err == nil {
		t.Fatal("expected validateFlags to reject --static-ip with --skip-network")
	}
	if !strings.Contains(err.Error(), "static-ip cannot be used with --skip-network") {
		t.Errorf("got %v, want error containing %q", err, "static-ip cannot be used with --skip-network")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_TraceFileOpenError(t *testing.T) {
	// Stage the backup in a writable base, then wrap it read-only: the backup read
	// still succeeds while attachTrace's Create fails, covering run's
	// attachTrace-error branch before any device I/O.
	base := afero.NewMemMapFs()
	const backupFile = "/test/trace-run.json"
	config.SetFs(base)
	writeValidBackup(t, backupFile)
	config.SetFs(afero.NewReadOnlyFs(base))
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	opts := &Options{Factory: tf.Factory, Device: "dst", FilePath: backupFile, TraceFile: "/test/trace.log"}
	err := run(ctx, opts)
	if err == nil {
		t.Fatal("expected run to fail opening a trace file over a directory")
	}
	if !strings.Contains(err.Error(), "failed to open trace file") {
		t.Errorf("got %v, want error containing %q", err, "failed to open trace file")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_DryRun_WithStaticIPOverride(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	const backupFile = "/test/dry-override.json"
	writeValidBackup(t, backupFile)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Dry-run with a static-IP override exercises the override branch of run and
	// the override-info line, all without device I/O.
	opts := &Options{
		Factory:  tf.Factory,
		Device:   "dst",
		FilePath: backupFile,
		DryRun:   true,
		StaticIP: "10.0.0.9",
		Gateway:  "10.0.0.1",
		Netmask:  "255.255.255.0",
	}
	if err := run(ctx, opts); err != nil {
		t.Fatalf("dry-run restore: %v", err)
	}
	out := tf.OutString()
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry-run preview, got %q", out)
	}
	if !strings.Contains(out, "10.0.0.9") {
		t.Errorf("expected static-IP override note, got %q", out)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestAttachTrace_CreateError(t *testing.T) {
	// A read-only FS makes Create fail, covering attachTrace's open-error branch.
	config.SetFs(afero.NewReadOnlyFs(afero.NewMemMapFs()))
	t.Cleanup(func() { config.SetFs(nil) })

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, TraceFile: "/test/trace.log"}
	restoreOpts := &clibackup.RestoreOptions{}
	cleanup, err := opts.attachTrace(restoreOpts)
	if err == nil {
		t.Fatal("expected attachTrace to fail on a read-only filesystem")
	}
	if cleanup != nil {
		t.Error("cleanup must be nil when attachTrace errors")
	}
	if !strings.Contains(err.Error(), "failed to open trace file") {
		t.Errorf("got %v, want error containing %q", err, "failed to open trace file")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestAttachTrace(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	t.Run("no trace file is a no-op cleanup", func(t *testing.T) {
		opts := &Options{Factory: f}
		restoreOpts := &clibackup.RestoreOptions{}
		cleanup, err := opts.attachTrace(restoreOpts)
		if err != nil {
			t.Fatalf("attachTrace: %v", err)
		}
		if cleanup == nil {
			t.Fatal("cleanup must never be nil")
		}
		cleanup() // must not panic
		if restoreOpts.StepTrace != nil {
			t.Error("StepTrace should stay nil without --trace-file")
		}
	})

	t.Run("trace file wires a writable sink", func(t *testing.T) {
		opts := &Options{Factory: f, TraceFile: "/test/trace.log"}
		restoreOpts := &clibackup.RestoreOptions{}
		cleanup, err := opts.attachTrace(restoreOpts)
		if err != nil {
			t.Fatalf("attachTrace: %v", err)
		}
		if restoreOpts.StepTrace == nil {
			t.Fatal("StepTrace was not wired from --trace-file")
		}
		if _, err := restoreOpts.StepTrace.Write([]byte("step=mqtt ok\n")); err != nil {
			t.Errorf("trace sink not writable: %v", err)
		}
		cleanup()
		data, err := afero.ReadFile(config.Fs(), "/test/trace.log")
		if err != nil {
			t.Fatalf("read trace file: %v", err)
		}
		if !strings.Contains(string(data), "step=mqtt") {
			t.Errorf("trace file missing written content: %q", data)
		}
	})
}
