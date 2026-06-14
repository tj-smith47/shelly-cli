package preview

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

const themeNord = "nord"

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
			args:    []string{themeNord},
			wantErr: false,
		},
		{
			name:    "two args is invalid",
			args:    []string{themeNord, "dracula"},
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

//nolint:paralleltest // run() mutates the process-global theme singleton; concurrent theme tests would race on it.
func TestRun_CurrentTheme(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Ensure a theme is set
	theme.SetTheme("dracula")

	opts := &Options{Factory: f, ThemeName: ""}
	err := run(opts)
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

//nolint:paralleltest // run() mutates the process-global theme singleton.
func TestRun_SpecificTheme(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Set a different current theme first
	theme.SetTheme("dracula")

	opts := &Options{Factory: f, ThemeName: themeNord}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should show the previewed theme name
	if !strings.Contains(output, themeNord) {
		t.Error("Output should contain theme name 'nord'")
	}
}

//nolint:paralleltest // run() reads the process-global theme singleton.
func TestRun_InvalidTheme(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, ThemeName: "nonexistent-theme-xyz"}
	err := run(opts)
	if err == nil {
		t.Error("run() should return error for nonexistent theme")
	}

	if !strings.Contains(err.Error(), "theme not found") {
		t.Errorf("Error message should contain 'theme not found', got: %v", err)
	}
}

//nolint:paralleltest // run() mutates the process-global theme singleton.
func TestRun_ColorSections(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	opts := &Options{Factory: f, ThemeName: ""}
	err := run(opts)
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
	opts := &Options{Factory: f, ThemeName: themeNord}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	// Theme should be restored after preview
	currentTheme := theme.CurrentThemeName()
	if currentTheme != originalTheme {
		t.Errorf("Theme not restored: got %q, want %q", currentTheme, originalTheme)
	}
}
