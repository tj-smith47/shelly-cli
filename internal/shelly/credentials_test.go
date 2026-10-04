package shelly_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

const (
	credsUser     = "admin"
	credsRight    = "r1ght-pass"
	credsWrong    = "wr0ng-pass"
	credsLocked   = "locked"
	credsUnlocked = "unlocked"
)

// credentialsServer serves one password-protected and one open mock device.
func credentialsServer(t *testing.T, generation int) *mock.DeviceServer {
	t.Helper()
	deviceType := "SNSW-001P16EU"
	if generation == 1 {
		deviceType = "SHSW-1"
	}
	srv := mock.NewDeviceServer(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{Devices: []mock.DeviceFixture{
			{
				Name: credsLocked, MAC: "AA:BB:CC:DD:EE:01", Type: deviceType, Model: deviceType,
				Generation: generation, AuthEnabled: true, AuthUser: credsUser, AuthPass: credsRight,
			},
			{
				Name: credsUnlocked, MAC: "AA:BB:CC:DD:EE:02", Type: deviceType, Model: deviceType,
				Generation: generation,
			},
		}},
	})
	t.Cleanup(srv.Close)
	return srv
}

func credentialsDevice(srv *mock.DeviceServer, name string, generation int, password string) model.Device {
	dev := model.Device{Name: name, Address: srv.DeviceURL(name), Generation: generation}
	if password != "" {
		dev.Auth = &model.Auth{Username: credsUser, Password: password}
	}
	return dev
}

func TestVerifyCredentials(t *testing.T) {
	t.Parallel()

	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", generation), func(t *testing.T) {
			t.Parallel()
			srv := credentialsServer(t, generation)
			svc := shelly.New(&testutil.Resolver{})
			ctx := context.Background()

			t.Run("right password accepted", func(t *testing.T) {
				t.Parallel()
				if err := svc.VerifyCredentials(ctx, credentialsDevice(srv, credsLocked, generation, credsRight)); err != nil {
					t.Fatalf("VerifyCredentials() error = %v, want nil", err)
				}
			})

			t.Run("wrong password rejected", func(t *testing.T) {
				t.Parallel()
				err := svc.VerifyCredentials(ctx, credentialsDevice(srv, credsLocked, generation, credsWrong))
				if !errors.Is(err, shelly.ErrCredentialsRejected) {
					t.Fatalf("VerifyCredentials() error = %v, want ErrCredentialsRejected", err)
				}
				if got, want := err.Error(), "device rejected the password for user admin"; got != want {
					t.Errorf("error = %q, want %q", got, want)
				}
				if ratelimit.IsConnectivityFailure(err) {
					t.Error("a rejected password was classed as a connectivity failure")
				}
				if strings.Contains(err.Error(), credsWrong) {
					t.Error("the error contains the password")
				}
			})

			t.Run("no credentials for a protected device", func(t *testing.T) {
				t.Parallel()
				err := svc.VerifyCredentials(ctx, credentialsDevice(srv, credsLocked, generation, ""))
				if !errors.Is(err, model.ErrAuthRequired) || errors.Is(err, shelly.ErrCredentialsRejected) {
					t.Fatalf("VerifyCredentials() error = %v, want ErrAuthRequired", err)
				}
			})

			t.Run("device without authentication accepts anything", func(t *testing.T) {
				t.Parallel()
				if err := svc.VerifyCredentials(ctx, credentialsDevice(srv, credsUnlocked, generation, credsWrong)); err != nil {
					t.Fatalf("VerifyCredentials() error = %v, want nil", err)
				}
			})

			t.Run("probe answers whatever the password", func(t *testing.T) {
				t.Parallel()
				info, err := svc.ProbeDevice(ctx, credentialsDevice(srv, credsLocked, generation, credsWrong))
				if err != nil {
					t.Fatalf("ProbeDevice() error = %v", err)
				}
				if !info.AuthEn {
					t.Error("the device did not report authentication enabled")
				}
			})
		})
	}
}

func TestVerifyCredentials_UnreachableIsNotARejection(t *testing.T) {
	t.Parallel()

	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", generation), func(t *testing.T) {
			t.Parallel()
			srv := credentialsServer(t, generation)
			dev := credentialsDevice(srv, credsLocked, generation, credsRight)
			srv.Close()

			// The transport retries a refused connection with backoff; the
			// deadline ends the wait after the first refusal.
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			err := shelly.New(&testutil.Resolver{}).VerifyCredentials(ctx, dev)
			if err == nil {
				t.Fatal("VerifyCredentials() = nil for a device that cannot be reached")
			}
			if errors.Is(err, shelly.ErrCredentialsRejected) || errors.Is(err, model.ErrAuthRequired) {
				t.Errorf("error = %v, want a connection error", err)
			}
			if strings.Contains(err.Error(), credsRight) {
				t.Error("the error contains the password")
			}
		})
	}
}
