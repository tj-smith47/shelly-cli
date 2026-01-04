package iostreams_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestSystem(t *testing.T) {
	t.Parallel()

	ios := iostreams.System()

	if ios == nil {
		t.Fatal("System() returned nil")
	}

	if ios.In == nil {
		t.Error("In is nil")
	}
	if ios.Out == nil {
		t.Error("Out is nil")
	}
	if ios.ErrOut == nil {
		t.Error("ErrOut is nil")
	}
}

func TestTest(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	ios := iostreams.Test(in, out, errOut)

	if ios.In != in {
		t.Error("In not set correctly")
	}
	if ios.Out != out {
		t.Error("Out not set correctly")
	}
	if ios.ErrOut != errOut {
		t.Error("ErrOut not set correctly")
	}

	// Test streams should not be TTY
	if ios.IsStdinTTY() {
		t.Error("Test stdin should not be TTY")
	}
	if ios.IsStdoutTTY() {
		t.Error("Test stdout should not be TTY")
	}
	if ios.IsStderrTTY() {
		t.Error("Test stderr should not be TTY")
	}

	// Color should be disabled in tests
	if ios.ColorEnabled() {
		t.Error("Color should be disabled in tests")
	}
}

func TestIOStreams_TTYSetters(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// Test SetStdinTTY
	ios.SetStdinTTY(true)
	if !ios.IsStdinTTY() {
		t.Error("SetStdinTTY(true) failed")
	}
	ios.SetStdinTTY(false)
	if ios.IsStdinTTY() {
		t.Error("SetStdinTTY(false) failed")
	}

	// Test SetStdoutTTY
	ios.SetStdoutTTY(true)
	if !ios.IsStdoutTTY() {
		t.Error("SetStdoutTTY(true) failed")
	}
	ios.SetStdoutTTY(false)
	if ios.IsStdoutTTY() {
		t.Error("SetStdoutTTY(false) failed")
	}

	// Test SetStderrTTY
	ios.SetStderrTTY(true)
	if !ios.IsStderrTTY() {
		t.Error("SetStderrTTY(true) failed")
	}
	ios.SetStderrTTY(false)
	if ios.IsStderrTTY() {
		t.Error("SetStderrTTY(false) failed")
	}
}

func TestIOStreams_ColorEnabled(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// Initially disabled
	if ios.ColorEnabled() {
		t.Error("Color should be disabled by default in tests")
	}

	// Enable color
	ios.SetColorEnabled(true)
	if !ios.ColorEnabled() {
		t.Error("SetColorEnabled(true) failed")
	}

	// Disable color
	ios.SetColorEnabled(false)
	if ios.ColorEnabled() {
		t.Error("SetColorEnabled(false) failed")
	}
}

func TestIOStreams_Quiet(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// Initially not quiet
	if ios.IsQuiet() {
		t.Error("Should not be quiet by default")
	}

	// Enable quiet
	ios.SetQuiet(true)
	if !ios.IsQuiet() {
		t.Error("SetQuiet(true) failed")
	}

	// Disable quiet
	ios.SetQuiet(false)
	if ios.IsQuiet() {
		t.Error("SetQuiet(false) failed")
	}
}

func TestIOStreams_CanPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stdinTTY  bool
		stdoutTTY bool
		want      bool
	}{
		{
			name:      "both TTY",
			stdinTTY:  true,
			stdoutTTY: true,
			want:      true,
		},
		{
			name:      "stdin not TTY",
			stdinTTY:  false,
			stdoutTTY: true,
			want:      false,
		},
		{
			name:      "stdout not TTY",
			stdinTTY:  true,
			stdoutTTY: false,
			want:      false,
		},
		{
			name:      "neither TTY",
			stdinTTY:  false,
			stdoutTTY: false,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ios := iostreams.Test(nil, nil, nil)
			ios.SetStdinTTY(tt.stdinTTY)
			ios.SetStdoutTTY(tt.stdoutTTY)

			if got := ios.CanPrompt(); got != tt.want {
				t.Errorf("CanPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIOStreams_Printf(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Printf("Hello %s", "World")

	if got := out.String(); got != "Hello World" {
		t.Errorf("Printf() output = %q, want %q", got, "Hello World")
	}
}

func TestIOStreams_Println(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	ios.Println("Hello", "World")

	if got := out.String(); got != "Hello World\n" {
		t.Errorf("Println() output = %q, want %q", got, "Hello World\n")
	}
}

func TestIOStreams_Errorf(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	ios.Errorf("Error: %s", "test")

	if got := errOut.String(); got != "Error: test" {
		t.Errorf("Errorf() output = %q, want %q", got, "Error: test")
	}
}

func TestIOStreams_Errorln(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	ios.Errorln("Error", "message")

	if got := errOut.String(); got != "Error message\n" {
		t.Errorf("Errorln() output = %q, want %q", got, "Error message\n")
	}
}

func TestIOStreams_StartProgress_NonTTY(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)
	ios.SetStderrTTY(false)

	ios.StartProgress("Loading...")

	// For non-TTY, should print message directly
	if got := errOut.String(); got != "Loading...\n" {
		t.Errorf("StartProgress() output = %q, want %q", got, "Loading...\n")
	}

	// Stop should be safe to call
	ios.StopProgress()
}

func TestIOStreams_StartProgress_Quiet(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)
	ios.SetQuiet(true)

	ios.StartProgress("Loading...")

	// In quiet mode, should not print anything
	if got := errOut.String(); got != "" {
		t.Errorf("StartProgress() in quiet mode should not output, got %q", got)
	}
}

func TestIOStreams_Progress_TTY(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)
	ios.SetStderrTTY(true)

	// Start progress - this starts a spinner goroutine
	ios.StartProgress("Loading...")

	// Update progress
	ios.UpdateProgress("Still loading...")

	// Stop with success - verifies no panic and spinner state handling
	ios.StopProgressWithSuccess("Done!")

	// For TTY mode with real spinner, output depends on timing
	// The important thing is these calls don't panic
}

func TestIOStreams_StopProgressWithError(t *testing.T) {
	t.Parallel()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)
	ios.SetStderrTTY(true)

	ios.StartProgress("Loading...")
	ios.StopProgressWithError("Failed!")

	// For TTY mode with real spinner, output depends on timing
	// The important thing is these calls don't panic
}

func TestIOStreams_StopProgress_NilIndicator(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// These should not panic when called without StartProgress
	ios.StopProgress()
	ios.StopProgressWithSuccess("Done")
	ios.StopProgressWithError("Failed")
	ios.UpdateProgress("Update")
}

// Tests for color environment variables - these test the internal
// isColorDisabled and isColorForced functions indirectly via System().
// Note: These tests can't run in parallel because they modify environment variables.

func TestSystem_ColorDisabled_NO_COLOR(t *testing.T) {
	// Set NO_COLOR env var
	t.Setenv("NO_COLOR", "1")

	ios := iostreams.System()

	// Color should be disabled when NO_COLOR is set
	if ios.ColorEnabled() {
		t.Error("Color should be disabled when NO_COLOR is set")
	}
}

func TestSystem_ColorDisabled_SHELLY_NO_COLOR(t *testing.T) {
	t.Setenv("SHELLY_NO_COLOR", "1")

	ios := iostreams.System()

	if ios.ColorEnabled() {
		t.Error("Color should be disabled when SHELLY_NO_COLOR is set")
	}
}

func TestSystem_ColorDisabled_TERM_dumb(t *testing.T) {
	t.Setenv("TERM", "dumb")

	ios := iostreams.System()

	if ios.ColorEnabled() {
		t.Error("Color should be disabled when TERM=dumb")
	}
}

func TestSystem_ColorForced_FORCE_COLOR(t *testing.T) {
	// First disable via NO_COLOR
	t.Setenv("NO_COLOR", "1")
	// Then force via FORCE_COLOR
	t.Setenv("FORCE_COLOR", "1")

	ios := iostreams.System()

	// FORCE_COLOR should override NO_COLOR
	if !ios.ColorEnabled() {
		t.Error("Color should be enabled when FORCE_COLOR is set")
	}
}

func TestSystem_ColorForced_SHELLY_FORCE_COLOR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("SHELLY_FORCE_COLOR", "1")

	ios := iostreams.System()

	if !ios.ColorEnabled() {
		t.Error("Color should be enabled when SHELLY_FORCE_COLOR is set")
	}
}

func TestIOStreams_PlainMode(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)

	// IsPlainMode only returns true when explicitly set via --plain flag
	// Non-TTY streams use no-color style (ASCII borders) instead
	if ios.IsPlainMode() {
		t.Error("Test streams should not be in plain mode by default")
	}

	// Test SetPlainMode
	ios.SetPlainMode(true)
	if !ios.IsPlainMode() {
		t.Error("SetPlainMode(true) should result in IsPlainMode() returning true")
	}

	// After setting plain to false, IsPlainMode should be false
	ios.SetPlainMode(false)
	if ios.IsPlainMode() {
		t.Error("SetPlainMode(false) should result in IsPlainMode() returning false")
	}
}

func TestIOStreams_IsPlainMode_Conditions(t *testing.T) {
	t.Parallel()

	// IsPlainMode only returns true when explicitly set via --plain flag
	// Non-TTY and no-color use ASCII borders (no-color style), not plain mode
	tests := []struct {
		name         string
		plainMode    bool
		stdoutTTY    bool
		colorEnabled bool
		want         bool
	}{
		{
			name:         "plain mode explicitly set",
			plainMode:    true,
			stdoutTTY:    true,
			colorEnabled: true,
			want:         true,
		},
		{
			name:         "non-TTY stdout - not plain mode",
			plainMode:    false,
			stdoutTTY:    false,
			colorEnabled: true,
			want:         false, // Non-TTY uses no-color style, not plain
		},
		{
			name:         "color disabled - not plain mode",
			plainMode:    false,
			stdoutTTY:    true,
			colorEnabled: false,
			want:         false, // No-color uses ASCII borders, not plain
		},
		{
			name:         "TTY with color - not plain mode",
			plainMode:    false,
			stdoutTTY:    true,
			colorEnabled: true,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ios := iostreams.Test(nil, nil, nil)
			ios.SetPlainMode(tt.plainMode)
			ios.SetStdoutTTY(tt.stdoutTTY)
			ios.SetColorEnabled(tt.colorEnabled)

			if got := ios.IsPlainMode(); got != tt.want {
				t.Errorf("IsPlainMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// IsColorDisabled Tests
// =============================================================================

//nolint:paralleltest // Tests modify global viper state
func TestIsColorDisabled_NoColorFlag(t *testing.T) {
	viper.Reset()
	viper.Set("no-color", true)
	defer viper.Reset()

	if !iostreams.IsColorDisabled() {
		t.Error("IsColorDisabled() should return true when no-color flag is set")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestIsColorDisabled_PlainFlag(t *testing.T) {
	viper.Reset()
	viper.Set("plain", true)
	defer viper.Reset()

	if !iostreams.IsColorDisabled() {
		t.Error("IsColorDisabled() should return true when plain flag is set")
	}
}

func TestIsColorDisabled_TermDumb(t *testing.T) {
	viper.Reset()
	t.Setenv("TERM", "dumb")
	defer viper.Reset()

	if !iostreams.IsColorDisabled() {
		t.Error("IsColorDisabled() should return true when TERM=dumb")
	}
}

func TestIsColorDisabled_NoColorEnv(t *testing.T) {
	viper.Reset()
	t.Setenv("NO_COLOR", "1")
	defer viper.Reset()

	if !iostreams.IsColorDisabled() {
		t.Error("IsColorDisabled() should return true when NO_COLOR is set")
	}
}

func TestIsColorDisabled_ShellyNoColorEnv(t *testing.T) {
	viper.Reset()
	t.Setenv("SHELLY_NO_COLOR", "1")
	defer viper.Reset()

	if !iostreams.IsColorDisabled() {
		t.Error("IsColorDisabled() should return true when SHELLY_NO_COLOR is set")
	}
}

func TestIsColorDisabled_NothingSet(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Unset environment variables that might be set
	// Must use os.Unsetenv because t.Setenv("VAR", "") still counts as "set"
	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Logf("warning: could not unset NO_COLOR: %v", err)
	}
	if err := os.Unsetenv("SHELLY_NO_COLOR"); err != nil {
		t.Logf("warning: could not unset SHELLY_NO_COLOR: %v", err)
	}
	t.Setenv("TERM", "xterm-256color")

	if iostreams.IsColorDisabled() {
		t.Error("IsColorDisabled() should return false when no color-disable settings are present")
	}
}

// =============================================================================
// ClearScreen Tests
// =============================================================================

func TestIOStreams_ClearScreen_NonTTY(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	// Test() sets isStdoutTTY to false by default

	ios.ClearScreen()

	// For non-TTY, ClearScreen should be a no-op
	if out.Len() != 0 {
		t.Errorf("ClearScreen() should not output for non-TTY, got %q", out.String())
	}
}

func TestIOStreams_ClearScreen_TTY(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	ios.SetStdoutTTY(true)

	ios.ClearScreen()

	// For TTY, should output ANSI escape codes
	if !strings.Contains(out.String(), "\033[") {
		t.Errorf("ClearScreen() should output ANSI escape codes for TTY, got %q", out.String())
	}
}

// =============================================================================
// PromptTypedInput Tests
// =============================================================================

func TestIOStreams_PromptTypedInput_NonTTY(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(nil, nil, nil)
	// Test() creates non-TTY iostreams, so CanPrompt() returns false

	result, err := ios.PromptTypedInput("Enter value:", "default", "string")

	if err != nil {
		t.Errorf("PromptTypedInput() error = %v, want nil", err)
	}
	// Non-TTY should return the default value
	if result != "default" {
		t.Errorf("PromptTypedInput() = %v, want 'default'", result)
	}
}

// =============================================================================
// PromptScriptVariables Tests
// =============================================================================

func TestIOStreams_PromptScriptVariables_NoVariables(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	variables := []config.ScriptVariable{}
	result := ios.PromptScriptVariables(variables, true)

	if len(result) != 0 {
		t.Errorf("PromptScriptVariables() should return empty map for no variables, got %v", result)
	}
}

func TestIOStreams_PromptScriptVariables_WithDefaults(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)

	variables := []config.ScriptVariable{
		{Name: "THRESHOLD", Default: 75, Type: "number"},
		{Name: "ENABLED", Default: true, Type: "boolean"},
		{Name: "NAME", Default: "test", Type: "string"},
	}
	result := ios.PromptScriptVariables(variables, false) // configure = false means use defaults

	if result["THRESHOLD"] != 75 {
		t.Errorf("result[THRESHOLD] = %v, want 75", result["THRESHOLD"])
	}
	if result["ENABLED"] != true {
		t.Errorf("result[ENABLED] = %v, want true", result["ENABLED"])
	}
	if result["NAME"] != "test" {
		t.Errorf("result[NAME] = %v, want 'test'", result["NAME"])
	}
}

func TestIOStreams_PromptScriptVariables_NonTTY(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, nil)
	// Test() creates non-TTY iostreams, so CanPrompt() returns false

	variables := []config.ScriptVariable{
		{Name: "VAR1", Default: "value1"},
	}
	result := ios.PromptScriptVariables(variables, true) // configure = true, but non-TTY

	// Non-TTY should return defaults even when configure is true
	if result["VAR1"] != "value1" {
		t.Errorf("result[VAR1] = %v, want 'value1'", result["VAR1"])
	}
}

// =============================================================================
// LogLevel Additional Tests
// =============================================================================

func TestLogLevel_String_AllLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level    iostreams.LogLevel
		expected string
	}{
		{iostreams.LevelTrace, "trace"},
		{iostreams.LevelDebug, "debug"},
		{iostreams.LevelInfo, "info"},
		{iostreams.LevelWarn, "warn"},
		{iostreams.LevelError, "error"},
		{iostreams.LevelNone, "none"},
		{iostreams.LogLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel_AllVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected iostreams.LogLevel
	}{
		{"trace", iostreams.LevelTrace},
		{"TRACE", iostreams.LevelTrace},
		{"debug", iostreams.LevelDebug},
		{"DEBUG", iostreams.LevelDebug},
		{"info", iostreams.LevelInfo},
		{"INFO", iostreams.LevelInfo},
		{"warn", iostreams.LevelWarn},
		{"WARN", iostreams.LevelWarn},
		{"warning", iostreams.LevelWarn},
		{"WARNING", iostreams.LevelWarn},
		{"error", iostreams.LevelError},
		{"ERROR", iostreams.LevelError},
		{"none", iostreams.LevelNone},
		{"NONE", iostreams.LevelNone},
		{"off", iostreams.LevelNone},
		{"OFF", iostreams.LevelNone},
		{"silent", iostreams.LevelNone},
		{"SILENT", iostreams.LevelNone},
		{"unknown", iostreams.LevelDebug},
		{"", iostreams.LevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := iostreams.ParseLogLevel(tt.input); got != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// ConfigureLogger Tests
// =============================================================================

//nolint:paralleltest // Tests modify global viper state
func TestConfigureLogger_Verbosity(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 3)
	defer viper.Reset()

	// ConfigureLogger should not panic
	iostreams.ConfigureLogger()
}

//nolint:paralleltest // Tests modify global viper state
func TestConfigureLogger_LogLevel(t *testing.T) {
	viper.Reset()
	viper.Set("log.level", "trace")
	defer viper.Reset()

	// ConfigureLogger should not panic
	iostreams.ConfigureLogger()
}

//nolint:paralleltest // Tests modify global viper state
func TestConfigureLogger_JSONMode(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 1)
	viper.Set("log.json", true)
	defer viper.Reset()

	// ConfigureLogger should not panic
	iostreams.ConfigureLogger()
}

//nolint:paralleltest // Tests modify global viper state
func TestConfigureLogger_Categories(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 1)
	viper.Set("log.categories", "network,device,api")
	defer viper.Reset()

	// ConfigureLogger should not panic
	iostreams.ConfigureLogger()
}

// =============================================================================
// Logger.SetCategories Tests
// =============================================================================

func TestLogger_SetCategories_NilClearsFilter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := iostreams.NewLogger(&buf)

	// Set some categories
	logger.SetCategories([]iostreams.LogCategory{iostreams.CategoryDevice})

	// Clear with nil
	logger.SetCategories(nil)

	// Log should work for any category now
	logger.Log(iostreams.LevelInfo, iostreams.CategoryNetwork, "test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Error("Log should work for any category after SetCategories(nil)")
	}
}

// =============================================================================
// Logger.LogWithData Text Mode Tests
// =============================================================================

func TestLogger_LogWithData_TextMode(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := iostreams.NewLogger(&buf)
	logger.SetJSONMode(false)

	data := map[string]string{"key": "value"}
	logger.LogWithData(iostreams.LevelInfo, iostreams.CategoryDevice, "message with data", data)

	output := buf.String()
	// In text mode, data is part of the LogEntry but writeTextEntry doesn't display it
	if !strings.Contains(output, "message with data") {
		t.Errorf("output should contain message, got %q", output)
	}
}

// =============================================================================
// GetVerbosity and IsVerbose Tests
// =============================================================================

//nolint:paralleltest // Tests modify global viper state
func TestGetVerbosity(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 5)
	defer viper.Reset()

	if got := iostreams.GetVerbosity(); got != 5 {
		t.Errorf("GetVerbosity() = %d, want 5", got)
	}
}

// =============================================================================
// Package-level Log Functions Tests
// =============================================================================

//nolint:paralleltest // Tests modify global viper state
func TestPackageLevel_LogErr(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 1)
	defer viper.Reset()

	// Should not panic
	err := &testError{msg: "test error"}
	iostreams.LogErr(iostreams.LevelError, iostreams.CategoryDevice, "operation failed", err)
}

//nolint:paralleltest // Tests modify global viper state
func TestPackageLevel_LogErr_NotVerbose(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 0)
	defer viper.Reset()

	// Should not panic and produce no output when not verbose
	err := &testError{msg: "test error"}
	iostreams.LogErr(iostreams.LevelError, iostreams.CategoryDevice, "operation failed", err)
}

//nolint:paralleltest // Tests modify global viper state
func TestPackageLevel_LogWithData(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 1)
	defer viper.Reset()

	// Should not panic
	data := map[string]int{"count": 42}
	iostreams.LogWithData(iostreams.LevelInfo, iostreams.CategoryAPI, "response", data)
}

//nolint:paralleltest // Tests modify global viper state
func TestPackageLevel_LogWithData_NotVerbose(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 0)
	defer viper.Reset()

	// Should not panic and produce no output when not verbose
	data := map[string]int{"count": 42}
	iostreams.LogWithData(iostreams.LevelInfo, iostreams.CategoryAPI, "response", data)
}

// =============================================================================
// IOStreams.Logger Tests
// =============================================================================

//nolint:paralleltest // Tests modify global viper state
func TestIOStreams_Logger_VerboseMode(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 2)
	defer viper.Reset()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	logger := ios.Logger()
	if logger == nil {
		t.Fatal("Logger() should not return nil")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestIOStreams_Logger_NotVerbose(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 0)
	defer viper.Reset()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	logger := ios.Logger()
	if logger == nil {
		t.Fatal("Logger() should not return nil")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestIOStreams_LogWithData(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 1)
	viper.Set("log.json", true)
	defer viper.Reset()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	data := map[string]int{"count": 42}
	ios.LogWithData(iostreams.LevelInfo, iostreams.CategoryAPI, "response", data)

	// Check output contains the data in JSON format
	output := errOut.String()
	if !strings.Contains(output, `"data"`) {
		t.Errorf("LogWithData output should contain data, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestIOStreams_LogWithData_NotVerbose(t *testing.T) {
	viper.Reset()
	viper.Set("verbosity", 0)
	defer viper.Reset()

	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, nil, errOut)

	data := map[string]int{"count": 42}
	ios.LogWithData(iostreams.LevelInfo, iostreams.CategoryAPI, "response", data)

	if errOut.Len() > 0 {
		t.Errorf("LogWithData should produce no output when not verbose, got %q", errOut.String())
	}
}

// testError is a simple error implementation for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
