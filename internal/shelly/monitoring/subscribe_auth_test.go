package monitoring

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

const subscribeAuthPassword = "s3cret-monitor-pass"

// resolveOnly is a ShellyConnector that only resolves; SubscribeEvents uses
// nothing else.
type resolveOnly struct {
	ShellyConnector
	dev model.Device
}

func (r resolveOnly) Resolve(string) (model.Device, error) { return r.dev, nil }

func subscribeAuthDevice(t *testing.T, auth *model.Auth, handler func(context.CancelFunc) EventHandler) error {
	t.Helper()
	d := testutil.NewWSDevice(t, subscribeAuthPassword, "AAAAAABnMonNonce==")
	svc := NewService(resolveOnly{dev: model.Device{Name: "bath", Address: d.Addr, Generation: 2, Auth: auth}})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	return svc.SubscribeEvents(ctx, "bath", handler(cancel))
}

func TestSubscribeEvents_PasswordProtectedDevice(t *testing.T) {
	t.Parallel()
	var got model.DeviceEvent
	err := subscribeAuthDevice(t, &model.Auth{Username: "admin", Password: subscribeAuthPassword},
		func(cancel context.CancelFunc) EventHandler {
			return func(e model.DeviceEvent) error {
				got = e
				cancel()
				return nil
			}
		})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want the stream to run until cancelled", err)
	}
	if got.Event != "NotifyStatus" {
		t.Errorf("event = %+v, want NotifyStatus", got)
	}
}

func TestSubscribeEvents_RejectedOrMissingCredentials(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		auth *model.Auth
		want error
	}{
		{&model.Auth{Username: "admin", Password: "wrong"}, model.ErrCredentialsRejected},
		{nil, model.ErrAuthRequired},
	} {
		err := subscribeAuthDevice(t, tc.auth, func(context.CancelFunc) EventHandler {
			return func(e model.DeviceEvent) error {
				t.Errorf("event delivered without valid credentials: %+v", e)
				return nil
			}
		})
		if !errors.Is(err, tc.want) {
			t.Errorf("auth %+v: err = %v, want %v", tc.auth, err, tc.want)
		}
	}
}
