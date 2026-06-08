package shelly

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// recordingResolver captures the identifier passed to Resolve so tests can
// assert which target ProvisionDevice connects to, then aborts the connection
// by returning an error before any network call is attempted.
type recordingResolver struct {
	lastIdentifier string
}

func (r *recordingResolver) Resolve(identifier string) (model.Device, error) {
	r.lastIdentifier = identifier
	return model.Device{}, errors.New("resolve aborted for test")
}

func TestProvisionDeviceTarget(t *testing.T) {
	t.Parallel()

	wifi := &model.ProvisionWiFiConfig{SSID: "net", Password: "pw"}

	t.Run("explicit address is used as connection target", func(t *testing.T) {
		t.Parallel()

		resolver := &recordingResolver{}
		svc := New(resolver)

		err := svc.ProvisionDevice(context.Background(), model.DeviceProvisionConfig{
			Name:    "bedroom",
			Address: "192.168.1.100",
		}, wifi)

		if err == nil {
			t.Fatal("expected error from aborted resolve")
		}
		if resolver.lastIdentifier != "192.168.1.100" {
			t.Errorf("expected connection target %q, got %q", "192.168.1.100", resolver.lastIdentifier)
		}
	})

	t.Run("name is used as connection target when address is empty", func(t *testing.T) {
		t.Parallel()

		resolver := &recordingResolver{}
		svc := New(resolver)

		err := svc.ProvisionDevice(context.Background(), model.DeviceProvisionConfig{
			Name: "bedroom",
		}, wifi)

		if err == nil {
			t.Fatal("expected error from aborted resolve")
		}
		if resolver.lastIdentifier != "bedroom" {
			t.Errorf("expected connection target %q, got %q", "bedroom", resolver.lastIdentifier)
		}
	})
}

func TestValidateBulkProvisionConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid config with addresses", func(t *testing.T) {
		t.Parallel()

		cfg := &model.BulkProvisionConfig{
			Devices: []model.DeviceProvisionConfig{
				{Name: "device1", Address: "192.168.1.100"},
				{Name: "device2", Address: "192.168.1.101"},
			},
		}

		// Always return false - devices not registered
		err := ValidateBulkProvisionConfig(cfg, func(name string) bool { return false })

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid config with registered devices", func(t *testing.T) {
		t.Parallel()

		cfg := &model.BulkProvisionConfig{
			Devices: []model.DeviceProvisionConfig{
				{Name: "kitchen-switch"},
				{Name: "living-room"},
			},
		}

		// Always return true - devices are registered
		err := ValidateBulkProvisionConfig(cfg, func(name string) bool { return true })

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unregistered device without address", func(t *testing.T) {
		t.Parallel()

		cfg := &model.BulkProvisionConfig{
			Devices: []model.DeviceProvisionConfig{
				{Name: "unknown-device"},
			},
		}

		err := ValidateBulkProvisionConfig(cfg, func(name string) bool { return false })

		if err == nil {
			t.Error("expected error for unregistered device without address")
		}
	})

	t.Run("invalid device name with path separator", func(t *testing.T) {
		t.Parallel()

		cfg := &model.BulkProvisionConfig{
			Devices: []model.DeviceProvisionConfig{
				{Name: "invalid/name", Address: "192.168.1.100"},
			},
		}

		err := ValidateBulkProvisionConfig(cfg, func(name string) bool { return false })

		if err == nil {
			t.Error("expected error for invalid device name with path separator")
		}
	})

	t.Run("empty device name", func(t *testing.T) {
		t.Parallel()

		cfg := &model.BulkProvisionConfig{
			Devices: []model.DeviceProvisionConfig{
				{Name: "", Address: "192.168.1.100"},
			},
		}

		err := ValidateBulkProvisionConfig(cfg, func(name string) bool { return false })

		if err == nil {
			t.Error("expected error for empty device name")
		}
	})

	t.Run("mixed valid and invalid", func(t *testing.T) {
		t.Parallel()

		cfg := &model.BulkProvisionConfig{
			Devices: []model.DeviceProvisionConfig{
				{Name: "valid-device", Address: "192.168.1.100"},
				{Name: "invalid:name"},
			},
		}

		err := ValidateBulkProvisionConfig(cfg, func(name string) bool { return false })

		if err == nil {
			t.Error("expected error for mixed config")
		}
	})
}

func TestProvisionDevicesParallelClamp(t *testing.T) {
	t.Parallel()

	cfg := &model.BulkProvisionConfig{
		WiFi: &model.ProvisionWiFiConfig{SSID: "net", Password: "pw"},
		Devices: []model.DeviceProvisionConfig{
			{Name: "a", Address: "192.168.1.1"},
			{Name: "b", Address: "192.168.1.2"},
		},
	}

	for _, parallel := range []int{0, -1} {
		t.Run("does not deadlock", func(t *testing.T) {
			t.Parallel()

			svc := New(&recordingResolver{})

			done := make(chan []model.ProvisionResult, 1)
			go func() {
				done <- svc.ProvisionDevices(context.Background(), cfg, parallel)
			}()

			select {
			case results := <-done:
				if len(results) != len(cfg.Devices) {
					t.Errorf("expected %d results, got %d", len(cfg.Devices), len(results))
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("ProvisionDevices deadlocked with parallel=%d", parallel)
			}
		})
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestParseBulkProvisionFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	t.Run("valid YAML file", func(t *testing.T) {
		tmpFile := "/test/provision.yaml"
		content := `wifi:
  ssid: "MyNetwork"
  password: "secret123"
devices:
  - name: device1
    address: 192.168.1.100
  - name: device2
    address: 192.168.1.101
`
		if err := afero.WriteFile(config.Fs(), tmpFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		cfg, err := ParseBulkProvisionFile(tmpFile)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected config to be non-nil")
		}
		if len(cfg.Devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(cfg.Devices))
		}
		if cfg.WiFi == nil {
			t.Error("expected WiFi config to be set")
		} else if cfg.WiFi.SSID != "MyNetwork" {
			t.Errorf("expected SSID 'MyNetwork', got %q", cfg.WiFi.SSID)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := ParseBulkProvisionFile("/nonexistent/file.yaml")

		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		tmpFile := "/test/invalid.yaml"
		content := `{invalid yaml: [unterminated`
		if err := afero.WriteFile(config.Fs(), tmpFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		_, err := ParseBulkProvisionFile(tmpFile)

		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpFile := "/test/empty.yaml"
		if err := afero.WriteFile(config.Fs(), tmpFile, []byte(""), 0o600); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		cfg, err := ParseBulkProvisionFile(tmpFile)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected config to be non-nil")
		}
		if len(cfg.Devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(cfg.Devices))
		}
	})
}
