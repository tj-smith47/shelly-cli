package shelly

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
)

// The in-scope --to-ap restore / at-AP firmware / reset-honesty functions reach a
// real device only through Service.WithConnection / WithGen1Connection, which dial
// the address the resolver hands back. Pointing that resolver at an httptest server
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
}

// apdevGen1 holds the mutable per-route behaviour of a fake Gen1 device.
type apdevGen1 struct {
	mu sync.Mutex
	// fw is the build string returned at /settings (".FW"); its first 8 chars are a
	// YYYYMMDD date that drives the downgrade decision in ensureGen1FirmwareAtAP. It
	// is guarded by mu because the /ota handler flips it while /settings reads it.
	fw string
	// uptime is the value returned at /status; >= 12 reads as "stable" to the
	// pre-station-write gate (gen1StableUptime), < 12 reads as a reboot loop.
	uptime int
	// rebootErr / resetErr / settingsErr, when true, make the matching endpoint
	// answer 500 so the production call sees a real (non-connectivity) failure.
	rebootErr   bool
	resetErr    bool
	settingsErr bool

	rebootHits int32
	resetHits  int32

	// otaFlipFW, when non-empty, is the build the device "boots onto" the moment its
	// /ota endpoint is triggered — modelling a successful flash so the post-OTA wait
	// observes a changed build on its first poll instead of running the full budget.
	otaFlipFW string
	otaErr    bool
	otaHits   int32
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
	d.srv = httptest.NewServer(mux)
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
	return &generationAwareResolver{device: model.Device{
		Name:       "apdev",
		Address:    d.addr(),
		Generation: generation,
	}}
}

func (g *apdevGen1) currentFW() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.fw
}

func (g *apdevGen1) setFW(fw string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.fw = fw
}

func (d *apdevServer) registerGen1(mux *http.ServeMux) {
	// /shelly identifies the device for ConnectGen1.
	mux.HandleFunc("/shelly", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"type": "SHBDUO-1", "mac": "AABBCCDDEEFF", "fw": d.gen1.currentFW(), "gen": 1,
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
		if d.gen1.settingsErr {
			http.Error(w, "settings refused", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"fw": d.gen1.currentFW(), "device": map[string]any{"type": "SHBDUO-1"}})
	})
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
	mux.HandleFunc("/ota", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&d.gen1.otaHits, 1)
		// Model a flash that takes: the device "boots onto" the new build, so the
		// post-OTA wait sees a changed FW (and the already-stable uptime) at once.
		if d.gen1.otaFlipFW != "" {
			d.gen1.setFW(d.gen1.otaFlipFW)
		}
		if d.gen1.otaErr {
			http.Error(w, "ota refused", http.StatusInternalServerError)
			return
		}
		const statusKey = "status" // JSON key in the fake Gen1 OTA status response
		writeJSON(w, map[string]any{statusKey: "updating"})
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
			ID     any    `json:"id"`
			Method string `json:"method"`
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

// writeJSON encodes v as the response body, ignoring encode errors (the test fails
// downstream if the device response is unusable, which is the real signal).
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v) //nolint:errchkjson,errcheck // test stub response
}

// apdevService builds a Service whose resolver points at the fake device, with a
// real rate limiter so the connection manager is fully wired.
func apdevService(d *apdevServer, generation int) *Service {
	return New(d.resolver(generation), WithRateLimiter(ratelimit.New()))
}
