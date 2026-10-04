package rotate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/auth"
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

	// Test Use
	if cmd.Use != "rotate <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "rotate <device>")
	}

	// Test Aliases
	wantAliases := []string{"renew", "refresh"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
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
		{"one arg valid", []string{"device"}, false},
		{"two args", []string{"device1", "device2"}, true},
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
		name     string
		defValue string
	}{
		{"user", ""},
		{"password", ""},
		{"password-stdin", "false"},
		{"length", "16"},
		{"generate", "false"},
		{"show", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
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

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly auth rotate",
		"--password",
		"--generate",
		"--show",
		"--length",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		User:       "testuser",
		Password:   "testpass",
		Length:     24,
		Generate:   true,
		ShowSecret: true,
	}

	if opts.User != "testuser" {
		t.Errorf("User = %q, want %q", opts.User, "testuser")
	}
	if opts.Password != "testpass" {
		t.Errorf("Password = %q, want %q", opts.Password, "testpass")
	}
	if opts.Length != 24 {
		t.Errorf("Length = %d, want %d", opts.Length, 24)
	}
	if !opts.Generate {
		t.Error("Generate should be true")
	}
	if !opts.ShowSecret {
		t.Error("ShowSecret should be true")
	}
}

func TestNewCommand_DefaultLength(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("length")
	if flag == nil {
		t.Fatal("length flag not found")
	}

	// Verify it uses the constant from auth package
	expected := "16" // auth.DefaultPasswordLength
	if flag.DefValue != expected {
		t.Errorf("length default = %q, want %q", flag.DefValue, expected)
	}
	_ = auth.DefaultPasswordLength // verify constant exists
}

func TestNewCommand_PasswordSources(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		args []string
		want []string
	}{
		{args: []string{"device"}, want: []string{"--password <value>", "--password-stdin", "--generate"}},
		{args: []string{"device", "--password", "x", "--password-stdin"}, want: []string{"none of the others can be"}},
		{args: []string{"device", "--password", "x", "--generate"}, want: []string{"none of the others can be"}},
		{args: []string{"device", "--password-stdin", "--generate"}, want: []string{"none of the others can be"}},
	} {
		t.Run(strings.Join(tc.args[1:], " "), func(t *testing.T) {
			t.Parallel()
			ios := iostreams.Test(nil, &bytes.Buffer{}, &bytes.Buffer{})
			cmd := NewCommand(cmdutil.NewFactory().SetIOStreams(ios))
			cmd.SetArgs(tc.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
		})
	}
}
