package shelly

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

// The reboot and factory-reset functions reach a real device only through
// Service.WithConnection / WithGen1Connection, which dial the address the resolver
// hands back. Pointing that resolver at an httptest server
// turns every device round-trip into in-process HTTP, so NO real network or host
// state is ever touched. NONE of these helpers run a WiFi scan, dhcpcd, nmcli, or
// reach discovery.DefaultAPIP — the resolver substitution is the entire safety seam.

// apdevServer is a configurable in-process stand-in for a Shelly device. It serves
// the handful of endpoints the in-scope functions exercise (Gen1 REST and Gen2 RPC)
// and records the paths/methods it was asked for, so a test can both drive a code
// path and assert which device call it made.
type apdevServer struct {
	srv  *httptest.Server
	gen1 *apdevGen1
	gen2 *apdevGen2
	// staSSID is the configured WiFi station SSID both generations report.
	staSSID string
	// sta1SSID, when set, is the configured secondary station SSID.
	sta1SSID string
	// staStatic, when set, is the station's static address; gateway and
	// netmask are fixed test values.
	staStatic string
	// noStation makes the device report no station at all.
	noStation bool
	// wifiReadFail makes every WiFi settings read fail.
	wifiReadFail bool

	wifiReads atomic.Int32
	mu        sync.Mutex
	// writes records each station write: a Gen1 /settings/sta or /settings/sta1
	// query, or a Gen2 WiFi.SetConfig / Shelly.SetConfig params object.
	writes []apdevWrite
	// hits counts requests per URL path.
	hits map[string]int
}

// hitsFor returns how many requests reached path.
func (d *apdevServer) hitsFor(path string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.hits[path]
}

// apdevWrite is one recorded write to the fake device.
type apdevWrite struct {
	method string
	query  url.Values
	params map[string]any
}

func (d *apdevServer) record(w apdevWrite) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.writes = append(d.writes, w)
}

// written returns the recorded writes.
func (d *apdevServer) written() []apdevWrite {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]apdevWrite(nil), d.writes...)
}

// gen1Station is the wifi_sta object the fake Gen1 device reports.
func (d *apdevServer) gen1Station(ssid string) map[string]any {
	sta := map[string]any{"enabled": true, "ssid": ssid, "ipv4_method": "dhcp"}
	if d.staStatic != "" {
		sta["ipv4_method"] = "static"
		sta["ip"] = d.staStatic
		sta["gw"] = testGateway
		sta["mask"] = testNetmask
	}
	return sta
}

// gen2WiFiConfig is the WiFi config the fake Gen2 device reports.
func (d *apdevServer) gen2WiFiConfig() map[string]any {
	cfg := map[string]any{}
	if !d.noStation {
		sta := map[string]any{"ssid": d.staSSID, "enable": true, "ipv4mode": "dhcp", "is_open": false}
		if d.staStatic != "" {
			sta["ipv4mode"] = "static"
			sta["ip"] = d.staStatic
			sta["gw"] = testGateway
			sta["netmask"] = testNetmask
		}
		cfg["sta"] = sta
	}
	if d.sta1SSID != "" {
		cfg["sta1"] = map[string]any{"ssid": d.sta1SSID, "enable": true, "is_open": false}
	}
	return cfg
}

const (
	testGateway = "10.0.0.1"
	testNetmask = "255.255.255.0"
)

// apdevGen1 holds the mutable per-route behaviour of a fake Gen1 device.
type apdevGen1 struct {
	// fw is the build string returned at /shelly and /settings.
	fw string
	// uptime is the value returned at /status.
	uptime int
	// rebootErr / resetErr, when true, make the matching endpoint answer 500 so
	// the production call sees a real (non-connectivity) failure.
	rebootErr bool
	resetErr  bool

	rebootHits int32
	resetHits  int32
}

// apdevGen2 holds the mutable per-route behaviour of a fake Gen2 device.
type apdevGen2 struct {
	// rebootErr makes Shelly.Reboot answer with a JSON-RPC error so the action
	// surfaces a real failure rather than a dropped-connection success signal.
	rebootErr  bool
	rebootHits int32
}

// newAPDevServer starts a fake device of the given generation and registers cleanup.
func newAPDevServer(t *testing.T, generation int) *apdevServer {
	t.Helper()
	d := &apdevServer{
		gen1: &apdevGen1{fw: "20210101-000000/v1.0", uptime: 99},
		gen2: &apdevGen2{},
	}
	mux := http.NewServeMux()
	if generation == 1 {
		d.registerGen1(mux)
	} else {
		d.registerGen2(t, mux)
	}
	d.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.mu.Lock()
		if d.hits == nil {
			d.hits = map[string]int{}
		}
		d.hits[r.URL.Path]++
		d.mu.Unlock()
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(d.srv.Close)
	return d
}

// addr returns the server's host:port, suitable for model.Device.Address.
func (d *apdevServer) addr() string {
	return strings.TrimPrefix(d.srv.URL, "http://")
}

// resolver returns a generation-aware resolver that maps every identifier to this
// fake device, so WithConnection / WithGen1Connection dial it.
func (d *apdevServer) resolver(generation int) DeviceResolver {
	return &testutil.Resolver{Device: model.Device{
		Name:       "apdev",
		Address:    d.addr(),
		Generation: generation,
	}}
}

func (d *apdevServer) registerGen1(mux *http.ServeMux) {
	// /shelly identifies the device for ConnectGen1.
	mux.HandleFunc("/shelly", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"type": "SHBDUO-1", "mac": "AABBCCDDEEFF", "fw": d.gen1.fw, "gen": 1,
		})
	})
	mux.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("reset") == "true" {
			atomic.AddInt32(&d.gen1.resetHits, 1)
			if d.gen1.resetErr {
				http.Error(w, "reset refused", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]any{"ok": true})
			return
		}
		d.wifiReads.Add(1)
		if d.wifiReadFail {
			http.Error(w, "settings unavailable", http.StatusInternalServerError)
			return
		}
		settings := map[string]any{"fw": d.gen1.fw, "device": map[string]any{"type": "SHBDUO-1"}}
		if !d.noStation {
			settings["wifi_sta"] = d.gen1Station(d.staSSID)
		}
		if d.sta1SSID != "" {
			settings["wifi_sta1"] = map[string]any{"enabled": true, "ssid": d.sta1SSID}
		}
		writeJSON(w, settings)
	})
	for _, path := range []string{"/settings/sta", "/settings/sta1"} {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			d.record(apdevWrite{method: path, query: r.URL.Query()})
			writeJSON(w, map[string]any{"ok": true})
		})
	}
	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"uptime": d.gen1.uptime, "unixtime": 1700000000})
	})
	mux.HandleFunc("/reboot", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&d.gen1.rebootHits, 1)
		if d.gen1.rebootErr {
			http.Error(w, "reboot refused", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"ok": true})
	})
}

func (d *apdevServer) registerGen2(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	mux.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read rpc body: %v", err)
			return
		}
		var req struct {
			ID     any            `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if uerr := json.Unmarshal(body, &req); uerr != nil {
			t.Errorf("decode rpc body: %v", uerr)
			return
		}
		switch req.Method {
		case "Shelly.GetDeviceInfo":
			d.writeRPCResult(w, req.ID, map[string]any{
				"id": "shellyplus1-aabbcc", "mac": "AABBCCDDEEFF", "gen": 2,
				"model": "SNSW-001P16EU", "fw_id": "20230101-000000",
			})
		case "WiFi.GetConfig", "Wifi.GetConfig":
			d.wifiReads.Add(1)
			if d.wifiReadFail {
				d.writeRPCError(w, req.ID)
				return
			}
			d.writeRPCResult(w, req.ID, d.gen2WiFiConfig())
		case "Shelly.GetConfig":
			d.wifiReads.Add(1)
			d.writeRPCResult(w, req.ID, map[string]any{
				"sys": map[string]any{"device": map[string]any{"name": "apdev", "mac": "AABBCCDDEEFF"}}, "wifi": d.gen2WiFiConfig(),
			})
		case "WiFi.SetConfig", "Wifi.SetConfig", "Shelly.SetConfig":
			d.record(apdevWrite{method: req.Method, params: req.Params})
			d.writeRPCResult(w, req.ID, map[string]any{"restart_required": false})
		case "Shelly.Reboot":
			atomic.AddInt32(&d.gen2.rebootHits, 1)
			if d.gen2.rebootErr {
				d.writeRPCError(w, req.ID)
				return
			}
			d.writeRPCResult(w, req.ID, map[string]any{})
		default:
			d.writeRPCResult(w, req.ID, map[string]any{})
		}
	})
}

func (d *apdevServer) writeRPCResult(w http.ResponseWriter, id, result any) {
	writeJSON(w, map[string]any{"id": id, "jsonrpc": "2.0", "result": result})
}

func (d *apdevServer) writeRPCError(w http.ResponseWriter, id any) {
	writeJSON(w, map[string]any{
		"id": id, "jsonrpc": "2.0",
		"error": map[string]any{"code": 500, "message": "reboot refused"},
	})
}

// writeJSON encodes v as the response body; an encode failure answers 500 so
// the caller sees an unusable device response.
func writeJSON(w http.ResponseWriter, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// apdevService builds a Service whose resolver points at the fake device, with a
// real rate limiter so the connection manager is fully wired.
func apdevService(d *apdevServer, generation int) *Service {
	return New(d.resolver(generation), WithRateLimiter(ratelimit.New()))
}

// refusingAddr returns a localhost address whose port is closed, so connection
// attempts fail fast and deterministically with "connection refused".
func refusingAddr(t *testing.T) string {
	t.Helper()
	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	if cerr := ln.Close(); cerr != nil {
		t.Fatalf("close listener: %v", cerr)
	}
	return addr
}
