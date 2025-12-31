package create

import (
	"bytes"
	"os"
	"path/filepath"
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

	// Test Use
	if cmd.Use != "create <name>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create <name>")
	}

	// Test Aliases
	wantAliases := []string{"new", "init", "scaffold"}
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
		{"one arg valid", []string{"myext"}, false},
		{"two args", []string{"ext1", "ext2"}, true},
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
		shorthand string
		defValue  string
	}{
		{"lang", "l", "bash"},
		{"output", "o", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
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

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly extension create",
		"--lang",
		"--output",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"bash",
		"go",
		"python",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long description to contain %q", pattern)
		}
	}
}

func TestExecute_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "create <name>") {
		t.Error("help output should contain usage")
	}
	if !strings.Contains(output, "--lang") {
		t.Error("help output should contain --lang flag")
	}
	if !strings.Contains(output, "--output") {
		t.Error("help output should contain --output flag")
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with no args")
	}
}

func TestExecute_BashScaffold(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"mytest", "--lang", "bash", "--output", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check output
	output := tf.OutString()
	if !strings.Contains(output, "Created extension scaffold") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Next steps") {
		t.Errorf("expected next steps message, got: %s", output)
	}

	// Verify files were created
	extDir := filepath.Join(tmpDir, "shelly-mytest")
	scriptPath := filepath.Join(extDir, "shelly-mytest")
	readmePath := filepath.Join(extDir, "README.md")

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("expected bash script to be created")
	}
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("expected README.md to be created")
	}
}

func TestExecute_GoScaffold(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"goext", "--lang", "go", "--output", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify files were created
	extDir := filepath.Join(tmpDir, "shelly-goext")
	mainPath := filepath.Join(extDir, "main.go")
	modPath := filepath.Join(extDir, "go.mod")
	makefilePath := filepath.Join(extDir, "Makefile")
	readmePath := filepath.Join(extDir, "README.md")

	for _, path := range []string{mainPath, modPath, makefilePath, readmePath} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to be created", filepath.Base(path))
		}
	}
}

func TestExecute_PythonScaffold(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"pyext", "--lang", "python", "--output", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify files were created
	extDir := filepath.Join(tmpDir, "shelly-pyext")
	scriptPath := filepath.Join(extDir, "shelly-pyext")
	readmePath := filepath.Join(extDir, "README.md")

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("expected python script to be created")
	}
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("expected README.md to be created")
	}
}

func TestExecute_UnsupportedLanguage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"myext", "--lang", "ruby", "--output", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unsupported language")
	}
	if !strings.Contains(err.Error(), "unsupported language") {
		t.Errorf("error should mention unsupported language: %v", err)
	}
}

func TestExecute_LanguageAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		lang string
	}{
		{"sh alias", "sh"},
		{"golang alias", "golang"},
		{"py alias", "py"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()

			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)

			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs([]string{"test" + tt.lang, "--lang", tt.lang, "--output", tmpDir})

			err := cmd.Execute()
			if err != nil {
				t.Errorf("Execute() with lang %q error = %v", tt.lang, err)
			}
		})
	}
}

func TestExecute_DefaultOutputDir(t *testing.T) {
	t.Parallel()

	// Create a temp dir and change to it for this test
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Logf("warning: failed to restore working directory: %v", err)
		}
	}()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"defaultext"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify extension was created in current directory
	extDir := filepath.Join(tmpDir, "shelly-defaultext")
	if _, err := os.Stat(extDir); os.IsNotExist(err) {
		t.Error("expected extension directory in current directory")
	}
}

func TestExecute_NameNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase", "MYEXT", "shelly-myext"},
		{"mixed case", "MyExt", "shelly-myext"},
		{"with prefix", "shelly-myext", "shelly-myext"},
		{"with uppercase prefix", "SHELLY-MYEXT", "shelly-myext"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()

			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)

			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs([]string{tt.input, "--output", tmpDir})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			extDir := filepath.Join(tmpDir, tt.expected)
			if _, err := os.Stat(extDir); os.IsNotExist(err) {
				entries, readErr := os.ReadDir(tmpDir)
				if readErr != nil {
					t.Logf("warning: failed to read directory: %v", readErr)
				}
				var names []string
				for _, e := range entries {
					names = append(names, e.Name())
				}
				t.Errorf("expected directory %q, found: %v", tt.expected, names)
			}
		})
	}
}

func TestExecute_InvalidOutputDir(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	// Use a path that cannot be created
	cmd.SetArgs([]string{"myext", "--output", "/dev/null/invalid/path"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid output directory")
	}
	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("error should mention failed to create directory: %v", err)
	}
}

func TestExecute_OutputMessages(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"msgtest", "--output", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()

	// Check for success message
	if !strings.Contains(output, "Created extension scaffold") {
		t.Error("expected success message")
	}

	// Check for next steps
	if !strings.Contains(output, "Next steps") {
		t.Error("expected 'Next steps' message")
	}

	// Check for edit instruction
	if !strings.Contains(output, "Edit the extension code") {
		t.Error("expected edit instruction")
	}

	// Check for test instruction
	if !strings.Contains(output, "Test locally") {
		t.Error("expected test instruction")
	}

	// Check for install instruction
	if !strings.Contains(output, "Install:") {
		t.Error("expected install instruction")
	}
}

func TestRun_BashLanguage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, Name: "testbash", Lang: "bash", OutputDir: tmpDir}
	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify bash script content
	scriptPath := filepath.Join(tmpDir, "shelly-testbash", "shelly-testbash")
	content, err := os.ReadFile(scriptPath) //nolint:gosec // G304: scriptPath from t.TempDir()
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}
	if !strings.Contains(string(content), "#!/usr/bin/env bash") {
		t.Error("bash script should have shebang")
	}
}

func TestRun_GoLanguage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, Name: "testgo", Lang: "go", OutputDir: tmpDir}
	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify main.go content
	mainPath := filepath.Join(tmpDir, "shelly-testgo", "main.go")
	content, err := os.ReadFile(mainPath) //nolint:gosec // G304: mainPath from t.TempDir()
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	if !strings.Contains(string(content), "package main") {
		t.Error("main.go should have package main")
	}
}

func TestRun_PythonLanguage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, Name: "testpy", Lang: "python", OutputDir: tmpDir}
	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify python script content
	scriptPath := filepath.Join(tmpDir, "shelly-testpy", "shelly-testpy")
	content, err := os.ReadFile(scriptPath) //nolint:gosec // G304: scriptPath from t.TempDir()
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}
	if !strings.Contains(string(content), "#!/usr/bin/env python3") {
		t.Error("python script should have shebang")
	}
}

func TestRun_ShAlias(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, Name: "testsh", Lang: "sh", OutputDir: tmpDir}
	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// sh should create bash scaffold
	scriptPath := filepath.Join(tmpDir, "shelly-testsh", "shelly-testsh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("expected bash script to be created with sh alias")
	}
}

func TestRun_GolangAlias(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, Name: "testgolang", Lang: "golang", OutputDir: tmpDir}
	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// golang should create go scaffold
	mainPath := filepath.Join(tmpDir, "shelly-testgolang", "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Error("expected main.go to be created with golang alias")
	}
}

func TestRun_PyAlias(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, Name: "testpy2", Lang: "py", OutputDir: tmpDir}
	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// py should create python scaffold
	scriptPath := filepath.Join(tmpDir, "shelly-testpy2", "shelly-testpy2")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("expected python script to be created with py alias")
	}
}

func TestRun_UnsupportedLanguage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, Name: "testjava", Lang: "java", OutputDir: tmpDir}
	err := run(opts)
	if err == nil {
		t.Error("expected error for unsupported language")
	}
	if !strings.Contains(err.Error(), "unsupported language") {
		t.Errorf("error should mention unsupported language: %v", err)
	}
	if !strings.Contains(err.Error(), "java") {
		t.Errorf("error should include the unsupported language name: %v", err)
	}
}

func TestRun_DirectoryCreationError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	// Use a path that cannot be created
	opts := &Options{Factory: tf.Factory, Name: "test", Lang: "bash", OutputDir: "/dev/null/invalid"}
	err := run(opts)
	if err == nil {
		t.Error("expected error for invalid output directory")
	}
	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("error should mention failed to create directory: %v", err)
	}
}
