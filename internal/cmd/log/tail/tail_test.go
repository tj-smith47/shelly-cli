package tail

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const testConfigDir = "/test/config"

// setupTestEnv sets up an isolated config environment for tests using afero.
// Returns the filesystem and the config directory path.
func setupTestEnv(t *testing.T) (fs afero.Fs, configDir string) {
	t.Helper()
	fs = afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	configDir = testConfigDir
	t.Setenv("HOME", testConfigDir)
	t.Setenv("XDG_CONFIG_HOME", testConfigDir)
	return fs, configDir
}

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "tail" {
		t.Errorf("Use = %q, want %q", cmd.Use, "tail")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"follow", "f"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
		return
	}

	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept no arguments
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("Expected no error with no args, got: %v", err)
	}

	// Should reject any arguments
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("Expected error when extra args provided")
	}
}

func TestNewCommand_FollowFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	followFlag := cmd.Flags().Lookup("follow")
	if followFlag == nil {
		t.Fatal("--follow flag not defined")
	}

	if followFlag.Shorthand != "f" {
		t.Errorf("follow flag shorthand = %q, want %q", followFlag.Shorthand, "f")
	}

	if followFlag.DefValue != "false" {
		t.Errorf("follow flag default = %q, want %q", followFlag.DefValue, "false")
	}
}

func TestNewCommand_FlagValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantFollow bool
	}{
		{
			name:       "default values",
			args:       []string{},
			wantFollow: false,
		},
		{
			name:       "follow flag long",
			args:       []string{"--follow"},
			wantFollow: true,
		},
		{
			name:       "follow flag short",
			args:       []string{"-f"},
			wantFollow: true,
		},
		{
			name:       "follow flag explicit true",
			args:       []string{"--follow=true"},
			wantFollow: true,
		},
		{
			name:       "follow flag explicit false",
			args:       []string{"--follow=false"},
			wantFollow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(cmdutil.NewFactory())
			cmd.SetArgs(tt.args)

			// Parse flags only (don't execute RunE)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("ParseFlags error: %v", err)
			}

			gotFollow, err := cmd.Flags().GetBool("follow")
			if err != nil {
				t.Fatalf("GetBool(follow) error: %v", err)
			}
			if gotFollow != tt.wantFollow {
				t.Errorf("follow = %v, want %v", gotFollow, tt.wantFollow)
			}
		})
	}
}

func TestOptions_Struct(t *testing.T) {
	t.Parallel()

	// Verify Options struct has expected fields
	opts := &Options{
		Follow: true,
	}

	if !opts.Follow {
		t.Error("Options.Follow should be settable to true")
	}

	opts.Follow = false
	if opts.Follow {
		t.Error("Options.Follow should be settable to false")
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_NoLogFile(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory but no log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Follow: false}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should output info message about no log file
	// The message goes to errOut since ios.Info writes to stderr
	combined := out.String() + errOut.String()
	if combined == "" {
		t.Error("Expected some output when log file doesn't exist")
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_WithLogFile(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory and log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	// Create log file with some content
	logPath := shellyDir + "/shelly.log"
	logContent := "line1\nline2\nline3\nline4\nline5\n"
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Follow: false}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should output the log lines
	output := out.String()
	if output == "" {
		t.Error("Expected log content in output")
	}

	// Verify all lines are present
	for i := 1; i <= 5; i++ {
		expected := fmt.Sprintf("line%d", i)
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_ShowsLast20Lines(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory and log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	// Create log file with 30 lines using proper formatting
	var logContent string
	for i := 1; i <= 30; i++ {
		logContent += fmt.Sprintf("line%02d\n", i)
	}
	logPath := shellyDir + "/shelly.log"
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Follow: false}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := out.String()

	// Should NOT contain first 10 lines (lines 1-10)
	if bytes.Contains([]byte(output), []byte("line01")) {
		t.Error("Output should not contain line01 (should only show last 20)")
	}

	// Should contain last 20 lines (lines 11-30)
	if !bytes.Contains([]byte(output), []byte("line11")) {
		t.Error("Output should contain line11")
	}
	if !bytes.Contains([]byte(output), []byte("line30")) {
		t.Error("Output should contain line30")
	}
}

func TestNewCommand_CommandChain(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Verify command can be properly attached to parent
	if cmd.HasParent() {
		t.Error("Fresh command should not have parent")
	}

	// Verify command has RunE set
	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}

	// Verify Args validator is set
	if cmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_EmptyLogFile(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory and empty log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	// Create empty log file
	logPath := shellyDir + "/shelly.log"
	if err := afero.WriteFile(fs, logPath, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Follow: false}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// With empty file, output should be empty (no lines to print)
	output := out.String()
	if output != "" {
		t.Errorf("Expected empty output for empty log file, got: %q", output)
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_FewerThan20Lines(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory and log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	// Create log file with only 5 lines (fewer than the 20 line limit)
	var logContent string
	for i := 1; i <= 5; i++ {
		logContent += fmt.Sprintf("log entry %d\n", i)
	}
	logPath := shellyDir + "/shelly.log"
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Follow: false}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should output all 5 lines
	output := out.String()
	for i := 1; i <= 5; i++ {
		expected := fmt.Sprintf("log entry %d", i)
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_Exactly20Lines(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory and log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	// Create log file with exactly 20 lines
	var logContent string
	for i := 1; i <= 20; i++ {
		logContent += fmt.Sprintf("log line %02d\n", i)
	}
	logPath := shellyDir + "/shelly.log"
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Follow: false}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should output all 20 lines
	output := out.String()
	for i := 1; i <= 20; i++ {
		expected := fmt.Sprintf("log line %02d", i)
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestCommand_Execute(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory and log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	// Create log file with some content
	logPath := shellyDir + "/shelly.log"
	if err := afero.WriteFile(fs, logPath, []byte("test line\n"), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{}) // No arguments

	// Execute the command (this tests the RunE callback)
	err := cmd.Execute()

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Should output the log content
	output := out.String()
	if !bytes.Contains([]byte(output), []byte("test line")) {
		t.Errorf("Expected output to contain 'test line', got: %q", output)
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_SingleLine(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory and log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	// Create log file with a single line
	logPath := shellyDir + "/shelly.log"
	if err := afero.WriteFile(fs, logPath, []byte("single line\n"), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Follow: false}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should output the single line
	output := out.String()
	if !bytes.Contains([]byte(output), []byte("single line")) {
		t.Errorf("Expected output to contain 'single line', got: %q", output)
	}
}

//nolint:paralleltest // Modifies global state via config.SetFs
func TestRun_NoTrailingNewline(t *testing.T) {
	fs, configDir := setupTestEnv(t)

	// Create shelly config directory and log file
	shellyDir := configDir + "/shelly"
	if err := fs.MkdirAll(shellyDir, 0o750); err != nil {
		t.Fatalf("failed to create shelly config dir: %v", err)
	}

	// Create log file without trailing newline
	logPath := shellyDir + "/shelly.log"
	if err := afero.WriteFile(fs, logPath, []byte("line1\nline2"), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewWithIOStreams(ios)

	opts := &Options{Factory: f, Follow: false}
	err := run(opts)

	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should output both lines
	output := out.String()
	if !bytes.Contains([]byte(output), []byte("line1")) {
		t.Errorf("Expected output to contain 'line1', got: %q", output)
	}
	if !bytes.Contains([]byte(output), []byte("line2")) {
		t.Errorf("Expected output to contain 'line2', got: %q", output)
	}
}

func TestNewCommand_Properties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		check  func(*cobra.Command) bool
		errMsg string
	}{
		{
			name: "Use field",
			check: func(cmd *cobra.Command) bool {
				return cmd.Use == "tail"
			},
			errMsg: "Use should be 'tail'",
		},
		{
			name: "Short description",
			check: func(cmd *cobra.Command) bool {
				return cmd.Short == "Tail log file"
			},
			errMsg: "Short should be 'Tail log file'",
		},
		{
			name: "Long description exists",
			check: func(cmd *cobra.Command) bool {
				return cmd.Long != ""
			},
			errMsg: "Long description should not be empty",
		},
		{
			name: "Example contains shelly log tail",
			check: func(cmd *cobra.Command) bool {
				return bytes.Contains([]byte(cmd.Example), []byte("shelly log tail"))
			},
			errMsg: "Example should contain 'shelly log tail'",
		},
		{
			name: "Has follow alias",
			check: func(cmd *cobra.Command) bool {
				for _, alias := range cmd.Aliases {
					if alias == "follow" {
						return true
					}
				}
				return false
			},
			errMsg: "Should have 'follow' alias",
		},
		{
			name: "Has f alias",
			check: func(cmd *cobra.Command) bool {
				for _, alias := range cmd.Aliases {
					if alias == "f" {
						return true
					}
				}
				return false
			},
			errMsg: "Should have 'f' alias",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if !tt.check(cmd) {
				t.Error(tt.errMsg)
			}
		})
	}
}
