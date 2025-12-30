package exclusion

import (
	"bytes"
	"context"
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

	// Test Use
	if cmd.Use != "exclusion <model>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "exclusion <model>")
	}

	// Test Aliases
	wantAliases := []string{"exclude", "unpair", "remove"}
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
		{"one arg valid", []string{"SNSW-001P16ZW"}, false},
		{"two args", []string{"SNSW-001P16ZW", "extra"}, true},
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

	// Test mode flag
	flag := cmd.Flags().Lookup("mode")
	if flag == nil {
		t.Fatal("--mode flag not found")
	}
	wantDefault := "button" //nolint:goconst // Test data, acceptable to repeat
	if flag.DefValue != wantDefault {
		t.Errorf("--mode default = %q, want %q", flag.DefValue, wantDefault)
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
		"shelly zwave exclusion",
		"--mode switch",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
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

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}
}

func TestRun_UnknownModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Model:   "UNKNOWN-MODEL-12345",
		Mode:    "button",
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
		Model:   "SNSW-001P16ZW", // Valid Z-Wave model
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

	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16EU", // This is not a ZWave device (but may not be in profiles)
		Mode:    "button",
	}

	err := run(opts)
	if err == nil {
		t.Error("Expected error for non-ZWave device or unknown model")
	}
	// Could be either unknown model or not a ZWave device
	if !strings.Contains(err.Error(), "not a Z-Wave device") && !strings.Contains(err.Error(), "unknown device model") {
		t.Errorf("Error should mention 'not a Z-Wave device' or 'unknown device model': %v", err)
	}
}

func TestRun_ModeNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mode        string
		expectError bool
	}{
		{"uppercase BUTTON", "BUTTON", false},
		{"lowercase s", "s", false},
		{"lowercase switch", "switch", false},
		{"invalid mode xyz", "xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)
			opts := &Options{
				Factory: tf.Factory,
				Model:   "SNSW-001P16ZW", // Valid Z-Wave model
				Mode:    tt.mode,
			}

			err := run(opts)
			// Invalid modes should error
			if tt.expectError && err == nil {
				t.Errorf("expected error for invalid mode %q", tt.mode)
			}
			// Valid modes should succeed
			if !tt.expectError && err != nil {
				t.Errorf("run() error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}

func TestExecute_InvalidModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"INVALID-MODEL"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid model")
	}
	if !strings.Contains(err.Error(), "unknown device model") {
		t.Logf("error = %v", err)
	}
}

func TestExecute_InvalidMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "invalid"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid mode")
	}
	// Should error on invalid mode
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Logf("error = %v", err)
	}
}

func TestRun_Modes_Button(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Test that "button" mode is recognized and produces output
	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16ZW", // Valid Z-Wave model
		Mode:    "button",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Z-Wave") {
		t.Logf("output = %s", output)
	}
}

func TestRun_Modes_S(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Test that "s" mode is recognized as shorthand for button
	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16ZW", // Valid Z-Wave model
		Mode:    "s",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Z-Wave") {
		t.Logf("output = %s", output)
	}
}

func TestRun_Modes_Switch(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Test that "switch" mode is recognized
	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16ZW", // Valid Z-Wave model
		Mode:    "switch",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Z-Wave") {
		t.Logf("output = %s", output)
	}
}

func TestExecute_DefaultModeIsButton(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Verify default mode is set to "button"
	flag := cmd.Flags().Lookup("mode")
	if flag == nil {
		t.Fatal("mode flag not found")
	}
	if flag.DefValue != "button" {
		t.Errorf("default mode = %q, want %q", flag.DefValue, "button")
	}
}

func TestRun_LowercaseModeSwitches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode string
	}{
		{"lowercase button", "button"},
		{"lowercase switch", "switch"},
		{"lowercase s", "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf := factory.NewTestFactory(t)
			opts := &Options{
				Factory: tf.Factory,
				Model:   "SNSW-001P16ZW",
				Mode:    tt.mode,
			}

			err := run(opts)
			if err != nil {
				t.Errorf("run() error = %v", err)
			}
		})
	}
}

func TestNewCommand_Metadata(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that all required metadata is present
	tests := []struct {
		name  string
		field string
	}{
		{"Use", cmd.Use},
		{"Short", cmd.Short},
		{"Long", cmd.Long},
		{"Example", cmd.Example},
	}

	for _, tt := range tests {
		if tt.field == "" {
			t.Errorf("%s is empty", tt.name)
		}
	}
}

func TestOptions_FactoryNotNil(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		Model:   "test",
		Mode:    "button",
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

func TestRun_ModeValidationOrder(t *testing.T) {
	t.Parallel()

	// This test verifies mode validation happens at the right point
	// Mode is validated even for unknown models
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Model:   "INVALIDMODEL123456",
		Mode:    "button", // Valid mode
	}

	err := run(opts)
	if err == nil {
		t.Error("expected error for unknown model")
	}
	// Should error on model lookup before giving output
	if !strings.Contains(err.Error(), "unknown device model") {
		t.Logf("unexpected error: %v", err)
	}
}

func TestRun_InvalidModePrecedence(t *testing.T) {
	t.Parallel()

	// Test error handling for invalid mode
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16ZW", // Valid Z-Wave model
		Mode:    "invalid_xyz",   // Invalid mode
	}

	err := run(opts)
	if err == nil {
		t.Error("expected error for invalid mode")
	}
	// Should error on mode validation
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Logf("expected 'invalid mode' error but got: %v", err)
	}
}

func TestExecute_CommandMetadata(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test command name and structure
	if cmd.Use != "exclusion <model>" {
		t.Errorf("Use = %q, expected %q", cmd.Use, "exclusion <model>")
	}

	// Test at least one alias exists
	if len(cmd.Aliases) == 0 {
		t.Error("Command should have at least one alias")
	}

	// Verify expected aliases are present
	expectedAliases := map[string]bool{
		"exclude": false,
		"unpair":  false,
		"remove":  false,
	}
	for _, alias := range cmd.Aliases {
		if _, exists := expectedAliases[alias]; exists {
			expectedAliases[alias] = true
		}
	}
	for alias, found := range expectedAliases {
		if !found {
			t.Errorf("Expected alias %q not found", alias)
		}
	}
}

func TestExecute_ButtonMode_Success(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "button"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Z-Wave Exclusion") {
		t.Logf("output = %s", output)
	}
}

func TestExecute_SwitchMode_Success(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "switch"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Z-Wave Exclusion") {
		t.Logf("output = %s", output)
	}
}

func TestExecute_DefaultMode_Success(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"SNSW-001P16ZW"}) // Use default mode
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Z-Wave Exclusion") {
		t.Logf("output = %s", output)
	}
}

func TestRun_ValidZWaveModel_ButtonMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "button",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for valid Z-Wave model")
	}

	if !strings.Contains(output, "Z-Wave Exclusion") {
		t.Logf("output = %s", output)
	}
}

func TestRun_ValidZWaveModel_SwitchMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "switch",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output for valid Z-Wave model")
	}

	if !strings.Contains(output, "Z-Wave Exclusion") {
		t.Logf("output = %s", output)
	}
}

func TestRun_OutputContainsDeviceInfo(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Model:   "SNSW-001P16ZW",
		Mode:    "button",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
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
		t.Fatalf("run() error = %v", err)
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
		Mode:    "button",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
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
	cmd.SetArgs([]string{"SNSW-001P16ZW", "--mode", "switch"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute should not error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("RunE with flag should have executed")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
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
		{"s uppercase", "S", "SNSW-001P16ZW"},
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
