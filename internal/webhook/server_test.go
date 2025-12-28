package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

//nolint:gocritic,unparam // helper function returns multiple values for API consistency
func testIOStreams() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return iostreams.Test(in, out, errOut), out, errOut
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	server := NewServer(ios, false)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if server.ios != ios {
		t.Error("ios not set correctly")
	}
	if server.logJSON {
		t.Error("logJSON should be false")
	}
}

func TestNewServer_JSONMode(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	server := NewServer(ios, true)

	if !server.logJSON {
		t.Error("logJSON should be true")
	}
}

func TestServer_Handler(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	server := NewServer(ios, false)
	handler := server.Handler()

	if handler == nil {
		t.Fatal("Handler returned nil")
	}
}

func TestServer_HandleHealth(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	server := NewServer(ios, false)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, `"status":"healthy"`) {
		t.Errorf("body = %q, should contain status:healthy", body)
	}
	if !strings.Contains(body, `"webhooks_received"`) {
		t.Errorf("body = %q, should contain webhooks_received", body)
	}
}

func TestServer_HandleWebhook(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	server := NewServer(ios, false)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodPost, "/webhook?event=switch.on", strings.NewReader(`{"switch":{"id":0,"output":true}}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if body != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", body, `{"status":"ok"}`)
	}
}

func TestServer_HandleWebhook_JSON(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	server := NewServer(ios, true)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{"event":"button.push"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	// In JSON mode, output should contain the webhook entry
	if !strings.Contains(out.String(), `"method":"POST"`) {
		t.Errorf("output should contain JSON entry, got: %s", out.String())
	}
}

func TestServer_HandleWebhook_RootPath(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	server := NewServer(ios, false)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestServer_CountIncrement(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	server := NewServer(ios, true)
	handler := server.Handler()

	// Send three webhooks
	for range 3 {
		req := httptest.NewRequest(http.MethodPost, "/webhook", http.NoBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Check health to see count
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, `"webhooks_received":3`) {
		t.Errorf("health body = %q, should show 3 webhooks", body)
	}
}

func TestFormatBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantPretty bool
	}{
		{"valid JSON", `{"key":"value"}`, true},
		{"nested JSON", `{"outer":{"inner":true}}`, true},
		{"not JSON", "plain text", false},
		{"empty", "", false},
		{"invalid JSON", `{"key":}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := FormatBody(tt.input)

			if tt.wantPretty {
				// Pretty-printed JSON should contain newlines
				if !strings.Contains(result, "\n") && len(tt.input) > 10 {
					t.Errorf("FormatBody(%q) should contain newlines for pretty printing", tt.input)
				}
			} else {
				// Non-JSON or invalid should be returned as-is
				if result != tt.input {
					t.Errorf("FormatBody(%q) = %q, want %q", tt.input, result, tt.input)
				}
			}
		})
	}
}

func TestGetLocalIP(t *testing.T) {
	t.Parallel()

	ip := GetLocalIP()

	// Should return something (either an IP or "localhost")
	if ip == "" {
		t.Error("GetLocalIP() returned empty string")
	}

	// Should be either "localhost" or look like an IP address
	if ip != "localhost" && !strings.Contains(ip, ".") { //nolint:goconst // test value
		t.Errorf("GetLocalIP() = %q, doesn't look like IP or localhost", ip)
	}
}

func TestEntry(t *testing.T) {
	t.Parallel()

	entry := Entry{
		RemoteAddr: "192.168.1.1:12345",
		Method:     "POST",
		Path:       "/webhook",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Query:      map[string]string{"event": "switch.on"},
		Body:       `{"id":0}`,
	}

	if entry.RemoteAddr != "192.168.1.1:12345" {
		t.Errorf("RemoteAddr = %q", entry.RemoteAddr)
	}
	if entry.Method != http.MethodPost {
		t.Errorf("Method = %q", entry.Method)
	}
	if entry.Path != "/webhook" {
		t.Errorf("Path = %q", entry.Path)
	}
	if entry.Headers["Content-Type"] != "application/json" {
		t.Errorf("Headers[Content-Type] = %q", entry.Headers["Content-Type"])
	}
	if entry.Query["event"] != "switch.on" {
		t.Errorf("Query[event] = %q", entry.Query["event"])
	}
	if entry.Body != `{"id":0}` {
		t.Errorf("Body = %q", entry.Body)
	}
}

func TestServer_logWebhookPretty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()

	server := NewServer(ios, false)
	entry := Entry{
		RemoteAddr: "127.0.0.1:8080",
		Method:     "POST",
		Path:       "/test",
		Query:      map[string]string{"key": "value"},
		Body:       `{"data":"test"}`,
	}

	server.logWebhookPretty(1, entry)

	// Check output contains expected elements
	output := out.String()
	if !strings.Contains(output, "[#1]") {
		t.Error("output should contain webhook number")
	}
}
