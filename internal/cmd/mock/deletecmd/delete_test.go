package deletecmd

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
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

	if cmd.Use != "delete <name>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <name>")
	}

	wantAliases := []string{"rm", "remove"}
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
		"shelly mock delete",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestRun_NotFound(t *testing.T) {
	// Set up in-memory filesystem
	factory.SetupTestFs(t)

	tf := factory.NewTestFactory(t)

	err := run(context.Background(), tf.Factory, "nonexistent-device-12345")
	if err == nil {
		t.Error("Expected error for nonexistent mock device")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found': %v", err)
	}
}

func TestRun_DeleteExisting(t *testing.T) {
	// Set up in-memory filesystem
	fs := factory.SetupTestFs(t)

	tf := factory.NewTestFactory(t)

	// Create a mock device file in the in-memory filesystem
	mockDir, err := mock.Dir()
	if err != nil {
		t.Fatalf("mock.Dir: %v", err)
	}

	// Ensure mock directory exists in the in-memory filesystem
	if err := fs.MkdirAll(mockDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	testName := "test-delete-device"
	filename := filepath.Join(mockDir, testName+".json")

	// Create the file in the in-memory filesystem
	if err := afero.WriteFile(fs, filename, []byte(`{"name":"test"}`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Delete it
	err = run(context.Background(), tf.Factory, testName)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify it's gone from the in-memory filesystem
	if _, err := fs.Stat(filename); err == nil {
		t.Error("File should have been deleted")
	}

	out := tf.OutString()
	if !strings.Contains(out, "Deleted") {
		t.Errorf("Output should contain 'Deleted', got: %s", out)
	}
}
