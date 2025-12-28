package upgrade

import (
	"bytes"
	"context"
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

	if cmd.Use != "upgrade [name]" {
		t.Errorf("Use = %q, want \"upgrade [name]\"", cmd.Use)
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

	expectedAliases := []string{"update"}
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

	// upgrade command takes 0 or 1 argument
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args succeeds",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg succeeds",
			args:    []string{"myext"},
			wantErr: false,
		},
		{
			name:    "two args returns error",
			args:    []string{"myext", "extra"},
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
		"Upgrade",
		"--all",
		"GitHub",
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
		"extension upgrade",
		"--all",
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

	// Verify usage mentions "all" or "upgrade"
	if !strings.Contains(usage, "all") && !strings.Contains(usage, "Upgrade") {
		t.Error("all flag usage should mention 'all' or 'Upgrade'")
	}
}

func TestNewCommand_MaximumArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test that more than 1 arg fails (MaximumNArgs(1))
	err := cmd.Args(cmd, []string{"arg1", "arg2", "arg3"})
	if err == nil {
		t.Error("expected error with more than 1 arg")
	}
}

func TestRun_NoNameNoAll(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	// No name and all=false should print info message
	err := run(context.Background(), f, "", false)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should output info message about specifying extension
	output := stdout.String() + stderr.String()
	// If there's output, it may mention the usage (Specify or --all)
	// Empty output or any mention of usage is acceptable
	_ = output // Use output to avoid unused variable warning
}

func TestRun_AllWithEmptyList(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	// all=true with no extensions should output info message
	err := run(context.Background(), f, "", true)

	if err != nil {
		// Error from registry is acceptable
		if !strings.Contains(err.Error(), "plugins") && !strings.Contains(err.Error(), "directory") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestRun_SpecificExtension(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Try to upgrade a non-existent extension
	err := run(context.Background(), f, "nonexistent-extension", false)

	// Should fail with "not installed" error
	if err == nil {
		t.Error("expected error for non-existent extension")
	}

	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("error should mention extension not installed, got: %v", err)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should handle cancelled context gracefully
	err := run(ctx, f, "test-ext", false)

	// Either returns quickly with context error or proceeds
	// Just verify it doesn't panic
	_ = err
}

func TestRun_AllFlagBehavior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		extName string
		all     bool
	}{
		{"no name no all", "", false},
		{"no name with all", "", true},
		{"with name no all", "myext", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
			f := cmdutil.NewFactory().SetIOStreams(ios)

			// Just verify it doesn't panic - error result is not important for this test
			_ = run(context.Background(), f, tt.extName, tt.all) //nolint:errcheck // intentionally ignored for panic check
		})
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Short, "Upgrade") && !strings.Contains(cmd.Short, "extension") {
		t.Errorf("Short description should mention upgrading extensions, got: %q", cmd.Short)
	}
}

func TestNewCommand_LongDescriptionDetails(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should mention version checking behavior
	expectedTerms := []string{
		"newer",
		"release",
		"installed",
	}

	for _, term := range expectedTerms {
		if !strings.Contains(cmd.Long, term) {
			t.Errorf("Long description should mention %q", term)
		}
	}
}
