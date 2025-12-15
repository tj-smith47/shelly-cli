package iostreams

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

// Test string constants to satisfy goconst linter.
const (
	testLevelInfo    = "info"
	testLevelError   = "error"
	testCatDevice    = "device"
	testCatNetwork   = "network"
	testMessage      = "test message"
	testErrorMessage = "test error"
)

func TestLogLevel_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, testLevelInfo},
		{LevelWarn, "warn"},
		{LevelError, testLevelError},
		{LogLevel(99), "unknown"},
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

func TestParseLogLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{testLevelInfo, LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{testLevelError, LevelError},
		{"ERROR", LevelError},
		{"invalid", LevelDebug}, // defaults to debug
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := ParseLogLevel(tt.input); got != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLogger_Log(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	logger.Log(LevelInfo, CategoryDevice, testMessage)

	output := buf.String()
	if !strings.Contains(output, "info:device: test message") {
		t.Errorf("Log output = %q, want info:device prefix", output)
	}
}

func TestLogger_LogErr(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	err := &testError{msg: testErrorMessage}
	logger.LogErr(LevelError, CategoryNetwork, "connection failed", err)

	output := buf.String()
	if !strings.Contains(output, "error:network: connection failed: test error") {
		t.Errorf("LogErr output = %q, want error:network prefix with error", output)
	}
}

func TestLogger_LogErr_NilError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	logger.LogErr(LevelError, CategoryNetwork, "connection failed", nil)

	if buf.Len() > 0 {
		t.Errorf("LogErr with nil error should produce no output, got %q", buf.String())
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	logger.SetLevel(LevelWarn)

	logger.Log(LevelDebug, CategoryDevice, "debug message")
	logger.Log(LevelInfo, CategoryDevice, "info message")
	logger.Log(LevelWarn, CategoryDevice, "warn message")
	logger.Log(LevelError, CategoryDevice, "error message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be present")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be present")
	}
}

func TestLogger_CategoryFiltering(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	logger.SetCategories([]LogCategory{CategoryDevice})

	logger.Log(LevelInfo, CategoryDevice, "device message")
	logger.Log(LevelInfo, CategoryNetwork, "network message")
	logger.Log(LevelInfo, CategoryConfig, "config message")

	output := buf.String()
	if !strings.Contains(output, "device message") {
		t.Error("Device message should be present")
	}
	if strings.Contains(output, "network message") {
		t.Error("Network message should be filtered out")
	}
	if strings.Contains(output, "config message") {
		t.Error("Config message should be filtered out")
	}
}

func TestLogger_CategoryFiltering_AllCategories(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	logger.SetCategories(nil) // nil means all categories

	logger.Log(LevelInfo, CategoryDevice, "device message")
	logger.Log(LevelInfo, CategoryNetwork, "network message")

	output := buf.String()
	if !strings.Contains(output, "device message") {
		t.Error("Device message should be present")
	}
	if !strings.Contains(output, "network message") {
		t.Error("Network message should be present")
	}
}

func TestLogger_JSONMode(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	logger.SetJSONMode(true)

	logger.Log(LevelInfo, CategoryDevice, testMessage)

	output := buf.String()
	var entry LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry.Level != testLevelInfo {
		t.Errorf("Level = %q, want %q", entry.Level, testLevelInfo)
	}
	if entry.Category != testCatDevice {
		t.Errorf("Category = %q, want %q", entry.Category, testCatDevice)
	}
	if entry.Message != testMessage {
		t.Errorf("Message = %q, want %q", entry.Message, testMessage)
	}
	if entry.Time.IsZero() {
		t.Error("Time should be set")
	}
}

func TestLogger_JSONMode_WithError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	logger.SetJSONMode(true)

	err := &testError{msg: testErrorMessage}
	logger.LogErr(LevelError, CategoryNetwork, "operation failed", err)

	output := buf.String()
	var entry LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry.Error != testErrorMessage {
		t.Errorf("Error = %q, want %q", entry.Error, testErrorMessage)
	}
}

func TestLogger_LogWithData(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	logger.SetJSONMode(true)

	data := map[string]string{"key": "value"}
	logger.LogWithData(LevelInfo, CategoryAPI, "response received", data)

	output := buf.String()
	var entry LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry.Data == nil {
		t.Error("Data should be present")
	}
}

func TestLogger_EmptyCategory(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	logger.Log(LevelInfo, "", "message without category")

	output := buf.String()
	if !strings.Contains(output, "info: message without category") {
		t.Errorf("Output = %q, want simple 'info:' prefix without category", output)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestPackageLevel_Log(t *testing.T) {
	viper.Set("verbosity", 1)
	defer viper.Set("verbosity", 0)

	// Package-level functions use defaultLogger which writes to os.Stderr
	// We can't easily capture that, so just ensure no panic
	Log(LevelInfo, CategoryDevice, "test message")
}

//nolint:paralleltest // Tests modify global viper state
func TestPackageLevel_Log_NotVerbose(t *testing.T) {
	viper.Set("verbosity", 0)

	// Should not panic and produce no output when not verbose
	Log(LevelInfo, CategoryDevice, "test message")
}

//nolint:paralleltest // Tests modify global viper state
func TestIOStreams_Log(t *testing.T) {
	viper.Set("verbosity", 1)
	defer viper.Set("verbosity", 0)

	var buf bytes.Buffer
	ios := &IOStreams{
		ErrOut: &buf,
	}

	ios.Log(LevelInfo, CategoryDevice, "test message")

	output := buf.String()
	if !strings.Contains(output, "info:device: test message") {
		t.Errorf("IOStreams.Log output = %q, want info:device prefix", output)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestIOStreams_Log_NotVerbose(t *testing.T) {
	viper.Set("verbosity", 0)

	var buf bytes.Buffer
	ios := &IOStreams{
		ErrOut: &buf,
	}

	ios.Log(LevelInfo, CategoryDevice, "test message")

	if buf.Len() > 0 {
		t.Errorf("IOStreams.Log should produce no output when not verbose, got %q", buf.String())
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestIOStreams_LogErr(t *testing.T) {
	viper.Set("verbosity", 1)
	defer viper.Set("verbosity", 0)

	var buf bytes.Buffer
	ios := &IOStreams{
		ErrOut: &buf,
	}

	err := &testError{msg: "test error"}
	ios.LogErr(LevelError, CategoryNetwork, "connection failed", err)

	output := buf.String()
	if !strings.Contains(output, "error:network: connection failed: test error") {
		t.Errorf("IOStreams.LogErr output = %q, want error:network prefix", output)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestIOStreams_LogWithData(t *testing.T) {
	viper.Set("verbosity", 1)
	viper.Set("log.json", true)
	defer func() {
		viper.Set("verbosity", 0)
		viper.Set("log.json", false)
	}()

	var buf bytes.Buffer
	ios := &IOStreams{
		ErrOut: &buf,
	}

	data := map[string]int{"count": 42}
	ios.LogWithData(LevelInfo, CategoryAPI, "response", data)

	output := buf.String()
	var entry LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if entry.Data == nil {
		t.Error("Data should be present in JSON output")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestVerbosityToLevel(t *testing.T) {
	tests := []struct {
		verbosity int
		expected  LogLevel
	}{
		{0, LevelNone},
		{-1, LevelNone},
		{1, LevelInfo},
		{2, LevelDebug},
		{3, LevelTrace},
		{10, LevelTrace}, // anything >= 3 is trace
	}

	for _, tt := range tests {
		if got := VerbosityToLevel(tt.verbosity); got != tt.expected {
			t.Errorf("VerbosityToLevel(%d) = %v, want %v", tt.verbosity, got, tt.expected)
		}
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestIsVerbose(t *testing.T) {
	viper.Set("verbosity", 0)
	if IsVerbose() {
		t.Error("IsVerbose() should be false when verbosity=0")
	}

	viper.Set("verbosity", 1)
	if !IsVerbose() {
		t.Error("IsVerbose() should be true when verbosity=1")
	}
	viper.Set("verbosity", 0)
}

func TestLogEntry_JSON_Serialization(t *testing.T) {
	t.Parallel()
	entry := LogEntry{
		Time:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Level:    testLevelInfo,
		Category: testCatDevice,
		Message:  testMessage,
		Error:    testErrorMessage,
		Data:     map[string]string{"key": "value"},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal LogEntry: %v", err)
	}

	var parsed LogEntry
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal LogEntry: %v", err)
	}

	if parsed.Level != entry.Level {
		t.Errorf("Level = %q, want %q", parsed.Level, entry.Level)
	}
	if parsed.Category != entry.Category {
		t.Errorf("Category = %q, want %q", parsed.Category, entry.Category)
	}
	if parsed.Message != entry.Message {
		t.Errorf("Message = %q, want %q", parsed.Message, entry.Message)
	}
	if parsed.Error != entry.Error {
		t.Errorf("Error = %q, want %q", parsed.Error, entry.Error)
	}
}

// testError is a simple error implementation for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
