package model

import "testing"

// Test data constants.
const (
	testKitchenLight = "Kitchen Light"
)

func TestComponentType_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		ctype ComponentType
		want  string
	}{
		{"switch", ComponentSwitch, "switch"},
		{"cover", ComponentCover, "cover"},
		{"light", ComponentLight, "light"},
		{"rgb", ComponentRGB, "rgb"},
		{"input", ComponentInput, "input"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.ctype) != tt.want {
				t.Errorf("ComponentType = %q, want %q", tt.ctype, tt.want)
			}
		})
	}
}

func TestComponent_Fields(t *testing.T) {
	t.Parallel()

	comp := Component{
		Type: ComponentSwitch,
		ID:   0,
		Key:  "switch:0",
	}

	if comp.Type != ComponentSwitch {
		t.Errorf("Type = %q, want %q", comp.Type, ComponentSwitch)
	}
	if comp.ID != 0 {
		t.Errorf("ID = %d, want %d", comp.ID, 0)
	}
	if comp.Key != "switch:0" {
		t.Errorf("Key = %q, want %q", comp.Key, "switch:0")
	}
}

func TestSwitchStatus_Fields(t *testing.T) {
	t.Parallel()

	power := 25.5
	voltage := 230.0
	current := 0.11
	status := SwitchStatus{
		ID:          0,
		Output:      true,
		Source:      "button",
		Power:       &power,
		Voltage:     &voltage,
		Current:     &current,
		Overtemp:    false,
		Overpower:   false,
		Overvoltage: false,
	}

	if status.ID != 0 {
		t.Errorf("ID = %d, want 0", status.ID)
	}
	if !status.Output {
		t.Error("Output = false, want true")
	}
	if status.Source != "button" {
		t.Errorf("Source = %q, want button", status.Source)
	}
	if status.Power == nil || *status.Power != 25.5 {
		t.Errorf("Power = %v, want 25.5", status.Power)
	}
	if status.Voltage == nil || *status.Voltage != 230.0 {
		t.Errorf("Voltage = %v, want 230.0", status.Voltage)
	}
	if status.Current == nil || *status.Current != 0.11 {
		t.Errorf("Current = %v, want 0.11", status.Current)
	}
}

func TestSwitchStatus_NilPointers(t *testing.T) {
	t.Parallel()

	status := SwitchStatus{
		ID:     0,
		Output: true,
	}

	if status.Power != nil {
		t.Errorf("Power = %v, want nil", status.Power)
	}
	if status.Voltage != nil {
		t.Errorf("Voltage = %v, want nil", status.Voltage)
	}
	if status.Current != nil {
		t.Errorf("Current = %v, want nil", status.Current)
	}
	if status.Energy != nil {
		t.Errorf("Energy = %v, want nil", status.Energy)
	}
}

func TestSwitchConfig_Fields(t *testing.T) {
	t.Parallel()

	name := testKitchenLight
	cfg := SwitchConfig{
		ID:           0,
		Name:         &name,
		InitialState: "on",
		AutoOn:       true,
		AutoOnDelay:  5.0,
		AutoOff:      true,
		AutoOffDelay: 300.0,
	}

	if cfg.ID != 0 {
		t.Errorf("ID = %d, want 0", cfg.ID)
	}
	if cfg.Name == nil || *cfg.Name != testKitchenLight {
		t.Errorf("Name = %v, want %s", cfg.Name, testKitchenLight)
	}
	if cfg.InitialState != "on" {
		t.Errorf("InitialState = %q, want on", cfg.InitialState)
	}
	if !cfg.AutoOn {
		t.Error("AutoOn = false, want true")
	}
	if cfg.AutoOnDelay != 5.0 {
		t.Errorf("AutoOnDelay = %f, want 5.0", cfg.AutoOnDelay)
	}
	if !cfg.AutoOff {
		t.Error("AutoOff = false, want true")
	}
	if cfg.AutoOffDelay != 300.0 {
		t.Errorf("AutoOffDelay = %f, want 300.0", cfg.AutoOffDelay)
	}
}

func TestCoverStatus_Fields(t *testing.T) {
	t.Parallel()

	position := 50
	target := 100
	power := 10.5
	status := CoverStatus{
		ID:              0,
		State:           "opening",
		Source:          "button",
		CurrentPosition: &position,
		TargetPosition:  &target,
		MoveTimeout:     false,
		Calibrating:     false,
		Power:           &power,
	}

	if status.State != "opening" {
		t.Errorf("State = %q, want opening", status.State)
	}
	if status.CurrentPosition == nil || *status.CurrentPosition != 50 {
		t.Errorf("CurrentPosition = %v, want 50", status.CurrentPosition)
	}
	if status.TargetPosition == nil || *status.TargetPosition != 100 {
		t.Errorf("TargetPosition = %v, want 100", status.TargetPosition)
	}
}

func TestCoverSafety_Fields(t *testing.T) {
	t.Parallel()

	safety := CoverSafety{
		Obstacle:    true,
		Overpower:   false,
		Overtemp:    false,
		Overvoltage: false,
	}

	if !safety.Obstacle {
		t.Error("Obstacle = false, want true")
	}
	if safety.Overpower {
		t.Error("Overpower = true, want false")
	}
}

func TestCoverConfig_Fields(t *testing.T) {
	t.Parallel()

	name := "Bedroom Blinds"
	maxTime := 60.0
	cfg := CoverConfig{
		ID:               0,
		Name:             &name,
		InitialState:     "open",
		InvertDirections: true,
		MaxTime:          &maxTime,
		SwapInputs:       false,
	}

	if cfg.Name == nil || *cfg.Name != "Bedroom Blinds" {
		t.Errorf("Name = %v, want Bedroom Blinds", cfg.Name)
	}
	if !cfg.InvertDirections {
		t.Error("InvertDirections = false, want true")
	}
	if cfg.MaxTime == nil || *cfg.MaxTime != 60.0 {
		t.Errorf("MaxTime = %v, want 60.0", cfg.MaxTime)
	}
}

func TestLightStatus_Fields(t *testing.T) {
	t.Parallel()

	brightness := 75
	power := 15.0
	status := LightStatus{
		ID:         0,
		Output:     true,
		Brightness: &brightness,
		Source:     "ui",
		Power:      &power,
		Overtemp:   false,
		Overpower:  false,
	}

	if !status.Output {
		t.Error("Output = false, want true")
	}
	if status.Brightness == nil || *status.Brightness != 75 {
		t.Errorf("Brightness = %v, want 75", status.Brightness)
	}
}

func TestLightConfig_Fields(t *testing.T) {
	t.Parallel()

	name := "Desk Lamp"
	cfg := LightConfig{
		ID:              0,
		Name:            &name,
		DefaultBright:   50,
		NightModeEnable: true,
		NightModeBright: 10,
	}

	if cfg.DefaultBright != 50 {
		t.Errorf("DefaultBright = %d, want 50", cfg.DefaultBright)
	}
	if !cfg.NightModeEnable {
		t.Error("NightModeEnable = false, want true")
	}
	if cfg.NightModeBright != 10 {
		t.Errorf("NightModeBright = %d, want 10", cfg.NightModeBright)
	}
}

func TestRGBStatus_Fields(t *testing.T) {
	t.Parallel()

	brightness := 100
	status := RGBStatus{
		ID:         0,
		Output:     true,
		Brightness: &brightness,
		RGB: &RGBColor{
			Red:   255,
			Green: 128,
			Blue:  0,
		},
		Source: "app",
	}

	if !status.Output {
		t.Error("Output = false, want true")
	}
	if status.RGB == nil {
		t.Fatal("RGB = nil, want non-nil")
	}
	if status.RGB.Red != 255 {
		t.Errorf("RGB.Red = %d, want 255", status.RGB.Red)
	}
	if status.RGB.Green != 128 {
		t.Errorf("RGB.Green = %d, want 128", status.RGB.Green)
	}
	if status.RGB.Blue != 0 {
		t.Errorf("RGB.Blue = %d, want 0", status.RGB.Blue)
	}
}

func TestRGBColor_Fields(t *testing.T) {
	t.Parallel()

	color := RGBColor{Red: 255, Green: 0, Blue: 128}

	if color.Red != 255 {
		t.Errorf("Red = %d, want 255", color.Red)
	}
	if color.Green != 0 {
		t.Errorf("Green = %d, want 0", color.Green)
	}
	if color.Blue != 128 {
		t.Errorf("Blue = %d, want 128", color.Blue)
	}
}

func TestRGBConfig_Fields(t *testing.T) {
	t.Parallel()

	name := "LED Strip"
	cfg := RGBConfig{
		ID:              0,
		Name:            &name,
		DefaultBright:   80,
		NightModeEnable: false,
		NightModeBright: 5,
	}

	if cfg.Name == nil || *cfg.Name != "LED Strip" {
		t.Errorf("Name = %v, want LED Strip", cfg.Name)
	}
	if cfg.DefaultBright != 80 {
		t.Errorf("DefaultBright = %d, want 80", cfg.DefaultBright)
	}
}

func TestEnergyCounter_Fields(t *testing.T) {
	t.Parallel()

	energy := EnergyCounter{
		Total:    1234.56,
		ByMinute: []float64{0.5, 0.6, 0.4},
		MinuteTs: 1609459200,
	}

	if energy.Total != 1234.56 {
		t.Errorf("Total = %f, want 1234.56", energy.Total)
	}
	if len(energy.ByMinute) != 3 {
		t.Errorf("len(ByMinute) = %d, want 3", len(energy.ByMinute))
	}
	if energy.MinuteTs != 1609459200 {
		t.Errorf("MinuteTs = %d, want 1609459200", energy.MinuteTs)
	}
}

func TestInputStatus_Fields(t *testing.T) {
	t.Parallel()

	status := InputStatus{
		ID:    0,
		State: true,
		Type:  "button",
	}

	if status.ID != 0 {
		t.Errorf("ID = %d, want 0", status.ID)
	}
	if !status.State {
		t.Error("State = false, want true")
	}
	if status.Type != "button" {
		t.Errorf("Type = %q, want button", status.Type)
	}
}

func TestInputConfig_Fields(t *testing.T) {
	t.Parallel()

	name := "Wall Switch"
	cfg := InputConfig{
		ID:     0,
		Name:   &name,
		Type:   "switch",
		Invert: true,
	}

	if cfg.Name == nil || *cfg.Name != "Wall Switch" {
		t.Errorf("Name = %v, want Wall Switch", cfg.Name)
	}
	if cfg.Type != "switch" { //nolint:goconst // test uses ComponentSwitch constant elsewhere
		t.Errorf("Type = %q, want switch", cfg.Type)
	}
	if !cfg.Invert {
		t.Error("Invert = false, want true")
	}
}
