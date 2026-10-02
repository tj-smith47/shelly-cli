package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
)

// Gen2NameDevice is a loopback Gen2 device reporting a fixed MAC. It answers
// every RPC with an empty result and records every Set method called and the
// names written through Sys.SetConfig.
type Gen2NameDevice struct {
	// Addr is the device's host:port.
	Addr  string
	mu    sync.Mutex
	names []string
	sets  []string
}

// NewGen2NameDevice starts a Gen2NameDevice reporting mac, closed when t ends.
func NewGen2NameDevice(t *testing.T, mac string) *Gen2NameDevice {
	t.Helper()
	d := &Gen2NameDevice{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
			Params struct {
				Config struct {
					Device struct {
						Name *string `json:"name"`
					} `json:"device"`
				} `json:"config"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode rpc: %v", err)
			return
		}
		if strings.Contains(req.Method, ".Set") {
			d.mu.Lock()
			d.sets = append(d.sets, req.Method)
			if n := req.Params.Config.Device.Name; req.Method == "Sys.SetConfig" && n != nil {
				d.names = append(d.names, *n)
			}
			d.mu.Unlock()
		}
		result := map[string]any{}
		if req.Method == "Shelly.GetDeviceInfo" {
			result = map[string]any{"id": "shellyplus1-test", "mac": mac, "gen": 2, "model": "SNSW-001X16EU"}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"id": req.ID, "jsonrpc": "2.0", "result": result}); err != nil {
			t.Errorf("encode rpc: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	d.Addr = strings.TrimPrefix(srv.URL, "http://")
	return d
}

// Written returns the names written through Sys.SetConfig, in order.
func (d *Gen2NameDevice) Written() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return slices.Clone(d.names)
}

// SetCalls returns every RPC method containing ".Set" that was called, in order.
func (d *Gen2NameDevice) SetCalls() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return slices.Clone(d.sets)
}
