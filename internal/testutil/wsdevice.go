package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// WSDeviceRealm is the realm a WSDevice puts in its digest challenge.
const WSDeviceRealm = "shellyplus1-wsfake"

// WSDeviceNotification is the notification a WSDevice sends after each
// request frame it answers.
const WSDeviceNotification = `{"src":"shellyplus1-wsfake","dst":"cli","method":"NotifyStatus","params":{"ts":1,"switch:0":{"output":true}}}`

// WSDevice is a loopback Gen2+ device websocket at ws://Addr/rpc. With a
// password it behaves like a device with authentication enabled: on each
// connection it answers request frames with error 401 and a digest challenge
// until one frame carries an auth object that RPCDigestAuthorized accepts for
// user "admin", and only then answers and starts notifications. Without a
// password it answers every frame. Each answered frame gets the result
// {"ok":true} followed by WSDeviceNotification.
type WSDevice struct {
	// Addr is the device's host:port.
	Addr     string
	password string
	nonce    any
	frames   chan string
	mu       sync.Mutex
	conns    []*websocket.Conn
}

// NewWSDevice starts a WSDevice closed when t ends. An empty password turns
// authentication off. nonce is the challenge nonce, with its JSON type.
func NewWSDevice(t *testing.T, password string, nonce any) *WSDevice {
	t.Helper()
	d := &WSDevice{password: password, nonce: nonce, frames: make(chan string, 64)}
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc" || !websocket.IsWebSocketUpgrade(r) {
			http.NotFound(w, r)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		d.mu.Lock()
		d.conns = append(d.conns, conn)
		d.mu.Unlock()
		d.serve(t, conn)
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(d.Drop)
	d.Addr = strings.TrimPrefix(srv.URL, "http://")
	return d
}

func (d *WSDevice) serve(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	authenticated := d.password == ""
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		d.frames <- string(data)
		var frame struct {
			ID   json.RawMessage `json:"id"`
			Auth json.RawMessage `json:"auth"`
		}
		if err := json.Unmarshal(data, &frame); err != nil {
			t.Errorf("WSDevice: decode frame %s: %v", data, err)
			return
		}
		if !authenticated && len(frame.Auth) > 0 {
			authenticated = RPCDigestAuthorized(frame.Auth, "admin", WSDeviceRealm, d.nonce, d.password)
		}
		var reply string
		if authenticated {
			reply = `{"id":` + string(frame.ID) + `,"src":"shellyplus1-wsfake","result":{"ok":true}}`
		} else {
			msg, err := json.Marshal(RPCDigestChallenge(WSDeviceRealm, d.nonce))
			if err != nil {
				t.Errorf("WSDevice: %v", err)
				return
			}
			reply = `{"id":` + string(frame.ID) + `,"src":"shellyplus1-wsfake","error":{"code":401,"message":` + string(msg) + `}}`
		}
		if conn.WriteMessage(websocket.TextMessage, []byte(reply)) != nil {
			return
		}
		if authenticated && conn.WriteMessage(websocket.TextMessage, []byte(WSDeviceNotification)) != nil {
			return
		}
	}
}

// NextFrame returns the next request frame the device received, failing t
// when none arrives within timeout.
func (d *WSDevice) NextFrame(t *testing.T, timeout time.Duration) string {
	t.Helper()
	select {
	case f := <-d.frames:
		return f
	case <-time.After(timeout):
		t.Fatalf("WSDevice: no request frame within %s", timeout)
		return ""
	}
}

// Frames returns the request frames received and not yet read by NextFrame.
func (d *WSDevice) Frames() []string {
	var out []string
	for {
		select {
		case f := <-d.frames:
			out = append(out, f)
		default:
			return out
		}
	}
}

// Drop closes every open connection, as a device that reboots or loses its
// network does.
func (d *WSDevice) Drop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, c := range d.conns {
		if err := c.Close(); err != nil {
			continue
		}
	}
	d.conns = nil
}
