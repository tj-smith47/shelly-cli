package status

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// carrierFixtures registers one device per kind of component that meters
// power, plus a relay that meters nothing.
func carrierFixtures() *mock.Fixtures {
	gen2 := func(name, typ string) mock.DeviceFixture {
		return mock.DeviceFixture{Name: name, Address: "192.0.2.1", MAC: "AA:BB:CC:00:00:" + name[:2], Type: typ, Model: typ, Generation: 2}
	}
	meter := func(id int, w float64) map[string]any {
		return map[string]any{"id": id, "output": true, "apower": w, "voltage": 230.0, "current": w / 230,
			"aenergy": map[string]any{"total": 1000.0 + w}}
	}
	return &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{Devices: []mock.DeviceFixture{
			gen2("p1pm", "SNSW-001P16EU"),
			gen2("p2pm", "SNSW-102P16EU"),
			gen2("cv", "SNSW-102P16EU"),
			gen2("dim", "SNDM-0013US"),
			gen2("rgbwpm", "SNDC-0D4P10WW"),
			gen2("mini", "SNPM-001PCEU16"),
			gen2("p3em", "SPEM-003CEBEU"),
			gen2("pem", "SPEM-002CEBEU50"),
			gen2("relay", "SNSW-001X16EU"),
			{Name: "g1pm", Address: "192.0.2.2", MAC: "AA:BB:CC:00:01:01", Type: "SHSW-PM", Model: "SHSW-PM", Generation: 1},
			{Name: "g1em", Address: "192.0.2.3", MAC: "AA:BB:CC:00:01:02", Type: "SHEM", Model: "SHEM", Generation: 1},
		}},
		DeviceStates: map[string]mock.DeviceState{
			"p1pm":   {"switch:0": meter(0, 48.5)},
			"p2pm":   {"switch:0": meter(0, 10), "switch:1": meter(1, 20)},
			"cv":     {"cover:0": meter(0, 35)},
			"dim":    {"light:0": meter(0, 7.5)},
			"rgbwpm": {"rgbw:0": meter(0, 6)},
			"mini":   {"pm1:0": meter(0, 12.5)},
			"p3em":   {"em:0": map[string]any{"id": 0, "total_current": 4.5, "total_act_power": 1035.0}},
			"pem": {
				"em1:0": map[string]any{"id": 0, "voltage": 230.0, "current": 2.5, "act_power": 575.0},
				"em1:1": map[string]any{"id": 1, "voltage": 231.0, "current": 0.5, "act_power": 100.0},
			},
			"relay": {"switch:0": map[string]any{"id": 0, "output": true}},
			"g1pm":  {"meters": []any{map[string]any{"power": 60.5, "total": 600, "is_valid": true}}},
			"g1em": {"emeters": []any{
				map[string]any{"power": 300.0, "voltage": 231.0, "current": 1.3, "total": 5000.0},
				map[string]any{"power": 40.0, "voltage": 231.0, "current": 0.2, "total": 700.0},
			}},
		},
	}
}

func execute(t *testing.T, format string, args ...string) (*factory.TestFactory, error) {
	t.Helper()
	demo, err := mock.StartWithFixtures(carrierFixtures())
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)
	if format != "" {
		viper.Set("output", format)
		t.Cleanup(viper.Reset)
	}
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	return tf, cmd.Execute()
}

//nolint:paralleltest // uses the global default config manager and viper
func TestRun_EveryCarrier(t *testing.T) {
	tests := []struct {
		args      []string
		wantType  string
		wantID    int
		wantPower float64
		wantKey   string
	}{
		{[]string{"p1pm"}, "switch", 0, 48.5, "meter"},
		{[]string{"p2pm", "1"}, "switch", 1, 20, "meter"},
		{[]string{"cv"}, "cover", 0, 35, "meter"},
		{[]string{"dim"}, "light", 0, 7.5, "meter"},
		{[]string{"rgbwpm"}, "rgbw", 0, 6, "meter"},
		{[]string{"mini", "--type", "pm1"}, "pm1", 0, 12.5, "meter"},
		{[]string{"p3em"}, "em", 0, 1035, "em"},
		{[]string{"pem", "1"}, "em1", 1, 100, "em1"},
		{[]string{"g1pm"}, "meter", 0, 60.5, "meter"},
		{[]string{"g1em", "1"}, "emeter", 1, 40, "meter"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			tf, err := execute(t, "json", tt.args...)
			if err != nil {
				t.Fatalf("Execute: %v\nstderr: %s", err, tf.ErrString())
			}
			var raw map[string]any
			if err := json.Unmarshal([]byte(tf.OutString()), &raw); err != nil {
				t.Fatalf("stdout is not one JSON object: %v\n%s", err, tf.OutString())
			}
			var r model.PowerReading
			if err := json.Unmarshal([]byte(tf.OutString()), &r); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if r.Name != tt.args[0] || r.Type != tt.wantType || r.ID != tt.wantID || r.Power != tt.wantPower {
				t.Errorf("reading = %+v, want %s %s:%d %v W", r, tt.args[0], tt.wantType, tt.wantID, tt.wantPower)
			}
			if _, ok := raw[tt.wantKey]; !ok {
				t.Errorf("JSON has no %q key:\n%s", tt.wantKey, tf.OutString())
			}
		})
	}
}

//nolint:paralleltest // uses the global default config manager and viper
func TestRun_Gen1MeterTotalIsWattHours(t *testing.T) {
	tf, err := execute(t, "json", "g1pm")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var r model.PowerReading
	if err := json.Unmarshal([]byte(tf.OutString()), &r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The device reports 600 watt-minutes.
	if r.Meter == nil || r.Meter.AEnergy == nil || r.Meter.AEnergy.Total != 10 {
		t.Errorf("meter = %+v, want aenergy.total 10 Wh", r.Meter)
	}
}

//nolint:paralleltest // uses the global default config manager and viper
func TestRun_Table(t *testing.T) {
	tf, err := execute(t, "", "p2pm", "1")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := tf.OutString()
	for _, want := range []string{"Switch #1", "Voltage: 230.00 V", "Power:   20.00 W", "Total: 1020.00 Wh"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

//nolint:paralleltest // uses the global default config manager and viper
func TestRun_Errors(t *testing.T) {
	tests := []struct {
		args    []string
		wantErr string
	}{
		{[]string{"relay"}, "relay: no component on this device reports power"},
		{[]string{"p2pm", "5"}, "p2pm has no component 5 that reports power; its power readings are: switch:0, switch:1"},
		{[]string{"p1pm", "--type", "pm1"}, "p1pm has no pm1 that reports power; its power readings are: switch:0"},
		{[]string{"p1pm", "x"}, `invalid component ID "x"`},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, err := execute(t, "", tt.args...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

//nolint:paralleltest // uses the global default config manager and viper
func TestRun_All(t *testing.T) {
	tf, err := execute(t, "json", "--all")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var readings []model.PowerReading
	if err := json.Unmarshal([]byte(tf.OutString()), &readings); err != nil {
		t.Fatalf("stdout is not one JSON list: %v\n%s", err, tf.OutString())
	}
	got := map[string]bool{}
	for _, r := range readings {
		got[r.Name+" "+r.Type] = true
	}
	for _, want := range []string{"p1pm switch", "p2pm switch", "cv cover", "dim light", "rgbwpm rgbw",
		"mini pm1", "p3em em", "pem em1", "g1pm meter", "g1em emeter"} {
		if !got[want] {
			t.Errorf("--all is missing %q; got %v", want, got)
		}
	}
	if len(readings) != 13 {
		t.Errorf("got %d readings, want 13", len(readings))
	}
	if stderr := tf.ErrString(); !strings.Contains(stderr, "Skipped relay: no component on this device reports power") {
		t.Errorf("stderr does not name the relay:\n%s", stderr)
	}
}

func TestNewCommand_Shape(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())
	if cmd.Use != "status [device] [id]" || len(cmd.Aliases) == 0 || cmd.Example == "" {
		t.Errorf("Use=%q Aliases=%v Example empty=%v", cmd.Use, cmd.Aliases, cmd.Example == "")
	}
	for _, name := range []string{"type", "all"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s", name)
		}
	}
}
