package connection

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestIsConnectionError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"connection refused", errors.New("connection refused"), true},
		{"no route to host", errors.New("no route to host"), true},
		{"i/o timeout", errors.New("i/o timeout"), true},
		{"network is unreachable", errors.New("network is unreachable"), true},
		{"no such host", errors.New("no such host"), true},
		{"dial tcp", errors.New("dial tcp 192.168.1.1:80"), true},
		{"ErrConnectionFailed", model.ErrConnectionFailed, true},
		{"wrapped connection failed", fmt.Errorf("failed: %w", model.ErrConnectionFailed), true},
		{"auth error", errors.New("authentication failed"), false},
		{"generic error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isConnectionError(tt.err)
			if got != tt.want {
				t.Errorf("isConnectionError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsConnectionError_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"empty error message", errors.New(""), false},
		{"partial match not enough", errors.New("refused but not connection"), false},
		{"case sensitive connection", errors.New("Connection Refused"), false},
		{"multiple errors combined", errors.New("dial tcp: connection refused"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isConnectionError(tt.err)
			if got != tt.want {
				t.Errorf("isConnectionError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
