package client

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

const (
	wsTestPassword = "s3cret-ws-pass"
	wsTestNonce    = "AAAAAABnwsNonceBase64=="
	wsTestWait     = 5 * time.Second
)

// connectWSDevice opens a reconnecting websocket to d with a notification
// channel subscribed.
func connectWSDevice(t *testing.T, d *testutil.WSDevice, auth *model.Auth) (*DeviceWebSocket, <-chan json.RawMessage) {
	t.Helper()
	ws, err := NewDeviceWebSocket("ws://"+d.Addr+"/rpc", auth,
		transport.WithReconnect(true), transport.WithRetry(3, 10*time.Millisecond))
	if err != nil {
		t.Fatalf("NewDeviceWebSocket: %v", err)
	}
	t.Cleanup(func() {
		if err := ws.Close(); err != nil {
			t.Logf("close: %v", err)
		}
	})
	if err := ws.Connect(t.Context()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	notes := make(chan json.RawMessage, 16)
	if err := ws.Subscribe(func(m json.RawMessage) { notes <- m }); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	return ws, notes
}

func waitNotification(t *testing.T, notes <-chan json.RawMessage) {
	t.Helper()
	select {
	case m := <-notes:
		if !strings.Contains(string(m), "NotifyStatus") {
			t.Fatalf("notification = %s, want NotifyStatus", m)
		}
	case <-time.After(wsTestWait):
		t.Fatal("no notification arrived")
	}
}

func assertNoPassword(t *testing.T, texts ...string) {
	t.Helper()
	for _, s := range texts {
		if strings.Contains(s, wsTestPassword) {
			t.Fatalf("password leaked: %s", s)
		}
	}
}

func TestStartDeviceNotifications_DigestAuth(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		nonce any
	}{
		{"string nonce (firmware 2.0.0+)", wsTestNonce},
		{"numeric nonce (older firmware)", 1625053638},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := testutil.NewWSDevice(t, wsTestPassword, tc.nonce)
			ws, notes := connectWSDevice(t, d, &model.Auth{Username: "admin", Password: wsTestPassword})

			status, err := StartDeviceNotifications(t.Context(), ws)
			if err != nil {
				t.Fatalf("StartDeviceNotifications: %v", err)
			}
			if string(status) != `{"ok":true}` {
				t.Errorf("status = %s", status)
			}
			waitNotification(t, notes)

			first, second := d.NextFrame(t, wsTestWait), d.NextFrame(t, wsTestWait)
			if strings.Contains(first, `"auth"`) || !strings.Contains(second, `"auth"`) {
				t.Errorf("frames = %s / %s, want unauthenticated then authenticated", first, second)
			}
			assertNoPassword(t, first, second)
		})
	}
}

func TestCallDeviceWebSocket_UsernameDefaultsToAdmin(t *testing.T) {
	t.Parallel()
	d := testutil.NewWSDevice(t, wsTestPassword, wsTestNonce)
	ws, _ := connectWSDevice(t, d, &model.Auth{Password: wsTestPassword})
	if _, err := CallDeviceWebSocket(t.Context(), ws, "Sys.GetStatus", nil); err != nil {
		t.Fatalf("CallDeviceWebSocket: %v", err)
	}
}

func TestCallDeviceWebSocket_WrongPassword(t *testing.T) {
	t.Parallel()
	d := testutil.NewWSDevice(t, wsTestPassword, wsTestNonce)
	wrong := "not-" + wsTestPassword
	ws, notes := connectWSDevice(t, d, &model.Auth{Username: "admin", Password: wrong})

	_, err := StartDeviceNotifications(t.Context(), ws)
	if !errors.Is(err, model.ErrCredentialsRejected) {
		t.Fatalf("err = %v, want ErrCredentialsRejected", err)
	}
	if err.Error() != "device rejected the password for user admin" {
		t.Errorf("err = %q", err)
	}
	if frames := d.Frames(); len(frames) != 2 {
		t.Errorf("frames = %d, want 2 (one retry): %v", len(frames), frames)
	} else {
		assertNoPassword(t, frames...)
		for _, f := range frames {
			if strings.Contains(f, wrong) {
				t.Errorf("frame carries the password: %s", f)
			}
		}
	}
	select {
	case m := <-notes:
		t.Errorf("notification after a rejected password: %s", m)
	default:
	}
}

func TestCallDeviceWebSocket_NoCredentials(t *testing.T) {
	t.Parallel()
	for _, auth := range []*model.Auth{nil, {Username: "admin"}} {
		d := testutil.NewWSDevice(t, wsTestPassword, wsTestNonce)
		ws, _ := connectWSDevice(t, d, auth)
		_, err := CallDeviceWebSocket(t.Context(), ws, "Shelly.GetStatus", nil)
		if !errors.Is(err, model.ErrAuthRequired) || errors.Is(err, model.ErrCredentialsRejected) {
			t.Fatalf("auth %+v: err = %v, want ErrAuthRequired", auth, err)
		}
		if frames := d.Frames(); len(frames) != 1 {
			t.Errorf("frames = %v, want 1 (no retry without credentials)", frames)
		}
	}
}

func TestStartDeviceNotifications_AuthDisabled(t *testing.T) {
	t.Parallel()
	d := testutil.NewWSDevice(t, "", wsTestNonce)
	ws, notes := connectWSDevice(t, d, &model.Auth{Username: "admin", Password: wsTestPassword})

	if _, err := StartDeviceNotifications(t.Context(), ws); err != nil {
		t.Fatalf("StartDeviceNotifications: %v", err)
	}
	waitNotification(t, notes)
	frames := d.Frames()
	if len(frames) != 1 || strings.Contains(frames[0], `"auth"`) {
		t.Errorf("frames = %v, want one frame without auth", frames)
	}
}

func TestStartDeviceNotifications_ResubscribesAfterReconnect(t *testing.T) {
	t.Parallel()
	d := testutil.NewWSDevice(t, wsTestPassword, wsTestNonce)
	ws, notes := connectWSDevice(t, d, &model.Auth{Username: "admin", Password: wsTestPassword})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if _, err := StartDeviceNotifications(ctx, ws); err != nil {
		t.Fatalf("StartDeviceNotifications: %v", err)
	}
	waitNotification(t, notes)
	d.NextFrame(t, wsTestWait)
	d.NextFrame(t, wsTestWait)

	d.Drop()

	// The transport keeps the device's nonce, so the frame sent after a
	// reconnect authenticates without a second challenge.
	if frame := d.NextFrame(t, wsTestWait); !strings.Contains(frame, `"auth"`) {
		t.Errorf("frame after reconnect = %s, want it authenticated", frame)
	}
	waitNotification(t, notes)
}
