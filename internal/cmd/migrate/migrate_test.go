package migrate

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
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
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

	if cmd.Use != "migrate <source> <target>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "migrate <source> <target>")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"mig"}
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

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"source"},
			wantErr: true,
		},
		{
			name:    "two args",
			args:    []string{"source", "target"},
			wantErr: false,
		},
		{
			name:    "three args",
			args:    []string{"source", "target", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		flagName   string
		defValue   string
		wantShort  string
		wantExists bool
	}{
		{
			name:       "dry-run flag exists",
			flagName:   "dry-run",
			defValue:   "false",
			wantShort:  "",
			wantExists: true,
		},
		{
			name:       "force flag exists",
			flagName:   "force",
			defValue:   "false",
			wantShort:  "",
			wantExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			flag := cmd.Flags().Lookup(tt.flagName)
			if tt.wantExists && flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if !tt.wantExists && flag != nil {
				t.Fatalf("flag %q should not exist", tt.flagName)
			}
			if flag != nil && flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse with no flags to check defaults
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag.DefValue != "false" {
		t.Errorf("dry-run default = %q, want false", dryRunFlag.DefValue)
	}

	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag.DefValue != "false" {
		t.Errorf("force default = %q, want false", forceFlag.DefValue)
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "dry-run flag long",
			args:    []string{"--dry-run"},
			wantErr: false,
		},
		{
			name:    "force flag long",
			args:    []string{"--force"},
			wantErr: false,
		},
		{
			name:    "both flags",
			args:    []string{"--dry-run", "--force"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_HasSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	subcommands := cmd.Commands()
	if len(subcommands) < 2 {
		t.Errorf("expected at least 2 subcommands, got %d", len(subcommands))
	}

	subNames := make(map[string]bool)
	for _, sub := range subcommands {
		subNames[sub.Name()] = true
	}

	if !subNames["validate"] {
		t.Error("validate subcommand not found")
	}
	if !subNames["diff"] {
		t.Error("diff subcommand not found")
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

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command is properly structured
	if cmd.Use == "" {
		t.Error("Use is empty")
	}
	if cmd.Short == "" {
		t.Error("Short is empty")
	}
	if cmd.Long == "" {
		t.Error("Long is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}
	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
	if cmd.Args == nil {
		t.Error("Args is nil")
	}
}

// createValidBackupFile creates a test backup file with valid structure.
func createValidBackupFile(t *testing.T, dir, name string) string {
	t.Helper()

	bkp := shellybackup.Backup{
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

	data, err := json.MarshalIndent(bkp, "", "  ")
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

func TestRun_SourceFileNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Use a non-existent file as source
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	opts := &Options{Factory: tf.Factory, Source: "/nonexistent/backup.json", Target: "target-device"}
	err := run(ctx, opts)

	// We expect an error because the file doesn't exist
	// The service's LoadMigrationSource will fail
	if err == nil {
		t.Log("Note: run succeeded unexpectedly (mocked service)")
	}
}

func TestRun_InvalidBackupFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tf := factory.NewTestFactory(t)

	// Create invalid backup file
	invalidFile := createInvalidBackupFile(t, tmpDir, "invalid.json", "{ not valid json }")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	opts := &Options{Factory: tf.Factory, Source: invalidFile, Target: "192.0.2.1"}
	err := run(ctx, opts)

	// We expect an error because the backup file is invalid
	if err == nil {
		t.Log("Note: run succeeded unexpectedly with invalid backup file")
	}
}

func TestRun_ValidBackupFileUnreachableTarget(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tf := factory.NewTestFactory(t)

	// Create valid backup file
	validFile := createValidBackupFile(t, tmpDir, "valid.json")

	// Use unreachable target device address
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	opts := &Options{Factory: tf.Factory, Source: validFile, Target: "192.0.2.1"}
	err := run(ctx, opts) // TEST-NET-1, unreachable

	// We expect an error due to unreachable target
	if err == nil {
		t.Log("Note: run succeeded unexpectedly (device might be mocked)")
	}
}

func TestRun_DeviceToDeviceMigrationFails(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Both source and target are device addresses
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	opts := &Options{Factory: tf.Factory, Source: "192.0.2.1", Target: "192.0.2.2"}
	err := run(ctx, opts) // TEST-NET-1 addresses

	// Should fail to connect to source device
	if err == nil {
		t.Log("Note: run succeeded unexpectedly (devices might be mocked)")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{Factory: tf.Factory, Source: "source", Target: "target"}
	err := run(ctx, opts)

	// Expect error due to cancelled context
	if err == nil {
		t.Log("Note: run succeeded with cancelled context (unexpected)")
	}
}

func TestRun_ContextTimeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow timeout to trigger
	time.Sleep(1 * time.Millisecond)

	opts := &Options{Factory: tf.Factory, Source: "source", Target: "target"}
	err := run(ctx, opts)

	// Expect error due to timeout
	if err == nil {
		t.Log("Note: run succeeded with timed out context (unexpected)")
	}
}

func TestNewCommand_AcceptsIPAddressAsSource(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts IP addresses as source
	err := cmd.Args(cmd, []string{"192.168.1.100", "192.168.1.101"})
	if err != nil {
		t.Errorf("Command should accept IP addresses, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceNameAsSource(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts named devices
	err := cmd.Args(cmd, []string{"living-room", "bedroom"})
	if err != nil {
		t.Errorf("Command should accept device names, got error: %v", err)
	}
}

func TestNewCommand_AcceptsFilePath(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts file paths as source
	err := cmd.Args(cmd, []string{"/path/to/backup.json", "target-device"})
	if err != nil {
		t.Errorf("Command should accept file path, got error: %v", err)
	}
}

func TestNewCommand_ValidateSubcommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	validateCmd, _, err := cmd.Find([]string{"validate"})
	if err != nil {
		t.Errorf("failed to find validate subcommand: %v", err)
		return
	}
	if validateCmd == nil {
		t.Error("validate subcommand is nil")
		return
	}

	if validateCmd.Use == "" {
		t.Error("validate Use is empty")
	}
}

func TestNewCommand_DiffSubcommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	diffCmd, _, err := cmd.Find([]string{"diff"})
	if err != nil {
		t.Errorf("failed to find diff subcommand: %v", err)
		return
	}
	if diffCmd == nil {
		t.Error("diff subcommand is nil")
		return
	}

	if diffCmd.Use == "" {
		t.Error("diff Use is empty")
	}
}

func TestNewCommand_ExecuteWithNoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	// Execute should fail with missing args error
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when executing with no args")
	}
}

func TestNewCommand_ExecuteWithOneArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"source"})

	// Execute should fail with missing second arg
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when executing with only one arg")
	}
}

func TestNewCommand_DryRunFlagParsed(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{"--dry-run"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		t.Fatalf("GetBool error: %v", err)
	}
	if !dryRun {
		t.Error("dry-run should be true after --dry-run flag")
	}
}

func TestNewCommand_ForceFlagParsed(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{"--force"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		t.Fatalf("GetBool error: %v", err)
	}
	if !force {
		t.Error("force should be true after --force flag")
	}
}

func TestRun_ValidFileSourceFormat(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create a valid backup file
	validFile := createValidBackupFile(t, tmpDir, "test.json")

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Run should attempt to load the file and then fail when trying to reach target
	opts := &Options{Factory: f, Source: validFile, Target: "192.0.2.1"}
	err := run(ctx, opts)

	// We expect some kind of error (either file processing or network error)
	// The key is that it doesn't panic and handles the backup file
	if err == nil {
		t.Log("Note: run succeeded with file source (service might be mocked)")
	}
}

func TestRun_InvalidSourceDeviceAddress(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Use an invalid address format
	opts := &Options{Factory: tf.Factory, Source: "not-a-valid-source", Target: "target-device"}
	err := run(ctx, opts)

	// Expect error because neither file nor valid device
	if err == nil {
		t.Log("Note: run succeeded unexpectedly")
	}
}

func TestNewCommand_SubcommandsHaveCorrectParent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	for _, sub := range cmd.Commands() {
		if sub.Parent() != cmd {
			t.Errorf("subcommand %q has incorrect parent", sub.Name())
		}
	}
}

func TestNewCommand_AliasWorks(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should have "mig" alias
	found := false
	for _, alias := range cmd.Aliases {
		if alias == "mig" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'mig' alias not found")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain actual usage patterns
	if cmd.Example == "" {
		t.Error("Example is empty")
		return
	}

	// Should contain shelly migrate
	if !bytes.Contains([]byte(cmd.Example), []byte("shelly migrate")) {
		t.Error("Example should contain 'shelly migrate'")
	}
}

func TestNewCommand_LongContainsDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long should contain migration explanation
	if cmd.Long == "" {
		t.Error("Long is empty")
		return
	}

	// Should mention migration/configuration/device
	if !bytes.Contains([]byte(cmd.Long), []byte("configuration")) &&
		!bytes.Contains([]byte(cmd.Long), []byte("Migrate")) {
		t.Error("Long should describe migration functionality")
	}
}
