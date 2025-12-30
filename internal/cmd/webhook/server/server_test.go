package server

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

	if cmd.Use != "server" {
		t.Errorf("Use = %q, want %q", cmd.Use, "server")
	}

	wantAliases := []string{"serve", "listen", "receiver"}
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

	// Server command takes no args
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("Args should accept no arguments: %v", err)
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("Args should reject arguments")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		defValue string
	}{
		{"port", "8080"},
		{"interface", "0.0.0.0"},
		{"log-json", "false"},
		{"auto-config", "false"},
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

	deviceFlag := cmd.Flags().Lookup("device")
	if deviceFlag == nil {
		t.Error("--device flag not found")
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
		"shelly webhook server",
		"--port",
		"--log-json",
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
		Port:       9000,
		Interface:  "localhost",
		LogJSON:    true,
		AutoConfig: true,
		Devices:    []string{"dev1", "dev2"},
	}

	if opts.Port != 9000 {
		t.Errorf("Port = %d, want %d", opts.Port, 9000)
	}

	if opts.Interface != "localhost" {
		t.Errorf("Interface = %q, want %q", opts.Interface, "localhost")
	}

	if !opts.LogJSON {
		t.Error("LogJSON should be true")
	}

	if !opts.AutoConfig {
		t.Error("AutoConfig should be true")
	}

	if len(opts.Devices) != 2 {
		t.Errorf("Devices length = %d, want %d", len(opts.Devices), 2)
	}
}
