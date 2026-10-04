package shelly_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/auth"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

const (
	setAuthNewPassword  = "n3w-device-pass"
	setAuthNextPassword = "n3xt-device-pass"
	setAuthGen1User     = "bob"
)

// setAuthService returns a service resolving every identifier to the
// unlocked mock device of the given generation. The mock keeps whatever login
// SetAuth gives it and checks every later request against that login only.
func setAuthService(t *testing.T, generation int) (*shelly.Service, *testutil.Resolver) {
	t.Helper()
	srv := credentialsServer(t, generation)
	res := &testutil.Resolver{Device: credentialsDevice(srv, credsUnlocked, generation, "")}
	return shelly.New(res), res
}

// withAuth returns dev carrying user and password.
func withAuth(dev model.Device, user, password string) model.Device {
	dev.Auth = &model.Auth{Username: user, Password: password}
	return dev
}

func TestSetAuth_RotateDisableAndStatus(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		generation int
		user       string
	}{
		{"gen1 any user", 1, setAuthGen1User},
		{"gen1 admin", 1, auth.DefaultUser},
		{"gen2 admin", 2, auth.DefaultUser},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc, res := setAuthService(t, tc.generation)
			ctx := context.Background()
			name := res.Device.Name
			open := res.Device

			assertStatus(ctx, t, svc, name, false)

			user, stored, err := svc.SetAuth(ctx, name, tc.user, setAuthNewPassword)
			if err != nil {
				t.Fatalf("SetAuth: %v", err)
			}
			if user != tc.user {
				t.Errorf("SetAuth set user %q, want %q", user, tc.user)
			}
			if stored {
				t.Error("stored = true for a device that is not in the config")
			}
			assertStatus(ctx, t, svc, name, true)
			assertAccepted(ctx, t, svc, withAuth(open, tc.user, setAuthNewPassword))
			assertRejected(ctx, t, svc, withAuth(open, tc.user, credsWrong))
			assertRejected(ctx, t, svc, open)

			// Every later request needs the new password, as the CLI would
			// send from the stored credentials. Rotating without a user keeps
			// the stored one.
			res.Device = withAuth(open, tc.user, setAuthNewPassword)
			if user, _, err := svc.SetAuth(ctx, name, "", setAuthNextPassword); err != nil || user != tc.user {
				t.Fatalf("rotate: user %q, err %v; want user %q", user, err, tc.user)
			}
			assertRejected(ctx, t, svc, withAuth(open, tc.user, setAuthNewPassword))
			assertAccepted(ctx, t, svc, withAuth(open, tc.user, setAuthNextPassword))

			res.Device = withAuth(open, tc.user, setAuthNextPassword)
			if err := svc.DisableAuth(ctx, name); err != nil {
				t.Fatalf("DisableAuth: %v", err)
			}
			assertStatus(ctx, t, svc, name, false)
			assertAccepted(ctx, t, svc, open)
		})
	}
}

func TestSetAuth_Gen2RejectsUserOtherThanAdmin(t *testing.T) {
	t.Parallel()
	svc, res := setAuthService(t, 2)
	ctx := context.Background()

	_, _, err := svc.SetAuth(ctx, res.Device.Name, setAuthGen1User, setAuthNewPassword)
	if !errors.Is(err, auth.ErrInvalidAuthParams) {
		t.Fatalf("err = %v, want ErrInvalidAuthParams", err)
	}
	if !strings.Contains(err.Error(), "Gen2+ devices have a single user, admin") {
		t.Errorf("error %q does not say Gen2+ devices have a single user, admin", err)
	}
	assertStatus(ctx, t, svc, res.Device.Name, false)
	assertAccepted(ctx, t, svc, res.Device)
}

func assertStatus(ctx context.Context, t *testing.T, svc *shelly.Service, name string, want bool) {
	t.Helper()
	st, err := svc.GetAuthStatus(ctx, name)
	if err != nil {
		t.Fatalf("GetAuthStatus: %v", err)
	}
	if st.Enabled != want {
		t.Errorf("auth enabled = %v, want %v", st.Enabled, want)
	}
}

func assertAccepted(ctx context.Context, t *testing.T, svc *shelly.Service, dev model.Device) {
	t.Helper()
	if err := svc.VerifyCredentials(ctx, dev); err != nil {
		t.Errorf("credentials %+v refused: %v", dev.Auth, err)
	}
}

func assertRejected(ctx context.Context, t *testing.T, svc *shelly.Service, dev model.Device) {
	t.Helper()
	err := svc.VerifyCredentials(ctx, dev)
	if !errors.Is(err, shelly.ErrCredentialsRejected) && !errors.Is(err, model.ErrAuthRequired) {
		t.Errorf("credentials %+v: err = %v, want a rejection", dev.Auth, err)
	}
}
