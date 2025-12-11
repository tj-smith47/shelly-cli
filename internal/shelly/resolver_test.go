// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestNewConfigResolver(t *testing.T) {
	t.Parallel()
	resolver := NewConfigResolver()
	if resolver == nil {
		t.Fatal("NewConfigResolver() returned nil")
	}
}

func TestConfigDeviceToModel_Basic(t *testing.T) {
	t.Parallel()
	cfg := config.Device{
		Name:       "test-device",
		Address:    "192.168.1.100",
		Generation: 2,
		Type:       "switch",
		Model:      "SHSW-1",
	}

	result := configDeviceToModel(cfg)

	if result.Name != cfg.Name {
		t.Errorf("Name = %q, want %q", result.Name, cfg.Name)
	}
	if result.Address != cfg.Address {
		t.Errorf("Address = %q, want %q", result.Address, cfg.Address)
	}
	if result.Generation != cfg.Generation {
		t.Errorf("Generation = %d, want %d", result.Generation, cfg.Generation)
	}
	if result.Type != cfg.Type {
		t.Errorf("Type = %q, want %q", result.Type, cfg.Type)
	}
	if result.Model != cfg.Model {
		t.Errorf("Model = %q, want %q", result.Model, cfg.Model)
	}
	if result.Auth != nil {
		t.Error("Auth should be nil when not set in config")
	}
}

func TestConfigDeviceToModel_WithAuth(t *testing.T) {
	t.Parallel()
	cfg := config.Device{
		Name:    "secure-device",
		Address: "192.168.1.101",
		Auth: &config.Auth{
			Username: "admin",
			Password: "secret123",
		},
	}

	result := configDeviceToModel(cfg)

	if result.Auth == nil {
		t.Fatal("Auth should not be nil when set in config")
	}
	if result.Auth.Username != cfg.Auth.Username {
		t.Errorf("Auth.Username = %q, want %q", result.Auth.Username, cfg.Auth.Username)
	}
	if result.Auth.Password != cfg.Auth.Password {
		t.Errorf("Auth.Password = %q, want %q", result.Auth.Password, cfg.Auth.Password)
	}
}

func TestConfigDeviceToModel_EmptyAuth(t *testing.T) {
	t.Parallel()
	cfg := config.Device{
		Name:    "device-no-auth",
		Address: "192.168.1.102",
		Auth:    nil,
	}

	result := configDeviceToModel(cfg)

	if result.Auth != nil {
		t.Error("Auth should be nil when config auth is nil")
	}
}

func TestConfigDeviceToModel_PreservesAllFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  config.Device
	}{
		{
			name: "gen1 device",
			cfg: config.Device{
				Name:       "gen1",
				Address:    "10.0.0.1",
				Generation: 1,
				Type:       "relay",
				Model:      "SHSW-25",
			},
		},
		{
			name: "gen2 device",
			cfg: config.Device{
				Name:       "gen2",
				Address:    "10.0.0.2",
				Generation: 2,
				Type:       "switch",
				Model:      "SNSW-001P16EU",
			},
		},
		{
			name: "gen3 device",
			cfg: config.Device{
				Name:       "gen3",
				Address:    "10.0.0.3",
				Generation: 3,
				Type:       "pm",
				Model:      "S3PM-001PCEU16",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := configDeviceToModel(tt.cfg)

			if result.Name != tt.cfg.Name {
				t.Errorf("Name = %q, want %q", result.Name, tt.cfg.Name)
			}
			if result.Address != tt.cfg.Address {
				t.Errorf("Address = %q, want %q", result.Address, tt.cfg.Address)
			}
			if result.Generation != tt.cfg.Generation {
				t.Errorf("Generation = %d, want %d", result.Generation, tt.cfg.Generation)
			}
			if result.Type != tt.cfg.Type {
				t.Errorf("Type = %q, want %q", result.Type, tt.cfg.Type)
			}
			if result.Model != tt.cfg.Model {
				t.Errorf("Model = %q, want %q", result.Model, tt.cfg.Model)
			}
		})
	}
}

func TestDeviceResolver_Interface(t *testing.T) {
	t.Parallel()
	// Verify ConfigResolver implements DeviceResolver interface
	var _ DeviceResolver = (*ConfigResolver)(nil)
}

func TestDeviceResolver_MockImplementation(t *testing.T) {
	t.Parallel()
	// Verify mockResolver implements DeviceResolver interface
	var _ DeviceResolver = (*mockResolver)(nil)

	mock := &mockResolver{
		device: model.Device{
			Name:    "mock-device",
			Address: "192.168.1.50",
		},
	}

	device, err := mock.Resolve("any")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if device.Name != "mock-device" {
		t.Errorf("Name = %q, want %q", device.Name, "mock-device")
	}
}
