// Package coiot provides CoIoT discovery command.
package coiot

import (
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"testing"
	"time"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "coiot" {
		t.Errorf("Use = %q, want %q", cmd.Use, "coiot")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
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
	case timeout.DefValue != "10s":
		t.Errorf("timeout default = %q, want %q", timeout.DefValue, "10s")
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

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()
	expected := 10 * time.Second
	if DefaultTimeout != expected {
		t.Errorf("DefaultTimeout = %v, want %v", DefaultTimeout, expected)
	}
}
