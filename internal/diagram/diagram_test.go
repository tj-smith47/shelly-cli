package diagram

import (
	"strings"
	"testing"
)

func TestParseGeneration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"1", 1, false},
		{"2", 2, false},
		{"3", 3, false},
		{"4", 4, false},
		{"gen1", 1, false},
		{"gen2", 2, false},
		{"gen3", 3, false},
		{"gen4", 4, false},
		{"Gen1", 1, false},
		{"Gen2", 2, false},
		{"GEN3", 3, false},
		{"", 0, true},
		{"5", 0, true},
		{"genX", 0, true},
		{"gen0", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseGeneration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGeneration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseGeneration(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    Style
		wantErr bool
	}{
		{"schematic", StyleSchematic, false},
		{"compact", StyleCompact, false},
		{"detailed", StyleDetailed, false},
		{"Schematic", StyleSchematic, false},
		{"COMPACT", StyleCompact, false},
		{"Detailed", StyleDetailed, false},
		{"", 0, true},
		{"unknown", 0, true},
		{"simple", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseStyle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStyle(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseStyle(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidStyles(t *testing.T) {
	t.Parallel()

	styles := ValidStyles()
	if len(styles) != 3 {
		t.Errorf("ValidStyles() returned %d styles, want 3", len(styles))
	}
	expected := []string{"schematic", "compact", "detailed"}
	for i, s := range expected {
		if styles[i] != s {
			t.Errorf("ValidStyles()[%d] = %q, want %q", i, styles[i], s)
		}
	}
}

func TestValidGenerations(t *testing.T) {
	t.Parallel()

	gens := ValidGenerations()
	if len(gens) != 8 {
		t.Errorf("ValidGenerations() returned %d values, want 8", len(gens))
	}
}

func TestLookupModelUnambiguous(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("plus-1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Shelly Plus 1" {
		t.Errorf("got name %q, want %q", m.Name, "Shelly Plus 1")
	}
	if m.Generation != 2 {
		t.Errorf("got generation %d, want 2", m.Generation)
	}
}

func TestLookupModelMultiGenDefaultsToLatest(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Generation != 4 {
		t.Errorf("got generation %d, want 4 (latest)", m.Generation)
	}
	if m.Name != "Shelly 1 Gen4" {
		t.Errorf("got name %q, want %q", m.Name, "Shelly 1 Gen4")
	}
}

func TestLookupModelWithGeneration(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Shelly 1" {
		t.Errorf("got name %q, want %q", m.Name, "Shelly 1")
	}
	if m.Generation != 1 {
		t.Errorf("got generation %d, want 1", m.Generation)
	}
}

func TestLookupModelNotFound(t *testing.T) {
	t.Parallel()

	_, err := LookupModel("nonexistent", 0)
	if err == nil {
		t.Fatal("expected error for nonexistent model")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain 'not found', got: %v", err)
	}
}

func TestLookupModelNotFoundWithGeneration(t *testing.T) {
	t.Parallel()

	_, err := LookupModel("plus-1", 1)
	if err == nil {
		t.Fatal("expected error for plus-1 gen1")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain 'not found', got: %v", err)
	}
}

func TestLookupModelAltSlug(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("shelly25", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Shelly 2.5" {
		t.Errorf("got name %q, want %q", m.Name, "Shelly 2.5")
	}
}

func TestLookupModelCaseInsensitive(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("PLUS-1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Shelly Plus 1" {
		t.Errorf("got name %q, want %q", m.Name, "Shelly Plus 1")
	}
}

func TestModelSlugs(t *testing.T) {
	t.Parallel()

	t.Run("all slugs", func(t *testing.T) {
		t.Parallel()
		slugs := ModelSlugs(0)
		if len(slugs) == 0 {
			t.Fatal("ModelSlugs(0) returned empty")
		}
		found := false
		for _, s := range slugs {
			if s == "plus-1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("ModelSlugs(0) should contain 'plus-1'")
		}
	})

	t.Run("filtered by generation", func(t *testing.T) {
		t.Parallel()
		gen1Slugs := ModelSlugs(1)
		for _, s := range gen1Slugs {
			m, err := LookupModel(s, 1)
			if err != nil {
				t.Errorf("slug %q from ModelSlugs(1) failed lookup: %v", s, err)
				continue
			}
			if m.Generation != 1 {
				t.Errorf("slug %q has generation %d, expected 1", s, m.Generation)
			}
		}
	})

	t.Run("gen1 count", func(t *testing.T) {
		t.Parallel()
		slugs := ModelSlugs(1)
		if len(slugs) != 10 {
			t.Errorf("ModelSlugs(1) returned %d slugs, want 10", len(slugs))
		}
	})

	t.Run("sorted", func(t *testing.T) {
		t.Parallel()
		slugs := ModelSlugs(0)
		for i := 1; i < len(slugs); i++ {
			if slugs[i] < slugs[i-1] {
				t.Errorf("ModelSlugs not sorted: %q before %q", slugs[i-1], slugs[i])
				break
			}
		}
	})
}

func TestListModels(t *testing.T) {
	t.Parallel()

	t.Run("all models", func(t *testing.T) {
		t.Parallel()
		models := ListModels(0)
		if len(models) == 0 {
			t.Fatal("ListModels(0) returned empty")
		}
	})

	t.Run("filtered", func(t *testing.T) {
		t.Parallel()
		models := ListModels(1)
		for _, m := range models {
			if m.Generation != 1 {
				t.Errorf("ListModels(1) returned model with generation %d", m.Generation)
			}
		}
	})
}

// renderTestCase defines a test for rendering a topology with a given style.
type renderTestCase struct {
	name         string
	slug         string
	generation   int
	style        Style
	wantContains []string
}

func TestRenderTopologies(t *testing.T) {
	t.Parallel()

	tests := []renderTestCase{
		// SingleRelay - all 3 styles
		{name: "SingleRelay/schematic", slug: "plus-1", style: StyleSchematic,
			wantContains: []string{"Shelly Plus 1", "L", "N", "SW", "O"}},
		{name: "SingleRelay/compact", slug: "plus-1", style: StyleCompact,
			wantContains: []string{"Shelly Plus 1", "L", "N", "SW", "O"}},
		{name: "SingleRelay/detailed", slug: "plus-1", style: StyleDetailed,
			wantContains: []string{"Shelly Plus 1", "L", "N", "SW", "O"}},

		// DualRelay - all 3 styles
		{name: "DualRelay/schematic", slug: "2.5", generation: 1, style: StyleSchematic,
			wantContains: []string{"Shelly 2.5", "L", "N", "S1", "S2", "O1", "O2"}},
		{name: "DualRelay/compact", slug: "2.5", generation: 1, style: StyleCompact,
			wantContains: []string{"Shelly 2.5", "L", "N", "S1", "S2", "O1", "O2"}},
		{name: "DualRelay/detailed", slug: "2.5", generation: 1, style: StyleDetailed,
			wantContains: []string{"Shelly 2.5", "L", "N", "S1", "S2", "O1", "O2"}},

		// QuadRelay - all 3 styles
		{name: "QuadRelay/schematic", slug: "pro-4pm", style: StyleSchematic,
			wantContains: []string{"Shelly Pro 4PM", "L", "N", "S1", "S4", "O1", "O4"}},
		{name: "QuadRelay/compact", slug: "pro-4pm", style: StyleCompact,
			wantContains: []string{"Shelly Pro 4PM", "L", "N", "S1", "S4", "O1", "O4"}},
		{name: "QuadRelay/detailed", slug: "pro-4pm", style: StyleDetailed,
			wantContains: []string{"Shelly Pro 4PM", "L", "N", "S1", "S4", "O1", "O4"}},

		// Dimmer - all 3 styles
		{name: "Dimmer/schematic", slug: "dimmer-2", style: StyleSchematic,
			wantContains: []string{"Shelly Dimmer 2", "L", "N", "SW1", "SW2", "O"}},
		{name: "Dimmer/compact", slug: "dimmer-2", style: StyleCompact,
			wantContains: []string{"Shelly Dimmer 2", "L", "N", "SW1", "SW2", "O"}},
		{name: "Dimmer/detailed", slug: "dimmer-2", style: StyleDetailed,
			wantContains: []string{"Shelly Dimmer 2", "L", "N", "SW1", "SW2", "O"}},

		// InputOnly - all 3 styles
		{name: "InputOnly/schematic", slug: "i3", style: StyleSchematic,
			wantContains: []string{"Shelly i3", "L", "N", "SW1", "SW2", "SW3"}},
		{name: "InputOnly/compact", slug: "i3", style: StyleCompact,
			wantContains: []string{"Shelly i3", "L", "N", "SW1"}},
		{name: "InputOnly/detailed", slug: "i3", style: StyleDetailed,
			wantContains: []string{"Shelly i3", "L", "N", "SW1", "SW2", "SW3"}},

		// Plug - all 3 styles
		{name: "Plug/schematic", slug: "plug-s", generation: 1, style: StyleSchematic,
			wantContains: []string{"Shelly Plug S", "Relay", "Plug"}},
		{name: "Plug/compact", slug: "plug-s", generation: 1, style: StyleCompact,
			wantContains: []string{"Shelly Plug S", "Plug"}},
		{name: "Plug/detailed", slug: "plug-s", generation: 1, style: StyleDetailed,
			wantContains: []string{"Shelly Plug S", "Relay", "Plug"}},

		// EnergyMonitor - all 3 styles
		{name: "EnergyMonitor/schematic", slug: "em", style: StyleSchematic,
			wantContains: []string{"Shelly EM", "L", "N", "CT1", "CT2"}},
		{name: "EnergyMonitor/compact", slug: "em", style: StyleCompact,
			wantContains: []string{"Shelly EM", "L", "N", "CT"}},
		{name: "EnergyMonitor/detailed", slug: "em", style: StyleDetailed,
			wantContains: []string{"Shelly EM", "L", "N", "CT1", "CT2"}},

		// RGBW - all 3 styles
		{name: "RGBW/schematic", slug: "rgbw2", style: StyleSchematic,
			wantContains: []string{"Shelly RGBW2", "V+", "GND", "R", "G", "B", "W"}},
		{name: "RGBW/compact", slug: "rgbw2", style: StyleCompact,
			wantContains: []string{"Shelly RGBW2", "V+", "GND", "R", "G", "B", "W"}},
		{name: "RGBW/detailed", slug: "rgbw2", style: StyleDetailed,
			wantContains: []string{"Shelly RGBW2", "V+", "GND", "R", "G", "B", "W"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := LookupModel(tt.slug, tt.generation)
			if err != nil {
				t.Fatalf("LookupModel(%q, %d): %v", tt.slug, tt.generation, err)
			}

			renderer := NewRenderer(tt.style)
			output := renderer.Render(m)

			if output == "" {
				t.Fatal("Render returned empty string")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q\n\nFull output:\n%s", want, output)
				}
			}
		})
	}
}

func TestRenderContainsSpecs(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("plus-1pm", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, style := range []Style{StyleSchematic, StyleCompact, StyleDetailed} {
		renderer := NewRenderer(style)
		output := renderer.Render(m)

		if !strings.Contains(output, "Specs:") {
			t.Errorf("style %d: output missing 'Specs:'", style)
		}
		if !strings.Contains(output, "Voltage:") {
			t.Errorf("style %d: output missing 'Voltage:'", style)
		}
		if !strings.Contains(output, "Metering:") {
			t.Errorf("style %d: output missing 'Metering:' for PM model", style)
		}
	}
}

func TestNewRendererDefault(t *testing.T) {
	t.Parallel()

	r := NewRenderer(StyleSchematic)
	if r == nil {
		t.Fatal("NewRenderer(StyleSchematic) returned nil")
	}
	if _, ok := r.(*schematicRenderer); !ok {
		t.Error("NewRenderer(StyleSchematic) did not return schematicRenderer")
	}

	r = NewRenderer(StyleCompact)
	if _, ok := r.(*compactRenderer); !ok {
		t.Error("NewRenderer(StyleCompact) did not return compactRenderer")
	}

	r = NewRenderer(StyleDetailed)
	if _, ok := r.(*detailedRenderer); !ok {
		t.Error("NewRenderer(StyleDetailed) did not return detailedRenderer")
	}
}

func TestRenderInputOnlyFourInputs(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("plus-i4", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	renderer := NewRenderer(StyleSchematic)
	output := renderer.Render(m)

	if !strings.Contains(output, "SW4") {
		t.Errorf("plus-i4 should have SW4 in output:\n%s", output)
	}
}

func TestRenderEnergyMonitor3Phase(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("3em", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	renderer := NewRenderer(StyleSchematic)
	output := renderer.Render(m)

	if !strings.Contains(output, "CT3") {
		t.Errorf("3em should have CT3 in output:\n%s", output)
	}
}

func TestNewRendererUnknownStyle(t *testing.T) {
	t.Parallel()

	// Unknown style should fall back to schematic
	r := NewRenderer(Style(99))
	if r == nil {
		t.Fatal("NewRenderer with unknown style returned nil")
	}
	if _, ok := r.(*schematicRenderer); !ok {
		t.Error("unknown style should default to schematicRenderer")
	}
}

func TestHeader(t *testing.T) {
	t.Parallel()

	m := DeviceModel{Name: "Test Device", Generation: 2}
	h := header(m)
	if !strings.Contains(h, "Test Device") {
		t.Error("header should contain device name")
	}
	if !strings.Contains(h, "Gen2") {
		t.Error("header should contain generation")
	}
	if !strings.Contains(h, "─") {
		t.Error("header should contain underline")
	}
}

func TestSpecsFooter(t *testing.T) {
	t.Parallel()

	t.Run("full specs", func(t *testing.T) {
		t.Parallel()
		m := DeviceModel{
			Specs: DeviceSpecs{
				Voltage:         "110-240V AC",
				MaxAmps:         16,
				PowerMonitoring: true,
				NeutralRequired: true,
				Notes:           "Test note",
			},
		}
		f := specsFooter(m)
		for _, want := range []string{"Voltage:", "Max Amps:", "Metering:", "Neutral:    Required", "Note:"} {
			if !strings.Contains(f, want) {
				t.Errorf("specsFooter missing %q, got:\n%s", want, f)
			}
		}
	})

	t.Run("minimal specs", func(t *testing.T) {
		t.Parallel()
		m := DeviceModel{
			Specs: DeviceSpecs{
				Voltage: "12V DC",
			},
		}
		f := specsFooter(m)
		if !strings.Contains(f, "Voltage:") {
			t.Error("specsFooter should contain Voltage")
		}
		if strings.Contains(f, "Max Amps:") {
			t.Error("specsFooter should not contain Max Amps when 0")
		}
		if strings.Contains(f, "Metering:") {
			t.Error("specsFooter should not contain Metering when false")
		}
		if !strings.Contains(f, "Neutral:    Optional") {
			t.Error("specsFooter should show Neutral as Optional when not required")
		}
		if strings.Contains(f, "Note:") {
			t.Error("specsFooter should not contain Note when empty")
		}
	})
}

func TestInputCount(t *testing.T) {
	t.Parallel()

	t.Run("i3 device", func(t *testing.T) {
		t.Parallel()
		m := DeviceModel{Slug: "i3"}
		if got := inputCount(m); got != 3 {
			t.Errorf("inputCount(i3) = %d, want 3", got)
		}
	})

	t.Run("i4 device", func(t *testing.T) {
		t.Parallel()
		m := DeviceModel{Slug: "plus-i4"}
		if got := inputCount(m); got != 4 {
			t.Errorf("inputCount(plus-i4) = %d, want 4", got)
		}
	})
}

func TestEmChannelCount(t *testing.T) {
	t.Parallel()

	t.Run("2-channel", func(t *testing.T) {
		t.Parallel()
		m := DeviceModel{Slug: "em"}
		if got := emChannelCount(m); got != 2 {
			t.Errorf("emChannelCount(em) = %d, want 2", got)
		}
	})

	t.Run("3-channel", func(t *testing.T) {
		t.Parallel()
		m := DeviceModel{Slug: "3em"}
		if got := emChannelCount(m); got != 3 {
			t.Errorf("emChannelCount(3em) = %d, want 3", got)
		}
	})
}

func TestLookupModel1PMDefaultsToGen4(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("1pm", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Generation != 4 {
		t.Errorf("got generation %d, want 4 (latest)", m.Generation)
	}
	if m.Name != "Shelly 1PM Gen4" {
		t.Errorf("got name %q, want %q", m.Name, "Shelly 1PM Gen4")
	}
}

func TestLookupModel1MiniDefaultsToGen4(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("1-mini", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Generation != 4 {
		t.Errorf("got generation %d, want 4 (latest)", m.Generation)
	}
}

func TestModelGenerations(t *testing.T) {
	t.Parallel()

	t.Run("multi-gen slug", func(t *testing.T) {
		t.Parallel()
		gens := ModelGenerations("1")
		if len(gens) < 3 {
			t.Fatalf("slug '1' should match 3+ generations, got %d", len(gens))
		}
		// Should be sorted
		for i := 1; i < len(gens); i++ {
			if gens[i] <= gens[i-1] {
				t.Errorf("ModelGenerations not sorted: %v", gens)
				break
			}
		}
	})

	t.Run("single-gen slug", func(t *testing.T) {
		t.Parallel()
		gens := ModelGenerations("plus-1")
		if len(gens) != 1 {
			t.Errorf("slug 'plus-1' should match 1 generation, got %d", len(gens))
		}
		if gens[0] != 2 {
			t.Errorf("plus-1 should be Gen2, got Gen%d", gens[0])
		}
	})

	t.Run("1pm multi-gen", func(t *testing.T) {
		t.Parallel()
		gens := ModelGenerations("1pm")
		if len(gens) < 3 {
			t.Fatalf("slug '1pm' should match 3+ generations, got %d", len(gens))
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gens := ModelGenerations("nonexistent")
		if len(gens) != 0 {
			t.Errorf("nonexistent slug should return empty, got %v", gens)
		}
	})
}

func TestLatestGeneration(t *testing.T) {
	t.Parallel()

	matches := []DeviceModel{
		{Name: "Gen1", Generation: 1},
		{Name: "Gen3", Generation: 3},
		{Name: "Gen2", Generation: 2},
	}
	got := latestGeneration(matches)
	if got.Generation != 3 {
		t.Errorf("latestGeneration() = Gen%d, want Gen3", got.Generation)
	}
	if got.Name != "Gen3" {
		t.Errorf("latestGeneration() name = %q, want %q", got.Name, "Gen3")
	}
}

func TestLookupModelGen3WithFilter(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("1", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Generation != 3 {
		t.Errorf("got generation %d, want 3", m.Generation)
	}
	if m.Name != "Shelly 1 Gen3" {
		t.Errorf("got name %q, want %q", m.Name, "Shelly 1 Gen3")
	}
}

func TestLookupModelGen4WithFilter(t *testing.T) {
	t.Parallel()

	m, err := LookupModel("1", 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Generation != 4 {
		t.Errorf("got generation %d, want 4", m.Generation)
	}
	if m.Name != "Shelly 1 Gen4" {
		t.Errorf("got name %q, want %q", m.Name, "Shelly 1 Gen4")
	}
}
