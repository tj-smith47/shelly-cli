// Package discover provides device discovery commands.
package discover

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "discover" {
		t.Errorf("Use = %q, want %q", cmd.Use, "discover")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify subcommands are registered
	subcommands := cmd.Commands()
	expectedSubcmds := map[string]bool{
		"mdns":          false,
		"ble":           false,
		"coiot":         false,
		"http [subnet]": false,
	}

	for _, sub := range subcommands {
		if _, ok := expectedSubcmds[sub.Use]; ok {
			expectedSubcmds[sub.Use] = true
		}
	}

	for name, found := range expectedSubcmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test timeout flag exists
	timeout := cmd.Flags().Lookup("timeout")
	switch {
	case timeout == nil:
		t.Error("timeout flag not found")
	case timeout.Shorthand != "t":
		t.Errorf("timeout shorthand = %q, want %q", timeout.Shorthand, "t")
	case timeout.DefValue != "2m0s":
		t.Errorf("timeout default = %q, want %q", timeout.DefValue, "2m0s")
	}

	// Test register flag exists
	register := cmd.Flags().Lookup("register")
	if register == nil {
		t.Error("register flag not found")
	} else if register.DefValue != "false" {
		t.Errorf("register default = %q, want %q", register.DefValue, "false")
	}

	// Test skip-existing flag exists
	skipExisting := cmd.Flags().Lookup("skip-existing")
	if skipExisting == nil {
		t.Error("skip-existing flag not found")
	} else if skipExisting.DefValue != "true" {
		t.Errorf("skip-existing default = %q, want %q", skipExisting.DefValue, "true")
	}
}

func TestDefaultScanTimeout(t *testing.T) {
	t.Parallel()
	expected := 2 * time.Minute
	if DefaultScanTimeout != expected {
		t.Errorf("DefaultScanTimeout = %v, want %v", DefaultScanTimeout, expected)
	}
}

func TestNewCommand_SubcommandCount(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should have exactly 4 subcommands
	if len(cmd.Commands()) != 4 {
		t.Errorf("subcommand count = %d, want 4", len(cmd.Commands()))
	}
}
