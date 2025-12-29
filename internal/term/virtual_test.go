package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func ptrFloat(f float64) *float64 { return &f }
func ptrString(s string) *string  { return &s }
func ptrBool(b bool) *bool        { return &b }

func TestDisplayVirtualComponents_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayVirtualComponents(ios, []shelly.VirtualComponent{})

	output := out.String()
	if !strings.Contains(output, "virtual components") {
		t.Error("expected no results message")
	}
}

func TestDisplayVirtualComponents_WithComponents(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	components := []shelly.VirtualComponent{
		{
			Key:       "boolean:200",
			Type:      shelly.VirtualBoolean,
			ID:        200,
			Name:      "Light State",
			BoolValue: ptrBool(true),
		},
		{
			Key:      "number:201",
			Type:     shelly.VirtualNumber,
			ID:       201,
			Name:     "Temperature",
			NumValue: ptrFloat(22.5),
			Unit:     ptrString("°C"),
		},
	}
	DisplayVirtualComponents(ios, components)

	output := out.String()
	if !strings.Contains(output, "Virtual Components") {
		t.Error("expected title")
	}
	if !strings.Contains(output, "boolean:200") {
		t.Error("expected component key")
	}
	if !strings.Contains(output, "Light State") {
		t.Error("expected component name")
	}
	if !strings.Contains(output, "2") {
		t.Error("expected count")
	}
}

func TestDisplayVirtualComponent_Boolean(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	comp := &shelly.VirtualComponent{
		Key:       "boolean:100",
		Type:      shelly.VirtualBoolean,
		ID:        100,
		Name:      "Switch State",
		BoolValue: ptrBool(true),
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "Virtual Component") {
		t.Error("expected title")
	}
	if !strings.Contains(output, "boolean:100") {
		t.Error("expected key")
	}
	if !strings.Contains(output, "Type") {
		t.Error("expected type label")
	}
	if !strings.Contains(output, "Switch State") {
		t.Error("expected name")
	}
	if !strings.Contains(output, "true") {
		t.Error("expected boolean value")
	}
}

func TestDisplayVirtualComponent_Number(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	minVal := float64(0)
	maxVal := float64(100)
	comp := &shelly.VirtualComponent{
		Key:      "number:101",
		Type:     shelly.VirtualNumber,
		ID:       101,
		Name:     "Humidity",
		NumValue: ptrFloat(65.5),
		Unit:     ptrString("%"),
		Min:      &minVal,
		Max:      &maxVal,
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "65.50") {
		t.Error("expected numeric value")
	}
	if !strings.Contains(output, "%") {
		t.Error("expected unit")
	}
	if !strings.Contains(output, "Range") {
		t.Error("expected range label")
	}
	if !strings.Contains(output, "0.00 - 100.00") {
		t.Error("expected range values")
	}
}

func TestDisplayVirtualComponent_Text(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	comp := &shelly.VirtualComponent{
		Key:      "text:102",
		Type:     shelly.VirtualText,
		ID:       102,
		Name:     "Status Message",
		StrValue: ptrString("All good"),
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "All good") {
		t.Error("expected text value")
	}
}

func TestDisplayVirtualComponent_Enum(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	comp := &shelly.VirtualComponent{
		Key:      "enum:103",
		Type:     shelly.VirtualEnum,
		ID:       103,
		Name:     "Mode",
		StrValue: ptrString("auto"),
		Options:  []string{"off", "auto", "manual"},
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "auto") {
		t.Error("expected enum value")
	}
	if !strings.Contains(output, "Options") {
		t.Error("expected options label")
	}
}

func TestDisplayVirtualComponent_Button(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	comp := &shelly.VirtualComponent{
		Key:  "button:104",
		Type: shelly.VirtualButton,
		ID:   104,
		Name: "Trigger",
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "(button)") {
		t.Error("expected button indicator")
	}
}

func TestDisplayVirtualComponent_Group(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	comp := &shelly.VirtualComponent{
		Key:  "group:105",
		Type: shelly.VirtualGroup,
		ID:   105,
		Name: "Settings",
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "(group)") {
		t.Error("expected group indicator")
	}
}

func TestDisplayVirtualComponent_NoValue(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	comp := &shelly.VirtualComponent{
		Key:  "boolean:106",
		Type: shelly.VirtualBoolean,
		ID:   106,
		Name: "Unknown",
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "(no value)") {
		t.Error("expected no value indicator")
	}
}

func TestDisplayVirtualComponent_MinOnly(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	minVal := float64(0)
	comp := &shelly.VirtualComponent{
		Key:      "number:107",
		Type:     shelly.VirtualNumber,
		ID:       107,
		NumValue: ptrFloat(50),
		Min:      &minVal,
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, ">=") {
		t.Error("expected min-only range indicator")
	}
}

func TestDisplayVirtualComponent_MaxOnly(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	maxVal := float64(100)
	comp := &shelly.VirtualComponent{
		Key:      "number:108",
		Type:     shelly.VirtualNumber,
		ID:       108,
		NumValue: ptrFloat(50),
		Max:      &maxVal,
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "<=") {
		t.Error("expected max-only range indicator")
	}
}

func TestDisplayVirtualComponent_RawValue(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	comp := &shelly.VirtualComponent{
		Key:   "unknown:109",
		Type:  "custom",
		ID:    109,
		Value: "raw-value",
	}
	DisplayVirtualComponent(ios, comp)

	output := out.String()
	if !strings.Contains(output, "raw-value") {
		t.Error("expected raw value fallback")
	}
}
