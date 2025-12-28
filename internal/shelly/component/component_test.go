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
