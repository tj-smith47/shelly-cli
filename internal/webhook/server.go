// Package webhook provides webhook server functionality.
package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Entry represents a single received webhook.
type Entry struct {
	Timestamp  time.Time         `json:"timestamp"`
	RemoteAddr string            `json:"remote_addr"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Query      map[string]string `json:"query"`
	Body       string            `json:"body,omitempty"`
}

// Server handles incoming webhooks.
type Server struct {
	ios     *iostreams.IOStreams
	logJSON bool
	mu      sync.Mutex
	count   int
}

// NewServer creates a new webhook server.
func NewServer(ios *iostreams.IOStreams, logJSON bool) *Server {
	return &Server{
		ios:     ios,
		logJSON: logJSON,
	}
}

// Handler returns an HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWebhook)
	mux.HandleFunc("/webhook", s.handleWebhook)
	mux.HandleFunc("/health", s.handleHealth)
	return mux
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()

	// Build webhook entry
	entry := Entry{
		Timestamp:  time.Now(),
		RemoteAddr: r.RemoteAddr,
		Method:     r.Method,
		Path:       r.URL.Path,
		Headers:    make(map[string]string),
		Query:      make(map[string]string),
	}

	// Capture headers
	for key := range r.Header {
		entry.Headers[key] = r.Header.Get(key)
	}

	// Capture query params
	for key := range r.URL.Query() {
		entry.Query[key] = r.URL.Query().Get(key)
	}

	// Capture body
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err == nil && len(body) > 0 {
			entry.Body = string(body)
		}
	}

	// Log the webhook
	if s.logJSON {
		s.logWebhookJSON(entry)
	} else {
		s.logWebhookPretty(count, entry)
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		s.ios.DebugErr("write response", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"healthy","webhooks_received":` + fmt.Sprint(s.count) + `}`)); err != nil {
		s.ios.DebugErr("write health response", err)
	}
}

func (s *Server) logWebhookJSON(entry Entry) {
	data, err := json.Marshal(entry)
	if err != nil {
		s.ios.DebugErr("marshal webhook", err)
		return
	}
	s.ios.Println(string(data))
}

func (s *Server) logWebhookPretty(count int, entry Entry) {
	s.ios.Success("[#%d] Webhook Received", count)
	s.ios.Printf("  Time:   %s\n", entry.Timestamp.Format(time.RFC3339))
	s.ios.Printf("  From:   %s\n", entry.RemoteAddr)
	s.ios.Printf("  Method: %s\n", entry.Method)
	s.ios.Printf("  Path:   %s\n", entry.Path)

	if len(entry.Query) > 0 {
		s.ios.Printf("  Query:\n")
		for k, v := range entry.Query {
			s.ios.Printf("    %s: %s\n", k, v)
		}
	}

	if entry.Body != "" {
		s.ios.Printf("  Body:\n")
		s.ios.Printf("    %s\n", FormatBody(entry.Body))
	}

	s.ios.Println("")
}

// FormatBody attempts to pretty-print JSON, falls back to raw body.
func FormatBody(body string) string {
	var jsonBody map[string]any
	if err := json.Unmarshal([]byte(body), &jsonBody); err != nil {
		return body
	}
	prettyBody, err := json.MarshalIndent(jsonBody, "    ", "  ")
	if err != nil {
		return body
	}
	return string(prettyBody)
}

// GetLocalIP returns the local IP address for display.
func GetLocalIP() string {
	// Try to get a non-loopback IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return "localhost"
}

// ConfigureDevices configures devices to send webhooks to the server.
func ConfigureDevices(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, devices []string, serverURL string) {
	webhookURL := serverURL + "/webhook"

	for _, device := range devices {
		ios.Printf("  Configuring %s... ", device)

		conn, err := svc.Connect(ctx, device)
		if err != nil {
			ios.Printf("failed (connect: %v)\n", err)
			continue
		}

		// Create a webhook for all events
		params := map[string]any{
			"cid":    0,
			"enable": true,
			"event":  "*",
			"urls":   []string{webhookURL},
		}

		_, err = conn.Call(ctx, "Webhook.Create", params)
		iostreams.CloseWithDebug("closing auto-config connection", conn)

		if err != nil {
			ios.Printf("failed (%v)\n", err)
			continue
		}

		ios.Printf("done\n")
	}
}
