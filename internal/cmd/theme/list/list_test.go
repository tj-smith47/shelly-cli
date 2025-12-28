package list

import (
	"bytes"
	"encoding/json"
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

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want 'list'", cmd.Use)
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

	expectedAliases := []string{"ls", "l"}
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

func TestNewCommand_FilterFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	filterFlag := cmd.Flags().Lookup("filter")
	if filterFlag == nil {
		t.Fatal("filter flag not found")
	}

	if filterFlag.DefValue != "" {
		t.Errorf("filter default = %q, want empty string", filterFlag.DefValue)
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_ListAllThemes(t *testing.T) {
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Set current theme
	theme.SetTheme("dracula")

	err := run(f, "")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should contain header showing theme count
	if !strings.Contains(output, "Available Themes") {
		t.Error("Output should contain 'Available Themes'")
	}

	// Should list some themes
	if !strings.Contains(output, "dracula") {
		t.Error("Output should contain 'dracula' theme")
	}

	// Should show current indicator
	if !strings.Contains(output, "Theme") && !strings.Contains(output, "Current") {
		t.Error("Output should contain table headers")
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_FilterThemes(t *testing.T) {
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	// Filter for "nord" themes
	err := run(f, "nord")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should contain nord theme
	if !strings.Contains(strings.ToLower(output), "nord") {
		t.Error("Output should contain filtered 'nord' theme")
	}

	// Should NOT contain dracula (unless filter matches substring)
	if strings.Contains(strings.ToLower(output), "dracula") {
		t.Error("Output should not contain 'dracula' when filtering for 'nord'")
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_FilterCaseInsensitive(t *testing.T) {
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	// Filter with uppercase (should still match)
	err := run(f, "NORD")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should still match nord themes (case insensitive)
	if !strings.Contains(strings.ToLower(output), "nord") {
		t.Error("Filter should be case-insensitive")
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_FilterNoMatches(t *testing.T) {
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	// Filter for non-existent pattern
	err := run(f, "nonexistent-xyz-pattern")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should show "no themes found" message
	if !strings.Contains(strings.ToLower(output), "no themes found") {
		t.Error("Output should indicate no themes found")
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_JSONOutput(t *testing.T) {
	viper.Set("output", "json")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	err := run(f, "")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	// Parse JSON output
	var result []map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v\nOutput: %s", err, outBuf.String())
	}

	// Should have many themes (280+)
	if len(result) < 100 {
		t.Errorf("Expected 100+ themes, got %d", len(result))
	}

	// Check first item has expected fields
	if len(result) > 0 {
		item := result[0]
		if _, ok := item["id"]; !ok {
			t.Error("JSON items should have 'id' field")
		}
		if _, ok := item["current"]; !ok {
			t.Error("JSON items should have 'current' field")
		}
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_JSONOutputWithFilter(t *testing.T) {
	viper.Set("output", "json")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	// Filter for "dracula"
	err := run(f, "dracula")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v\nOutput: %s", err, outBuf.String())
	}

	// Should have at least one dracula theme
	if len(result) == 0 {
		t.Error("Filter for 'dracula' should return at least one theme")
	}

	// All returned themes should match filter
	for _, item := range result {
		id, ok := item["id"].(string)
		if !ok {
			continue
		}
		if !strings.Contains(strings.ToLower(id), "dracula") {
			t.Errorf("Theme %q should match filter 'dracula'", id)
		}
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_CurrentThemeMarked(t *testing.T) {
	viper.Set("output", "json")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Set specific theme
	theme.SetTheme("nord")

	err := run(f, "nord")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(outBuf.Bytes(), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	// Find the nord theme and check it's marked current
	found := false
	for _, item := range result {
		if id, ok := item["id"].(string); ok && id == "nord" {
			found = true
			current, ok := item["current"].(bool)
			if !ok || !current {
				t.Error("Current theme 'nord' should have current=true")
			}
		}
	}

	if !found {
		t.Error("Should find 'nord' theme in output")
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_YAMLOutput(t *testing.T) {
	viper.Set("output", "yaml")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	err := run(f, "dracula")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// YAML output should contain expected fields
	if !strings.Contains(output, "id:") {
		t.Error("YAML output should contain 'id:' field")
	}
	if !strings.Contains(output, "current:") {
		t.Error("YAML output should contain 'current:' field")
	}
}

//nolint:paralleltest // Cannot run in parallel due to viper global state
func TestRun_TableOutputShowsCount(t *testing.T) {
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	err := run(f, "")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should show theme count in header
	// e.g., "Available Themes (280 themes)"
	if !strings.Contains(output, "themes)") && !strings.Contains(output, "theme)") {
		t.Error("Output should show theme count")
	}
}
