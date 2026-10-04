package automation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/events"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

const eventsAuthPassword = "s3cret-tui-pass"

// connectAuthDevice runs connectDevice against a WSDevice requiring
// eventsAuthPassword, with auth as the stored credentials, and returns the
// device and a channel of every event published.
func connectAuthDevice(t *testing.T, auth *model.Auth) (*testutil.WSDevice, <-chan events.Event) {
	t.Helper()
	d := testutil.NewWSDevice(t, eventsAuthPassword, "AAAAAABnTuiNonce==")
	es := NewEventStream(&mockConnectionProvider{
		resolveWithGenerationFn: func(context.Context, string) (model.Device, error) {
			return model.Device{Name: "bath", Address: d.Addr, Generation: 2, Auth: auth}, nil
		},
	})
	t.Cleanup(es.Stop)
	got := make(chan events.Event, 64)
	es.Subscribe(func(e events.Event) { got <- e })
	es.connectDevice("bath", d.Addr)
	return d, got
}

func TestEventStream_PasswordProtectedDevice(t *testing.T) {
	t.Parallel()
	d, got := connectAuthDevice(t, &model.Auth{Username: "admin", Password: eventsAuthPassword})

	var sawStatus, sawNotification bool
	deadline := time.After(5 * time.Second)
	for !sawStatus || !sawNotification {
		select {
		case e := <-got:
			switch e.(type) {
			case *events.FullStatusEvent:
				sawStatus = true
			case *events.StatusChangeEvent:
				sawNotification = true
			}
		case <-deadline:
			t.Fatalf("full status %v, notification %v: want both", sawStatus, sawNotification)
		}
	}
	for _, f := range d.Frames() {
		if strings.Contains(f, eventsAuthPassword) {
			t.Fatalf("password sent in a frame: %s", f)
		}
	}
}

func TestEventStream_WrongPasswordPublishesNoStatus(t *testing.T) {
	t.Parallel()
	d, got := connectAuthDevice(t, &model.Auth{Username: "admin", Password: "wrong"})

	if frames := d.Frames(); len(frames) != 2 {
		t.Errorf("frames = %v, want the challenge and one authenticated retry", frames)
	}
	for {
		select {
		case e := <-got:
			if _, ok := e.(*events.FullStatusEvent); ok {
				t.Fatal("full status published although the device rejected the password")
			}
		default:
			return
		}
	}
}
