// Package tuierrors provides error categorization for TUI components.
package tuierrors

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestCategorize_Nil(t *testing.T) {
	t.Parallel()

	result := Categorize(nil)
	if result.Original != nil {
		t.Error("Categorize(nil).Original should be nil")
	}
	if result.Category != CategoryUnknown {
		t.Errorf("Categorize(nil).Category = %d, want %d", result.Category, CategoryUnknown)
	}
}

func TestCategorize_Timeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"context deadline exceeded", context.DeadlineExceeded},
		{"timeout in message", errors.New("connection timeout")},
		{"timed out in message", errors.New("operation timed out")},
		{"deadline exceeded message", errors.New("deadline exceeded")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Categorize(tt.err)
			if result.Category != CategoryTimeout {
				t.Errorf("Categorize(%q).Category = %d, want %d (Timeout)",
					tt.err.Error(), result.Category, CategoryTimeout)
			}
			if result.Message != "Request timed out" {
				t.Errorf("Categorize(%q).Message = %q, want 'Request timed out'",
					tt.err.Error(), result.Message)
			}
			if result.Hint == "" {
				t.Error("Hint should not be empty")
			}
		})
	}
}

func TestCategorize_Network(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"connection refused", errors.New("connection refused")},
		{"no route to host", errors.New("no route to host")},
		{"network unreachable", errors.New("network unreachable")},
		{"host unreachable", errors.New("host unreachable")},
		{"no such host", errors.New("no such host")},
		{"dial tcp", errors.New("dial tcp: connection failed")},
		// Note: "i/o timeout" is categorized as Timeout, not Network, because it contains "timeout"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Categorize(tt.err)
			if result.Category != CategoryNetwork {
				t.Errorf("Categorize(%q).Category = %d, want %d (Network)",
					tt.err.Error(), result.Category, CategoryNetwork)
			}
			if result.Message != "Network error" {
				t.Errorf("Categorize(%q).Message = %q, want 'Network error'",
					tt.err.Error(), result.Message)
			}
		})
	}
}

func TestCategorize_Auth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"401 status", errors.New("HTTP 401 response")},
		{"403 status", errors.New("HTTP 403 forbidden")},
		{"unauthorized", errors.New("request unauthorized")},
		{"forbidden", errors.New("access forbidden")},
		{"authentication", errors.New("authentication required")},
		{"invalid password", errors.New("invalid password")},
		{"access denied", errors.New("access denied")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Categorize(tt.err)
			if result.Category != CategoryAuth {
				t.Errorf("Categorize(%q).Category = %d, want %d (Auth)",
					tt.err.Error(), result.Category, CategoryAuth)
			}
			if result.Message != "Authentication failed" {
				t.Errorf("Categorize(%q).Message = %q, want 'Authentication failed'",
					tt.err.Error(), result.Message)
			}
		})
	}
}

func TestCategorize_Device(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"device error", errors.New("device not responding")},
		{"shelly error", errors.New("shelly returned error")},
		{"component error", errors.New("component unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Categorize(tt.err)
			if result.Category != CategoryDevice {
				t.Errorf("Categorize(%q).Category = %d, want %d (Device)",
					tt.err.Error(), result.Category, CategoryDevice)
			}
		})
	}
}

func TestCategorize_Unknown(t *testing.T) {
	t.Parallel()

	err := errors.New("some random unexpected error")
	result := Categorize(err)

	if result.Category != CategoryUnknown {
		t.Errorf("Categorize().Category = %d, want %d (Unknown)", result.Category, CategoryUnknown)
	}
	if !errors.Is(result.Original, err) {
		t.Error("Original error should be preserved")
	}
}

func TestCategorizedError_Error(t *testing.T) {
	t.Parallel()

	ce := CategorizedError{
		Category: CategoryNetwork,
		Original: errors.New("original"),
		Message:  "Network error",
		Hint:     "Check network",
	}

	if ce.Error() != "Network error" {
		t.Errorf("Error() = %q, want %q", ce.Error(), "Network error")
	}
}

func TestCategorizedError_Unwrap(t *testing.T) {
	t.Parallel()

	original := errors.New("original error")
	ce := CategorizedError{
		Category: CategoryNetwork,
		Original: original,
		Message:  "Network error",
	}

	if !errors.Is(ce.Unwrap(), original) {
		t.Error("Unwrap() should return original error")
	}

	// Test that errors.Is works through Unwrap
	if !errors.Is(ce, original) {
		t.Error("errors.Is should work through Unwrap")
	}
}

func TestFormatError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		wantMessage string
		wantHint    bool
	}{
		{
			name:        "timeout error",
			err:         context.DeadlineExceeded,
			wantMessage: "Request timed out",
			wantHint:    true,
		},
		{
			name:        "network error",
			err:         errors.New("connection refused"),
			wantMessage: "Network error",
			wantHint:    true,
		},
		{
			name:        "auth error",
			err:         errors.New("unauthorized"),
			wantMessage: "Authentication failed",
			wantHint:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			message, hint := FormatError(tt.err)

			if message != tt.wantMessage {
				t.Errorf("FormatError().message = %q, want %q", message, tt.wantMessage)
			}
			if tt.wantHint && hint == "" {
				t.Error("FormatError().hint should not be empty")
			}
		})
	}
}

// mockNetError implements net.Error interface for testing.
type mockNetError struct {
	msg       string
	timeout   bool
	temporary bool
}

func (e mockNetError) Error() string   { return e.msg }
func (e mockNetError) Timeout() bool   { return e.timeout }
func (e mockNetError) Temporary() bool { return e.temporary }

func TestCategorize_NetError(t *testing.T) {
	t.Parallel()

	// Ensure mockNetError implements net.Error
	var _ net.Error = mockNetError{}

	err := mockNetError{msg: "network failure", timeout: false, temporary: false}
	result := Categorize(err)

	if result.Category != CategoryNetwork {
		t.Errorf("Categorize(net.Error).Category = %d, want %d (Network)",
			result.Category, CategoryNetwork)
	}
}

func TestCategory_Values(t *testing.T) {
	t.Parallel()

	// Verify category constants are distinct
	categories := []Category{
		CategoryUnknown,
		CategoryNetwork,
		CategoryTimeout,
		CategoryAuth,
		CategoryDevice,
	}

	seen := make(map[Category]bool)
	for _, c := range categories {
		if seen[c] {
			t.Errorf("Category %d is duplicated", c)
		}
		seen[c] = true
	}
}
