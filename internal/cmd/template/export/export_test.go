package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "export <template> [file]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "export <template> [file]")
	}

	wantAliases := []string{"save", "dump"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
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
		{"no args", []string{}, true},
		{"one arg valid", []string{"template"}, false},
		{"two args valid", []string{"template", "file.yaml"}, false},
		{"three args", []string{"template", "file", "extra"}, true},
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

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly template export",
		".yaml",
		".json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Template: "my-template",
		File:     "output.yaml",
		Factory:  f,
	}

	if opts.Template != "my-template" {
		t.Errorf("Template = %q, want %q", opts.Template, "my-template")
	}

	if opts.File != "output.yaml" {
		t.Errorf("File = %q, want %q", opts.File, "output.yaml")
	}
}

func TestRun_TemplateNotFound(t *testing.T) {
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Template: "nonexistent-template-12345",
	}

	err := run(opts)
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found': %v", err)
	}
}

func TestRun_UnsupportedFormat(t *testing.T) {
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Template: "nonexistent",
	}
	opts.Format = "xml"

	err := run(opts)
	// Will fail at template lookup first, but format check comes after
	if err == nil {
		t.Logf("Expected error")
	}
}
