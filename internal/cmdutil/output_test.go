package cmdutil_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// Note: Tests in this file modify global state (viper)
// and cannot run in parallel.

// setupViper creates a clean viper instance for testing.
func setupViper(t *testing.T) {
	t.Helper()
	// Reset viper for each test
	viper.Reset()
}

//nolint:paralleltest // Tests modify global state (viper)
func TestGetOutputConfig(t *testing.T) {
	setupViper(t)

	cfg := cmdutil.GetOutputConfig()

	if cfg.Format != output.FormatTable {
		t.Errorf("Format = %v, want %v", cfg.Format, output.FormatTable)
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestFormatOutput_JSON(t *testing.T) {
	setupViper(t)
	viper.Set("output", "json")

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	data := map[string]string{"key": "value"}
	err := cmdutil.FormatOutput(ios, data)

	if err != nil {
		t.Errorf("FormatOutput() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "key") || !strings.Contains(got, "value") {
		t.Errorf("FormatOutput() output = %q, should contain key and value", got)
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestFormatOutput_YAML(t *testing.T) {
	setupViper(t)
	viper.Set("output", "yaml")

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	data := map[string]string{"key": "value"}
	err := cmdutil.FormatOutput(ios, data)

	if err != nil {
		t.Errorf("FormatOutput() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "key") || !strings.Contains(got, "value") {
		t.Errorf("FormatOutput() output = %q, should contain key and value", got)
	}
}

func TestFormatTable(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"key1", "value1"},
		{"key2", "value2"},
	}

	if err := cmdutil.FormatTable(ios, headers, rows); err != nil {
		t.Fatalf("FormatTable failed: %v", err)
	}

	got := out.String()
	if got == "" {
		t.Error("FormatTable() should produce output")
	}
	if !strings.Contains(got, "Name") {
		t.Error("FormatTable() should contain header 'Name'")
	}
	if !strings.Contains(got, "key1") {
		t.Error("FormatTable() should contain 'key1'")
	}
}

func TestPrintFormatted_JSON(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	data := struct{ Name string }{Name: "test"}
	err := cmdutil.PrintFormatted(ios, output.FormatJSON, data)

	if err != nil {
		t.Errorf("PrintFormatted() error = %v", err)
	}
	if !strings.Contains(out.String(), "test") {
		t.Error("PrintFormatted() should contain 'test'")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIsQuiet_FalseByDefault(t *testing.T) {
	setupViper(t)
	if cmdutil.IsQuiet() {
		t.Error("IsQuiet() should be false by default")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIsQuiet_TrueWhenSet(t *testing.T) {
	setupViper(t)
	viper.Set("quiet", true)
	if !cmdutil.IsQuiet() {
		t.Error("IsQuiet() should be true when quiet flag is set")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIsVerbose_FalseByDefault(t *testing.T) {
	setupViper(t)
	if cmdutil.IsVerbose() {
		t.Error("IsVerbose() should be false by default")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestIsVerbose_TrueWhenSet(t *testing.T) {
	setupViper(t)
	viper.Set("verbose", true)
	if !cmdutil.IsVerbose() {
		t.Error("IsVerbose() should be true when verbose flag is set")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsJSON_FalseByDefault(t *testing.T) {
	setupViper(t)
	if cmdutil.WantsJSON() {
		t.Error("WantsJSON() should be false by default")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsJSON_TrueWhenJSON(t *testing.T) {
	setupViper(t)
	viper.Set("output", "json")
	if !cmdutil.WantsJSON() {
		t.Error("WantsJSON() should be true when output is json")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsYAML_FalseByDefault(t *testing.T) {
	setupViper(t)
	if cmdutil.WantsYAML() {
		t.Error("WantsYAML() should be false by default")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsYAML_TrueWhenYAML(t *testing.T) {
	setupViper(t)
	viper.Set("output", "yaml")
	if !cmdutil.WantsYAML() {
		t.Error("WantsYAML() should be true when output is yaml")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsTable_TrueByDefault(t *testing.T) {
	setupViper(t)
	if !cmdutil.WantsTable() {
		t.Error("WantsTable() should be true by default")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsTable_FalseWhenJSON(t *testing.T) {
	setupViper(t)
	viper.Set("output", "json")
	if cmdutil.WantsTable() {
		t.Error("WantsTable() should be false when output is json")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsStructured_FalseByDefault(t *testing.T) {
	setupViper(t)
	if cmdutil.WantsStructured() {
		t.Error("WantsStructured() should be false by default")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsStructured_TrueForJSON(t *testing.T) {
	setupViper(t)
	viper.Set("output", "json")
	if !cmdutil.WantsStructured() {
		t.Error("WantsStructured() should be true for json")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestWantsStructured_TrueForYAML(t *testing.T) {
	setupViper(t)
	viper.Set("output", "yaml")
	if !cmdutil.WantsStructured() {
		t.Error("WantsStructured() should be true for yaml")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestShouldShowProgress_FalseInQuietMode(t *testing.T) {
	setupViper(t)
	viper.Set("quiet", true)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	if cmdutil.ShouldShowProgress(ios) {
		t.Error("ShouldShowProgress() should be false in quiet mode")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestShouldShowProgress_FalseForStructured(t *testing.T) {
	setupViper(t)
	viper.Set("output", "json")

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	if cmdutil.ShouldShowProgress(ios) {
		t.Error("ShouldShowProgress() should be false for structured output")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestShouldShowProgress_FalseForNonTTY(t *testing.T) {
	setupViper(t)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	// Test() creates non-TTY iostreams

	if cmdutil.ShouldShowProgress(ios) {
		t.Error("ShouldShowProgress() should be false for non-TTY")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestConditionalPrint_PrintsWhenNotQuiet(t *testing.T) {
	setupViper(t)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	cmdutil.ConditionalPrint(ios, "test message")

	if !strings.Contains(out.String(), "test message") {
		t.Error("ConditionalPrint() should print when not quiet")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestConditionalPrint_SilentWhenQuiet(t *testing.T) {
	setupViper(t)
	viper.Set("quiet", true)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	cmdutil.ConditionalPrint(ios, "test message")

	if out.String() != "" {
		t.Error("ConditionalPrint() should be silent when quiet")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestConditionalSuccess_PrintsWhenNotQuiet(t *testing.T) {
	setupViper(t)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	cmdutil.ConditionalSuccess(ios, "success message")

	if out.String() == "" {
		t.Error("ConditionalSuccess() should print when not quiet")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestConditionalSuccess_SilentWhenQuiet(t *testing.T) {
	setupViper(t)
	viper.Set("quiet", true)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	cmdutil.ConditionalSuccess(ios, "success message")

	if out.String() != "" {
		t.Error("ConditionalSuccess() should be silent when quiet")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestConditionalInfo_PrintsWhenNotQuiet(t *testing.T) {
	setupViper(t)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	cmdutil.ConditionalInfo(ios, "info message")

	if out.String() == "" {
		t.Error("ConditionalInfo() should print when not quiet")
	}
}

//nolint:paralleltest // Tests modify global state (viper)
func TestConditionalInfo_SilentWhenQuiet(t *testing.T) {
	setupViper(t)
	viper.Set("quiet", true)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	cmdutil.ConditionalInfo(ios, "info message")

	if out.String() != "" {
		t.Error("ConditionalInfo() should be silent when quiet")
	}
}
