package cmdutil

import (
	"errors"
	"strings"
	"testing"
)

func TestEnhanceDeviceError_NilError(t *testing.T) {
	t.Parallel()
	result := EnhanceDeviceError(nil, "test-device")
	if result != nil {
		t.Errorf("EnhanceDeviceError(nil) = %v, want nil", result)
	}
}

func TestEnhanceDeviceError_DNSError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		err     error
		device  string
		wantMsg string
	}{
		{
			name:    "no such host",
			err:     errors.New("dial tcp: lookup test-device: no such host"),
			device:  "test-device",
			wantMsg: "not found (DNS lookup failed)",
		},
		{
			name:    "server misbehaving",
			err:     errors.New("dial tcp: lookup test-device on 127.0.0.53:53: server misbehaving"),
			device:  "test-device",
			wantMsg: "not found (DNS lookup failed)",
		},
		{
			name:    "lookup with dial tcp",
			err:     errors.New("dial tcp: lookup baddevice: some error"),
			device:  "baddevice",
			wantMsg: "not found (DNS lookup failed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := EnhanceDeviceError(tt.err, tt.device)
			if result == nil {
				t.Fatal("EnhanceDeviceError returned nil for DNS error")
			}
			if !strings.Contains(result.Error(), tt.wantMsg) {
				t.Errorf("error message %q does not contain %q", result.Error(), tt.wantMsg)
			}
		})
	}
}

func TestEnhanceDeviceError_NonDNSError(t *testing.T) {
	t.Parallel()
	originalErr := errors.New("connection refused")
	result := EnhanceDeviceError(originalErr, "test-device")
	// For non-DNS errors, we return the original error unchanged
	if result.Error() != originalErr.Error() {
		t.Errorf("EnhanceDeviceError returned %v, want original error %v", result, originalErr)
	}
}

func TestIsSimilar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a    string
		b    string
		want bool
	}{
		{"master-bathroom", "master-bathoom", true},
		{"master-bathroom", "master-bedroom", true},
		{"kitchen", "kitche", true},
		{"kitchen", "bedroom", false},
		{"abc", "xyz", false},
		{"living-room", "living", true},
		{"test", "testing", true},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			t.Parallel()
			got := isSimilar(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("isSimilar(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
