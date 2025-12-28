package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-go/profiles"
	"github.com/tj-smith47/shelly-go/types"
)

func TestDisplayProfile(t *testing.T) {
	t.Parallel()

	t.Run("basic profile", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		profile := &profiles.Profile{
			Model:       "SNSW-001P16EU",
			Name:        "Shelly Plus 1PM",
			Generation:  types.Gen2,
			Series:      profiles.SeriesPlus,
			FormFactor:  profiles.FormFactorDIN,
			PowerSource: profiles.PowerSourceMains,
			Components: profiles.Components{
				Switches: 1,
				Inputs:   1,
			},
			Protocols: profiles.Protocols{
				HTTP:      true,
				WebSocket: true,
				MQTT:      true,
			},
			Capabilities: profiles.Capabilities{
				PowerMetering: true,
				Scripting:     true,
			},
		}

		DisplayProfile(ios, profile)

		output := out.String()
		if !strings.Contains(output, "SNSW-001P16EU") {
			t.Error("output should contain model")
		}
		if !strings.Contains(output, "Shelly Plus 1PM") {
			t.Error("output should contain name")
		}
		if !strings.Contains(output, "Switches: 1") {
			t.Error("output should contain switches count")
		}
	})

	t.Run("with limits", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		profile := &profiles.Profile{
			Model:      "SNSW-001P16EU",
			Name:       "Test Device",
			Generation: types.Gen2,
			Series:     profiles.SeriesPlus,
			Limits: profiles.Limits{
				MaxScripts:   10,
				MaxSchedules: 20,
				MaxPower:     3500,
			},
		}

		DisplayProfile(ios, profile)

		output := out.String()
		if !strings.Contains(output, "Max Scripts: 10") {
			t.Error("output should contain max scripts")
		}
		if !strings.Contains(output, "Max Power: 3500W") {
			t.Error("output should contain max power")
		}
	})

	t.Run("with app name", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		profile := &profiles.Profile{
			Model:      "SNSW-001X16EU",
			Name:       "Shelly 1 Mini",
			App:        "Mini1G3",
			Generation: types.Gen3,
		}

		DisplayProfile(ios, profile)

		output := out.String()
		if !strings.Contains(output, "Mini1G3") {
			t.Error("output should contain app name")
		}
	})
}

func TestBoolYesNo(t *testing.T) {
	t.Parallel()

	t.Run("true", func(t *testing.T) {
		t.Parallel()

		result := boolYesNo(true)
		if !strings.Contains(result, "Yes") {
			t.Errorf("expected 'Yes', got %q", result)
		}
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()

		result := boolYesNo(false)
		if !strings.Contains(result, "No") {
			t.Errorf("expected 'No', got %q", result)
		}
	})
}

func TestProfileProtocolList(t *testing.T) {
	t.Parallel()

	t.Run("multiple protocols", func(t *testing.T) {
		t.Parallel()

		p := &profiles.Profile{
			Protocols: profiles.Protocols{
				HTTP:      true,
				WebSocket: true,
				MQTT:      true,
				BLE:       true,
			},
		}

		result := profileProtocolList(p)

		if len(result) != 4 {
			t.Errorf("expected 4 protocols, got %d", len(result))
		}
	})

	t.Run("all protocols", func(t *testing.T) {
		t.Parallel()

		p := &profiles.Profile{
			Protocols: profiles.Protocols{
				HTTP:      true,
				WebSocket: true,
				MQTT:      true,
				CoIoT:     true,
				BLE:       true,
				Matter:    true,
				Zigbee:    true,
				ZWave:     true,
				Ethernet:  true,
			},
		}

		result := profileProtocolList(p)

		if len(result) != 9 {
			t.Errorf("expected 9 protocols, got %d", len(result))
		}
	})
}

func TestProfileCapabilityList(t *testing.T) {
	t.Parallel()

	t.Run("with capabilities", func(t *testing.T) {
		t.Parallel()

		p := &profiles.Profile{
			Capabilities: profiles.Capabilities{
				PowerMetering: true,
				Scripting:     true,
			},
		}

		result := profileCapabilityList(p)

		found := false
		for _, c := range result {
			if c == "Power Metering" {
				found = true
				break
			}
		}
		if !found {
			t.Error("result should contain 'Power Metering'")
		}
	})

	t.Run("no capabilities", func(t *testing.T) {
		t.Parallel()

		p := &profiles.Profile{}

		result := profileCapabilityList(p)

		if len(result) != 1 || result[0] != "Basic functionality" {
			t.Errorf("expected 'Basic functionality', got %v", result)
		}
	})
}

func TestProfileHasLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		limits profiles.Limits
		want   bool
	}{
		{
			name:   "no limits",
			limits: profiles.Limits{},
			want:   false,
		},
		{
			name:   "has max scripts",
			limits: profiles.Limits{MaxScripts: 10},
			want:   true,
		},
		{
			name:   "has max power",
			limits: profiles.Limits{MaxPower: 3500},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &profiles.Profile{Limits: tt.limits}
			got := profileHasLimits(p)
			if got != tt.want {
				t.Errorf("profileHasLimits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseProfileGeneration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  types.Generation
	}{
		{"1", types.Gen1},
		{"gen1", types.Gen1},
		{"Gen1", types.Gen1},
		{"2", types.Gen2},
		{"gen2", types.Gen2},
		{"Gen2", types.Gen2},
		{"3", types.Gen3},
		{"4", types.Gen4},
		{"invalid", types.GenerationUnknown},
		{"", types.GenerationUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := ParseProfileGeneration(tt.input)
			if got != tt.want {
				t.Errorf("ParseProfileGeneration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseProfileSeries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  profiles.Series
	}{
		{"classic", profiles.SeriesClassic},
		{"plus", profiles.SeriesPlus},
		{"pro", profiles.SeriesPro},
		{"mini", profiles.SeriesMini},
		{"blu", profiles.SeriesBLU},
		{"wave", profiles.SeriesWave},
		{"wave_pro", profiles.SeriesWavePro},
		{"standard", profiles.SeriesStandard},
		{"invalid", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := ParseProfileSeries(tt.input)
			if got != tt.want {
				t.Errorf("ParseProfileSeries(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
