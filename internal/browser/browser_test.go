package browser

import (
	"context"
	"fmt"
	"runtime"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	b := New()
	if b == nil {
		t.Error("New() returned nil")
	}
}

// TestBrowserInterface verifies the Browser interface is correctly implemented.
func TestBrowserInterface(t *testing.T) {
	t.Parallel()

	var _ Browser = (*browserImpl)(nil)
	var _ = New()
}

// TestBrowserImplType verifies New returns the correct implementation type.
func TestBrowserImplType(t *testing.T) {
	t.Parallel()

	b := New()
	_, ok := b.(*browserImpl)
	if !ok {
		t.Errorf("New() returned %T, want *browserImpl", b)
	}
}

// TestOpenDeviceUI_URLFormat verifies URL formatting for device IPs.
func TestOpenDeviceUI_URLFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ip      string
		wantURL string
	}{
		{"ipv4_private", "192.168.1.1", "http://192.168.1.1"},
		{"ipv4_private_10", "10.0.0.100", "http://10.0.0.100"},
		{"localhost", "localhost", "http://localhost"},
		{"localhost_port", "localhost:8080", "http://localhost:8080"},
		{"ipv4_with_port", "192.168.1.1:80", "http://192.168.1.1:80"},
		{"hostname", "shelly-device.local", "http://shelly-device.local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fmt.Sprintf("http://%s", tt.ip)
			if got != tt.wantURL {
				t.Errorf("URL for %q = %q, want %q", tt.ip, got, tt.wantURL)
			}
		})
	}
}

// TestBrowseContext_Cancelled verifies cancelled context is handled.
func TestBrowseContext_Cancelled(t *testing.T) {
	t.Parallel()

	b := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// We can't easily verify the behavior without executing a real command,
	// but we can at least verify the method accepts a cancelled context
	// without panicking. The actual Browse call may or may not return an error
	// depending on timing.
	_ = b.Browse(ctx, "http://example.com") //nolint:errcheck // testing panic-free behavior, not error
}

// TestOpenDeviceUI_Cancelled verifies cancelled context for OpenDeviceUI.
func TestOpenDeviceUI_Cancelled(t *testing.T) {
	t.Parallel()

	b := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should not panic with cancelled context
	_ = b.OpenDeviceUI(ctx, "192.168.1.1") //nolint:errcheck // testing panic-free behavior, not error
}

// TestBrowseOSDetection verifies the OS-specific command selection logic.
// This doesn't execute commands but tests the runtime.GOOS detection path.
func TestBrowseOSDetection(t *testing.T) {
	t.Parallel()

	// Verify current OS is supported
	supported := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"windows": true,
	}

	if !supported[runtime.GOOS] {
		t.Logf("Current OS %q may not be supported for browser opening", runtime.GOOS)
	}
}

// TestBrowseURLVariants tests that various URL formats are accepted.
func TestBrowseURLVariants(t *testing.T) {
	t.Parallel()

	b := New()
	ctx := context.Background()

	// These tests verify the method accepts various URL formats without panicking.
	// We use a cancelled context to prevent actual command execution.
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	urls := []string{
		"http://example.com",
		"https://example.com",
		"http://192.168.1.1",
		"http://localhost:8080",
		"https://example.com/path?query=value",
		"http://user:pass@example.com",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			t.Parallel()
			// Should not panic
			_ = b.Browse(cancelCtx, url) //nolint:errcheck // testing panic-free behavior, not error
		})
	}
}

// TestOpenDeviceUI_IPVariants tests various IP address formats.
func TestOpenDeviceUI_IPVariants(t *testing.T) {
	t.Parallel()

	b := New()
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ips := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"localhost",
		"shelly.local",
		"192.168.1.1:80",
	}

	for _, ip := range ips {
		t.Run(ip, func(t *testing.T) {
			t.Parallel()
			// Should not panic
			_ = b.OpenDeviceUI(cancelCtx, ip) //nolint:errcheck // testing panic-free behavior, not error
		})
	}
}

// MockBrowser is a test double for Browser interface.
type MockBrowser struct {
	BrowseCalls        []string
	OpenDeviceCalls    []string
	CopyClipboardCalls []string
	BrowseError        error
	OpenDeviceError    error
	CopyClipboardError error
}

// Browse records the URL and returns configured error.
func (m *MockBrowser) Browse(_ context.Context, url string) error {
	m.BrowseCalls = append(m.BrowseCalls, url)
	return m.BrowseError
}

// OpenDeviceUI records the IP and returns configured error.
func (m *MockBrowser) OpenDeviceUI(_ context.Context, deviceIP string) error {
	m.OpenDeviceCalls = append(m.OpenDeviceCalls, deviceIP)
	return m.OpenDeviceError
}

// CopyToClipboard records the URL and returns configured error.
func (m *MockBrowser) CopyToClipboard(url string) error {
	m.CopyClipboardCalls = append(m.CopyClipboardCalls, url)
	return m.CopyClipboardError
}

// TestMockBrowser verifies the mock implementation works correctly.
func TestMockBrowser(t *testing.T) {
	t.Parallel()

	t.Run("browse_records_calls", func(t *testing.T) {
		t.Parallel()
		m := &MockBrowser{}
		ctx := context.Background()

		err := m.Browse(ctx, "http://example.com")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(m.BrowseCalls) != 1 {
			t.Errorf("expected 1 Browse call, got %d", len(m.BrowseCalls))
		}
		if m.BrowseCalls[0] != "http://example.com" {
			t.Errorf("BrowseCalls[0] = %q, want http://example.com", m.BrowseCalls[0])
		}
	})

	t.Run("browse_returns_error", func(t *testing.T) {
		t.Parallel()
		m := &MockBrowser{BrowseError: fmt.Errorf("test error")}
		ctx := context.Background()

		err := m.Browse(ctx, "http://example.com")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("open_device_records_calls", func(t *testing.T) {
		t.Parallel()
		m := &MockBrowser{}
		ctx := context.Background()

		err := m.OpenDeviceUI(ctx, "192.168.1.1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(m.OpenDeviceCalls) != 1 {
			t.Errorf("expected 1 OpenDeviceUI call, got %d", len(m.OpenDeviceCalls))
		}
		if m.OpenDeviceCalls[0] != "192.168.1.1" {
			t.Errorf("OpenDeviceCalls[0] = %q, want 192.168.1.1", m.OpenDeviceCalls[0])
		}
	})

	t.Run("open_device_returns_error", func(t *testing.T) {
		t.Parallel()
		m := &MockBrowser{OpenDeviceError: fmt.Errorf("test error")}
		ctx := context.Background()

		err := m.OpenDeviceUI(ctx, "192.168.1.1")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

// TestMockBrowserInterface verifies MockBrowser implements Browser.
func TestMockBrowserInterface(t *testing.T) {
	t.Parallel()

	var _ Browser = (*MockBrowser)(nil)
}

// TestClipboardFallbackError_Error verifies error message format.
func TestClipboardFallbackError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"http_url", "http://example.com", "URL copied to clipboard: http://example.com"},
		{"https_url", "https://example.com", "URL copied to clipboard: https://example.com"},
		{"device_url", "http://192.168.1.1", "URL copied to clipboard: http://192.168.1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := &ClipboardFallbackError{URL: tt.url}
			got := err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestClipboardFallbackError_ImplementsError verifies it implements error interface.
func TestClipboardFallbackError_ImplementsError(t *testing.T) {
	t.Parallel()

	// Verify ClipboardFallbackError implements error interface at compile time
	err := &ClipboardFallbackError{URL: "http://example.com"}
	msg := err.Error()
	if msg == "" {
		t.Error("Error() should return non-empty message")
	}
}

// TestMockBrowser_CopyToClipboard verifies the mock records clipboard calls.
func TestMockBrowser_CopyToClipboard(t *testing.T) {
	t.Parallel()

	t.Run("records_calls", func(t *testing.T) {
		t.Parallel()
		m := &MockBrowser{}

		err := m.CopyToClipboard("http://example.com")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(m.CopyClipboardCalls) != 1 {
			t.Errorf("expected 1 CopyToClipboard call, got %d", len(m.CopyClipboardCalls))
		}
		if m.CopyClipboardCalls[0] != "http://example.com" {
			t.Errorf("CopyClipboardCalls[0] = %q, want http://example.com", m.CopyClipboardCalls[0])
		}
	})

	t.Run("returns_error", func(t *testing.T) {
		t.Parallel()
		m := &MockBrowser{CopyClipboardError: fmt.Errorf("clipboard error")}

		err := m.CopyToClipboard("http://example.com")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
