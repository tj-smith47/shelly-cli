// Package auth provides authentication configuration for Shelly devices.
package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// mockConnectionProvider is a test double for ConnectionProvider.
type mockConnectionProvider struct {
	withConnectionFn func(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

func (m *mockConnectionProvider) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionProvider) WithGen1Connection(context.Context, string, func(*client.Gen1Client) error) error {
	return nil
}

// mockDeviceInfoProvider is a test double for DeviceInfoProvider.
type mockDeviceInfoProvider struct {
	authEnabled bool
	err         error
}

func (m *mockDeviceInfoProvider) GetAuthEnabled(ctx context.Context, identifier string) (bool, error) {
	return m.authEnabled, m.err
}

func TestNew(t *testing.T) {
	t.Parallel()

	connProvider := &mockConnectionProvider{}
	infoProvider := &mockDeviceInfoProvider{}

	svc := New(connProvider, infoProvider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != connProvider {
		t.Error("expected provider to be set")
	}
	if svc.infoProvider != infoProvider {
		t.Error("expected infoProvider to be set")
	}
}

func TestGetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		authEnabled bool
		err         error
		wantEnabled bool
		wantErr     bool
	}{
		{
			name:        "auth enabled",
			authEnabled: true,
			wantEnabled: true,
		},
		{
			name:        "auth disabled",
			authEnabled: false,
			wantEnabled: false,
		},
		{
			name:    "error getting auth status",
			err:     errors.New("device unreachable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			infoProvider := &mockDeviceInfoProvider{
				authEnabled: tt.authEnabled,
				err:         tt.err,
			}
			svc := New(&mockConnectionProvider{}, infoProvider)

			status, err := svc.GetStatus(context.Background(), "test-device")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status.Enabled != tt.wantEnabled {
				t.Errorf("got Enabled=%v, want %v", status.Enabled, tt.wantEnabled)
			}
		})
	}
}

func TestCalculateHA1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		user     string
		realm    string
		password string
		want     string
	}{
		{
			name:     "standard credentials",
			user:     DefaultUser,
			realm:    "shelly",
			password: "password123",
			// SHA-256 of "admin:shelly:password123"
			want: "517ec270b36ca59e99d5d0a2ec926151176cbbb6e12c18e8629714cf59c6823b",
		},
		{
			name:     "empty password",
			user:     DefaultUser,
			realm:    "shelly",
			password: "",
			// SHA-256 of "admin:shelly:"
			want: "0a24e34f1b156fb73da4ba13fefe4577c17c86af78e7d1e649e66a0b8a9922f2",
		},
		{
			name:     "empty realm",
			user:     DefaultUser,
			realm:    "",
			password: "test",
			// SHA-256 of "admin::test"
			want: "828881b3a0c927154a60bf439710c5d35c577ddcef89640b604d9b2367dd30dc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CalculateHA1(tt.user, tt.realm, tt.password)

			if got != tt.want {
				t.Errorf("CalculateHA1(%q, %q, %q) = %q, want %q",
					tt.user, tt.realm, tt.password, got, tt.want)
			}
		})
	}
}

func TestStatus_Fields(t *testing.T) {
	t.Parallel()

	status := Status{
		Enabled: true,
		User:    DefaultUser,
		Realm:   "shelly-device",
	}

	if !status.Enabled {
		t.Error("expected Enabled to be true")
	}
	if status.User != DefaultUser {
		t.Errorf("got User=%q, want %q", status.User, DefaultUser)
	}
	if status.Realm != "shelly-device" {
		t.Errorf("got Realm=%q, want %q", status.Realm, "shelly-device")
	}
}
