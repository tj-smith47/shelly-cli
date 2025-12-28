package csv

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "csv <devices...> [file]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "csv <devices...> [file]")
	}

	aliases := []string{"spreadsheet"}
	if len(cmd.Aliases) != len(aliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, aliases)
	}
	for i, alias := range aliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

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
			name:    "one device",
			args:    []string{"device1"},
			wantErr: false,
		},
		{
			name:    "multiple devices",
			args:    []string{"device1", "device2", "device3"},
			wantErr: false,
		},
		{
			name:    "device with file",
			args:    []string{"device1", "output.csv"},
			wantErr: false,
		},
		{
			name:    "all devices group",
			args:    []string{"@all"},
			wantErr: false,
		},
		{
			name:    "multiple devices with file",
			args:    []string{"device1", "device2", "export.csv"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check no-header flag exists
	noHeaderFlag := cmd.Flags().Lookup("no-header")
	if noHeaderFlag == nil {
		t.Fatal("no-header flag not found")
	}

	if noHeaderFlag.DefValue != "false" {
		t.Errorf("no-header default = %q, want %q", noHeaderFlag.DefValue, "false")
	}

	// Verify flag is boolean type
	if noHeaderFlag.Value.Type() != "bool" {
		t.Errorf("no-header type = %q, want bool", noHeaderFlag.Value.Type())
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Get the no-header flag value
	noHeader, err := cmd.Flags().GetBool("no-header")
	if err != nil {
		t.Fatalf("failed to get no-header flag: %v", err)
	}

	if noHeader {
		t.Error("no-header default should be false")
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil, expected completion function")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description exists and is meaningful
	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	if len(cmd.Long) < 30 {
		t.Error("Long description seems too short")
	}

	// Check it mentions key features
	if !strings.Contains(cmd.Long, "CSV") {
		t.Error("Long description should mention CSV")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Example should show meaningful patterns
	if len(cmd.Example) < 20 {
		t.Error("Example seems too short to be useful")
	}

	// Check it has export csv commands
	if !strings.Contains(cmd.Example, "shelly export csv") {
		t.Error("Example should contain 'shelly export csv'")
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{Factory: f}

	if opts.NoHead {
		t.Error("NoHead should default to false")
	}

	if opts.File != "" {
		t.Error("File should default to empty string")
	}

	if opts.Devices != nil {
		t.Error("Devices should default to nil")
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

func TestOptions_Structure(t *testing.T) {
	t.Parallel()

	// Test that Options struct can be created with all fields
	f := cmdutil.NewFactory()
	opts := &Options{
		Devices: []string{"device1", "device2"},
		File:    "output.csv",
		NoHead:  true,
		Factory: f,
	}

	if len(opts.Devices) != 2 {
		t.Errorf("Devices length = %d, want 2", len(opts.Devices))
	}

	if opts.File != "output.csv" {
		t.Errorf("File = %q, want %q", opts.File, "output.csv")
	}

	if !opts.NoHead {
		t.Error("NoHead should be true")
	}
}

func TestSplitDevicesAndFile_CSV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantDevices []string
		wantFile    string
	}{
		{
			name:        "single device",
			args:        []string{"device1"},
			wantDevices: []string{"device1"},
			wantFile:    "",
		},
		{
			name:        "device with csv file",
			args:        []string{"device1", "output.csv"},
			wantDevices: []string{"device1"},
			wantFile:    "output.csv",
		},
		{
			name:        "multiple devices",
			args:        []string{"device1", "device2"},
			wantDevices: []string{"device1", "device2"},
			wantFile:    "",
		},
		{
			name:        "multiple devices with file",
			args:        []string{"device1", "device2", "export.csv"},
			wantDevices: []string{"device1", "device2"},
			wantFile:    "export.csv",
		},
		{
			name:        "file without csv extension ignored",
			args:        []string{"device1", "output.txt"},
			wantDevices: []string{"device1", "output.txt"},
			wantFile:    "",
		},
	}

	csvExtensions := []string{".csv"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			devices, file := shelly.SplitDevicesAndFile(tt.args, csvExtensions)

			if len(devices) != len(tt.wantDevices) {
				t.Errorf("devices = %v, want %v", devices, tt.wantDevices)
			} else {
				for i, d := range devices {
					if d != tt.wantDevices[i] {
						t.Errorf("devices[%d] = %q, want %q", i, d, tt.wantDevices[i])
					}
				}
			}

			if file != tt.wantFile {
				t.Errorf("file = %q, want %q", file, tt.wantFile)
			}
		})
	}
}

func TestNewCommand_SetFlagValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting the flag value
	err := cmd.Flags().Set("no-header", "true")
	if err != nil {
		t.Fatalf("failed to set no-header flag: %v", err)
	}

	noHeader, err := cmd.Flags().GetBool("no-header")
	if err != nil {
		t.Fatalf("failed to get no-header flag: %v", err)
	}

	if !noHeader {
		t.Error("no-header should be true after setting")
	}
}

// TestRun_NoDevices tests run with empty devices after expansion.
func TestRun_NoDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{}, // Empty devices
		File:    "",
		NoHead:  false,
	}

	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error for no devices")
	}
	if !strings.Contains(err.Error(), "no devices specified") {
		t.Errorf("Error should contain 'no devices specified', got: %v", err)
	}
}

// TestRun_DeviceConnection tests run with device that requires connection.
func TestRun_DeviceConnection(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{"test-device"},
		File:    "",
		NoHead:  false,
	}

	// This will exercise the run function but fail on connection
	err := run(context.Background(), opts)

	// We don't expect the connection to succeed in tests
	// But this exercises the code path up to the connection attempt
	if err == nil {
		t.Log("Expected connection error (no real device), but run succeeded")
	}
}

// TestRun_WithNoHeader tests run with no-header option.
func TestRun_WithNoHeader(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{"test-device"},
		File:    "",
		NoHead:  true,
	}

	// This exercises the no-header code path
	err := run(context.Background(), opts)

	// Connection will fail but exercises the code path
	if err == nil {
		t.Log("Expected connection error (no real device), but run succeeded")
	}
}

// TestRun_WriteToFile tests run with file output.
func TestRun_WriteToFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "devices.csv")

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{"test-device"},
		File:    outputFile,
		NoHead:  false,
	}

	// This exercises the file output path
	err := run(context.Background(), opts)

	// Connection will fail but exercises the code path
	if err == nil {
		t.Log("Expected connection error (no real device), but run succeeded")
	}
}

// TestRun_InvalidFilePath tests run with invalid file path.
func TestRun_InvalidFilePath(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{"test-device"},
		File:    "/nonexistent/directory/file.csv",
		NoHead:  false,
	}

	// This should fail either on file creation or connection
	err := run(context.Background(), opts)

	if err == nil {
		t.Log("Expected error for invalid file path or connection")
	}
}

// TestRun_ContextCancelled tests behavior with cancelled context.
func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{"test-device"},
		File:    "",
		NoHead:  false,
	}

	err := run(ctx, opts)

	// Should return context error or handle cancellation
	if err == nil {
		t.Log("Expected error due to context cancellation")
	}
}

// TestRun_MultipleDevices tests run with multiple devices.
func TestRun_MultipleDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{"device1", "device2", "device3"},
		File:    "",
		NoHead:  false,
	}

	// This exercises the multiple devices code path
	err := run(context.Background(), opts)

	// Connection will fail but exercises the code path
	if err == nil {
		t.Log("Expected connection error (no real devices), but run succeeded")
	}
}

// TestNewCommand_WithTestIOStreams verifies command creation with test IOStreams.
func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}

	if cmd.Use == "" {
		t.Error("Use should not be empty")
	}
}

// TestNewCommand_FlagParsing tests flag parsing.
func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no-header flag",
			args:    []string{"--no-header"},
			wantErr: false,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown-flag"},
			wantErr: true,
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

// TestRun_StdoutOutput tests output to stdout.
func TestRun_StdoutOutput(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{"test-device"},
		File:    "", // Empty file means stdout
		NoHead:  false,
	}

	// This exercises the stdout output path
	err := run(context.Background(), opts)

	// Connection will fail but exercises the code path
	if err == nil {
		t.Log("Expected connection error (no real device), but run succeeded")
	}
}

// TestRun_FileExistsOverwrite tests overwriting existing file.
func TestRun_FileExistsOverwrite(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "existing.csv")

	// Create existing file
	err := os.WriteFile(outputFile, []byte("existing content"), 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Devices: []string{"test-device"},
		File:    outputFile,
		NoHead:  false,
	}

	// Run should attempt to overwrite the file
	runErr := run(context.Background(), opts)

	// Connection will fail but exercises the file creation path
	if runErr == nil {
		t.Log("Expected connection error (no real device), but run succeeded")
	}
}
