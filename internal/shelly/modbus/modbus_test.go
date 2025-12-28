// Package modbus provides Modbus configuration for Shelly devices.
package modbus

import (
	"context"
	"encoding/json"
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

func TestNew(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := New(provider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != provider {
		t.Error("expected provider to be set")
	}
}

func TestNewWithNilProvider(t *testing.T) {
	t.Parallel()

	svc := New(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != nil {
		t.Error("expected provider to be nil")
	}
}

func TestStatus_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  Status
		enabled bool
	}{
		{
			name:    "enabled",
			status:  Status{Enabled: true},
			enabled: true,
		},
		{
			name:    "disabled",
			status:  Status{Enabled: false},
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.status.Enabled != tt.enabled {
				t.Errorf("got Enabled=%v, want %v", tt.status.Enabled, tt.enabled)
			}
		})
	}
}

func TestStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var status Status

	if status.Enabled {
		t.Error("expected Enabled to be false by default")
	}
}

func TestConfig_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config Config
		enable bool
	}{
		{
			name:   "enabled",
			config: Config{Enable: true},
			enable: true,
		},
		{
			name:   "disabled",
			config: Config{Enable: false},
			enable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.config.Enable != tt.enable {
				t.Errorf("got Enable=%v, want %v", tt.config.Enable, tt.enable)
			}
		})
	}
}

func TestConfig_ZeroValue(t *testing.T) {
	t.Parallel()

	var config Config

	if config.Enable {
		t.Error("expected Enable to be false by default")
	}
}

func TestStatus_JSONSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   Status
		wantJSON string
	}{
		{
			name:     "enabled",
			status:   Status{Enabled: true},
			wantJSON: `{"enabled":true}`,
		},
		{
			name:     "disabled",
			status:   Status{Enabled: false},
			wantJSON: `{"enabled":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jsonBytes, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			if string(jsonBytes) != tt.wantJSON {
				t.Errorf("got JSON=%s, want %s", string(jsonBytes), tt.wantJSON)
			}
		})
	}
}

func TestStatus_JSONDeserialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		jsonStr     string
		wantEnabled bool
	}{
		{
			name:        "enabled",
			jsonStr:     `{"enabled":true}`,
			wantEnabled: true,
		},
		{
			name:        "disabled",
			jsonStr:     `{"enabled":false}`,
			wantEnabled: false,
		},
		{
			name:        "empty object",
			jsonStr:     `{}`,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var status Status
			if err := json.Unmarshal([]byte(tt.jsonStr), &status); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if status.Enabled != tt.wantEnabled {
				t.Errorf("got Enabled=%v, want %v", status.Enabled, tt.wantEnabled)
			}
		})
	}
}

func TestConfig_JSONSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   Config
		wantJSON string
	}{
		{
			name:     "enabled",
			config:   Config{Enable: true},
			wantJSON: `{"enable":true}`,
		},
		{
			name:     "disabled",
			config:   Config{Enable: false},
			wantJSON: `{"enable":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jsonBytes, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			if string(jsonBytes) != tt.wantJSON {
				t.Errorf("got JSON=%s, want %s", string(jsonBytes), tt.wantJSON)
			}
		})
	}
}

func TestConfig_JSONDeserialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		jsonStr    string
		wantEnable bool
	}{
		{
			name:       "enabled",
			jsonStr:    `{"enable":true}`,
			wantEnable: true,
		},
		{
			name:       "disabled",
			jsonStr:    `{"enable":false}`,
			wantEnable: false,
		},
		{
			name:       "empty object",
			jsonStr:    `{}`,
			wantEnable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var config Config
			if err := json.Unmarshal([]byte(tt.jsonStr), &config); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if config.Enable != tt.wantEnable {
				t.Errorf("got Enable=%v, want %v", config.Enable, tt.wantEnable)
			}
		})
	}
}

func TestGetStatus_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	status, err := svc.GetStatus(context.Background(), "test-device")

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if status != nil {
		t.Errorf("expected nil status, got %v", status)
	}
}

func TestGetConfig_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	config, err := svc.GetConfig(context.Background(), "test-device")

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if config != nil {
		t.Errorf("expected nil config, got %v", config)
	}
}

func TestSetConfig_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	err := svc.SetConfig(context.Background(), "test-device", true)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestSetConfig_EnableTrue(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			// Return an error to simulate the RPC call failing (we can't actually test success without network)
			return errors.New("rpc call not mocked")
		},
	}

	svc := New(provider)
	//nolint:errcheck // intentionally ignoring error to test identifier passthrough
	_ = svc.SetConfig(context.Background(), "my-device", true)

	if capturedIdentifier != "my-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "my-device")
	}
}

func TestSetConfig_EnableFalse(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc call not mocked")
		},
	}

	svc := New(provider)
	//nolint:errcheck // intentionally ignoring error to test identifier passthrough
	_ = svc.SetConfig(context.Background(), "another-device", false)

	if capturedIdentifier != "another-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "another-device")
	}
}

func TestGetStatus_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc call not mocked")
		},
	}

	svc := New(provider)
	//nolint:errcheck // intentionally ignoring error to test identifier passthrough
	_, _ = svc.GetStatus(context.Background(), "test-identifier")

	if capturedIdentifier != "test-identifier" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "test-identifier")
	}
}

func TestGetConfig_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc call not mocked")
		},
	}

	svc := New(provider)
	//nolint:errcheck // intentionally ignoring error to test identifier passthrough
	_, _ = svc.GetConfig(context.Background(), "config-identifier")

	if capturedIdentifier != "config-identifier" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "config-identifier")
	}
}

func TestConnectionProvider_Interface(t *testing.T) {
	t.Parallel()

	// This test verifies that mockConnectionProvider satisfies the ConnectionProvider interface
	var _ ConnectionProvider = (*mockConnectionProvider)(nil)
}
