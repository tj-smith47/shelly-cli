// Package ble provides BLE discovery command.
package ble

import (
	"errors"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly/wireless"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
	}

	if cmd.Use != "ble" {
		t.Errorf("Use = %q, want %q", cmd.Use, "ble")
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
	case timeout.DefValue != "15s":
		t.Errorf("timeout default = %q, want %q", timeout.DefValue, "15s")
	}

	// Test bthome flag exists
	bthome := cmd.Flags().Lookup("bthome")
	if bthome == nil {
		t.Error("bthome flag not found")
	} else if bthome.DefValue != "false" {
		t.Errorf("bthome default = %q, want %q", bthome.DefValue, "false")
	}

	// Test filter flag exists
	filter := cmd.Flags().Lookup("filter")
	if filter == nil {
		t.Error("filter flag not found")
	} else if filter.Shorthand != "f" {
		t.Errorf("filter shorthand = %q, want %q", filter.Shorthand, "f")
	}
}

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()
	expected := 15 * time.Second
	if DefaultTimeout != expected {
		t.Errorf("DefaultTimeout = %v, want %v", DefaultTimeout, expected)
	}
}

func TestIsBLENotSupportedError_NilError(t *testing.T) {
	t.Parallel()
	if wireless.IsBLENotSupportedError(nil) {
		t.Error("IsBLENotSupportedError(nil) = true, want false")
	}
}

func TestIsBLENotSupportedError_GenericError(t *testing.T) {
	t.Parallel()
	err := errors.New("some other error")
	if wireless.IsBLENotSupportedError(err) {
		t.Error("IsBLENotSupportedError(generic) = true, want false")
	}
}

func TestIsBLENotSupportedError_NotSupportedError(t *testing.T) {
	t.Parallel()
	if !wireless.IsBLENotSupportedError(discovery.ErrBLENotSupported) {
		t.Error("IsBLENotSupportedError(ErrBLENotSupported) = false, want true")
	}
}

func TestIsBLENotSupportedError_WrappedError(t *testing.T) {
	t.Parallel()
	wrappedErr := &discovery.BLEError{
		Message: "BLE not supported",
		Err:     discovery.ErrBLENotSupported,
	}
	if !wireless.IsBLENotSupportedError(wrappedErr) {
		t.Error("IsBLENotSupportedError(wrapped) = false, want true")
	}
}
