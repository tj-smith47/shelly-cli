package shelly

import (
	"testing"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestRGBWInfo_Fields(t *testing.T) {
	t.Parallel()

	info := RGBWInfo{
		ID:         0,
		Name:       "Living Room RGBW",
		Output:     true,
		Brightness: 75,
		Red:        255,
		Green:      128,
		Blue:       64,
		White:      50,
		Power:      10.5,
	}

	if info.ID != 0 {
		t.Errorf("expected ID 0, got %d", info.ID)
	}
	if info.Name != "Living Room RGBW" {
		t.Errorf("expected Name 'Living Room RGBW', got %q", info.Name)
	}
	if !info.Output {
		t.Error("expected Output to be true")
	}
	if info.Brightness != 75 {
		t.Errorf("expected Brightness 75, got %d", info.Brightness)
	}
	if info.Red != 255 {
		t.Errorf("expected Red 255, got %d", info.Red)
	}
	if info.Green != 128 {
		t.Errorf("expected Green 128, got %d", info.Green)
	}
	if info.Blue != 64 {
		t.Errorf("expected Blue 64, got %d", info.Blue)
	}
	if info.White != 50 {
		t.Errorf("expected White 50, got %d", info.White)
	}
	if info.Power != 10.5 {
		t.Errorf("expected Power 10.5, got %f", info.Power)
	}
}

func TestRGBWSetParams_Fields(t *testing.T) {
	t.Parallel()

	red := 200
	green := 100
	blue := 50
	white := 25
	brightness := 80
	on := true

	params := RGBWSetParams{
		Red:        &red,
		Green:      &green,
		Blue:       &blue,
		White:      &white,
		Brightness: &brightness,
		On:         &on,
	}

	if params.Red == nil || *params.Red != 200 {
		t.Error("expected Red to be 200")
	}
	if params.Green == nil || *params.Green != 100 {
		t.Error("expected Green to be 100")
	}
	if params.Blue == nil || *params.Blue != 50 {
		t.Error("expected Blue to be 50")
	}
	if params.White == nil || *params.White != 25 {
		t.Error("expected White to be 25")
	}
	if params.Brightness == nil || *params.Brightness != 80 {
		t.Error("expected Brightness to be 80")
	}
	if params.On == nil || *params.On != true {
		t.Error("expected On to be true")
	}
}

//nolint:gocyclo // table-driven test with multiple validation checks
func TestBuildRGBWSetParams(t *testing.T) {
	t.Parallel()

	t.Run("all values set", func(t *testing.T) {
		t.Parallel()

		params := BuildRGBWSetParams(255, 128, 64, 50, 75, true)

		if params.Red == nil || *params.Red != 255 {
			t.Error("expected Red 255")
		}
		if params.Green == nil || *params.Green != 128 {
			t.Error("expected Green 128")
		}
		if params.Blue == nil || *params.Blue != 64 {
			t.Error("expected Blue 64")
		}
		if params.White == nil || *params.White != 50 {
			t.Error("expected White 50")
		}
		if params.Brightness == nil || *params.Brightness != 75 {
			t.Error("expected Brightness 75")
		}
		if params.On == nil || *params.On != true {
			t.Error("expected On true")
		}
	})

	t.Run("negative values not set", func(t *testing.T) {
		t.Parallel()

		params := BuildRGBWSetParams(-1, -1, -1, -1, -1, false)

		if params.Red != nil {
			t.Error("expected Red to be nil for -1")
		}
		if params.Green != nil {
			t.Error("expected Green to be nil for -1")
		}
		if params.Blue != nil {
			t.Error("expected Blue to be nil for -1")
		}
		if params.White != nil {
			t.Error("expected White to be nil for -1")
		}
		if params.Brightness != nil {
			t.Error("expected Brightness to be nil for -1")
		}
		if params.On != nil {
			t.Error("expected On to be nil for false")
		}
	})

	t.Run("boundary values", func(t *testing.T) {
		t.Parallel()

		// Test valid boundary values
		params := BuildRGBWSetParams(0, 0, 0, 0, 0, false)

		if params.Red == nil || *params.Red != 0 {
			t.Error("expected Red 0")
		}
		if params.Brightness == nil || *params.Brightness != 0 {
			t.Error("expected Brightness 0")
		}
	})

	t.Run("out of range values", func(t *testing.T) {
		t.Parallel()

		// Values > 255 for colors should not be set
		params := BuildRGBWSetParams(256, 300, 500, 256, 101, false)

		if params.Red != nil {
			t.Error("expected Red to be nil for 256")
		}
		if params.Green != nil {
			t.Error("expected Green to be nil for 300")
		}
		if params.Blue != nil {
			t.Error("expected Blue to be nil for 500")
		}
		if params.White != nil {
			t.Error("expected White to be nil for 256")
		}
		if params.Brightness != nil {
			t.Error("expected Brightness to be nil for 101")
		}
	})

	t.Run("max valid values", func(t *testing.T) {
		t.Parallel()

		params := BuildRGBWSetParams(255, 255, 255, 255, 100, true)

		if params.Red == nil || *params.Red != 255 {
			t.Error("expected Red 255")
		}
		if params.Green == nil || *params.Green != 255 {
			t.Error("expected Green 255")
		}
		if params.Blue == nil || *params.Blue != 255 {
			t.Error("expected Blue 255")
		}
		if params.White == nil || *params.White != 255 {
			t.Error("expected White 255")
		}
		if params.Brightness == nil || *params.Brightness != 100 {
			t.Error("expected Brightness 100")
		}
	})
}

func TestGen1ColorStatusToRGBW(t *testing.T) {
	t.Parallel()

	t.Run("converts full status", func(t *testing.T) {
		t.Parallel()

		status := &gen1comp.ColorStatus{
			IsOn:  true,
			Red:   255,
			Green: 128,
			Blue:  64,
			White: 50,
			Gain:  75,
		}

		result := gen1ColorStatusToRGBW(0, status)

		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.ID != 0 {
			t.Errorf("expected ID 0, got %d", result.ID)
		}
		if !result.Output {
			t.Error("expected Output to be true")
		}
		if result.Brightness == nil || *result.Brightness != 75 {
			t.Error("expected Brightness 75")
		}
		if result.White == nil || *result.White != 50 {
			t.Error("expected White 50")
		}
		if result.RGB == nil {
			t.Fatal("expected RGB to be set")
		}
		if result.RGB.Red != 255 {
			t.Errorf("expected Red 255, got %d", result.RGB.Red)
		}
		if result.RGB.Green != 128 {
			t.Errorf("expected Green 128, got %d", result.RGB.Green)
		}
		if result.RGB.Blue != 64 {
			t.Errorf("expected Blue 64, got %d", result.RGB.Blue)
		}
	})

	t.Run("converts off status", func(t *testing.T) {
		t.Parallel()

		status := &gen1comp.ColorStatus{
			IsOn:  false,
			Red:   0,
			Green: 0,
			Blue:  0,
			White: 0,
			Gain:  0,
		}

		result := gen1ColorStatusToRGBW(1, status)

		if result.Output {
			t.Error("expected Output to be false")
		}
		if result.ID != 1 {
			t.Errorf("expected ID 1, got %d", result.ID)
		}
	})
}

func TestRGBWStatus_Model(t *testing.T) {
	t.Parallel()

	// Test model.RGBWStatus fields
	brightness := 80
	white := 50
	power := 5.5

	status := model.RGBWStatus{
		ID:         0,
		Output:     true,
		Brightness: &brightness,
		White:      &white,
		RGB: &model.RGBColor{
			Red:   200,
			Green: 100,
			Blue:  50,
		},
		Power: &power,
	}

	if status.ID != 0 {
		t.Errorf("expected ID 0, got %d", status.ID)
	}
	if !status.Output {
		t.Error("expected Output true")
	}
	if status.Brightness == nil || *status.Brightness != 80 {
		t.Error("expected Brightness 80")
	}
	if status.RGB == nil {
		t.Fatal("expected RGB to be set")
	}
	if status.RGB.Red != 200 {
		t.Errorf("expected Red 200, got %d", status.RGB.Red)
	}
	if status.Power == nil || *status.Power != 5.5 {
		t.Error("expected Power 5.5")
	}
}
