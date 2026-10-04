package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

// Authentication values of a mock device with AuthEnabled.
const (
	// gen2AuthUser is the only user a Gen2+ device accepts.
	gen2AuthUser = "admin"
	// mockAuthNonce is the nonce every digest challenge carries.
	mockAuthNonce = "5f6a7b8c"

	methodGetDeviceInfo = "Shelly.GetDeviceInfo"
)

// authorized reports whether the request may reach the device's handlers. A
// device without AuthEnabled accepts everything until a Gen2+ device is given
// a password through Shelly.SetAuth. A Gen1 device with it
// requires HTTP basic credentials on every endpoint except /shelly; a Gen2+
// device requires an HTTP digest response (SHA-256, qop=auth) on every RPC
// except Shelly.GetDeviceInfo, as real devices do. A refused request has
// already been answered with 401 and a WWW-Authenticate challenge.
func (ds *DeviceServer) authorized(w http.ResponseWriter, r *http.Request, device *DeviceFixture, endpoint string) bool {
	if endpoint == "/shelly" {
		return true
	}

	if device.Generation == 1 {
		login := ds.gen1LoginFor(device)
		if !login.enabled {
			return true
		}
		user, pass, ok := r.BasicAuth()
		if ok && user == login.user && pass == login.pass {
			return true
		}
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, device.Name))
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
		return false
	}

	ha1 := ds.gen2AuthHA1(device)
	realm := gen2DeviceID(device)
	if ha1 == "" || gen2PublicRPC(r, endpoint) || testutil.DigestAuthorizedHA1(r, gen2AuthUser, realm, mockAuthNonce, ha1) {
		return true
	}
	w.Header().Set("WWW-Authenticate", testutil.DigestChallenge(realm, mockAuthNonce))
	http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
	return false
}

// gen2AuthHA1 returns the HA1 a Gen2+ device checks digest responses
// against, or "" when its authentication is off. A device holds only this
// hash, never the password, so a hash computed the wrong way locks the
// password out exactly as it would on real hardware.
func (ds *DeviceServer) gen2AuthHA1(device *DeviceFixture) string {
	ds.mu.RLock()
	ha1, set := ds.authHA1[device.Name]
	ds.mu.RUnlock()
	if set {
		return ha1
	}
	if !device.AuthEnabled {
		return ""
	}
	return testutil.DigestHA1(gen2AuthUser, gen2DeviceID(device), device.AuthPass)
}

// gen2PublicRPC reports whether the request is Shelly.GetDeviceInfo, the one
// RPC a device answers without credentials. The body is put back for the
// handler that runs next.
func gen2PublicRPC(r *http.Request, endpoint string) bool {
	if endpoint == endpointRPC+"/"+methodGetDeviceInfo {
		return true
	}
	if endpoint != endpointRPC || r.Body == nil {
		return false
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var req struct {
		Method string `json:"method"`
	}
	return json.Unmarshal(body, &req) == nil && req.Method == methodGetDeviceInfo
}

// gen1Login is the login a Gen1 device requires when enabled.
type gen1Login struct {
	user, pass string
	enabled    bool
}

// gen1LoginFor returns the login device currently requires.
func (ds *DeviceServer) gen1LoginFor(device *DeviceFixture) gen1Login {
	ds.mu.RLock()
	login, set := ds.gen1Login[device.Name]
	ds.mu.RUnlock()
	if set {
		return login
	}
	return gen1Login{user: device.AuthUser, pass: device.AuthPass, enabled: device.AuthEnabled}
}

// handleGen1Login serves a Gen1 /settings/login request: enabled=1 with a
// username and password makes the device require them for basic auth,
// enabled=0 turns that off and keeps the stored login, as a real device does.
func (ds *DeviceServer) handleGen1Login(w http.ResponseWriter, r *http.Request, device *DeviceFixture) {
	login := ds.gen1LoginFor(device)
	q := r.URL.Query()
	if q.Has("username") {
		login.user = q.Get("username")
	}
	if q.Has("password") {
		login.pass = q.Get("password")
	}
	if q.Has("enabled") {
		login.enabled = q.Get("enabled") == "1" || q.Get("enabled") == "true"
	}
	if login.enabled && (login.user == "" || login.pass == "") {
		http.Error(w, `{"error":"username and password are required"}`, http.StatusBadRequest)
		return
	}
	ds.mu.Lock()
	ds.gen1Login[device.Name] = login
	ds.mu.Unlock()
	ds.writeJSON(w, map[string]any{"enabled": login.enabled, "unprotected": false, "username": login.user})
}
