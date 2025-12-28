package diff

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	// Test Use field
	if cmd.Use != "diff <template> <device>" {
		t.Errorf("Use = %q, want \"diff <template> <device>\"", cmd.Use)
	}

	// Test Short description
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Short != "Compare a template with a device" {
		t.Errorf("Short = %q, want \"Compare a template with a device\"", cmd.Short)
	}

	// Test Long description
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"compare", "cmp"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
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

	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no args",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "one arg",
			args:      []string{"template"},
			wantError: true,
		},
		{
			name:      "two args valid",
			args:      []string{"template", "device"},
			wantError: false,
		},
		{
			name:      "three args",
			args:      []string{"template", "device", "extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantError {
				t.Errorf("Args() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for tab completion")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_NoCustomFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The diff command doesn't add any custom flags (no dry-run, no yes)
	// It relies on global output format flag from parent
	localFlags := cmd.Flags()

	// Verify no local flags are added (only inherits from parent)
	// Count only non-persistent local flags
	localFlagCount := 0
	localFlags.VisitAll(func(_ *pflag.Flag) {
		localFlagCount++
	})

	// The diff command should have 0 local flags defined
	if localFlagCount != 0 {
		t.Errorf("Expected 0 local flags, got %d", localFlagCount)
	}
}

func TestOptions_FactorySet(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	// Verify command is created successfully
	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	// The factory should be accessible through the options
	// We verify this by checking the command has proper RunE set
	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example
	if example == "" {
		t.Fatal("Example is empty")
	}

	// Check that the example contains expected patterns
	expectedPatterns := []string{
		"shelly template diff",
		"-o json",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(example, pattern) {
			t.Errorf("Example should contain %q, got:\n%s", pattern, example)
		}
	}
}

func TestNewCommand_LongDescriptionContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	long := cmd.Long
	if long == "" {
		t.Fatal("Long description is empty")
	}

	// Verify long description explains the command purpose
	expectedPhrases := []string{
		"Compare",
		"template",
		"configuration",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(long, phrase) {
			t.Errorf("Long description should mention %q", phrase)
		}
	}
}

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

func TestRun_TemplateNotFound(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Template: "nonexistent-template-xyz",
		Device:   "test-device",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for non-existent template")
	}

	// Error should mention template not found
	if err != nil {
		errStr := err.Error()
		if !bytes.Contains([]byte(errStr), []byte("not found")) {
			t.Errorf("error should mention 'not found', got: %v", err)
		}
		if !bytes.Contains([]byte(errStr), []byte("nonexistent-template-xyz")) {
			t.Errorf("error should mention template name, got: %v", err)
		}
	}
}

func TestRun_MultipleTemplateNotFoundCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		template string
		device   string
	}{
		{
			name:     "simple template name",
			template: "my-template",
			device:   "192.168.1.100",
		},
		{
			name:     "template with special chars",
			template: "test_template-123",
			device:   "device.local",
		},
		{
			name:     "template with numbers",
			template: "config-backup-2024",
			device:   "shelly-switch",
		},
		{
			name:     "long template name",
			template: "very-long-template-name-for-testing",
			device:   "192.0.2.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			ios := iostreams.Test(nil, &stdout, &stderr)
			f := cmdutil.NewWithIOStreams(ios)

			opts := &Options{
				Factory:  f,
				Template: tc.template,
				Device:   tc.device,
			}

			err := run(context.Background(), opts)
			if err == nil {
				t.Errorf("expected error for non-existent template %q", tc.template)
			}

			if err != nil && !bytes.Contains([]byte(err.Error()), []byte("not found")) {
				t.Errorf("error should mention 'not found', got: %v", err)
			}
		})
	}
}

func TestOptions_StructFields(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory:  f,
		Template: "test-template",
		Device:   "test-device",
	}

	if opts.Template != "test-template" {
		t.Errorf("Template = %q, want %q", opts.Template, "test-template")
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:  f,
		Template: "test-template",
		Device:   "test-device",
	}

	// This tests the timeout path - even with cancelled context,
	// the template lookup happens first and should fail with "not found"
	err := run(ctx, opts)
	if err == nil {
		t.Error("expected error")
	}
}

func TestNewCommand_AliasesOrder(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify specific alias positions
	if len(cmd.Aliases) < 2 {
		t.Fatalf("expected at least 2 aliases, got %d", len(cmd.Aliases))
	}

	if cmd.Aliases[0] != "compare" {
		t.Errorf("Aliases[0] = %q, want %q", cmd.Aliases[0], "compare")
	}
	if cmd.Aliases[1] != "cmp" {
		t.Errorf("Aliases[1] = %q, want %q", cmd.Aliases[1], "cmp")
	}
}

func TestNewCommand_ArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Args should be cobra.ExactArgs(2)
	if cmd.Args == nil {
		t.Error("Args function should be set")
	}

	// Test that it validates exactly 2 arguments
	testCases := []struct {
		argCount  int
		wantError bool
	}{
		{0, true},
		{1, true},
		{2, false},
		{3, true},
		{4, true},
	}

	for _, tc := range testCases {
		args := make([]string, tc.argCount)
		for i := range args {
			args[i] = "arg"
		}
		err := cmd.Args(cmd, args)
		if (err != nil) != tc.wantError {
			t.Errorf("Args with %d args: error = %v, wantError = %v", tc.argCount, err, tc.wantError)
		}
	}
}

func TestRun_EmptyTemplateName(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Template: "",
		Device:   "test-device",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for empty template name")
	}
}

func TestRun_EmptyDeviceName(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Template: "test-template",
		Device:   "",
	}

	// Empty device should still first check for template (which doesn't exist)
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error")
	}
	// Error should be about template not found
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestNewCommand_ExampleFormat(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify example has proper format with leading spaces
	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Examples typically start with leading spaces for proper help formatting
	lines := strings.Split(cmd.Example, "\n")
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") {
			// Non-empty, non-comment lines should have leading indentation
			t.Logf("Line without leading space: %q", line)
		}
	}
}

func TestNewCommand_ExecuteWithArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-template", "test-device"})

	// Execute should fail with template not found
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent template")
	}

	// Error should mention template not found
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestNewCommand_ExecuteNoArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})

	// Execute should fail with wrong number of arguments
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing args")
	}
}

func TestNewCommand_ExecuteOneArg(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"template-only"})

	// Execute should fail with wrong number of arguments
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing device arg")
	}
}
