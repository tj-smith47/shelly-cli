package errutil

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestIsTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"context deadline exceeded", context.DeadlineExceeded, true},
		{"timeout string", errors.New("connection timeout"), true},
		{"deadline exceeded string", errors.New("deadline exceeded"), true},
		{"timed out string", errors.New("request timed out"), true},
		{"unrelated error", errors.New("connection refused"), false},
		{"wrapped deadline", errors.New("failed: context deadline exceeded"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsTimeout(tt.err); got != tt.want {
				t.Errorf("IsTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDNS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"no such host", errors.New("dial tcp: lookup test-device: no such host"), true},
		{"server misbehaving", errors.New("lookup test on 127.0.0.53:53: server misbehaving"), true},
		{"lookup with dial tcp", errors.New("dial tcp: lookup baddevice: failed"), true},
		{"connection refused", errors.New("connection refused"), false},
		{"generic network error", errors.New("network unreachable"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsDNS(tt.err); got != tt.want {
				t.Errorf("IsDNS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNetwork(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"connection refused", errors.New("connection refused"), true},
		{"no route to host", errors.New("no route to host"), true},
		{"network unreachable", errors.New("network unreachable"), true},
		{"host unreachable", errors.New("host unreachable"), true},
		{"no such host", errors.New("no such host"), true},
		{"dial tcp", errors.New("dial tcp: connection failed"), true},
		{"i/o timeout", errors.New("i/o timeout"), true},
		{"auth error", errors.New("401 unauthorized"), false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsNetwork(tt.err); got != tt.want {
				t.Errorf("IsNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockNetError implements net.Error for testing.
type mockNetError struct {
	timeout   bool
	temporary bool
}

func (e *mockNetError) Error() string   { return "mock net error" }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return e.temporary }

func TestIsNetwork_NetError(t *testing.T) {
	t.Parallel()

	var netErr net.Error = &mockNetError{timeout: true}
	if !IsNetwork(netErr) {
		t.Error("IsNetwork should return true for net.Error")
	}
}

func TestIsAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"401 status", errors.New("HTTP 401"), true},
		{"403 status", errors.New("HTTP 403 forbidden"), true},
		{"unauthorized", errors.New("request unauthorized"), true},
		{"forbidden", errors.New("access forbidden"), true},
		{"authentication failed", errors.New("authentication failed"), true},
		{"invalid password", errors.New("invalid password"), true},
		{"access denied", errors.New("access denied"), true},
		{"network error", errors.New("connection refused"), false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsAuth(tt.err); got != tt.want {
				t.Errorf("IsAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDevice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"device error", errors.New("device error: switch failed"), true},
		{"device in message", errors.New("failed to connect to device"), true},
		{"shelly error", errors.New("shelly returned error"), true},
		{"component error", errors.New("component error: invalid id"), true},
		{"component in message", errors.New("switch component failed"), true},
		{"network error", errors.New("connection refused"), false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsDevice(tt.err); got != tt.want {
				t.Errorf("IsDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnsupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"404 status", errors.New("HTTP 404"), true},
		{"unknown method", errors.New("unknown method: Script.GetCode"), true},
		{"not found", errors.New("resource not found"), true},
		{"not supported", errors.New("feature not supported"), true},
		{"network error", errors.New("connection refused"), false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsUnsupported(tt.err); got != tt.want {
				t.Errorf("IsUnsupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategorize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want Category
	}{
		{"nil error", nil, CategoryUnknown},
		{"timeout", context.DeadlineExceeded, CategoryTimeout},
		{"dns error", errors.New("no such host"), CategoryDNS},
		{"network error", errors.New("connection refused"), CategoryNetwork},
		{"auth error", errors.New("401 unauthorized"), CategoryAuth},
		{"unsupported", errors.New("unknown method"), CategoryUnsupported},
		{"device error", errors.New("device error: failed"), CategoryDevice},
		{"shelly error", errors.New("shelly returned error"), CategoryDevice},
		{"unknown", errors.New("something random"), CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Categorize(tt.err); got != tt.want {
				t.Errorf("Categorize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategory_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cat  Category
		want string
	}{
		{CategoryUnknown, "unknown"},
		{CategoryTimeout, "timeout"},
		{CategoryDNS, "dns"},
		{CategoryNetwork, "network"},
		{CategoryAuth, "auth"},
		{CategoryDevice, "device"},
		{CategoryUnsupported, "unsupported"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.cat.String(); got != tt.want {
				t.Errorf("Category.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
