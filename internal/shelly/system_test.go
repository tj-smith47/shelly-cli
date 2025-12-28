package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-go/gen2/components"
)

func TestSysStatus_Fields(t *testing.T) {
	t.Parallel()

	status := SysStatus{
		MAC:             "AA:BB:CC:DD:EE:FF",
		Uptime:          86400,
		Time:            "12:30:45",
		Unixtime:        1700000000,
		RAMFree:         50000,
		RAMSize:         100000,
		FSFree:          200000,
		FSSize:          500000,
		RestartRequired: false,
		CfgRev:          5,
		UpdateAvailable: "1.2.1",
	}

	if status.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected MAC 'AA:BB:CC:DD:EE:FF', got %q", status.MAC)
	}
	if status.Uptime != 86400 {
		t.Errorf("expected Uptime 86400, got %d", status.Uptime)
	}
	if status.Time != "12:30:45" {
		t.Errorf("expected Time '12:30:45', got %q", status.Time)
	}
	if status.Unixtime != 1700000000 {
		t.Errorf("expected Unixtime 1700000000, got %d", status.Unixtime)
	}
	if status.RAMFree != 50000 {
		t.Errorf("expected RAMFree 50000, got %d", status.RAMFree)
	}
	if status.FSFree != 200000 {
		t.Errorf("expected FSFree 200000, got %d", status.FSFree)
	}
	if status.RestartRequired {
		t.Error("expected RestartRequired to be false")
	}
	if status.CfgRev != 5 {
		t.Errorf("expected CfgRev 5, got %d", status.CfgRev)
	}
	if status.UpdateAvailable != "1.2.1" {
		t.Errorf("expected UpdateAvailable '1.2.1', got %q", status.UpdateAvailable)
	}
}

func TestSysConfig_Fields(t *testing.T) {
	t.Parallel()

	cfg := SysConfig{
		Name:         "Kitchen Switch",
		Timezone:     "America/New_York",
		Lat:          40.7128,
		Lng:          -74.0060,
		EcoMode:      true,
		Discoverable: true,
		Profile:      "switch",
		SNTPServer:   "time.google.com",
	}

	if cfg.Name != "Kitchen Switch" {
		t.Errorf("expected Name 'Kitchen Switch', got %q", cfg.Name)
	}
	if cfg.Timezone != "America/New_York" {
		t.Errorf("expected Timezone 'America/New_York', got %q", cfg.Timezone)
	}
	if cfg.Lat != 40.7128 {
		t.Errorf("expected Lat 40.7128, got %f", cfg.Lat)
	}
	if cfg.Lng != -74.0060 {
		t.Errorf("expected Lng -74.0060, got %f", cfg.Lng)
	}
	if !cfg.EcoMode {
		t.Error("expected EcoMode to be true")
	}
	if !cfg.Discoverable {
		t.Error("expected Discoverable to be true")
	}
	if cfg.Profile != "switch" {
		t.Errorf("expected Profile 'switch', got %q", cfg.Profile)
	}
	if cfg.SNTPServer != "time.google.com" {
		t.Errorf("expected SNTPServer 'time.google.com', got %q", cfg.SNTPServer)
	}
}

func TestExtractDeviceConfig(t *testing.T) {
	t.Parallel()

	t.Run("extracts all fields", func(t *testing.T) {
		t.Parallel()

		name := "Test Device"
		ecoMode := true
		discoverable := false
		profile := "cover"

		device := &components.SysDeviceConfig{
			Name:         &name,
			EcoMode:      &ecoMode,
			Discoverable: &discoverable,
			Profile:      &profile,
		}

		result := &SysConfig{}
		extractDeviceConfig(device, result)

		if result.Name != "Test Device" {
			t.Errorf("expected Name 'Test Device', got %q", result.Name)
		}
		if !result.EcoMode {
			t.Error("expected EcoMode to be true")
		}
		if result.Discoverable {
			t.Error("expected Discoverable to be false")
		}
		if result.Profile != "cover" {
			t.Errorf("expected Profile 'cover', got %q", result.Profile)
		}
	})

	t.Run("handles nil device", func(t *testing.T) {
		t.Parallel()

		result := &SysConfig{Name: "original"}
		extractDeviceConfig(nil, result)

		// Should not change anything
		if result.Name != "original" {
			t.Error("nil device should not change result")
		}
	})

	t.Run("handles partial config", func(t *testing.T) {
		t.Parallel()

		name := "Partial Device"
		device := &components.SysDeviceConfig{
			Name: &name,
			// Other fields are nil
		}

		result := &SysConfig{}
		extractDeviceConfig(device, result)

		if result.Name != "Partial Device" {
			t.Errorf("expected Name 'Partial Device', got %q", result.Name)
		}
		// Default values for unset fields
		if result.EcoMode {
			t.Error("expected EcoMode default false")
		}
	})
}

func TestExtractLocationConfig(t *testing.T) {
	t.Parallel()

	t.Run("extracts all fields", func(t *testing.T) {
		t.Parallel()

		tz := "Europe/London"
		lat := 51.5074
		lng := -0.1278

		location := &components.SysLocationConfig{
			TZ:  &tz,
			Lat: &lat,
			Lng: &lng,
		}

		result := &SysConfig{}
		extractLocationConfig(location, result)

		if result.Timezone != "Europe/London" {
			t.Errorf("expected Timezone 'Europe/London', got %q", result.Timezone)
		}
		if result.Lat != 51.5074 {
			t.Errorf("expected Lat 51.5074, got %f", result.Lat)
		}
		if result.Lng != -0.1278 {
			t.Errorf("expected Lng -0.1278, got %f", result.Lng)
		}
	})

	t.Run("handles nil location", func(t *testing.T) {
		t.Parallel()

		result := &SysConfig{Timezone: "original"}
		extractLocationConfig(nil, result)

		// Should not change anything
		if result.Timezone != "original" {
			t.Error("nil location should not change result")
		}
	})

	t.Run("handles partial config", func(t *testing.T) {
		t.Parallel()

		tz := "Asia/Tokyo"
		location := &components.SysLocationConfig{
			TZ: &tz,
			// Lat, Lng are nil
		}

		result := &SysConfig{}
		extractLocationConfig(location, result)

		if result.Timezone != "Asia/Tokyo" {
			t.Errorf("expected Timezone 'Asia/Tokyo', got %q", result.Timezone)
		}
		if result.Lat != 0 {
			t.Error("expected Lat default 0")
		}
	})
}
