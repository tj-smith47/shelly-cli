package importcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "import <file>" {
		t.Errorf("Use = %q, want 'import <file>'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Fatal("Aliases are empty")
	}

	// Check for expected aliases
	expectedAliases := []string{"imp", "restore"}
	for _, expected := range expectedAliases {
		found := false
		for _, alias := range cmd.Aliases {
			if alias == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected alias %q not found in %v", expected, cmd.Aliases)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Check that example contains expected patterns
	expectedPatterns := []string{
		"shelly auth import",
		"credentials.json",
		"--dry-run",
	}
	for _, pattern := range expectedPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check dry-run flag
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("dry-run flag not found")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("dry-run flag default = %q, want 'false'", dryRunFlag.DefValue)
	}
}

func TestNewCommand_RequiresOneArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"credentials.json"},
			wantErr: false,
		},
		{
			name:    "two args",
			args:    []string{"file1.json", "file2.json"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		check  func() bool
		wantOK bool
		errMsg string
	}{
		{
			name:   "has Use",
			check:  func() bool { return NewCommand(cmdutil.NewFactory()).Use != "" },
			wantOK: true,
			errMsg: "Use should not be empty",
		},
		{
			name:   "has Short",
			check:  func() bool { return NewCommand(cmdutil.NewFactory()).Short != "" },
			wantOK: true,
			errMsg: "Short should not be empty",
		},
		{
			name:   "has Long",
			check:  func() bool { return NewCommand(cmdutil.NewFactory()).Long != "" },
			wantOK: true,
			errMsg: "Long should not be empty",
		},
		{
			name:   "has Example",
			check:  func() bool { return NewCommand(cmdutil.NewFactory()).Example != "" },
			wantOK: true,
			errMsg: "Example should not be empty",
		},
		{
			name:   "has Aliases",
			check:  func() bool { return len(NewCommand(cmdutil.NewFactory()).Aliases) > 0 },
			wantOK: true,
			errMsg: "Aliases should not be empty",
		},
		{
			name:   "has RunE",
			check:  func() bool { return NewCommand(cmdutil.NewFactory()).RunE != nil },
			wantOK: true,
			errMsg: "RunE should be set",
		},
		{
			name:   "has Args",
			check:  func() bool { return NewCommand(cmdutil.NewFactory()).Args != nil },
			wantOK: true,
			errMsg: "Args should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.check() != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "dry-run flag",
			args:    []string{"--dry-run"},
			wantErr: false,
		},
		{
			name:    "no flags",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_FileNotFound(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"/nonexistent/path/credentials.json"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "read file") {
		t.Errorf("expected 'read file' error, got: %v", err)
	}
}

func TestRun_InvalidJSON(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("not valid json"), 0o600); err != nil {
		t.Fatalf("failed to create invalid file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{invalidFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid JSON file")
	}

	if !strings.Contains(err.Error(), "parse file") {
		t.Errorf("expected 'parse file' error, got: %v", err)
	}
}

func TestRun_EmptyCredentials(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.json")

	// Create a valid JSON file with empty credentials
	export := map[string]any{
		"exported_at": "2024-01-15T10:30:00Z",
		"version":     "1.0.0",
		"credentials": map[string]any{},
	}
	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(emptyFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{emptyFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Warning output goes to stderr
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "No credentials found") {
		t.Errorf("expected 'No credentials found' warning in stderr, got: %s", errOutput)
	}
}

func TestRun_DryRun(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	credFile := filepath.Join(tmpDir, "credentials.json")

	// Create a valid credentials file
	export := map[string]any{
		"exported_at": "2024-01-15T10:30:00Z",
		"version":     "1.0.0",
		"credentials": map[string]any{
			"device1": map[string]string{
				"Username": "admin",
				"Password": "secret123",
			},
			"device2": map[string]string{
				"Username": "user",
				"Password": "pass456",
			},
		},
	}
	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(credFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--dry-run", credFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()

	// Should show dry-run indicator
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected 'dry-run' in output, got: %s", output)
	}

	// Should show credential count
	if !strings.Contains(output, "2") {
		t.Errorf("expected credential count '2' in output, got: %s", output)
	}
}

func TestRun_DryRunShowsDevices(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	credFile := filepath.Join(tmpDir, "credentials.json")

	// Create a valid credentials file
	export := map[string]any{
		"exported_at": "2024-01-15T10:30:00Z",
		"version":     "1.0.0",
		"credentials": map[string]any{
			"kitchen-switch": map[string]string{
				"Username": "admin",
				"Password": "secret123",
			},
		},
	}
	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(credFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--dry-run", credFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()

	// Should show device name
	if !strings.Contains(output, "kitchen-switch") {
		t.Errorf("expected 'kitchen-switch' in output, got: %s", output)
	}

	// Should show username (but not password)
	if !strings.Contains(output, "admin") {
		t.Errorf("expected 'admin' in output, got: %s", output)
	}
}

func TestRun_ImportWithTestFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	tmpDir := t.TempDir()
	credFile := filepath.Join(tmpDir, "credentials.json")

	// Create a valid credentials file
	export := map[string]any{
		"exported_at": "2024-01-15T10:30:00Z",
		"version":     "1.0.0",
		"credentials": map[string]any{
			"device1": map[string]string{
				"Username": "admin",
				"Password": "secret123",
			},
		},
	}
	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(credFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{credFile})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()

	// Should show success message
	if !strings.Contains(output, "Imported") {
		t.Errorf("expected 'Imported' in output, got: %s", output)
	}
}

func TestRun_DirectoryAsFile(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0o750); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{subDir})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when file is a directory")
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Input != "" {
		t.Errorf("Default Input = %q, want empty", opts.Input)
	}
	if opts.DryRun {
		t.Error("Default DryRun should be false")
	}
}

func TestOptions_FieldsSet(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Input:  "/path/to/creds.json",
		DryRun: true,
	}

	if opts.Input != "/path/to/creds.json" {
		t.Errorf("Input = %q, want '/path/to/creds.json'", opts.Input)
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
}

func TestRun_MissingExportedAt(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	tmpDir := t.TempDir()
	credFile := filepath.Join(tmpDir, "credentials.json")

	// Create a file with credentials but missing exported_at
	export := map[string]any{
		"version": "1.0.0",
		"credentials": map[string]any{
			"device1": map[string]string{
				"Username": "admin",
				"Password": "secret",
			},
		},
	}
	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(credFile, data, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--dry-run", credFile})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Should still work - exported_at is optional for display
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCommand_Long_Description(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	long := cmd.Long

	// Check that long description contains key information
	keywords := []string{
		"Import",
		"credentials",
		"export",
	}

	for _, kw := range keywords {
		if !strings.Contains(long, kw) {
			t.Errorf("expected long description to contain %q", kw)
		}
	}
}

func TestRun_ValidCredentialsFileFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content map[string]any
		wantErr bool
	}{
		{
			name: "standard format",
			content: map[string]any{
				"exported_at": "2024-01-15T10:30:00Z",
				"version":     "1.0.0",
				"credentials": map[string]any{
					"device1": map[string]string{
						"Username": "admin",
						"Password": "pass",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple devices",
			content: map[string]any{
				"exported_at": "2024-01-15T10:30:00Z",
				"version":     "2.0.0",
				"credentials": map[string]any{
					"device1": map[string]string{"Username": "user1", "Password": "pass1"},
					"device2": map[string]string{"Username": "user2", "Password": "pass2"},
					"device3": map[string]string{"Username": "user3", "Password": "pass3"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty credentials object",
			content: map[string]any{
				"exported_at": "2024-01-15T10:30:00Z",
				"version":     "1.0.0",
				"credentials": map[string]any{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			errOut := &bytes.Buffer{}
			ios := iostreams.Test(nil, out, errOut)
			f := cmdutil.NewFactory().SetIOStreams(ios)

			tmpDir := t.TempDir()
			credFile := filepath.Join(tmpDir, "credentials.json")

			data, err := json.Marshal(tt.content)
			if err != nil {
				t.Fatalf("failed to marshal JSON: %v", err)
			}
			if err := os.WriteFile(credFile, data, 0o600); err != nil {
				t.Fatalf("failed to write file: %v", err)
			}

			cmd := NewCommand(f)
			cmd.SetArgs([]string{"--dry-run", credFile})
			cmd.SetOut(out)
			cmd.SetErr(errOut)

			err = cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_InvalidJSONFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "invalid JSON syntax",
			content: "{not valid json}",
			wantErr: true,
		},
		{
			name:    "malformed braces",
			content: "{{}}",
			wantErr: true,
		},
		{
			name:    "plain text",
			content: "just some text",
			wantErr: true,
		},
		{
			name:    "array instead of object",
			content: "[1, 2, 3]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			errOut := &bytes.Buffer{}
			ios := iostreams.Test(nil, out, errOut)
			f := cmdutil.NewFactory().SetIOStreams(ios)

			tmpDir := t.TempDir()
			credFile := filepath.Join(tmpDir, "credentials.json")

			if err := os.WriteFile(credFile, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("failed to write file: %v", err)
			}

			cmd := NewCommand(f)
			cmd.SetArgs([]string{credFile})
			cmd.SetOut(out)
			cmd.SetErr(errOut)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
