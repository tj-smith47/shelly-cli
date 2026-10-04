package status

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func allFixtures() *mock.Fixtures {
	return &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{Name: "pro3em", Address: "192.168.1.20", MAC: "AA:BB:CC:00:00:20", Type: "SPEM-003CEBEU", Model: "Shelly Pro 3EM", Generation: 2},
				{Name: "proem", Address: "192.168.1.21", MAC: "AA:BB:CC:00:00:21", Type: "SPEM-002CEBEU50", Model: "Shelly Pro EM", Generation: 2},
				{Name: "plug", Address: "192.168.1.22", MAC: "AA:BB:CC:00:00:22", Type: "SNPL-00112EU", Model: "Shelly Plus Plug S", Generation: 2},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"pro3em": {"em:0": map[string]any{"id": 0, "total_current": 4.5, "total_act_power": 1035.0}},
			"proem": {
				"em1:0": map[string]any{"id": 0, "voltage": 230.0, "current": 2.5, "act_power": 575.0},
				"em1:1": map[string]any{"id": 1, "voltage": 231.0, "current": 0.5, "act_power": 100.0},
			},
			"plug": {"switch:0": map[string]any{"output": true}},
		},
	}
}

func runAll(t *testing.T, format string) *factory.TestFactory {
	t.Helper()
	demo, err := mock.StartWithFixtures(allFixtures())
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	if err := demo.ConfigMgr.RegisterDevice("dead", "127.0.0.1:1", 2, "SPEM-003CEBEU", "Shelly Pro 3EM", nil); err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	if format != "" {
		viper.Set("output", format)
		t.Cleanup(viper.Reset)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v\nstderr: %s", err, tf.ErrString())
	}
	return tf
}

//nolint:paralleltest // uses the global default config manager and viper
func TestAll_JSONIsOneList(t *testing.T) {
	tf := runAll(t, "json")

	var entries []model.EnergyStatusEntry
	if err := json.Unmarshal([]byte(tf.OutString()), &entries); err != nil {
		t.Fatalf("stdout is not one JSON list: %v\n%s", err, tf.OutString())
	}

	type key struct {
		name, typ string
		id        int
	}
	got := map[key]float64{}
	for _, e := range entries {
		got[key{e.Name, e.Type, e.ID}] = e.Power
	}
	want := map[key]float64{
		{"pro3em", "em", 0}: 1035,
		{"proem", "em1", 0}: 575,
		{"proem", "em1", 1}: 100,
	}
	if len(got) != len(want) {
		t.Fatalf("entries = %+v, want %d entries", entries, len(want))
	}
	for k, w := range want {
		if p, ok := got[k]; !ok || p != w {
			t.Errorf("entry %+v power = %v (present %v), want %v", k, p, ok, w)
		}
	}

	stderr := tf.ErrString()
	for _, want := range []string{"Skipped plug: no energy monitor", "Skipped dead: unreachable"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q:\n%s", want, stderr)
		}
	}
}

//nolint:paralleltest // uses the global default config manager and viper
func TestAll_YAMLIsOneList(t *testing.T) {
	tf := runAll(t, "yaml")

	out := tf.OutString()
	if !strings.HasPrefix(out, "- ") || strings.Count(out, "\n- ") != 2 {
		t.Errorf("stdout is not a 3-item YAML list:\n%s", out)
	}
}

//nolint:paralleltest // uses the global default config manager and viper
func TestAll_Table(t *testing.T) {
	tf := runAll(t, "")

	out := tf.OutString()
	for _, want := range []string{"pro3em", "EM #0", "proem", "EM1 #0", "EM1 #1", "230.00 V"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "plug") || strings.Contains(out, "dead") {
		t.Errorf("skipped devices leaked into the table:\n%s", out)
	}
}

//nolint:paralleltest // uses the global default config manager and viper
func TestAll_NoEnergyDevicesPrintsEmptyJSONList(t *testing.T) {
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{Devices: []mock.DeviceFixture{
			{Name: "plug", Address: "192.168.1.22", Type: "SNPL-00112EU", Model: "Shelly Plus Plug S", Generation: 2},
		}},
		DeviceStates: map[string]mock.DeviceState{"plug": {"switch:0": map[string]any{"output": true}}},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)
	viper.Set("output", "json")
	t.Cleanup(viper.Reset)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := strings.TrimSpace(tf.OutString()); got != "[]" {
		t.Errorf("stdout = %q, want []", got)
	}
}

func TestAll_ArgsAndFlagConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"all with device", []string{"--all", "pro3em"}, "do not pass a device"},
		{"no device no all", nil, "requires a device argument, or --all"},
		{"all with type", []string{"--all", "--type", "em"}, "none of the others can be"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)
			cmd.SetArgs(tt.args)
			cmd.SetOut(&strings.Builder{})
			cmd.SetErr(&strings.Builder{})
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
