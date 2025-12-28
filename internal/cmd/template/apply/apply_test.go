package apply

import (
	"bytes"
	"context"
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

	// Test Use field
	if cmd.Use != "apply <template> <device>" {
		t.Errorf("Use = %q, want \"apply <template> <device>\"", cmd.Use)
	}

	// Test Short description
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Short != "Apply a template to a device" {
		t.Errorf("Short = %q, want \"Apply a template to a device\"", cmd.Short)
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

	expectedAliases := []string{"set", "push"}
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{
			name:         "dry-run flag exists",
			flagName:     "dry-run",
			shorthand:    "",
			defaultValue: "false",
		},
		{
			name:         "yes flag exists",
			flagName:     "yes",
			shorthand:    "y",
			defaultValue: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defaultValue)
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

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	// Parse with no flags to verify defaults
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	// Verify dry-run default
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		t.Fatalf("GetBool(dry-run) failed: %v", err)
	}
	if dryRun {
		t.Error("dry-run should default to false")
	}

	// Verify yes default
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		t.Fatalf("GetBool(yes) failed: %v", err)
	}
	if yes {
		t.Error("yes should default to false")
	}
}

func TestOptions_FlagsCanBeSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    []string
		flagName string
		expected bool
	}{
		{
			name:     "dry-run flag set",
			flags:    []string{"--dry-run"},
			flagName: "dry-run",
			expected: true,
		},
		{
			name:     "yes flag short form",
			flags:    []string{"-y"},
			flagName: "yes",
			expected: true,
		},
		{
			name:     "yes flag long form",
			flags:    []string{"--yes"},
			flagName: "yes",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			if err := cmd.ParseFlags(tt.flags); err != nil {
				t.Fatalf("ParseFlags failed: %v", err)
			}

			got, err := cmd.Flags().GetBool(tt.flagName)
			if err != nil {
				t.Fatalf("GetBool(%s) failed: %v", tt.flagName, err)
			}
			if got != tt.expected {
				t.Errorf("flag %q = %v, want %v", tt.flagName, got, tt.expected)
			}
		})
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

func TestRun_TemplateNotFoundDryRun(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Template: "another-nonexistent-template",
		Device:   "192.0.2.1",
		DryRun:   true,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for non-existent template even in dry-run mode")
	}

	// Error should mention template not found
	if err != nil && !bytes.Contains([]byte(err.Error()), []byte("not found")) {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestRun_MultipleTemplateNotFoundCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		template string
		device   string
		dryRun   bool
		yes      bool
	}{
		{
			name:     "simple template name",
			template: "my-template",
			device:   "192.168.1.100",
			dryRun:   false,
			yes:      false,
		},
		{
			name:     "template with special chars",
			template: "test_template-123",
			device:   "device.local",
			dryRun:   true,
			yes:      false,
		},
		{
			name:     "with yes flag",
			template: "config-backup",
			device:   "shelly-switch",
			dryRun:   false,
			yes:      true,
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
				DryRun:   tc.dryRun,
			}
			opts.Yes = tc.yes

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
		DryRun:   true,
	}
	opts.Yes = true

	if opts.Template != "test-template" {
		t.Errorf("Template = %q, want %q", opts.Template, "test-template")
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
	if !opts.Yes {
		t.Error("Yes should be true")
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
		"shelly template apply",
		"--dry-run",
		"--yes",
	}

	for _, pattern := range expectedPatterns {
		found := false
		for _, line := range splitLines(example) {
			if containsSubstring(line, pattern) {
				found = true
				break
			}
		}
		if !found {
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
		"Apply",
		"template",
		"device",
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(long, phrase) {
			t.Errorf("Long description should mention %q", phrase)
		}
	}
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := range len(s) {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
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

func TestNewCommand_DryRunFlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("dry-run flag not found")
	}

	// Check flag usage description
	if flag.Usage == "" {
		t.Error("dry-run flag should have usage description")
	}
}

func TestNewCommand_YesFlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("yes flag not found")
	}

	// Check flag usage description
	if flag.Usage == "" {
		t.Error("yes flag should have usage description")
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
	if err != nil && !containsSubstring(err.Error(), "not found") {
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

func TestNewCommand_ExecuteWithDryRun(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-template", "test-device", "--dry-run"})

	// Execute should fail with template not found even in dry-run mode
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent template")
	}
}

func TestNewCommand_ExecuteWithYes(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-template", "test-device", "--yes"})

	// Execute should fail with template not found
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent template")
	}
}

func TestNewCommand_ExecuteWithBothFlags(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-template", "test-device", "--dry-run", "--yes"})

	// Execute should fail with template not found
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent template")
	}
}

func TestNewCommand_ExecuteTooManyArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"template", "device", "extra"})

	// Execute should fail with too many arguments
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for too many args")
	}
}

func TestNewCommand_AliasesOrder(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify specific alias positions
	if len(cmd.Aliases) < 2 {
		t.Fatalf("expected at least 2 aliases, got %d", len(cmd.Aliases))
	}

	if cmd.Aliases[0] != "set" {
		t.Errorf("Aliases[0] = %q, want %q", cmd.Aliases[0], "set")
	}
	if cmd.Aliases[1] != "push" {
		t.Errorf("Aliases[1] = %q, want %q", cmd.Aliases[1], "push")
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
	if err != nil && !containsSubstring(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}
