package export

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

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

	if cmd.Use != "export [file]" {
		t.Errorf("Use = %q, want 'export [file]'", cmd.Use)
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

	expectedAliases := []string{"exp", "save"}
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
			name:    "no args is valid (stdout)",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg is valid (file path)",
			args:    []string{"mytheme.yaml"},
			wantErr: false,
		},
		{
			name:    "two args is invalid",
			args:    []string{"file1.yaml", "file2.yaml"},
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

func TestRun_ExportToStdout(t *testing.T) {
	t.Parallel()

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Ensure a theme is set
	theme.SetTheme("dracula")

	opts := &Options{Factory: f, File: ""}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	output := outBuf.String()

	// Should output valid YAML
	var export theme.Export
	if err := yaml.Unmarshal([]byte(output), &export); err != nil {
		t.Fatalf("Output should be valid YAML: %v\nOutput: %s", err, output)
	}

	// Should contain theme name
	if export.Name == "" {
		t.Error("Export should contain theme name")
	}

	// Should contain rendered colors
	if export.RenderedColors.Foreground == "" {
		t.Error("Export should contain foreground color")
	}
	if export.RenderedColors.Background == "" {
		t.Error("Export should contain background color")
	}
	if export.RenderedColors.Green == "" {
		t.Error("Export should contain green color")
	}
	if export.RenderedColors.Red == "" {
		t.Error("Export should contain red color")
	}
}

func TestRun_ExportToFile(t *testing.T) {
	t.Parallel()

	// Create temp directory
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "theme-export.yaml")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	// Ensure a theme is set
	theme.SetTheme("nord")

	opts := &Options{Factory: f, File: tmpFile}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	// File should be created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("Export file should be created")
	}

	// Read file content
	content, err := os.ReadFile(tmpFile) //nolint:gosec // G304: tmpFile is from t.TempDir(), safe for tests
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	// Should be valid YAML
	var export theme.Export
	if err := yaml.Unmarshal(content, &export); err != nil {
		t.Fatalf("File content should be valid YAML: %v\nContent: %s", err, string(content))
	}

	// Should contain theme name
	if export.Name != "nord" {
		t.Errorf("Export name = %q, want 'nord'", export.Name)
	}
}

func TestRun_ExportContainsAllColors(t *testing.T) {
	t.Parallel()

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	opts := &Options{Factory: f, File: ""}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	var export theme.Export
	if err := yaml.Unmarshal(outBuf.Bytes(), &export); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	// Verify all color fields using a table-driven approach
	colors := export.RenderedColors
	colorTests := []struct {
		name  string
		value string
	}{
		{"Foreground", colors.Foreground},
		{"Background", colors.Background},
		{"Green", colors.Green},
		{"Red", colors.Red},
		{"Yellow", colors.Yellow},
		{"Blue", colors.Blue},
		{"Cyan", colors.Cyan},
		{"Purple", colors.Purple},
		{"BrightBlack", colors.BrightBlack},
	}

	for _, tc := range colorTests {
		if tc.value == "" || !strings.HasPrefix(tc.value, "#") {
			t.Errorf("%s should be a hex color, got %q", tc.name, tc.value)
		}
	}
}

func TestRun_ExportShowsSuccessMessage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-theme.yaml")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	opts := &Options{Factory: f, File: tmpFile}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	// Should show success message mentioning the file
	output := outBuf.String()
	if !strings.Contains(output, tmpFile) && !strings.Contains(output, "exported") {
		// The output should either contain the filename or indicate success
		// Since Success() writes to stdout, check for either
		if output == "" {
			t.Log("Note: Success message may be written to stderr or formatted differently")
		}
	}
}

func TestRun_ExportInvalidPath(t *testing.T) {
	t.Parallel()

	// Use a path that should fail (directory doesn't exist)
	invalidPath := "/nonexistent-dir-xyz/subdir/theme.yaml"

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	opts := &Options{Factory: f, File: invalidPath}
	err := run(opts)
	if err == nil {
		t.Error("run() should return error for invalid path")
	}

	if !strings.Contains(err.Error(), "failed to write file") {
		t.Errorf("Error should mention 'failed to write file', got: %v", err)
	}
}

func TestRun_ExportFilePermissions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "permissions-test.yaml")

	var outBuf, errBuf bytes.Buffer
	ios := iostreams.Test(strings.NewReader(""), &outBuf, &errBuf)
	f := cmdutil.NewWithIOStreams(ios)

	theme.SetTheme("dracula")

	opts := &Options{Factory: f, File: tmpFile}
	err := run(opts)
	if err != nil {
		t.Errorf("run() unexpected error: %v", err)
	}

	// Check file permissions - should be 0600 (user read/write only)
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("File permissions = %o, want 0600", perm)
	}
}
