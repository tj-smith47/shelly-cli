package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

const (
	testNameAlpha   = "alpha"
	testNameBravo   = "bravo"
	testNameCharlie = "charlie"
)

func TestDeviceListFilterOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := DeviceListFilterOptions{
		Generation: 2,
		DeviceType: "switch",
		Platform:   "shelly",
	}

	if opts.Generation != 2 {
		t.Errorf("Generation = %d, want 2", opts.Generation)
	}
	if opts.DeviceType != "switch" {
		t.Errorf("DeviceType = %q, want %q", opts.DeviceType, "switch")
	}
	if opts.Platform != "shelly" {
		t.Errorf("Platform = %q, want %q", opts.Platform, "shelly")
	}
}

func TestFilterDeviceList(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"switch1": {
			Address:    "192.168.1.10",
			Generation: 2,
			Type:       "switch",
			Model:      "Shelly Plus 1",
		},
		"switch2": {
			Address:    "192.168.1.11",
			Generation: 1,
			Type:       "switch",
			Model:      "Shelly 1",
		},
		"dimmer1": {
			Address:    "192.168.1.12",
			Generation: 2,
			Type:       "dimmer",
			Model:      "Shelly Plus Dimmer",
		},
	}

	t.Run("no filters returns all", func(t *testing.T) {
		t.Parallel()
		result, platforms := FilterDeviceList(devices, DeviceListFilterOptions{})
		if len(result) != 3 {
			t.Errorf("len(result) = %d, want 3", len(result))
		}
		if len(platforms) == 0 {
			t.Error("platforms should not be empty")
		}
	})

	t.Run("filter by generation", func(t *testing.T) {
		t.Parallel()
		result, _ := FilterDeviceList(devices, DeviceListFilterOptions{Generation: 2})
		if len(result) != 2 {
			t.Errorf("len(result) = %d, want 2 (Gen2 devices)", len(result))
		}
	})

	t.Run("filter by device type", func(t *testing.T) {
		t.Parallel()
		result, _ := FilterDeviceList(devices, DeviceListFilterOptions{DeviceType: "switch"})
		if len(result) != 2 {
			t.Errorf("len(result) = %d, want 2 (switches)", len(result))
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		t.Parallel()
		result, _ := FilterDeviceList(devices, DeviceListFilterOptions{
			Generation: 2,
			DeviceType: "switch",
		})
		if len(result) != 1 {
			t.Errorf("len(result) = %d, want 1 (Gen2 switch)", len(result))
		}
	})

	t.Run("empty devices returns empty result", func(t *testing.T) {
		t.Parallel()
		result, platforms := FilterDeviceList(map[string]model.Device{}, DeviceListFilterOptions{})
		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
		if len(platforms) != 0 {
			t.Errorf("len(platforms) = %d, want 0", len(platforms))
		}
	})
}

func TestMatchesDeviceFilters(t *testing.T) {
	t.Parallel()

	device := model.Device{
		Generation: 2,
		Type:       "switch",
		Platform:   "shelly",
	}

	tests := []struct {
		name string
		opts DeviceListFilterOptions
		want bool
	}{
		{
			name: "no filters matches",
			opts: DeviceListFilterOptions{},
			want: true,
		},
		{
			name: "matching generation",
			opts: DeviceListFilterOptions{Generation: 2},
			want: true,
		},
		{
			name: "non-matching generation",
			opts: DeviceListFilterOptions{Generation: 1},
			want: false,
		},
		{
			name: "matching type",
			opts: DeviceListFilterOptions{DeviceType: "switch"},
			want: true,
		},
		{
			name: "non-matching type",
			opts: DeviceListFilterOptions{DeviceType: "dimmer"},
			want: false,
		},
		{
			name: "all filters matching",
			opts: DeviceListFilterOptions{Generation: 2, DeviceType: "switch"},
			want: true,
		},
		{
			name: "one filter not matching",
			opts: DeviceListFilterOptions{Generation: 2, DeviceType: "dimmer"},
			want: false,
		},
		{
			name: "matching platform",
			opts: DeviceListFilterOptions{Platform: "shelly"},
			want: true,
		},
		{
			name: "non-matching platform",
			opts: DeviceListFilterOptions{Platform: "tasmota"},
			want: false,
		},
		{
			name: "all filters including platform",
			opts: DeviceListFilterOptions{Generation: 2, DeviceType: "switch", Platform: "shelly"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := matchesDeviceFilters(device, tt.opts); got != tt.want {
				t.Errorf("matchesDeviceFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortDeviceList(t *testing.T) {
	t.Parallel()

	t.Run("sort by name only", func(t *testing.T) {
		t.Parallel()
		devices := []model.DeviceListItem{
			{Name: testNameCharlie},
			{Name: testNameAlpha},
			{Name: testNameBravo},
		}
		SortDeviceList(devices, false)
		if devices[0].Name != testNameAlpha || devices[1].Name != testNameBravo || devices[2].Name != testNameCharlie {
			t.Errorf("devices not sorted by name: %v", devices)
		}
	})

	t.Run("updates first enabled", func(t *testing.T) {
		t.Parallel()
		devices := []model.DeviceListItem{
			{Name: testNameCharlie, HasUpdate: false},
			{Name: testNameAlpha, HasUpdate: true},
			{Name: testNameBravo, HasUpdate: false},
		}
		SortDeviceList(devices, true)
		if !devices[0].HasUpdate {
			t.Error("first device should have update")
		}
		if devices[0].Name != testNameAlpha {
			t.Errorf("first device should be alpha (has update), got %q", devices[0].Name)
		}
	})

	t.Run("multiple updates sorted by name", func(t *testing.T) {
		t.Parallel()
		devices := []model.DeviceListItem{
			{Name: testNameBravo, HasUpdate: true},
			{Name: testNameAlpha, HasUpdate: true},
			{Name: testNameCharlie, HasUpdate: false},
		}
		SortDeviceList(devices, true)
		if devices[0].Name != testNameAlpha || devices[1].Name != testNameBravo {
			t.Errorf("devices with updates not sorted by name: %v", devices)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		devices := []model.DeviceListItem{}
		SortDeviceList(devices, true) // Should not panic
		if len(devices) != 0 {
			t.Errorf("len(devices) = %d, want 0", len(devices))
		}
	})
}
