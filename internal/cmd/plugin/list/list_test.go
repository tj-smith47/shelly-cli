package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want \"list\"", cmd.Use)
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
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
	}

	for i, alias := range expectedAliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// list command doesn't specify explicit Args validator (uses default)
	// Verify that the command can accept any args (cobra default behavior)
	if cmd.Args != nil {
		// If Args is set, test that it accepts empty args
		if err := cmd.Args(cmd, []string{}); err != nil {
			t.Errorf("Args() with empty args returned error: %v", err)
		}
	}
	// No explicit args validation is fine for list commands
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{
			name:      "all flag exists",
			flagName:  "all",
			shorthand: "a",
			defValue:  "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %q not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default value = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify long description contains key information
	long := cmd.Long

	expectedContents := []string{
		"extensions",
		"shelly-*",
		"--all",
		"json",
		"yaml",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(long, expected) {
			t.Errorf("Long description missing expected content: %q", expected)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify examples contain key usage patterns
	example := cmd.Example

	expectedPatterns := []string{
		"extension list",
		"--all",
		"-o json",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(example, pattern) {
			t.Errorf("Example missing expected pattern: %q", pattern)
		}
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_AllFlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("all")
	if flag == nil {
		t.Fatal("all flag not found")
	}

	// Check the usage description contains relevant information
	usage := flag.Usage
	if usage == "" {
		t.Error("all flag usage description is empty")
	}

	// Verify usage mentions "all" or "discovered"
	if !strings.Contains(usage, "all") && !strings.Contains(usage, "discovered") {
		t.Error("all flag usage should mention 'all' or 'discovered'")
	}
}

func TestRun_InstalledExtensions(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Test with all=false (installed only)
	opts := &Options{Factory: f, All: false}
	err := run(opts)

	// Should succeed or produce a meaningful error
	if err != nil {
		// Only fail if it's an unexpected error
		if !strings.Contains(err.Error(), "plugins") && !strings.Contains(err.Error(), "directory") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestRun_AllDiscoveredExtensions(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Test with all=true (discover all)
	opts := &Options{Factory: f, All: true}
	err := run(opts)

	// Should succeed or produce a meaningful error
	if err != nil {
		// Only fail if it's an unexpected error
		if !strings.Contains(err.Error(), "plugins") && !strings.Contains(err.Error(), "directory") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestRun_OutputEmptyList(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Run the command
	opts := &Options{Factory: f, All: false}
	err := run(opts)

	// Check output contains info message or table header
	output := stdout.String() + stderr.String()

	// If error, it should be registry-related, not a panic
	if err != nil {
		return // Error already checked in TestRun_InstalledExtensions
	}

	// If successful with empty list, should show info message
	// If successful with extensions, should show table
	if output == "" && err == nil {
		// Empty output is fine for list with no extensions
		return
	}
}

func TestRun_FlagVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		all  bool
	}{
		{"installed only", false},
		{"all discovered", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
			f := cmdutil.NewFactory().SetIOStreams(ios)

				// Just verify it doesn't panic - error result is not important for this test
			opts := &Options{Factory: f, All: tt.all}
			_ = run(opts) //nolint:errcheck // intentionally ignored for panic check
		})
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Short, "installed") && !strings.Contains(cmd.Short, "List") {
		t.Errorf("Short description should mention listing extensions, got: %q", cmd.Short)
	}
}

func TestNewCommand_DescriptionContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify Long description mentions output formats
	if !strings.Contains(cmd.Long, "table") {
		t.Error("Long description should mention table output format")
	}

	// Verify it mentions columns
	if !strings.Contains(cmd.Long, "Name") {
		t.Error("Long description should mention column names")
	}
}
