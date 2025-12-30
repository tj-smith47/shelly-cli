package inclusion

import (
	"bytes"
	"strings"
	"testing"

	_ "github.com/tj-smith47/shelly-go/profiles/wave" // Register wave profiles

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

	if cmd.Use != "inclusion <model>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "inclusion <model>")
	}

	wantAliases := []string{"include", "pair", "add"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
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
		{"no args", []string{}, true},
		{"one arg valid", []string{"model"}, false},
		{"two args", []string{"model1", "model2"}, true},
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

	flag := cmd.Flags().Lookup("mode")
	if flag == nil {
		t.Fatal("--mode flag not found")
	}
	if flag.DefValue != "smart_start" {
		t.Errorf("--mode default = %q, want %q", flag.DefValue, "smart_start")
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
		"shelly zwave inclusion",
		"--mode button",
		"--mode switch",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "button",
		Factory: f,
	}

	if opts.Model != "SNSW-001P16ZW" {
		t.Errorf("Model = %q, want %q", opts.Model, "SNSW-001P16ZW")
	}

	if opts.Mode != "button" {
		t.Errorf("Mode = %q, want %q", opts.Mode, "button")
	}
}

func TestRun_UnknownModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Model:   "UNKNOWN-MODEL-12345",
		Mode:    "smart_start",
	}

	err := run(opts)
	if err == nil {
		t.Error("Expected error for unknown model")
	}
	if !strings.Contains(err.Error(), "unknown device model") {
		t.Errorf("Error should mention 'unknown device model': %v", err)
	}
}

func TestRun_InvalidMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16ZW", // Valid ZWave model
		Mode:    "invalid_mode",
	}

	err := run(opts)
	if err == nil {
		t.Error("Expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("Error should mention 'invalid mode': %v", err)
	}
}

func TestRun_NonZWaveDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// SNSW-001P16EU is not a ZWave device
	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16EU",
		Mode:    "smart_start",
	}

	err := run(opts)
	// Should fail at profile lookup since SNSW-001P16EU might not exist
	// or fail at ZWave check
	if err == nil {
		t.Error("Expected error for non-ZWave or unknown device")
	}
}

// Execute-based tests for coverage improvement

func TestExecute_SmartStartMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Test with valid ZWave model and smart_start mode (default)
	cmd.SetArgs([]string{"SNSW-001P16ZW"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for valid Z-Wave model")
	}

	// Verify output contains expected sections
	if !strings.Contains(output, "Z-Wave") {
		t.Error("Output should contain 'Z-Wave'")
	}
	if !strings.Contains(output, "Device") {
		t.Error("Output should contain device information")
	}
	if !strings.Contains(output, "Instructions") {
		t.Error("Output should contain inclusion instructions")
	}
}

func TestExecute_ButtonMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Test with button inclusion mode
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "button"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for valid Z-Wave model with button mode")
	}

	// Verify output contains mode information
	if !strings.Contains(output, "Inclusion Instructions") {
		t.Error("Output should contain inclusion instructions")
	}
}

func TestExecute_ButtonModeAlternative(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Test button mode with alternate name "s"
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "s"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output with 's' mode (button shorthand)")
	}
}

func TestExecute_SwitchMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Test with switch inclusion mode
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "switch"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for valid Z-Wave model with switch mode")
	}
}

func TestExecute_SmartStartAlternativeName(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Test smart_start mode with alternate name "smartstart"
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "smartstart"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output with 'smartstart' mode")
	}
}

func TestExecute_SmartStartQRName(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Test smart_start mode with alternate name "qr"
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "qr"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output with 'qr' mode")
	}
}

func TestExecute_UnknownModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetArgs([]string{"UNKNOWN-MODEL-XYZ"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute should error for unknown model")
	}
}

func TestExecute_InvalidMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "invalid_mode"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute should error for invalid mode")
	}

	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("Error should mention 'invalid mode', got: %v", err)
	}
}

func TestExecute_ModeParsingCaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		mode  string
		model string
	}{
		{"button uppercase", "BUTTON", "SNSW-001P16ZW"},
		{"switch uppercase", "SWITCH", "SNSW-001P16ZW"},
		{"qr uppercase", "QR", "SNSW-001P16ZW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)

			cmd.SetArgs([]string{tt.model, "--mode", tt.mode})

			err := cmd.Execute()
			if err != nil {
				t.Errorf("Execute error for mode %q: %v", tt.mode, err)
			}

			output := tf.OutString()
			if output == "" {
				t.Errorf("Expected output for mode %q", tt.mode)
			}
		})
	}
}

func TestRun_ValidZWaveModel_SmartStart(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "smart_start",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("Expected no error for valid Z-Wave model, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for valid Z-Wave model")
	}
}

func TestRun_ValidZWaveModel_Button(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "button",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("Expected no error for button mode, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for button mode")
	}
}

func TestRun_ValidZWaveModel_Switch(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "switch",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("Expected no error for switch mode, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for switch mode")
	}
}

func TestRun_OutputContainsDeviceInfo(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "smart_start",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := tf.OutString()

	// Verify output contains expected sections
	if !strings.Contains(output, "Z-Wave") {
		t.Error("Output should contain 'Z-Wave'")
	}
	if !strings.Contains(output, "Device") {
		t.Error("Output should contain 'Device'")
	}
}

func TestRun_OutputContainsMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "button",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := tf.OutString()

	// Verify output contains mode information
	if !strings.Contains(output, "Mode") {
		t.Error("Output should contain 'Mode'")
	}
}

func TestRun_OutputContainsInstructions(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "smart_start",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := tf.OutString()

	// Verify output contains steps
	if !strings.Contains(output, "Steps") {
		t.Error("Output should contain 'Steps'")
	}
}

func TestNewCommand_RunE_CallsRunWithModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Set a valid model as argument
	cmd.SetArgs([]string{"SNSW-001P16ZW"})

	// RunE should set the model and call run
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("RunE should have executed run function")
	}
}

func TestNewCommand_RunE_WithFlagMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Set model and mode flag
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "button"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("RunE with flag should have executed")
	}
}

func TestExecute_WithAlias_Include(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Verify the command name works
	if cmd.Name() != "inclusion" {
		t.Errorf("Command name = %q, want 'inclusion'", cmd.Name())
	}
}

func TestExecute_WithAlias_Pair(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify aliases are set
	hasAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "pair" {
			hasAlias = true
			break
		}
	}

	if !hasAlias {
		t.Error("Expected 'pair' alias")
	}
}

func TestExecute_WithAlias_Add(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify aliases are set
	hasAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "add" {
			hasAlias = true
			break
		}
	}

	if !hasAlias {
		t.Error("Expected 'add' alias")
	}
}
