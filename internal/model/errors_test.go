package model

import (
	"errors"
	"testing"
)

func TestErrors_AreDistinct(t *testing.T) {
	t.Parallel()

	errList := []error{
		ErrDeviceNotFound,
		ErrDeviceExists,
		ErrConnectionFailed,
		ErrComponentNotFound,
		ErrInvalidDeviceName,
		ErrAuthRequired,
		ErrTimeout,
		ErrNoDevices,
		ErrNoComponents,
	}

	// Ensure all errors are distinct
	for i, err1 := range errList {
		for j, err2 := range errList {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("errors should be distinct: %v and %v", err1, err2)
			}
		}
	}
}

func TestErrors_Messages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrDeviceNotFound", ErrDeviceNotFound, "device not found"},
		{"ErrDeviceExists", ErrDeviceExists, "device already exists"},
		{"ErrConnectionFailed", ErrConnectionFailed, "failed to connect to device"},
		{"ErrComponentNotFound", ErrComponentNotFound, "component not found"},
		{"ErrInvalidDeviceName", ErrInvalidDeviceName, "invalid device name"},
		{"ErrAuthRequired", ErrAuthRequired, "device requires authentication"},
		{"ErrTimeout", ErrTimeout, "operation timed out"},
		{"ErrNoDevices", ErrNoDevices, "no devices found"},
		{"ErrNoComponents", ErrNoComponents, "no matching components found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

func TestErrors_Wrapping(t *testing.T) {
	t.Parallel()

	wrapped := errors.Join(ErrConnectionFailed, errors.New("network unreachable"))

	if !errors.Is(wrapped, ErrConnectionFailed) {
		t.Error("wrapped error should match ErrConnectionFailed")
	}
}
