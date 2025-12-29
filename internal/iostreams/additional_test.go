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
