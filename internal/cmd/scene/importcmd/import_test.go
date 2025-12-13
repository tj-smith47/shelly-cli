package importcmd

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "import <file>" {
		t.Errorf("Use = %q, want \"import <file>\"", cmd.Use)
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

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test requires exactly 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"file.yaml"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"file1", "file2"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatal("name flag not found")
	}
	if nameFlag.Shorthand != "n" {
		t.Errorf("name shorthand = %q, want n", nameFlag.Shorthand)
	}

	overwriteFlag := cmd.Flags().Lookup("overwrite")
	if overwriteFlag == nil {
		t.Fatal("overwrite flag not found")
	}
	if overwriteFlag.DefValue != "false" {
		t.Errorf("overwrite default = %q, want false", overwriteFlag.DefValue)
	}
}

func TestParseUnknownFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "valid yaml",
			data:    "name: test\nactions: []",
			wantErr: false,
		},
		{
			name:    "valid json",
			data:    `{"name":"test","actions":[]}`,
			wantErr: false,
		},
		{
			name:    "invalid format",
			data:    "not valid yaml or json {{{",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var scene Scene
		err := parseUnknownFormat([]byte(tt.data), &scene)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: parseUnknownFormat() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}
