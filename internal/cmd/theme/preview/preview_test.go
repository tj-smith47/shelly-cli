package preview

import (
	"bytes"
	"strings"
	"testing"

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

	if cmd.Use != "preview [theme]" {
		t.Errorf("Use = %q, want 'preview [theme]'", cmd.Use)
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

	expectedAliases := []string{"show", "demo"}
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

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args is valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg is valid",
			args:    []string{"nord"},
			wantErr: false,
		},
		{
			name:    "two args is invalid",
			args:    []string{"nord", "dracula"},
			wantErr: true,
		},
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

func TestRun_CurrentTheme(t *testing.T) {
	t.Parallel()

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Ensure a theme is set
	theme.SetTheme("dracula")

	err := run(f, "")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should contain theme header
	if !strings.Contains(output, "Theme:") {
		t.Error("Output should contain 'Theme:'")
	}

	// Should contain color palette section
	if !strings.Contains(output, "Color Palette:") {
		t.Error("Output should contain 'Color Palette:'")
	}

	// Should contain status indicators section
	if !strings.Contains(output, "Status Indicators:") {
		t.Error("Output should contain 'Status Indicators:'")
	}

	// Should contain device status section
	if !strings.Contains(output, "Device Status:") {
		t.Error("Output should contain 'Device Status:'")
	}

	// Should contain switch state section
	if !strings.Contains(output, "Switch State:") {
		t.Error("Output should contain 'Switch State:'")
	}

	// Should contain power/energy section
	if !strings.Contains(output, "Power/Energy:") {
		t.Error("Output should contain 'Power/Energy:'")
	}

	// Should contain table header section
	if !strings.Contains(output, "Table Header:") {
		t.Error("Output should contain 'Table Header:'")
	}
}

func TestRun_SpecificTheme(t *testing.T) {
	t.Parallel()

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Set a different current theme first
	theme.SetTheme("dracula")

	err := run(f, "nord")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should show the previewed theme name
	if !strings.Contains(output, "nord") {
		t.Error("Output should contain theme name 'nord'")
	}
}

func TestRun_InvalidTheme(t *testing.T) {
	t.Parallel()

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f, "nonexistent-theme-xyz")
	if err == nil {
		t.Error("run() should return error for nonexistent theme")
	}

	if !strings.Contains(err.Error(), "theme not found") {
		t.Errorf("Error message should contain 'theme not found', got: %v", err)
	}
}

func TestRun_ColorSections(t *testing.T) {
	t.Parallel()

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	err := run(f, "")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Verify color names are in output
	colorNames := []string{
		"Foreground",
		"Background",
		"Red",
		"Green",
		"Yellow",
		"Blue",
		"Purple",
		"Cyan",
	}

	for _, name := range colorNames {
		if !strings.Contains(output, name) {
			t.Errorf("Output should contain color name %q", name)
		}
	}
}

//nolint:paralleltest // Cannot run in parallel due to global theme state
func TestRun_PreservesCurrentTheme(t *testing.T) {
	// Set initial theme
	theme.SetTheme("dracula")
	originalTheme := theme.CurrentThemeName()

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Preview a different theme
	err := run(f, "nord")
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	// Theme should be restored after preview
	currentTheme := theme.CurrentThemeName()
	if currentTheme != originalTheme {
		t.Errorf("Theme not restored: got %q, want %q", currentTheme, originalTheme)
	}
}
