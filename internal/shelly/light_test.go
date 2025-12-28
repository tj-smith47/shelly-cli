package shelly

import (
	"testing"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"
)

func TestLightInfo_Fields(t *testing.T) {
	t.Parallel()

	info := LightInfo{
		ID:         0,
		Name:       "Living Room Light",
		Output:     true,
		Brightness: 75,
		Power:      50.5,
	}

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "Living Room Light" {
		t.Errorf("Name = %q, want %q", info.Name, "Living Room Light")
	}
	if !info.Output {
		t.Error("Output = false, want true")
	}
	if info.Brightness != 75 {
		t.Errorf("Brightness = %d, want 75", info.Brightness)
	}
	if info.Power != 50.5 {
		t.Errorf("Power = %f, want 50.5", info.Power)
	}
}

func TestLightInfo_ZeroValues(t *testing.T) {
	t.Parallel()

	var info LightInfo

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("Name = %q, want empty", info.Name)
	}
	if info.Output {
		t.Error("Output = true, want false")
	}
	if info.Brightness != 0 {
		t.Errorf("Brightness = %d, want 0", info.Brightness)
	}
	if info.Power != 0 {
		t.Errorf("Power = %f, want 0", info.Power)
	}
}

func TestGen1LightStatusToLight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		id             int
		status         *gen1comp.LightStatus
		wantOn         bool
		wantBrightness int
	}{
		{
			name: "light on full brightness",
			id:   0,
			status: &gen1comp.LightStatus{
				IsOn:       true,
				Brightness: 100,
			},
			wantOn:         true,
			wantBrightness: 100,
		},
		{
			name: "light on dimmed",
			id:   1,
			status: &gen1comp.LightStatus{
				IsOn:       true,
				Brightness: 50,
			},
			wantOn:         true,
			wantBrightness: 50,
		},
		{
			name: "light off",
			id:   2,
			status: &gen1comp.LightStatus{
				IsOn:       false,
				Brightness: 0,
			},
			wantOn:         false,
			wantBrightness: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := gen1LightStatusToLight(tt.id, tt.status)

			if result.ID != tt.id {
				t.Errorf("ID = %d, want %d", result.ID, tt.id)
			}
			if result.Output != tt.wantOn {
				t.Errorf("Output = %v, want %v", result.Output, tt.wantOn)
			}
			if result.Brightness == nil {
				t.Fatal("Brightness is nil")
			}
			if *result.Brightness != tt.wantBrightness {
				t.Errorf("Brightness = %d, want %d", *result.Brightness, tt.wantBrightness)
			}
		})
	}
}
