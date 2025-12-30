package send

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "send <device> <data>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "send <device> <data>")
	}

	// Test Aliases
	wantAliases := []string{"tx", "transmit"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
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
		{"no args", []string{}, true},
		{"one arg", []string{"device"}, true},
		{"two args valid", []string{"device", "data"}, false},
		{"three args", []string{"device", "data", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test hex flag exists and defaults to false
	hexFlag := cmd.Flags().Lookup("hex")
	if hexFlag == nil {
		t.Fatal("--hex flag not found")
	}
	if hexFlag.DefValue != "false" {
		t.Errorf("--hex default = %q, want %q", hexFlag.DefValue, "false")
	}

	// Test id flag exists (set to 0 by AddComponentFlags, then overridden to 100 in NewCommand)
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Fatal("--id flag not found")
	}
	// The flag starts at 0, and opts.ID is set to 100, but that's on the Options struct
	// The flag itself defaults to 0
	if idFlag.DefValue != "0" {
		t.Errorf("--id default = %q, want %q", idFlag.DefValue, "0")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly lora send",
		"--hex",
		"--id",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	wantPatterns := []string{
		"LoRa",
		"data",
		"base64",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:  "test-device",
		Data:    "test-data",
		Factory: f,
		Hex:     false,
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}

	if opts.Data != "test-data" {
		t.Errorf("Data = %q, want %q", opts.Data, "test-data")
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}

	if opts.Hex {
		t.Error("Hex should be false")
	}
}

func TestExecute_InvalidDeviceNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-device", "data"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for nonexistent device")
	}
}

func TestExecute_InvalidDeviceWithHex(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-device", "deadbeef", "--hex"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for nonexistent device")
	}
}

func TestExecute_InvalidHexOddLength(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "48656c6c6f6", "--hex"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for odd-length hex string")
	}

	if !strings.Contains(err.Error(), "odd number of characters") {
		t.Errorf("expected 'odd number of characters' error, got: %v", err)
	}
}

func TestExecute_InvalidHexCharacter(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "48656c6c6fGG", "--hex"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid hex character")
	}

	if !strings.Contains(err.Error(), "invalid hex") {
		t.Errorf("expected 'invalid hex' error, got: %v", err)
	}
}

func TestExecute_HexParsingLowercase(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	// Lowercase hex should be parsed correctly
	cmd.SetArgs([]string{"test-device", "abcdef", "--hex"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	// Will fail due to device not found, but should parse hex successfully
	if err == nil {
		t.Fatal("Expected error for nonexistent device")
	}
	// Should not contain hex parsing error
	if strings.Contains(err.Error(), "invalid hex") {
		t.Errorf("should parse lowercase hex, got error: %v", err)
	}
}

func TestExecute_HexParsingUppercase(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	// Uppercase hex should be parsed correctly
	cmd.SetArgs([]string{"test-device", "ABCDEF", "--hex"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	// Will fail due to device not found, but should parse hex successfully
	if err == nil {
		t.Fatal("Expected error for nonexistent device")
	}
	// Should not contain hex parsing error
	if strings.Contains(err.Error(), "invalid hex") {
		t.Errorf("should parse uppercase hex, got error: %v", err)
	}
}
