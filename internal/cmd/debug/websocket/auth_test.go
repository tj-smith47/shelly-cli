package websocket

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const authTestPassword = "s3cret-cmd-pass"

func runAgainstWSDevice(t *testing.T, devicePassword string, auth *model.Auth) (*factory.TestFactory, *testutil.WSDevice, error) {
	t.Helper()
	d := testutil.NewWSDevice(t, devicePassword, "AAAAAABnCmdNonce==")
	tf := factory.NewTestFactory(t)
	tf.SetShellyService(shelly.New(&testutil.Resolver{Device: model.Device{
		Name: "bath", Address: d.Addr, Generation: 2, Model: "SNSW-001X16EU", Auth: auth,
	}}))
	err := run(t.Context(), &Options{Factory: tf.Factory, Device: "bath", Duration: 300 * time.Millisecond, Raw: true})
	return tf, d, err
}

func TestRun_StreamsEventsFromPasswordProtectedDevice(t *testing.T) {
	t.Parallel()
	tf, d, err := runAgainstWSDevice(t, authTestPassword, &model.Auth{Username: "admin", Password: authTestPassword})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	out := tf.OutString() + tf.ErrString()
	if !strings.Contains(out, "NotifyStatus") || !strings.Contains(out, "Received 1 events") {
		t.Errorf("output does not show the notification:\n%s", out)
	}
	frames := d.Frames()
	if strings.Contains(out+strings.Join(frames, "\n"), authTestPassword) {
		t.Errorf("password appears in output or frames")
	}
}

func TestRun_WrongPasswordFails(t *testing.T) {
	t.Parallel()
	tf, _, err := runAgainstWSDevice(t, authTestPassword, &model.Auth{Username: "admin", Password: "wrong-" + authTestPassword})
	if !errors.Is(err, model.ErrCredentialsRejected) {
		t.Fatalf("err = %v, want ErrCredentialsRejected", err)
	}
	if strings.Contains(tf.OutString()+tf.ErrString()+err.Error(), authTestPassword) {
		t.Errorf("password appears in output or error")
	}
}

func TestRun_NoCredentialsFails(t *testing.T) {
	t.Parallel()
	_, _, err := runAgainstWSDevice(t, authTestPassword, nil)
	if !errors.Is(err, model.ErrAuthRequired) {
		t.Fatalf("err = %v, want ErrAuthRequired", err)
	}
}

func TestRun_AuthDisabledDeviceStreams(t *testing.T) {
	t.Parallel()
	tf, d, err := runAgainstWSDevice(t, "", nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out := tf.OutString(); !strings.Contains(out, "NotifyStatus") {
		t.Errorf("output does not show the notification:\n%s", out)
	}
	for _, f := range d.Frames() {
		if strings.Contains(f, `"auth"`) {
			t.Errorf("frame to an auth-disabled device carries auth: %s", f)
		}
	}
}
