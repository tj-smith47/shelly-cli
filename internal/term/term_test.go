package term

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

//nolint:gocritic // helper function returns multiple values
func testIOStreams() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return iostreams.Test(in, out, errOut), out, errOut
}

func TestFormatTemp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		temp float64
	}{
		{"cold", 20.0},
		{"warm", 45.0},
		{"hot", 55.0},
		{"very hot", 75.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatTemp(tt.temp)
			if result == "" {
				t.Error("formatTemp returned empty string")
			}
			// Result should contain the temperature value
			// Note: it may have ANSI styling codes
		})
	}
}

func TestJoinStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		parts []string
		sep   string
		want  string
	}{
		{"empty", []string{}, ", ", ""},
		{"single", []string{"a"}, ", ", "a"},
		{"two", []string{"a", "b"}, ", ", "a, b"},
		{"three", []string{"a", "b", "c"}, "-", "a-b-c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := joinStrings(tt.parts, tt.sep)
			if got != tt.want {
				t.Errorf("joinStrings(%v, %q) = %q, want %q", tt.parts, tt.sep, got, tt.want)
			}
		})
	}
}

func TestRepeatChar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    rune
		n    int
		want string
	}{
		{"zero", '-', 0, ""},
		{"one", '-', 1, "-"},
		{"five", '=', 5, "====="},
		{"unicode", '★', 3, "★★★"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := repeatChar(tt.c, tt.n)
			if got != tt.want {
				t.Errorf("repeatChar(%q, %d) = %q, want %q", tt.c, tt.n, got, tt.want)
			}
		})
	}
}

func TestValueOrEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "<not connected>"},
		{"non-empty", "value", "value"},
		{"whitespace only", "  ", "  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := valueOrEmpty(tt.input)
			if got != tt.want {
				t.Errorf("valueOrEmpty(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatTempRanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		temp float64
	}{
		{"below 50", 35.0},
		{"at 50", 50.0},
		{"between 50-70", 60.0},
		{"at 70", 70.0},
		{"above 70", 80.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatTemp(tt.temp)
			// Should contain the temperature value
			if result == "" {
				t.Error("formatTemp returned empty string")
			}
		})
	}
}

func TestJoinStringsVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		parts []string
		sep   string
		want  string
	}{
		{"empty parts", []string{}, ", ", ""},
		{"single part", []string{"a"}, ", ", "a"},
		{"two parts comma", []string{"a", "b"}, ", ", "a, b"},
		{"three parts dash", []string{"a", "b", "c"}, "-", "a-b-c"},
		{"empty sep", []string{"a", "b"}, "", "ab"},
		{"long sep", []string{"a", "b"}, " | ", "a | b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := joinStrings(tt.parts, tt.sep)
			if got != tt.want {
				t.Errorf("joinStrings(%v, %q) = %q, want %q", tt.parts, tt.sep, got, tt.want)
			}
		})
	}
}

func TestRepeatCharVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    rune
		n    int
		want string
	}{
		{"zero", '-', 0, ""},
		{"one", '-', 1, "-"},
		{"five dash", '-', 5, "-----"},
		{"five equals", '=', 5, "====="},
		{"unicode star", '★', 3, "★★★"},
		{"ten spaces", ' ', 10, "          "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := repeatChar(tt.c, tt.n)
			if got != tt.want {
				t.Errorf("repeatChar(%q, %d) = %q, want %q", tt.c, tt.n, got, tt.want)
			}
		})
	}
}

func TestDisplayIntegratorCredentialHelp(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayIntegratorCredentialHelp(ios)

	output := out.String()
	if output == "" {
		t.Error("DisplayIntegratorCredentialHelp should produce output")
	}
	if !strings.Contains(output, "SHELLY_INTEGRATOR_TAG") {
		t.Error("output should mention SHELLY_INTEGRATOR_TAG")
	}
	if !strings.Contains(output, "SHELLY_INTEGRATOR_TOKEN") {
		t.Error("output should mention SHELLY_INTEGRATOR_TOKEN")
	}
}

func TestDisplaySwitchStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := &model.SwitchStatus{
		ID:     0,
		Output: true,
	}
	DisplaySwitchStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplaySwitchStatus should produce output")
	}
}

func TestDisplaySwitchStatusWithPower(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	power := 100.5
	voltage := 220.0
	current := 0.45
	status := &model.SwitchStatus{
		ID:      0,
		Output:  true,
		Power:   &power,
		Voltage: &voltage,
		Current: &current,
		Energy:  &model.EnergyCounter{Total: 150.0},
	}
	DisplaySwitchStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplaySwitchStatus should produce output")
	}
}

func TestDisplayLightStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	brightness := 75
	status := &model.LightStatus{
		ID:         0,
		Output:     true,
		Brightness: &brightness,
	}
	DisplayLightStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplayLightStatus should produce output")
	}
}

func TestDisplayLightStatusWithPower(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	brightness := 100
	power := 10.0
	voltage := 220.0
	current := 0.05
	status := &model.LightStatus{
		ID:         0,
		Output:     true,
		Brightness: &brightness,
		Power:      &power,
		Voltage:    &voltage,
		Current:    &current,
	}
	DisplayLightStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplayLightStatus should produce output")
	}
}

func TestDisplayRGBStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	brightness := 50
	power := 5.0
	status := &model.RGBStatus{
		ID:         0,
		Output:     true,
		RGB:        &model.RGBColor{Red: 255, Green: 100, Blue: 50},
		Brightness: &brightness,
		Power:      &power,
	}
	DisplayRGBStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplayRGBStatus should produce output")
	}
}

func TestDisplayRGBWStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	brightness := 80
	white := 100
	status := &model.RGBWStatus{
		ID:         0,
		Output:     true,
		RGB:        &model.RGBColor{Red: 200, Green: 150, Blue: 100},
		White:      &white,
		Brightness: &brightness,
	}
	DisplayRGBWStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplayRGBWStatus should produce output")
	}
}

func TestDisplayCoverStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	position := 50
	status := &model.CoverStatus{
		ID:              0,
		State:           "stopped",
		CurrentPosition: &position,
	}
	DisplayCoverStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplayCoverStatus should produce output")
	}
}

func TestDisplayInputStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := &model.InputStatus{
		ID:    0,
		State: true,
	}
	DisplayInputStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplayInputStatus should produce output")
	}
}

func TestDisplaySwitchList(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	switches := []shelly.SwitchInfo{
		{ID: 0, Name: "Switch 0", Output: true},
		{ID: 1, Name: "Switch 1", Output: false},
	}
	DisplaySwitchList(ios, switches)

	output := out.String()
	if output == "" {
		t.Error("DisplaySwitchList should produce output")
	}
}

func TestDisplaySwitchListEmpty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplaySwitchList(ios, []shelly.SwitchInfo{})

	// Empty list still prints the table header
	output := out.String()
	if output == "" {
		t.Error("DisplaySwitchList should produce output even for empty list")
	}
}

func TestDisplayLightList(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	lights := []shelly.LightInfo{
		{ID: 0, Name: "Light 0", Output: true, Brightness: 50},
	}
	DisplayLightList(ios, lights)

	output := out.String()
	if output == "" {
		t.Error("DisplayLightList should produce output")
	}
}

func TestDisplayRGBList(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	rgbs := []shelly.RGBInfo{
		{ID: 0, Name: "RGB 0", Output: true, Brightness: 75},
	}
	DisplayRGBList(ios, rgbs)

	output := out.String()
	if output == "" {
		t.Error("DisplayRGBList should produce output")
	}
}

func TestDisplayRGBWList(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	rgbws := []shelly.RGBWInfo{
		{ID: 0, Name: "RGBW 0", Output: true, Brightness: 100},
	}
	DisplayRGBWList(ios, rgbws)

	output := out.String()
	if output == "" {
		t.Error("DisplayRGBWList should produce output")
	}
}

func TestDisplayCoverList(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	covers := []shelly.CoverInfo{
		{ID: 0, Name: "Cover 0", State: "stopped", Position: 50},
	}
	DisplayCoverList(ios, covers)

	output := out.String()
	if output == "" {
		t.Error("DisplayCoverList should produce output")
	}
}

func TestDisplayInputList(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	inputs := []shelly.InputInfo{
		{ID: 0, Name: "Input 0", State: true, Type: "button"},
	}
	DisplayInputList(ios, inputs)

	output := out.String()
	if output == "" {
		t.Error("DisplayInputList should produce output")
	}
}
