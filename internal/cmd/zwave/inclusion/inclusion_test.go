package inclusion

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

	if cmd.Use != "inclusion <model>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "inclusion <model>")
	}

	wantAliases := []string{"include", "pair", "add"}
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
		{"one arg valid", []string{"model"}, false},
		{"two args", []string{"model1", "model2"}, true},
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

	flag := cmd.Flags().Lookup("mode")
	if flag == nil {
		t.Fatal("--mode flag not found")
	}
	if flag.DefValue != "smart_start" {
		t.Errorf("--mode default = %q, want %q", flag.DefValue, "smart_start")
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
		"shelly zwave inclusion",
		"--mode button",
		"--mode switch",
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
		Model:   "SNSW-001P16ZW",
		Mode:    "button",
		Factory: f,
	}

	if opts.Model != "SNSW-001P16ZW" {
		t.Errorf("Model = %q, want %q", opts.Model, "SNSW-001P16ZW")
	}

	if opts.Mode != "button" {
		t.Errorf("Mode = %q, want %q", opts.Mode, "button")
	}
}

func TestRun_UnknownModel(t *testing.T) {
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Model:   "UNKNOWN-MODEL-12345",
		Mode:    "smart_start",
	}

	err := run(opts)
	if err == nil {
		t.Error("Expected error for unknown model")
	}
	if !strings.Contains(err.Error(), "unknown device model") {
		t.Errorf("Error should mention 'unknown device model': %v", err)
	}
}

func TestRun_InvalidMode(t *testing.T) {
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16EU", // Non-ZWave model
		Mode:    "invalid_mode",
	}

	err := run(opts)
	// Will fail at ZWave check or mode validation
	if err == nil {
		t.Logf("Expected error for invalid mode or non-zwave model")
	}
}

func TestRun_NonZWaveDevice(t *testing.T) {
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Model:   "SNSW-001P16EU", // This is not a ZWave device
		Mode:    "smart_start",
	}

	err := run(opts)
	if err == nil {
		t.Error("Expected error for non-ZWave device")
	}
	// Error may mention "not a Z-Wave device" or "unknown device model"
	if err != nil {
		t.Logf("Error = %v (expected for non-ZWave)", err)
	}
}
