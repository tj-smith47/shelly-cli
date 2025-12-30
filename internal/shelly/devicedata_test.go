package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDeviceData_Fields(t *testing.T) {
	t.Parallel()

	data := model.DeviceData{
		Name:       "kitchen-switch",
		Address:    testIP,
		Model:      testModel,
		Generation: 2,
		App:        testApp,
		Online:     true,
	}

	if data.Name != "kitchen-switch" {
		t.Errorf("expected Name 'kitchen-switch', got %q", data.Name)
	}
	if data.Address != testIP {
		t.Errorf("expected Address %q, got %q", testIP, data.Address)
	}
	if data.Model != testModel {
		t.Errorf("expected Model %q, got %q", testModel, data.Model)
	}
	if data.Generation != 2 {
		t.Errorf("expected Generation 2, got %d", data.Generation)
	}
	if data.App != testApp {
		t.Errorf("expected App %q, got %q", testApp, data.App)
	}
	if !data.Online {
		t.Error("expected Online to be true")
	}
}

//nolint:gocyclo // table-driven test with multiple validation checks
func TestSplitDevicesAndFile(t *testing.T) {
	t.Parallel()

	validExts := []string{".yaml", ".yml", ".json"}

	t.Run("no args", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{}, validExts)

		if len(devices) != 0 {
			t.Error("expected empty devices")
		}
		if file != "" {
			t.Error("expected empty file")
		}
	})

	t.Run("single device no file", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{"device1"}, validExts)

		if len(devices) != 1 || devices[0] != "device1" {
			t.Error("expected single device")
		}
		if file != "" {
			t.Error("expected empty file")
		}
	})

	t.Run("multiple devices no file", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{"dev1", "dev2", "dev3"}, validExts)

		if len(devices) != 3 {
			t.Errorf("expected 3 devices, got %d", len(devices))
		}
		if file != "" {
			t.Error("expected empty file")
		}
	})

	t.Run("devices with yaml file", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{"dev1", "dev2", "config.yaml"}, validExts)

		if len(devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(devices))
		}
		if file != "config.yaml" {
			t.Errorf("expected file 'config.yaml', got %q", file)
		}
	})

	t.Run("devices with yml file", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{"dev1", "config.yml"}, validExts)

		if len(devices) != 1 {
			t.Errorf("expected 1 device, got %d", len(devices))
		}
		if file != "config.yml" {
			t.Errorf("expected file 'config.yml', got %q", file)
		}
	})

	t.Run("devices with json file", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{"dev1", "dev2", "template.json"}, validExts)

		if len(devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(devices))
		}
		if file != "template.json" {
			t.Errorf("expected file 'template.json', got %q", file)
		}
	})

	t.Run("only file", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{"config.yaml"}, validExts)

		if len(devices) != 1 || devices[0] != "config.yaml" {
			t.Error("single arg should remain in devices")
		}
		if file != "" {
			t.Error("expected empty file for single arg")
		}
	})

	t.Run("device with file-like name but not matching extension", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{"dev1", "device.txt"}, validExts)

		if len(devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(devices))
		}
		if file != "" {
			t.Error("expected empty file for non-matching extension")
		}
	})

	t.Run("file path with directory", func(t *testing.T) {
		t.Parallel()

		devices, file := SplitDevicesAndFile([]string{"dev1", "/path/to/config.yaml"}, validExts)

		if len(devices) != 1 {
			t.Errorf("expected 1 device, got %d", len(devices))
		}
		if file != "/path/to/config.yaml" {
			t.Errorf("expected file '/path/to/config.yaml', got %q", file)
		}
	})
}
