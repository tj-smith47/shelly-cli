package mock

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func authFixtures() *Fixtures {
	return &Fixtures{
		Version: "1",
		Config: ConfigFixture{Devices: []DeviceFixture{
			{Name: "gen2", MAC: "AA:BB:CC:DD:EE:01", Model: "SNSW-001P16EU", Generation: 2, AuthEnabled: true, AuthPass: "pw"},
			{Name: "gen1", MAC: "AA:BB:CC:DD:EE:02", Type: "SHSW-1", Generation: 1, AuthEnabled: true, AuthUser: "admin", AuthPass: "pw"},
		}},
	}
}

// request sends one request to the mock device and returns the status and the
// WWW-Authenticate header.
func request(t *testing.T, method, url, body string, basicPass string) (status int, challenge string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if basicPass != "" {
		req.SetBasicAuth("admin", basicPass)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Logf("warning: close body: %v", err)
	}
	return resp.StatusCode, resp.Header.Get("WWW-Authenticate")
}

func TestAuthorized_Gen2RequiresDigest(t *testing.T) {
	t.Parallel()
	ds := NewDeviceServer(authFixtures())
	defer ds.Close()
	rpc := ds.DeviceURL("gen2") + "/rpc"

	if status, _ := request(t, http.MethodPost, rpc, `{"id":1,"method":"Shelly.GetDeviceInfo"}`, ""); status != http.StatusOK {
		t.Errorf("Shelly.GetDeviceInfo without credentials: status %d, want 200", status)
	}
	status, challenge := request(t, http.MethodPost, rpc, `{"id":1,"method":"Sys.GetStatus"}`, "")
	if status != http.StatusUnauthorized || !strings.HasPrefix(challenge, "Digest ") || !strings.Contains(challenge, "SHA-256") {
		t.Errorf("Sys.GetStatus without credentials: status %d, challenge %q; want 401 with a SHA-256 digest challenge", status, challenge)
	}
	// A real Gen2+ device does not accept basic credentials, right or not.
	if status, _ := request(t, http.MethodPost, rpc, `{"id":1,"method":"Sys.GetStatus"}`, "pw"); status != http.StatusUnauthorized {
		t.Errorf("Sys.GetStatus with basic credentials: status %d, want 401", status)
	}
}

func TestAuthorized_Gen1RequiresBasic(t *testing.T) {
	t.Parallel()
	ds := NewDeviceServer(authFixtures())
	defer ds.Close()
	base := ds.DeviceURL("gen1")

	if status, _ := request(t, http.MethodGet, base+"/shelly", "", ""); status != http.StatusOK {
		t.Errorf("/shelly without credentials: status %d, want 200", status)
	}
	status, challenge := request(t, http.MethodGet, base+"/settings", "", "")
	if status != http.StatusUnauthorized || !strings.HasPrefix(challenge, "Basic ") {
		t.Errorf("/settings without credentials: status %d, challenge %q; want 401 with a basic challenge", status, challenge)
	}
	if status, _ := request(t, http.MethodGet, base+"/settings", "", "nope"); status != http.StatusUnauthorized {
		t.Errorf("/settings with a wrong password: status %d, want 401", status)
	}
	if status, _ := request(t, http.MethodGet, base+"/settings", "", "pw"); status != http.StatusOK {
		t.Errorf("/settings with the right password: status %d, want 200", status)
	}
}
