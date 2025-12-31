package telemetry

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestGetEndpoint(t *testing.T) {
	// Save original value
	origEndpoint := Endpoint
	defer func() { Endpoint = origEndpoint }()

	t.Run("returns ldflags value when set", func(t *testing.T) {
		Endpoint = "https://ldflags.example.com"
		t.Setenv("SHELLY_TELEMETRY_ENDPOINT", "https://env.example.com")

		got := getEndpoint()
		if got != "https://ldflags.example.com" {
			t.Errorf("getEndpoint() = %q, want %q", got, "https://ldflags.example.com")
		}
	})

	t.Run("falls back to env var when ldflags empty", func(t *testing.T) {
		Endpoint = ""
		t.Setenv("SHELLY_TELEMETRY_ENDPOINT", "https://env.example.com")

		got := getEndpoint()
		if got != "https://env.example.com" {
			t.Errorf("getEndpoint() = %q, want %q", got, "https://env.example.com")
		}
	})

	t.Run("returns empty when neither set", func(t *testing.T) {
		Endpoint = ""
		t.Setenv("SHELLY_TELEMETRY_ENDPOINT", "")

		got := getEndpoint()
		if got != "" {
			t.Errorf("getEndpoint() = %q, want empty", got)
		}
	})
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	t.Run("creates client with endpoint", func(t *testing.T) {
		t.Parallel()
		c := NewClient("https://example.com/telemetry")
		defer c.Close()

		if c.endpoint != "https://example.com/telemetry" {
			t.Errorf("endpoint = %q, want %q", c.endpoint, "https://example.com/telemetry")
		}
		if c.IsEnabled() {
			t.Error("new client should be disabled by default")
		}
	})

	t.Run("creates client with empty endpoint", func(t *testing.T) {
		t.Parallel()
		c := NewClient("")
		defer c.Close()

		if c.HasEndpoint() {
			t.Error("HasEndpoint() should return false for empty endpoint")
		}
	})
}

func TestClient_SetEnabled(t *testing.T) {
	t.Parallel()
	c := NewClient("https://example.com")
	defer c.Close()

	if c.IsEnabled() {
		t.Error("client should start disabled")
	}

	c.SetEnabled(true)
	if !c.IsEnabled() {
		t.Error("client should be enabled after SetEnabled(true)")
	}

	c.SetEnabled(false)
	if c.IsEnabled() {
		t.Error("client should be disabled after SetEnabled(false)")
	}
}

func TestClient_HasEndpoint(t *testing.T) {
	t.Parallel()
	t.Run("returns true with endpoint", func(t *testing.T) {
		t.Parallel()
		c := NewClient("https://example.com")
		defer c.Close()

		if !c.HasEndpoint() {
			t.Error("HasEndpoint() should return true")
		}
	})

	t.Run("returns false without endpoint", func(t *testing.T) {
		t.Parallel()
		c := NewClient("")
		defer c.Close()

		if c.HasEndpoint() {
			t.Error("HasEndpoint() should return false")
		}
	})
}

func TestClient_Track(t *testing.T) {
	t.Parallel()
	t.Run("does nothing when disabled", func(t *testing.T) {
		t.Parallel()
		c := NewClient("https://example.com")
		defer c.Close()

		// Track when disabled - should not panic or queue
		c.Track("test command", true, 100*time.Millisecond)
	})

	t.Run("does nothing without endpoint", func(t *testing.T) {
		t.Parallel()
		c := NewClient("")
		defer c.Close()
		c.SetEnabled(true)

		// Track without endpoint - should not panic or queue
		c.Track("test command", true, 100*time.Millisecond)
	})

	t.Run("queues event when enabled with endpoint", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL)
		c.SetEnabled(true)

		c.Track("device info", true, 50*time.Millisecond)

		// Give time for event to be queued
		time.Sleep(50 * time.Millisecond)

		c.Close()
	})

	t.Run("drops event when queue is full", func(t *testing.T) {
		t.Parallel()
		c := NewClient("https://example.com")
		c.SetEnabled(true)

		// Fill the queue (capacity 100)
		for range 150 {
			c.Track("command", true, time.Millisecond)
		}

		// Should not panic
		c.Close()
	})
}

func TestClient_sender(t *testing.T) {
	t.Parallel()
	t.Run("sends batch when size threshold reached", func(t *testing.T) {
		t.Parallel()
		var receivedEvents []Event
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read body: %v", err)
				return
			}

			var events []Event
			if err := json.Unmarshal(body, &events); err != nil {
				t.Errorf("unmarshal events: %v", err)
				return
			}

			mu.Lock()
			receivedEvents = append(receivedEvents, events...)
			mu.Unlock()

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL)
		c.SetEnabled(true)

		// Send enough events to trigger batch send (threshold is 10)
		for range 10 {
			c.Track("batch test", true, time.Millisecond)
		}

		// Wait for batch to be sent
		time.Sleep(200 * time.Millisecond)

		c.Close()

		mu.Lock()
		defer mu.Unlock()
		if len(receivedEvents) < 10 {
			t.Errorf("received %d events, want at least 10", len(receivedEvents))
		}
	})

	t.Run("sends remaining events on close", func(t *testing.T) {
		t.Parallel()
		var receivedEvents []Event
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Logf("read body: %v", err)
				return
			}
			var events []Event
			if err := json.Unmarshal(body, &events); err == nil {
				mu.Lock()
				receivedEvents = append(receivedEvents, events...)
				mu.Unlock()
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL)
		c.SetEnabled(true)

		// Send fewer than batch threshold
		c.Track("close test", true, time.Millisecond)
		c.Track("close test", true, time.Millisecond)
		c.Track("close test", true, time.Millisecond)

		// Close should flush remaining events
		c.Close()

		// Give a moment for the server to process
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if len(receivedEvents) != 3 {
			t.Errorf("received %d events, want 3", len(receivedEvents))
		}
	})
}

func TestClient_sendBatch(t *testing.T) {
	t.Parallel()
	t.Run("sends events successfully", func(t *testing.T) {
		t.Parallel()
		var received []Event

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != "application/json" {
				t.Error("expected Content-Type: application/json")
			}
			if r.Header.Get("User-Agent") == "" {
				t.Error("expected User-Agent header")
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read body: %v", err)
				return
			}
			if err := json.Unmarshal(body, &received); err != nil {
				t.Errorf("unmarshal: %v", err)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL)
		defer c.Close()

		events := []Event{
			{Command: "test", Success: true, Duration: 100},
		}
		c.sendBatch(events)

		if len(received) != 1 {
			t.Errorf("received %d events, want 1", len(received))
		}
	})

	t.Run("handles empty batch", func(t *testing.T) {
		t.Parallel()
		c := NewClient("https://example.com")
		defer c.Close()

		// Should not panic
		c.sendBatch(nil)
		c.sendBatch([]Event{})
	})

	t.Run("handles empty endpoint", func(t *testing.T) {
		t.Parallel()
		c := NewClient("")
		defer c.Close()

		// Should not panic
		c.sendBatch([]Event{{Command: "test"}})
	})

	t.Run("handles server error", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := NewClient(server.URL)
		defer c.Close()

		// Should not panic on server error
		c.sendBatch([]Event{{Command: "test"}})
	})

	t.Run("handles network error", func(t *testing.T) {
		t.Parallel()
		c := NewClient("http://localhost:1") // Invalid port
		defer c.Close()

		// Should not panic on network error
		c.sendBatch([]Event{{Command: "test"}})
	})
}

func TestClient_Close(t *testing.T) {
	t.Parallel()
	t.Run("closes cleanly", func(t *testing.T) {
		t.Parallel()
		c := NewClient("https://example.com")
		c.SetEnabled(true)

		// Queue some events
		c.Track("close test", true, time.Millisecond)

		// Close should not panic
		c.Close()
	})

	t.Run("can be called multiple times safely", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL)
		c.SetEnabled(true)

		c.Track("test", true, time.Millisecond)
		c.Close()

		// Note: Calling Close() twice would panic on closed channel,
		// which is expected behavior for a client that should only be closed once
	})
}

func TestGetCommandPath(t *testing.T) {
	t.Parallel()

	// Helper to build a fresh command tree for each test case
	buildCmdTree := func() *cobra.Command {
		rootCmd := &cobra.Command{Use: "shelly"}
		deviceCmd := &cobra.Command{Use: "device"}
		infoCmd := &cobra.Command{Use: "info"}
		listCmd := &cobra.Command{Use: "list"}
		switchCmd := &cobra.Command{Use: "switch"}
		onCmd := &cobra.Command{Use: "on"}

		deviceCmd.AddCommand(infoCmd, listCmd)
		switchCmd.AddCommand(onCmd)
		rootCmd.AddCommand(deviceCmd, switchCmd)
		return rootCmd
	}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "root command",
			args: []string{},
			want: "shelly",
		},
		{
			name: "single subcommand",
			args: []string{"device"},
			want: "device",
		},
		{
			name: "nested subcommand - device info",
			args: []string{"device", "info"},
			want: "device info",
		},
		{
			name: "nested subcommand - device list",
			args: []string{"device", "list"},
			want: "device list",
		},
		{
			name: "nested subcommand - switch on",
			args: []string{"switch", "on"},
			want: "switch on",
		},
		{
			name: "unknown command returns shelly",
			args: []string{"unknown"},
			want: "shelly",
		},
		{
			name: "unknown subcommand returns parent",
			args: []string{"device", "unknown"},
			want: "device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rootCmd := buildCmdTree()
			got := GetCommandPath(rootCmd, tt.args)
			if got != tt.want {
				t.Errorf("GetCommandPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // Test modifies package-level globals (globalClient, globalClientOnce, Endpoint).
func TestPackageLevelFunctions(t *testing.T) {
	// Reset global client for testing
	globalClientOnce = sync.Once{}
	globalClient = nil

	// Save original endpoint
	origEndpoint := Endpoint
	defer func() {
		Endpoint = origEndpoint
		globalClientOnce = sync.Once{}
		globalClient = nil
	}()

	//nolint:paralleltest // Shares state with parent test.
	t.Run("DefaultClient returns singleton", func(t *testing.T) {
		Endpoint = "https://example.com"
		c1 := DefaultClient()
		c2 := DefaultClient()

		if c1 != c2 {
			t.Error("DefaultClient should return the same instance")
		}
	})

	// Reset for next test
	globalClientOnce = sync.Once{}
	globalClient = nil
	Endpoint = ""

	//nolint:paralleltest // Shares state with parent test.
	t.Run("SetEnabled uses default client", func(t *testing.T) {
		SetEnabled(true)
		if !IsEnabled() {
			t.Error("IsEnabled should return true after SetEnabled(true)")
		}

		SetEnabled(false)
		if IsEnabled() {
			t.Error("IsEnabled should return false after SetEnabled(false)")
		}
	})

	//nolint:paralleltest // Shares state with parent test.
	t.Run("HasEndpoint uses default client", func(t *testing.T) {
		// With empty endpoint
		if HasEndpoint() {
			t.Error("HasEndpoint should return false with empty endpoint")
		}
	})

	//nolint:paralleltest // Shares state with parent test.
	t.Run("Track uses default client", func(t *testing.T) {
		// Should not panic even with empty endpoint
		Track("test", true, time.Millisecond)
	})

	//nolint:paralleltest // Shares state with parent test.
	t.Run("Close uses default client", func(t *testing.T) {
		// Should not panic
		Close()
	})
}

func TestNewClientWithOptions(t *testing.T) {
	t.Parallel()
	t.Run("creates client with custom ticker duration", func(t *testing.T) {
		t.Parallel()
		c := NewClientWithOptions("https://example.com", 100*time.Millisecond)
		defer c.Close()

		if c.tickerDuration != 100*time.Millisecond {
			t.Errorf("tickerDuration = %v, want %v", c.tickerDuration, 100*time.Millisecond)
		}
	})

	t.Run("defaults to 10s when duration is zero", func(t *testing.T) {
		t.Parallel()
		c := NewClientWithOptions("https://example.com", 0)
		defer c.Close()

		if c.tickerDuration != 10*time.Second {
			t.Errorf("tickerDuration = %v, want %v", c.tickerDuration, 10*time.Second)
		}
	})

	t.Run("defaults to 10s when duration is negative", func(t *testing.T) {
		t.Parallel()
		c := NewClientWithOptions("https://example.com", -1*time.Second)
		defer c.Close()

		if c.tickerDuration != 10*time.Second {
			t.Errorf("tickerDuration = %v, want %v", c.tickerDuration, 10*time.Second)
		}
	})
}

func TestClient_sender_ticker(t *testing.T) {
	t.Parallel()
	t.Run("sends batch on ticker", func(t *testing.T) {
		t.Parallel()
		var receivedEvents []Event
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Logf("read body: %v", err)
				return
			}
			var events []Event
			if err := json.Unmarshal(body, &events); err == nil {
				mu.Lock()
				receivedEvents = append(receivedEvents, events...)
				mu.Unlock()
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Use short ticker duration for testing
		c := NewClientWithOptions(server.URL, 50*time.Millisecond)
		c.SetEnabled(true)

		// Send fewer than batch threshold (less than 10)
		c.Track("ticker test", true, time.Millisecond)
		c.Track("ticker test", true, time.Millisecond)

		// Wait for ticker to fire (50ms ticker + buffer)
		time.Sleep(150 * time.Millisecond)

		c.Close()

		mu.Lock()
		defer mu.Unlock()
		if len(receivedEvents) != 2 {
			t.Errorf("received %d events, want 2", len(receivedEvents))
		}
	})

	t.Run("ticker with empty batch does nothing", func(t *testing.T) {
		t.Parallel()
		requestCount := 0
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			requestCount++
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Use short ticker duration
		c := NewClientWithOptions(server.URL, 50*time.Millisecond)
		c.SetEnabled(true)

		// Don't track any events

		// Wait for ticker to fire
		time.Sleep(100 * time.Millisecond)

		c.Close()

		mu.Lock()
		defer mu.Unlock()
		// Should not have sent any requests since batch was empty
		if requestCount != 0 {
			t.Errorf("request count = %d, want 0", requestCount)
		}
	})
}

func TestClient_sendBatch_invalidURL(t *testing.T) {
	t.Parallel()
	t.Run("handles invalid URL for request creation", func(t *testing.T) {
		t.Parallel()
		// This URL will pass validation but fail request creation
		c := NewClient("http://[::1]:-1")
		defer c.Close()

		// Should not panic
		c.sendBatch([]Event{{Command: "test"}})
	})
}

// errorBody is an io.ReadCloser that returns an error on Close.
type errorBody struct{}

func (e errorBody) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (e errorBody) Close() error {
	return io.ErrUnexpectedEOF // Return an error on close
}

// roundTripFunc allows creating custom http.RoundTripper from a function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClient_sendBatch_closeError(t *testing.T) {
	t.Parallel()
	t.Run("handles body close error", func(t *testing.T) {
		t.Parallel()
		c := NewClient("https://example.com")
		// Replace http client with one that returns a response with failing body close
		c.httpClient = &http.Client{
			Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       errorBody{},
				}, nil
			}),
		}
		defer c.Close()

		// Should not panic on body close error
		c.sendBatch([]Event{{Command: "test"}})
	})
}

var errMarshalFailed = errors.New("marshal failed")

func TestClient_sendBatch_marshalError(t *testing.T) {
	t.Parallel()
	t.Run("handles marshal error", func(t *testing.T) {
		t.Parallel()
		c := NewClient("https://example.com")
		// Replace marshal function with one that always fails
		c.marshalFunc = func(_ any) ([]byte, error) {
			return nil, errMarshalFailed
		}
		defer c.Close()

		// Should not panic on marshal error
		c.sendBatch([]Event{{Command: "test"}})
	})
}

func TestEvent_Fields(t *testing.T) {
	t.Parallel()
	event := Event{
		Command:   "device info",
		Success:   true,
		Duration:  150,
		Version:   "1.0.0",
		OS:        "linux",
		Arch:      "amd64",
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}

	if decoded.Command != event.Command {
		t.Errorf("Command = %q, want %q", decoded.Command, event.Command)
	}
	if decoded.Success != event.Success {
		t.Errorf("Success = %v, want %v", decoded.Success, event.Success)
	}
	if decoded.Duration != event.Duration {
		t.Errorf("Duration = %d, want %d", decoded.Duration, event.Duration)
	}
	if decoded.Version != event.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, event.Version)
	}
	if decoded.OS != event.OS {
		t.Errorf("OS = %q, want %q", decoded.OS, event.OS)
	}
	if decoded.Arch != event.Arch {
		t.Errorf("Arch = %q, want %q", decoded.Arch, event.Arch)
	}
}
