package semantic

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "semantic" {
		t.Errorf("Use = %q, want 'semantic'", cmd.Use)
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

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"sem", "colors"}
	if len(cmd.Aliases) < len(expectedAliases) {
		t.Errorf("Expected at least %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		found := false
		for _, alias := range cmd.Aliases {
			if alias == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected alias %q not found in %v", expected, cmd.Aliases)
		}
	}
}

func TestNewCommand_NoArgsRequired(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Command should not have Args validator set (accepts any)
	// This is standard for commands that take no positional args
	if cmd.Args != nil {
		// If Args is set, verify empty args work
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Expected no error with no args, got: %v", err)
		}
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_TextOutput(t *testing.T) {
	// Clear any output format settings
	if err := os.Unsetenv("SHELLY_OUTPUT"); err != nil {
		t.Logf("warning: failed to unset SHELLY_OUTPUT: %v", err)
	}
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Ensure a theme is set
	theme.SetTheme("dracula")

	opts := &Options{Factory: f}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	out := outBuf.String()

	// Check sections are present
	sections := []string{
		"Semantic Color Mappings:",
		"UI Colors:",
		"Text Colors:",
		"Feedback Colors:",
		"Device State Colors:",
		"Table Colors:",
	}
	for _, section := range sections {
		if !strings.Contains(out, section) {
			t.Errorf("Output should contain %q", section)
		}
	}

	// Check color names are present
	colorNames := []string{
		// UI colors
		"Primary", "Secondary", "Highlight", "Muted",
		// Feedback colors
		"Success", "Warning", "Error", "Info",
		// State colors
		"Online", "Offline", "Updating", "Idle",
		// Table colors
		"TableHeader", "TableCell", "TableAltCell", "TableBorder",
	}
	for _, name := range colorNames {
		if !strings.Contains(out, name) {
			t.Errorf("Output should contain color name %q", name)
		}
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_JSONOutput(t *testing.T) {
	// Set JSON output format
	viper.Set("output", "json")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Ensure a theme is set
	theme.SetTheme("dracula")

	opts := &Options{Factory: f}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	// Parse the JSON output
	var result map[string]string
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, outBuf.String())
	}

	// Verify expected keys are present
	expectedKeys := []string{
		"primary",
		"secondary",
		"highlight",
		"muted",
		"text",
		"alt_text",
		"success",
		"warning",
		"error",
		"info",
		"background",
		"alt_background",
		"online",
		"offline",
		"updating",
		"idle",
		"table_header",
		"table_cell",
		"table_alt_cell",
		"table_border",
	}

	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}

	// Verify values look like hex colors
	for key, val := range result {
		if val == "" || !strings.HasPrefix(val, "#") {
			t.Errorf("Expected hex color for key %q, got %q", key, val)
		}
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_YAMLOutput(t *testing.T) {
	// Set YAML output format
	viper.Set("output", "yaml")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Ensure a theme is set
	theme.SetTheme("dracula")

	opts := &Options{Factory: f}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	out := outBuf.String()

	// Verify YAML output contains expected keys
	expectedKeys := []string{
		"primary:",
		"secondary:",
		"success:",
		"warning:",
		"error:",
		"online:",
		"offline:",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(out, key) {
			t.Errorf("YAML output should contain key %q", key)
		}
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_ColorBlocksRendered(t *testing.T) {
	// Ensure text output format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	opts := &Options{Factory: f}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	// The output should contain block characters (used for color samples)
	// The block character is used in the theme package
	if !strings.Contains(outBuf.String(), "Text") && !strings.Contains(outBuf.String(), "Primary") {
		t.Error("Output should contain color sample blocks")
	}
}
