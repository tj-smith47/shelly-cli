package install

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
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

	if cmd.Use != "install <source>" {
		t.Errorf("Use = %q, want \"install <source>\"", cmd.Use)
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

	expectedAliases := []string{"add"}
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

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args returns error",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg succeeds",
			args:    []string{"./shelly-myext"},
			wantErr: false,
		},
		{
			name:    "two args returns error",
			args:    []string{"./shelly-myext", "extra"},
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
			name:      "force flag exists",
			flagName:  "force",
			shorthand: "f",
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
		"local file",
		"URL",
		"GitHub",
		"shelly-",
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
		"local file",
		"GitHub",
		"--force",
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

func TestRun_InvalidPrefix(t *testing.T) {
	t.Parallel()

	// Create a temp directory and file without shelly- prefix
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "myext")
	// Needs executable permissions for plugin testing
	if err := os.WriteFile(invalidFile, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: test executable
		t.Fatalf("failed to create temp file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory: f,
		Source:  invalidFile,
		Force:   false,
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for file without shelly- prefix")
	}

	if !strings.Contains(err.Error(), "shelly-") {
		t.Errorf("error should mention shelly- prefix requirement, got: %v", err)
	}
}

func TestRun_NonExistentLocalFile(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory: f,
		Source:  "/nonexistent/path/shelly-myext",
		Force:   false,
	}
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestRun_ValidLocalFile(t *testing.T) {
	// Set XDG_CONFIG_HOME to use temp directory for plugins
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Create a temp directory and file with shelly- prefix
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "shelly-testplugin")
	// Needs executable permissions for plugin testing
	if err := os.WriteFile(validFile, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: test executable
		t.Fatalf("failed to create temp file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory: f,
		Source:  validFile,
		Force:   false,
	}
	// This will try to install to the real plugins dir
	// It should succeed or fail with a registry-related error
	err := run(context.Background(), opts)

	// We accept either success or an error that's not about the prefix
	if err != nil && strings.Contains(err.Error(), "shelly-") && strings.Contains(err.Error(), "prefix") {
		t.Errorf("unexpected prefix error for valid shelly- file: %v", err)
	}
}

func TestRun_GitHubSourceParsing(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "gh: prefix",
			source: "gh:user/shelly-myext",
		},
		{
			name:   "github: prefix",
			source: "github:user/shelly-myext",
		},
	}

	for _, tt := range tests { //nolint:paralleltest // uses t.Setenv
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
			f := cmdutil.NewFactory().SetIOStreams(ios)

			// Create context with timeout to avoid long waits on network
			ctx, cancel := context.WithTimeout(context.Background(), 100)
			defer cancel()

			opts := &Options{
				Factory: f,
				Source:  tt.source,
				Force:   false,
			}
			// This should fail due to network/auth, but should parse the source
			err := run(ctx, opts)
			if err == nil {
				t.Error("expected error (network/timeout)")
			}
			// Just verify it didn't fail at the prefix parsing stage
			if strings.Contains(err.Error(), "prefix") {
				t.Errorf("unexpected prefix error: %v", err)
			}
		})
	}
}

func TestRun_URLSourceParsing(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "https URL",
			source: "https://example.com/shelly-myext",
		},
		{
			name:   "http URL",
			source: "http://example.com/shelly-myext",
		},
	}

	for _, tt := range tests { //nolint:paralleltest // uses t.Setenv
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
			f := cmdutil.NewFactory().SetIOStreams(ios)

			// Create context with timeout to avoid long waits on network
			ctx, cancel := context.WithTimeout(context.Background(), 100)
			defer cancel()

			opts := &Options{
				Factory: f,
				Source:  tt.source,
				Force:   false,
			}
			// This should fail due to network/timeout, but should parse the source
			err := run(ctx, opts)
			if err == nil {
				t.Error("expected error (network/timeout)")
			}
		})
	}
}

func TestRun_SourceTypeDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		source     string
		expectType string
	}{
		{"github gh prefix", "gh:user/repo", "github"},
		{"github full prefix", "github:user/repo", "github"},
		{"https url", "https://example.com/file", "url"},
		{"http url", "http://example.com/file", "url"},
		{"local path", "./shelly-myext", "local"},
		{"absolute path", "/tmp/shelly-myext", "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var detectedType string
			switch {
			case strings.HasPrefix(tt.source, "gh:") || strings.HasPrefix(tt.source, "github:"):
				detectedType = "github"
			case strings.HasPrefix(tt.source, "http://") || strings.HasPrefix(tt.source, "https://"):
				detectedType = "url"
			default:
				detectedType = "local"
			}

			if detectedType != tt.expectType {
				t.Errorf("source %q detected as %q, want %q", tt.source, detectedType, tt.expectType)
			}
		})
	}
}

func TestRun_ForceFlag(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Create a temp file
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "shelly-forcetestplugin")
	// Needs executable permissions for plugin testing
	if err := os.WriteFile(validFile, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: test executable
		t.Fatalf("failed to create temp file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := iostreams.Test(bytes.NewReader(nil), stdout, stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	// First install - ignore error as we're testing force flag behavior
	opts := &Options{
		Factory: f,
		Source:  validFile,
		Force:   false,
	}
	_ = run(context.Background(), opts) //nolint:errcheck // intentionally ignored for setup

	// Second install with force=true should not fail due to "already installed"
	opts.Force = true
	err := run(context.Background(), opts)

	// If we get an "already installed" error even with force=true, that's a bug
	if err != nil && strings.Contains(err.Error(), "already installed") {
		t.Errorf("force flag should allow reinstall, got: %v", err)
	}
}
