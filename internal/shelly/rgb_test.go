package shelly

import (
	"testing"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"
)

func TestRGBInfo_Fields(t *testing.T) {
	t.Parallel()

	info := RGBInfo{
		ID:         0,
		Name:       "Living Room Light",
		Output:     true,
		Brightness: 75,
		Red:        255,
		Green:      100,
		Blue:       50,
		Power:      10.5,
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
	if info.Red != 255 {
		t.Errorf("Red = %d, want 255", info.Red)
	}
	if info.Green != 100 {
		t.Errorf("Green = %d, want 100", info.Green)
	}
	if info.Blue != 50 {
		t.Errorf("Blue = %d, want 50", info.Blue)
	}
	if info.Power != 10.5 {
		t.Errorf("Power = %f, want 10.5", info.Power)
	}
}

func TestRGBInfo_ZeroValues(t *testing.T) {
	t.Parallel()

	var info RGBInfo

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
	if info.Red != 0 || info.Green != 0 || info.Blue != 0 {
		t.Error("RGB should be 0, 0, 0")
	}
}

func TestRGBSetParams_Fields(t *testing.T) {
	t.Parallel()

	red := 255
	green := 128
	blue := 64
	brightness := 80
	on := true

	params := RGBSetParams{
		Red:        &red,
		Green:      &green,
		Blue:       &blue,
		Brightness: &brightness,
		On:         &on,
	}

	if params.Red == nil || *params.Red != 255 {
		t.Errorf("Red = %v, want 255", params.Red)
	}
	if params.Green == nil || *params.Green != 128 {
		t.Errorf("Green = %v, want 128", params.Green)
	}
	if params.Blue == nil || *params.Blue != 64 {
		t.Errorf("Blue = %v, want 64", params.Blue)
	}
	if params.Brightness == nil || *params.Brightness != 80 {
		t.Errorf("Brightness = %v, want 80", params.Brightness)
	}
	if params.On == nil || !*params.On {
		t.Errorf("On = %v, want true", params.On)
	}
}

//nolint:gocyclo // table-driven test with multiple validation checks
func TestBuildRGBSetParams(t *testing.T) {
	t.Parallel()

	t.Run("all valid values", func(t *testing.T) {
		t.Parallel()
		params := BuildRGBSetParams(255, 128, 64, 80, true)

		if params.Red == nil || *params.Red != 255 {
			t.Errorf("Red = %v, want 255", params.Red)
		}
		if params.Green == nil || *params.Green != 128 {
			t.Errorf("Green = %v, want 128", params.Green)
		}
		if params.Blue == nil || *params.Blue != 64 {
			t.Errorf("Blue = %v, want 64", params.Blue)
		}
		if params.Brightness == nil || *params.Brightness != 80 {
			t.Errorf("Brightness = %v, want 80", params.Brightness)
		}
		if params.On == nil || !*params.On {
			t.Errorf("On = %v, want true", params.On)
		}
	})

	t.Run("invalid color values", func(t *testing.T) {
		t.Parallel()
		params := BuildRGBSetParams(-1, -1, -1, -1, false)

		if params.Red != nil {
			t.Errorf("Red = %v, want nil (invalid)", params.Red)
		}
		if params.Green != nil {
			t.Errorf("Green = %v, want nil (invalid)", params.Green)
		}
		if params.Blue != nil {
			t.Errorf("Blue = %v, want nil (invalid)", params.Blue)
		}
		if params.Brightness != nil {
			t.Errorf("Brightness = %v, want nil (invalid)", params.Brightness)
		}
		if params.On != nil {
			t.Errorf("On = %v, want nil (on=false)", params.On)
		}
	})

	t.Run("over max color values", func(t *testing.T) {
		t.Parallel()
		params := BuildRGBSetParams(256, 300, 1000, 101, false)

		if params.Red != nil {
			t.Errorf("Red = %v, want nil (over 255)", params.Red)
		}
		if params.Green != nil {
			t.Errorf("Green = %v, want nil (over 255)", params.Green)
		}
		if params.Blue != nil {
			t.Errorf("Blue = %v, want nil (over 255)", params.Blue)
		}
		if params.Brightness != nil {
			t.Errorf("Brightness = %v, want nil (over 100)", params.Brightness)
		}
	})

	t.Run("zero values are valid", func(t *testing.T) {
		t.Parallel()
		params := BuildRGBSetParams(0, 0, 0, 0, false)

		if params.Red == nil || *params.Red != 0 {
			t.Errorf("Red = %v, want 0", params.Red)
		}
		if params.Green == nil || *params.Green != 0 {
			t.Errorf("Green = %v, want 0", params.Green)
		}
		if params.Blue == nil || *params.Blue != 0 {
			t.Errorf("Blue = %v, want 0", params.Blue)
		}
		if params.Brightness == nil || *params.Brightness != 0 {
			t.Errorf("Brightness = %v, want 0", params.Brightness)
		}
	})

	t.Run("max values are valid", func(t *testing.T) {
		t.Parallel()
		params := BuildRGBSetParams(255, 255, 255, 100, true)

		if params.Red == nil || *params.Red != 255 {
			t.Errorf("Red = %v, want 255", params.Red)
		}
		if params.Brightness == nil || *params.Brightness != 100 {
			t.Errorf("Brightness = %v, want 100", params.Brightness)
		}
	})
}

func TestGen1ColorStatusToRGB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		id             int
		status         *gen1comp.ColorStatus
		wantOn         bool
		wantBrightness int
		wantRed        int
		wantGreen      int
		wantBlue       int
	}{
		{
			name: "on full brightness red",
			id:   0,
			status: &gen1comp.ColorStatus{
				IsOn:  true,
				Gain:  100,
				Red:   255,
				Green: 0,
				Blue:  0,
			},
			wantOn:         true,
			wantBrightness: 100,
			wantRed:        255,
			wantGreen:      0,
			wantBlue:       0,
		},
		{
			name: "dimmed white",
			id:   1,
			status: &gen1comp.ColorStatus{
				IsOn:  true,
				Gain:  50,
				Red:   255,
				Green: 255,
				Blue:  255,
			},
			wantOn:         true,
			wantBrightness: 50,
			wantRed:        255,
			wantGreen:      255,
			wantBlue:       255,
		},
		{
			name: "off",
			id:   0,
			status: &gen1comp.ColorStatus{
				IsOn:  false,
				Gain:  0,
				Red:   0,
				Green: 0,
				Blue:  0,
			},
			wantOn:         false,
			wantBrightness: 0,
			wantRed:        0,
			wantGreen:      0,
			wantBlue:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := gen1ColorStatusToRGB(tt.id, tt.status)

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
			if result.RGB == nil {
				t.Fatal("RGB is nil")
			}
			if result.RGB.Red != tt.wantRed || result.RGB.Green != tt.wantGreen || result.RGB.Blue != tt.wantBlue {
				t.Errorf("RGB = (%d, %d, %d), want (%d, %d, %d)",
					result.RGB.Red, result.RGB.Green, result.RGB.Blue,
					tt.wantRed, tt.wantGreen, tt.wantBlue)
			}
		})
	}
}
