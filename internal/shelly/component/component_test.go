// Package component provides component-level operations for Shelly devices.
package component

import (
	"context"
	"errors"
	"testing"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// mockConnectionProvider is a test double for ConnectionProvider.
type mockConnectionProvider struct {
	withConnectionFn     func(ctx context.Context, identifier string, fn func(*client.Client) error) error
	withGen1ConnectionFn func(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error
	withDeviceFn         func(ctx context.Context, identifier string, fn func(DeviceClient) error) error
	isGen1DeviceFn       func(ctx context.Context, identifier string) (bool, model.Device, error)
}

func (m *mockConnectionProvider) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionProvider) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	if m.withGen1ConnectionFn != nil {
		return m.withGen1ConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionProvider) WithDevice(ctx context.Context, identifier string, fn func(DeviceClient) error) error {
	if m.withDeviceFn != nil {
		return m.withDeviceFn(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionProvider) IsGen1Device(ctx context.Context, identifier string) (bool, model.Device, error) {
	if m.isGen1DeviceFn != nil {
		return m.isGen1DeviceFn(ctx, identifier)
	}
	return false, model.Device{}, nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	svc := New(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.parent != nil {
		t.Error("expected parent to be nil")
	}
}

func TestNewWithProvider(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := New(provider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.parent != provider {
		t.Error("expected parent to be set")
	}
}

func TestSwitchInfo_Fields(t *testing.T) {
	t.Parallel()

	info := SwitchInfo{
		ID:     0,
		Name:   "Living Room Switch",
		Output: true,
		Power:  100.5,
	}

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "Living Room Switch" {
		t.Errorf("got Name=%q, want %q", info.Name, "Living Room Switch")
	}
	if !info.Output {
		t.Error("expected Output to be true")
	}
	if info.Power != 100.5 {
		t.Errorf("got Power=%f, want 100.5", info.Power)
	}
}

func TestSwitchInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	var info SwitchInfo

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("got Name=%q, want empty", info.Name)
	}
	if info.Output {
		t.Error("expected Output to be false")
	}
	if info.Power != 0 {
		t.Errorf("got Power=%f, want 0", info.Power)
	}
}

func TestLightInfo_Fields(t *testing.T) {
	t.Parallel()

	info := LightInfo{
		ID:         0,
		Name:       "Bedroom Light",
		Output:     true,
		Brightness: 75,
		Power:      50.5,
	}

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "Bedroom Light" {
		t.Errorf("got Name=%q, want %q", info.Name, "Bedroom Light")
	}
	if !info.Output {
		t.Error("expected Output to be true")
	}
	if info.Brightness != 75 {
		t.Errorf("got Brightness=%d, want 75", info.Brightness)
	}
	if info.Power != 50.5 {
		t.Errorf("got Power=%f, want 50.5", info.Power)
	}
}

func TestLightInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	var info LightInfo

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("got Name=%q, want empty", info.Name)
	}
	if info.Output {
		t.Error("expected Output to be false")
	}
	if info.Brightness != 0 {
		t.Errorf("got Brightness=%d, want 0", info.Brightness)
	}
}

func TestRGBInfo_Fields(t *testing.T) {
	t.Parallel()

	info := RGBInfo{
		ID:         0,
		Name:       "RGB Strip",
		Output:     true,
		Brightness: 100,
		Red:        255,
		Green:      128,
		Blue:       64,
		Power:      25.5,
	}

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "RGB Strip" {
		t.Errorf("got Name=%q, want %q", info.Name, "RGB Strip")
	}
	if !info.Output {
		t.Error("expected Output to be true")
	}
	if info.Brightness != 100 {
		t.Errorf("got Brightness=%d, want 100", info.Brightness)
	}
	if info.Red != 255 {
		t.Errorf("got Red=%d, want 255", info.Red)
	}
	if info.Green != 128 {
		t.Errorf("got Green=%d, want 128", info.Green)
	}
	if info.Blue != 64 {
		t.Errorf("got Blue=%d, want 64", info.Blue)
	}
	if info.Power != 25.5 {
		t.Errorf("got Power=%f, want 25.5", info.Power)
	}
}

func TestRGBInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	var info RGBInfo

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("got Name=%q, want empty", info.Name)
	}
	if info.Output {
		t.Error("expected Output to be false")
	}
	if info.Red != 0 || info.Green != 0 || info.Blue != 0 {
		t.Error("expected RGB values to be 0")
	}
}

func TestCoverInfo_Fields(t *testing.T) {
	t.Parallel()

	info := CoverInfo{
		ID:       0,
		Name:     "Window Blind",
		State:    "stopped",
		Position: 50,
		Power:    75.5,
	}

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "Window Blind" {
		t.Errorf("got Name=%q, want %q", info.Name, "Window Blind")
	}
	if info.State != "stopped" {
		t.Errorf("got State=%q, want %q", info.State, "stopped")
	}
	if info.Position != 50 {
		t.Errorf("got Position=%d, want 50", info.Position)
	}
	if info.Power != 75.5 {
		t.Errorf("got Power=%f, want 75.5", info.Power)
	}
}

func TestCoverInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	var info CoverInfo

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("got Name=%q, want empty", info.Name)
	}
	if info.State != "" {
		t.Errorf("got State=%q, want empty", info.State)
	}
	if info.Position != 0 {
		t.Errorf("got Position=%d, want 0", info.Position)
	}
}

func TestInputInfo_Fields(t *testing.T) {
	t.Parallel()

	info := InputInfo{
		ID:    0,
		Name:  "Button Input",
		State: true,
		Type:  "button",
	}

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "Button Input" {
		t.Errorf("got Name=%q, want %q", info.Name, "Button Input")
	}
	if !info.State {
		t.Error("expected State to be true")
	}
	if info.Type != "button" {
		t.Errorf("got Type=%q, want %q", info.Type, "button")
	}
}

func TestInputInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	var info InputInfo

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("got Name=%q, want empty", info.Name)
	}
	if info.State {
		t.Error("expected State to be false")
	}
	if info.Type != "" {
		t.Errorf("got Type=%q, want empty", info.Type)
	}
}

func TestRGBSetParams_Fields(t *testing.T) {
	t.Parallel()

	red := 255
	green := 100
	blue := 50
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
		t.Errorf("got Red=%v, want 255", params.Red)
	}
	if params.Green == nil || *params.Green != 100 {
		t.Errorf("got Green=%v, want 100", params.Green)
	}
	if params.Blue == nil || *params.Blue != 50 {
		t.Errorf("got Blue=%v, want 50", params.Blue)
	}
	if params.Brightness == nil || *params.Brightness != 80 {
		t.Errorf("got Brightness=%v, want 80", params.Brightness)
	}
	if params.On == nil || !*params.On {
		t.Error("expected On to be true")
	}
}

func TestRGBSetParams_NilFields(t *testing.T) {
	t.Parallel()

	params := RGBSetParams{}

	if params.Red != nil {
		t.Error("expected Red to be nil")
	}
	if params.Green != nil {
		t.Error("expected Green to be nil")
	}
	if params.Blue != nil {
		t.Error("expected Blue to be nil")
	}
	if params.Brightness != nil {
		t.Error("expected Brightness to be nil")
	}
	if params.On != nil {
		t.Error("expected On to be nil")
	}
}

// checkIntParam is a helper to validate optional int parameters in tests.
func checkIntParam(t *testing.T, name string, param *int, expected int, shouldBeSet bool) {
	t.Helper()
	if shouldBeSet {
		if param == nil {
			t.Errorf("expected %s to be set", name)
			return
		}
		if *param != expected {
			t.Errorf("got %s=%d, want %d", name, *param, expected)
		}
		return
	}
	if param != nil {
		t.Errorf("expected %s to be nil, got %d", name, *param)
	}
}

// checkBoolParam is a helper to validate optional bool parameters in tests.
func checkBoolParam(t *testing.T, name string, param *bool, shouldBeSet bool) {
	t.Helper()
	if shouldBeSet {
		if param == nil {
			t.Errorf("expected %s to be set", name)
			return
		}
		if !*param {
			t.Errorf("expected %s to be true", name)
		}
		return
	}
	if param != nil {
		t.Errorf("expected %s to be nil", name)
	}
}

// checkFloat64Param is a helper to validate optional float64 parameters in tests.
func checkFloat64Param(t *testing.T, name string, param *float64, expected float64, shouldBeSet bool) {
	t.Helper()
	if shouldBeSet {
		if param == nil {
			t.Errorf("expected %s to be set", name)
			return
		}
		if *param != expected {
			t.Errorf("got %s=%f, want %f", name, *param, expected)
		}
		return
	}
	if param != nil {
		t.Errorf("expected %s to be nil, got %f", name, *param)
	}
}

func TestBuildRGBSetParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		red            int
		green          int
		blue           int
		brightness     int
		on             bool
		wantRed        bool
		wantGreen      bool
		wantBlue       bool
		wantBrightness bool
		wantOn         bool
	}{
		{
			name:           "full on with max values",
			red:            255,
			green:          255,
			blue:           255,
			brightness:     100,
			on:             true,
			wantRed:        true,
			wantGreen:      true,
			wantBlue:       true,
			wantBrightness: true,
			wantOn:         true,
		},
		{
			name:           "not set with -1",
			red:            -1,
			green:          -1,
			blue:           -1,
			brightness:     -1,
			on:             false,
			wantRed:        false,
			wantGreen:      false,
			wantBlue:       false,
			wantBrightness: false,
			wantOn:         false,
		},
		{
			name:           "red only",
			red:            255,
			green:          -1,
			blue:           -1,
			brightness:     -1,
			on:             true,
			wantRed:        true,
			wantGreen:      false,
			wantBlue:       false,
			wantBrightness: false,
			wantOn:         true,
		},
		{
			name:           "green only",
			red:            -1,
			green:          255,
			blue:           -1,
			brightness:     -1,
			on:             true,
			wantRed:        false,
			wantGreen:      true,
			wantBlue:       false,
			wantBrightness: false,
			wantOn:         true,
		},
		{
			name:           "blue only",
			red:            -1,
			green:          -1,
			blue:           255,
			brightness:     -1,
			on:             true,
			wantRed:        false,
			wantGreen:      false,
			wantBlue:       true,
			wantBrightness: false,
			wantOn:         true,
		},
		{
			name:           "brightness only",
			red:            -1,
			green:          -1,
			blue:           -1,
			brightness:     50,
			on:             false,
			wantRed:        false,
			wantGreen:      false,
			wantBlue:       false,
			wantBrightness: true,
			wantOn:         false,
		},
		{
			name:           "all zero values",
			red:            0,
			green:          0,
			blue:           0,
			brightness:     0,
			on:             false,
			wantRed:        true,
			wantGreen:      true,
			wantBlue:       true,
			wantBrightness: true,
			wantOn:         false,
		},
		{
			name:           "out of range high - color",
			red:            256,
			green:          256,
			blue:           256,
			brightness:     101,
			on:             true,
			wantRed:        false,
			wantGreen:      false,
			wantBlue:       false,
			wantBrightness: false,
			wantOn:         true,
		},
		{
			name:           "boundary - max valid",
			red:            255,
			green:          255,
			blue:           255,
			brightness:     100,
			on:             true,
			wantRed:        true,
			wantGreen:      true,
			wantBlue:       true,
			wantBrightness: true,
			wantOn:         true,
		},
		{
			name:           "mixed valid and invalid",
			red:            128,
			green:          -1,
			blue:           300,
			brightness:     50,
			on:             true,
			wantRed:        true,
			wantGreen:      false,
			wantBlue:       false,
			wantBrightness: true,
			wantOn:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := BuildRGBSetParams(tt.red, tt.green, tt.blue, tt.brightness, tt.on)

			// Check Red
			checkIntParam(t, "Red", params.Red, tt.red, tt.wantRed)

			// Check Green
			checkIntParam(t, "Green", params.Green, tt.green, tt.wantGreen)

			// Check Blue
			checkIntParam(t, "Blue", params.Blue, tt.blue, tt.wantBlue)

			// Check Brightness
			checkIntParam(t, "Brightness", params.Brightness, tt.brightness, tt.wantBrightness)

			// Check On
			checkBoolParam(t, "On", params.On, tt.wantOn)
		})
	}
}

func TestServiceWithNilParent(t *testing.T) {
	t.Parallel()

	// Service should be created even with nil parent
	svc := New(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestGen1RelayStatusToSwitch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		id     int
		status *gen1comp.RelayStatus
		want   *model.SwitchStatus
	}{
		{
			name: "relay on",
			id:   0,
			status: &gen1comp.RelayStatus{
				IsOn:      true,
				Source:    "input",
				Overpower: false,
			},
			want: &model.SwitchStatus{
				ID:        0,
				Output:    true,
				Source:    "input",
				Overpower: false,
			},
		},
		{
			name: "relay off",
			id:   1,
			status: &gen1comp.RelayStatus{
				IsOn:      false,
				Source:    "button",
				Overpower: false,
			},
			want: &model.SwitchStatus{
				ID:        1,
				Output:    false,
				Source:    "button",
				Overpower: false,
			},
		},
		{
			name: "overpower detected",
			id:   0,
			status: &gen1comp.RelayStatus{
				IsOn:      false,
				Source:    "overpower",
				Overpower: true,
			},
			want: &model.SwitchStatus{
				ID:        0,
				Output:    false,
				Source:    "overpower",
				Overpower: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := gen1RelayStatusToSwitch(tt.id, tt.status)

			if got.ID != tt.want.ID {
				t.Errorf("got ID=%d, want %d", got.ID, tt.want.ID)
			}
			if got.Output != tt.want.Output {
				t.Errorf("got Output=%v, want %v", got.Output, tt.want.Output)
			}
			if got.Source != tt.want.Source {
				t.Errorf("got Source=%q, want %q", got.Source, tt.want.Source)
			}
			if got.Overpower != tt.want.Overpower {
				t.Errorf("got Overpower=%v, want %v", got.Overpower, tt.want.Overpower)
			}
		})
	}
}

func TestGen1LightStatusToLight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		id     int
		status *gen1comp.LightStatus
		want   *model.LightStatus
	}{
		{
			name: "light on full brightness",
			id:   0,
			status: &gen1comp.LightStatus{
				IsOn:       true,
				Brightness: 100,
			},
			want: &model.LightStatus{
				ID:     0,
				Output: true,
			},
		},
		{
			name: "light off",
			id:   1,
			status: &gen1comp.LightStatus{
				IsOn:       false,
				Brightness: 0,
			},
			want: &model.LightStatus{
				ID:     1,
				Output: false,
			},
		},
		{
			name: "dimmed light",
			id:   0,
			status: &gen1comp.LightStatus{
				IsOn:       true,
				Brightness: 50,
			},
			want: &model.LightStatus{
				ID:     0,
				Output: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := gen1LightStatusToLight(tt.id, tt.status)

			if got.ID != tt.want.ID {
				t.Errorf("got ID=%d, want %d", got.ID, tt.want.ID)
			}
			if got.Output != tt.want.Output {
				t.Errorf("got Output=%v, want %v", got.Output, tt.want.Output)
			}
			if got.Brightness == nil {
				t.Error("expected Brightness to be set")
			} else if *got.Brightness != tt.status.Brightness {
				t.Errorf("got Brightness=%d, want %d", *got.Brightness, tt.status.Brightness)
			}
		})
	}
}

func TestGen1RollerStatusToCover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		id     int
		status *gen1comp.RollerStatus
	}{
		{
			name: "stopped at position 50",
			id:   0,
			status: &gen1comp.RollerStatus{
				State:       "stopped",
				CurrentPos:  50,
				Power:       0,
				IsValid:     true,
				Calibrating: false,
			},
		},
		{
			name: "opening",
			id:   0,
			status: &gen1comp.RollerStatus{
				State:       "open",
				CurrentPos:  75,
				Power:       100.5,
				IsValid:     true,
				Calibrating: false,
			},
		},
		{
			name: "closing",
			id:   1,
			status: &gen1comp.RollerStatus{
				State:       "close",
				CurrentPos:  25,
				Power:       95.3,
				IsValid:     true,
				Calibrating: false,
			},
		},
		{
			name: "calibrating",
			id:   0,
			status: &gen1comp.RollerStatus{
				State:       "stopped",
				CurrentPos:  0,
				Power:       0,
				IsValid:     false,
				Calibrating: true,
			},
		},
		{
			name: "invalid position",
			id:   0,
			status: &gen1comp.RollerStatus{
				State:       "stopped",
				CurrentPos:  -1,
				Power:       0,
				IsValid:     false,
				Calibrating: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := gen1RollerStatusToCover(tt.id, tt.status)

			if got.ID != tt.id {
				t.Errorf("got ID=%d, want %d", got.ID, tt.id)
			}
			if got.State != tt.status.State {
				t.Errorf("got State=%q, want %q", got.State, tt.status.State)
			}
			if got.Calibrating != tt.status.Calibrating {
				t.Errorf("got Calibrating=%v, want %v", got.Calibrating, tt.status.Calibrating)
			}

			// Position should only be set if valid and non-negative
			positionShouldBeSet := tt.status.CurrentPos >= 0 && tt.status.IsValid
			checkIntParam(t, "CurrentPosition", got.CurrentPosition, tt.status.CurrentPos, positionShouldBeSet)

			// Power should only be set if > 0
			powerShouldBeSet := tt.status.Power > 0
			checkFloat64Param(t, "Power", got.Power, tt.status.Power, powerShouldBeSet)
		})
	}
}

func TestGen1ColorStatusToRGB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		id     int
		status *gen1comp.ColorStatus
	}{
		{
			name: "full red",
			id:   0,
			status: &gen1comp.ColorStatus{
				IsOn:  true,
				Gain:  100,
				Red:   255,
				Green: 0,
				Blue:  0,
			},
		},
		{
			name: "off",
			id:   1,
			status: &gen1comp.ColorStatus{
				IsOn:  false,
				Gain:  0,
				Red:   0,
				Green: 0,
				Blue:  0,
			},
		},
		{
			name: "mixed color",
			id:   0,
			status: &gen1comp.ColorStatus{
				IsOn:  true,
				Gain:  75,
				Red:   128,
				Green: 64,
				Blue:  255,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := gen1ColorStatusToRGB(tt.id, tt.status)

			if got.ID != tt.id {
				t.Errorf("got ID=%d, want %d", got.ID, tt.id)
			}
			if got.Output != tt.status.IsOn {
				t.Errorf("got Output=%v, want %v", got.Output, tt.status.IsOn)
			}
			if got.Brightness == nil {
				t.Error("expected Brightness to be set")
			} else if *got.Brightness != tt.status.Gain {
				t.Errorf("got Brightness=%d, want %d", *got.Brightness, tt.status.Gain)
			}
			if got.RGB == nil {
				t.Fatal("expected RGB to be set")
			}
			if got.RGB.Red != tt.status.Red {
				t.Errorf("got RGB.Red=%d, want %d", got.RGB.Red, tt.status.Red)
			}
			if got.RGB.Green != tt.status.Green {
				t.Errorf("got RGB.Green=%d, want %d", got.RGB.Green, tt.status.Green)
			}
			if got.RGB.Blue != tt.status.Blue {
				t.Errorf("got RGB.Blue=%d, want %d", got.RGB.Blue, tt.status.Blue)
			}
		})
	}
}

func TestWithGenAwareAction_Gen1Device(t *testing.T) {
	t.Parallel()

	gen1Called := false
	gen2Called := false

	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			gen1Called = true
			return nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			gen2Called = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.withGenAwareAction(
		context.Background(),
		"test-device",
		func(_ *client.Gen1Client) error { return nil },
		func(_ *client.Client) error { return nil },
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !gen1Called {
		t.Error("expected Gen1 connection to be called")
	}
	if gen2Called {
		t.Error("expected Gen2 connection NOT to be called")
	}
}

func TestWithGenAwareAction_Gen2Device(t *testing.T) {
	t.Parallel()

	gen1Called := false
	gen2Called := false

	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			gen1Called = true
			return nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			gen2Called = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.withGenAwareAction(
		context.Background(),
		"test-device",
		func(_ *client.Gen1Client) error { return nil },
		func(_ *client.Client) error { return nil },
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if gen1Called {
		t.Error("expected Gen1 connection NOT to be called")
	}
	if !gen2Called {
		t.Error("expected Gen2 connection to be called")
	}
}

func TestWithGenAwareAction_IsGen1Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device not found")

	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{}, expectedErr
		},
	}

	svc := New(provider)
	err := svc.withGenAwareAction(
		context.Background(),
		"test-device",
		func(_ *client.Gen1Client) error { return nil },
		func(_ *client.Client) error { return nil },
	)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// ============== Switch Operation Tests ==============

func TestSwitchOn_Gen2(t *testing.T) {
	t.Parallel()

	switchOnCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			switchOnCalled = true
			// We don't have a real client, but the callback was invoked
			return nil
		},
	}

	svc := New(provider)
	err := svc.SwitchOn(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !switchOnCalled {
		t.Error("expected Gen2 switch on to be called")
	}
}

func TestSwitchOn_Gen1(t *testing.T) {
	t.Parallel()

	switchOnCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			switchOnCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.SwitchOn(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !switchOnCalled {
		t.Error("expected Gen1 switch on to be called")
	}
}

func TestSwitchOn_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{}, expectedErr
		},
	}

	svc := New(provider)
	err := svc.SwitchOn(context.Background(), "test-device", 0)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestSwitchOff_Gen2(t *testing.T) {
	t.Parallel()

	switchOffCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			switchOffCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.SwitchOff(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !switchOffCalled {
		t.Error("expected Gen2 switch off to be called")
	}
}

func TestSwitchOff_Gen1(t *testing.T) {
	t.Parallel()

	switchOffCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			switchOffCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.SwitchOff(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !switchOffCalled {
		t.Error("expected Gen1 switch off to be called")
	}
}

func TestSwitchToggle_WithDeviceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device error")
	provider := &mockConnectionProvider{
		withDeviceFn: func(_ context.Context, _ string, fn func(DeviceClient) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.SwitchToggle(context.Background(), "test-device", 0)

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestSwitchStatus_WithDeviceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device error")
	provider := &mockConnectionProvider{
		withDeviceFn: func(_ context.Context, _ string, fn func(DeviceClient) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.SwitchStatus(context.Background(), "test-device", 0)

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestSwitchList_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.SwitchList(context.Background(), "test-device")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// ============== Light Operation Tests ==============

func TestLightOn_Gen2(t *testing.T) {
	t.Parallel()

	lightOnCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			lightOnCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.LightOn(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !lightOnCalled {
		t.Error("expected Gen2 light on to be called")
	}
}

func TestLightOn_Gen1(t *testing.T) {
	t.Parallel()

	lightOnCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			lightOnCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.LightOn(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !lightOnCalled {
		t.Error("expected Gen1 light on to be called")
	}
}

func TestLightOff_Gen2(t *testing.T) {
	t.Parallel()

	lightOffCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			lightOffCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.LightOff(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !lightOffCalled {
		t.Error("expected Gen2 light off to be called")
	}
}

func TestLightOff_Gen1(t *testing.T) {
	t.Parallel()

	lightOffCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			lightOffCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.LightOff(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !lightOffCalled {
		t.Error("expected Gen1 light off to be called")
	}
}

func TestLightToggle_WithDeviceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device error")
	provider := &mockConnectionProvider{
		withDeviceFn: func(_ context.Context, _ string, fn func(DeviceClient) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.LightToggle(context.Background(), "test-device", 0)

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestLightBrightness_Gen2(t *testing.T) {
	t.Parallel()

	brightnessCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			brightnessCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.LightBrightness(context.Background(), "test-device", 0, 75)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !brightnessCalled {
		t.Error("expected Gen2 brightness to be called")
	}
}

func TestLightBrightness_Gen1(t *testing.T) {
	t.Parallel()

	brightnessCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			brightnessCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.LightBrightness(context.Background(), "test-device", 0, 75)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !brightnessCalled {
		t.Error("expected Gen1 brightness to be called")
	}
}

func TestLightStatus_WithDeviceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device error")
	provider := &mockConnectionProvider{
		withDeviceFn: func(_ context.Context, _ string, fn func(DeviceClient) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.LightStatus(context.Background(), "test-device", 0)

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestLightSet_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	brightness := 50
	on := true
	err := svc.LightSet(context.Background(), "test-device", 0, &brightness, &on)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestLightList_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.LightList(context.Background(), "test-device")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// ============== Cover Operation Tests ==============

func TestCoverOpen_Gen2(t *testing.T) {
	t.Parallel()

	coverOpenCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			coverOpenCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverOpen(context.Background(), "test-device", 0, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverOpenCalled {
		t.Error("expected Gen2 cover open to be called")
	}
}

func TestCoverOpen_Gen1(t *testing.T) {
	t.Parallel()

	coverOpenCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			coverOpenCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverOpen(context.Background(), "test-device", 0, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverOpenCalled {
		t.Error("expected Gen1 cover open to be called")
	}
}

func TestCoverOpen_Gen1WithDuration(t *testing.T) {
	t.Parallel()

	coverOpenCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			coverOpenCalled = true
			return nil
		},
	}

	svc := New(provider)
	duration := 10
	err := svc.CoverOpen(context.Background(), "test-device", 0, &duration)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverOpenCalled {
		t.Error("expected Gen1 cover open with duration to be called")
	}
}

func TestCoverClose_Gen2(t *testing.T) {
	t.Parallel()

	coverCloseCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			coverCloseCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverClose(context.Background(), "test-device", 0, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverCloseCalled {
		t.Error("expected Gen2 cover close to be called")
	}
}

func TestCoverClose_Gen1(t *testing.T) {
	t.Parallel()

	coverCloseCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			coverCloseCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverClose(context.Background(), "test-device", 0, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverCloseCalled {
		t.Error("expected Gen1 cover close to be called")
	}
}

func TestCoverClose_Gen1WithDuration(t *testing.T) {
	t.Parallel()

	coverCloseCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			coverCloseCalled = true
			return nil
		},
	}

	svc := New(provider)
	duration := 5
	err := svc.CoverClose(context.Background(), "test-device", 0, &duration)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverCloseCalled {
		t.Error("expected Gen1 cover close with duration to be called")
	}
}

func TestCoverStop_Gen2(t *testing.T) {
	t.Parallel()

	coverStopCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			coverStopCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverStop(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverStopCalled {
		t.Error("expected Gen2 cover stop to be called")
	}
}

func TestCoverStop_Gen1(t *testing.T) {
	t.Parallel()

	coverStopCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			coverStopCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverStop(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverStopCalled {
		t.Error("expected Gen1 cover stop to be called")
	}
}

func TestCoverPosition_Gen2(t *testing.T) {
	t.Parallel()

	coverPositionCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			coverPositionCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverPosition(context.Background(), "test-device", 0, 50)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverPositionCalled {
		t.Error("expected Gen2 cover position to be called")
	}
}

func TestCoverPosition_Gen1(t *testing.T) {
	t.Parallel()

	coverPositionCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			coverPositionCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverPosition(context.Background(), "test-device", 0, 50)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverPositionCalled {
		t.Error("expected Gen1 cover position to be called")
	}
}

func TestCoverStatus_WithDeviceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device error")
	provider := &mockConnectionProvider{
		withDeviceFn: func(_ context.Context, _ string, fn func(DeviceClient) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.CoverStatus(context.Background(), "test-device", 0)

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestCoverCalibrate_Gen2(t *testing.T) {
	t.Parallel()

	coverCalibrateCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			coverCalibrateCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverCalibrate(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverCalibrateCalled {
		t.Error("expected Gen2 cover calibrate to be called")
	}
}

func TestCoverCalibrate_Gen1(t *testing.T) {
	t.Parallel()

	coverCalibrateCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			coverCalibrateCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.CoverCalibrate(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !coverCalibrateCalled {
		t.Error("expected Gen1 cover calibrate to be called")
	}
}

func TestCoverList_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.CoverList(context.Background(), "test-device")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// ============== RGB Operation Tests ==============

func TestRGBOn_Gen2(t *testing.T) {
	t.Parallel()

	rgbOnCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			rgbOnCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBOn(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbOnCalled {
		t.Error("expected Gen2 RGB on to be called")
	}
}

func TestRGBOn_Gen1(t *testing.T) {
	t.Parallel()

	rgbOnCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			rgbOnCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBOn(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbOnCalled {
		t.Error("expected Gen1 RGB on to be called")
	}
}

func TestRGBOff_Gen2(t *testing.T) {
	t.Parallel()

	rgbOffCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			rgbOffCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBOff(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbOffCalled {
		t.Error("expected Gen2 RGB off to be called")
	}
}

func TestRGBOff_Gen1(t *testing.T) {
	t.Parallel()

	rgbOffCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			rgbOffCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBOff(context.Background(), "test-device", 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbOffCalled {
		t.Error("expected Gen1 RGB off to be called")
	}
}

func TestRGBToggle_WithDeviceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device error")
	provider := &mockConnectionProvider{
		withDeviceFn: func(_ context.Context, _ string, fn func(DeviceClient) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.RGBToggle(context.Background(), "test-device", 0)

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestRGBBrightness_Gen2(t *testing.T) {
	t.Parallel()

	rgbBrightnessCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			rgbBrightnessCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBBrightness(context.Background(), "test-device", 0, 75)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbBrightnessCalled {
		t.Error("expected Gen2 RGB brightness to be called")
	}
}

func TestRGBBrightness_Gen1(t *testing.T) {
	t.Parallel()

	rgbBrightnessCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			rgbBrightnessCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBBrightness(context.Background(), "test-device", 0, 75)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbBrightnessCalled {
		t.Error("expected Gen1 RGB brightness to be called")
	}
}

func TestRGBColor_Gen2(t *testing.T) {
	t.Parallel()

	rgbColorCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			rgbColorCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBColor(context.Background(), "test-device", 0, 255, 128, 64)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbColorCalled {
		t.Error("expected Gen2 RGB color to be called")
	}
}

func TestRGBColor_Gen1(t *testing.T) {
	t.Parallel()

	rgbColorCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			rgbColorCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBColor(context.Background(), "test-device", 0, 255, 128, 64)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbColorCalled {
		t.Error("expected Gen1 RGB color to be called")
	}
}

func TestRGBColorAndBrightness_Gen2(t *testing.T) {
	t.Parallel()

	rgbColorAndBrightnessCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return false, model.Device{Generation: 2}, nil
		},
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			rgbColorAndBrightnessCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBColorAndBrightness(context.Background(), "test-device", 0, 255, 128, 64, 80)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbColorAndBrightnessCalled {
		t.Error("expected Gen2 RGB color and brightness to be called")
	}
}

func TestRGBColorAndBrightness_Gen1(t *testing.T) {
	t.Parallel()

	rgbColorAndBrightnessCalled := false
	provider := &mockConnectionProvider{
		isGen1DeviceFn: func(_ context.Context, _ string) (bool, model.Device, error) {
			return true, model.Device{Generation: 1}, nil
		},
		withGen1ConnectionFn: func(_ context.Context, _ string, fn func(*client.Gen1Client) error) error {
			rgbColorAndBrightnessCalled = true
			return nil
		},
	}

	svc := New(provider)
	err := svc.RGBColorAndBrightness(context.Background(), "test-device", 0, 255, 128, 64, 80)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !rgbColorAndBrightnessCalled {
		t.Error("expected Gen1 RGB color and brightness to be called")
	}
}

func TestRGBStatus_WithDeviceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("device error")
	provider := &mockConnectionProvider{
		withDeviceFn: func(_ context.Context, _ string, fn func(DeviceClient) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.RGBStatus(context.Background(), "test-device", 0)

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestRGBSet_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	red := 255
	params := RGBSetParams{Red: &red}
	err := svc.RGBSet(context.Background(), "test-device", 0, params)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestRGBList_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.RGBList(context.Background(), "test-device")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// ============== Input Operation Tests ==============

func TestInputStatus_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.InputStatus(context.Background(), "test-device", 0)

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestInputTrigger_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	err := svc.InputTrigger(context.Background(), "test-device", 0, "single_push")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestInputList_WithConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(provider)
	result, err := svc.InputList(context.Background(), "test-device")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

// ============== Additional Edge Case Tests ==============

func TestRGBWInfo_Fields(t *testing.T) {
	t.Parallel()

	info := RGBInfo{
		ID:         0,
		Name:       "RGBW Strip",
		Output:     true,
		Brightness: 100,
		Red:        255,
		Green:      128,
		Blue:       64,
		Power:      25.5,
	}

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "RGBW Strip" {
		t.Errorf("got Name=%q, want %q", info.Name, "RGBW Strip")
	}
	if !info.Output {
		t.Error("expected Output to be true")
	}
	if info.Brightness != 100 {
		t.Errorf("got Brightness=%d, want 100", info.Brightness)
	}
}

func TestCoverInfo_WithMaxPosition(t *testing.T) {
	t.Parallel()

	info := CoverInfo{
		ID:       0,
		Name:     "Motorized Blind",
		State:    "open",
		Position: 100,
		Power:    150.0,
	}

	if info.Position != 100 {
		t.Errorf("got Position=%d, want 100", info.Position)
	}
	if info.Power != 150.0 {
		t.Errorf("got Power=%f, want 150.0", info.Power)
	}
}

func TestLightInfo_WithMaxBrightness(t *testing.T) {
	t.Parallel()

	info := LightInfo{
		ID:         0,
		Name:       "Ceiling Light",
		Output:     true,
		Brightness: 100,
		Power:      60.0,
	}

	if info.Brightness != 100 {
		t.Errorf("got Brightness=%d, want 100", info.Brightness)
	}
}

func TestBuildRGBSetParams_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		red        int
		green      int
		blue       int
		brightness int
		on         bool
		checkRed   bool
		checkGreen bool
		checkBlue  bool
		checkBri   bool
		checkOn    bool
	}{
		{
			name:       "exactly -1 for all",
			red:        -1,
			green:      -1,
			blue:       -1,
			brightness: -1,
			on:         false,
		},
		{
			name:       "exactly 0 for all colors",
			red:        0,
			green:      0,
			blue:       0,
			brightness: 0,
			on:         false,
			checkRed:   true,
			checkGreen: true,
			checkBlue:  true,
			checkBri:   true,
		},
		{
			name:       "max values for colors",
			red:        255,
			green:      255,
			blue:       255,
			brightness: 100,
			on:         true,
			checkRed:   true,
			checkGreen: true,
			checkBlue:  true,
			checkBri:   true,
			checkOn:    true,
		},
		{
			name:       "just above max colors",
			red:        256,
			green:      256,
			blue:       256,
			brightness: 101,
			on:         true,
			checkOn:    true,
		},
		{
			name:       "negative values other than -1",
			red:        -2,
			green:      -100,
			blue:       -50,
			brightness: -10,
			on:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params := BuildRGBSetParams(tt.red, tt.green, tt.blue, tt.brightness, tt.on)

			if tt.checkRed != (params.Red != nil) {
				t.Errorf("Red: expected set=%v, got set=%v", tt.checkRed, params.Red != nil)
			}
			if tt.checkGreen != (params.Green != nil) {
				t.Errorf("Green: expected set=%v, got set=%v", tt.checkGreen, params.Green != nil)
			}
			if tt.checkBlue != (params.Blue != nil) {
				t.Errorf("Blue: expected set=%v, got set=%v", tt.checkBlue, params.Blue != nil)
			}
			if tt.checkBri != (params.Brightness != nil) {
				t.Errorf("Brightness: expected set=%v, got set=%v", tt.checkBri, params.Brightness != nil)
			}
			if tt.checkOn != (params.On != nil) {
				t.Errorf("On: expected set=%v, got set=%v", tt.checkOn, params.On != nil)
			}
		})
	}
}
