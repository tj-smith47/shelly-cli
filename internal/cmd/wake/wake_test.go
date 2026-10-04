package wake

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "wake <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "wake <device>")
	}

	wantAliases := []string{"sunrise", "wakeup"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"device"}, false},
		{"two args", []string{"device", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("delay")
	if flag == nil {
		t.Fatal("--delay flag not found")
	}
	if flag.Shorthand != "d" {
		t.Errorf("--delay shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly wake",
		"-d",
		"--delay",
		"--simulate sunrise --duration",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestExecute_WithMock(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "--delay", "1ms"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestRun_ShortDelay(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Delay:   1 * time.Millisecond,
	}
	err = run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	out := tf.OutString()
	if !strings.Contains(out, "Good morning") && !strings.Contains(out, "on") {
		t.Errorf("Output should indicate success, got: %s", out)
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Delay: 10 * time.Minute,
	}

	if opts.Delay != 10*time.Minute {
		t.Errorf("Delay = %v, want %v", opts.Delay, 10*time.Minute)
	}
}

func TestNewCommand_SimulateFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	for name, def := range map[string]string{"simulate": "", "duration": "15m0s"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("--%s flag not found", name)
			continue
		}
		if flag.DefValue != def {
			t.Errorf("--%s default = %q, want %q", name, flag.DefValue, def)
		}
	}
}

func TestExecute_SimulateValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"duration without simulate", []string{"bedroom", "--duration", "15m"}, "add --simulate sunrise or drop --duration"},
		{"unknown simulation", []string{"bedroom", "--simulate", "sunset"}, `unsupported --simulate "sunset"`},
		{"zero duration", []string{"bedroom", "--simulate", "sunrise", "--duration", "0s"}, "--duration must be greater than zero"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)
			cmd.SetArgs(tt.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// startWakeMock registers one device with the given generation and state.
func startWakeMock(t *testing.T, gen int, typ string, state mock.DeviceState) (*factory.TestFactory, *mock.Demo) {
	t.Helper()
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{Devices: []mock.DeviceFixture{{
			Name: "bedroom", Address: "192.168.1.50", MAC: "AA:BB:CC:DD:EE:50",
			Type: typ, Model: typ, Generation: gen,
		}}},
		DeviceStates: map[string]mock.DeviceState{"bedroom": state},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)
	return tf, demo
}

func sunriseOpts(tf *factory.TestFactory, duration time.Duration) *Options {
	return &Options{
		Factory:      tf.Factory,
		Device:       "bedroom",
		Simulate:     "sunrise",
		Duration:     duration,
		RampInterval: time.Millisecond,
	}
}

func TestRun_SunriseGen2RampsToFull(t *testing.T) {
	t.Parallel()

	tf, demo := startWakeMock(t, 2, "SNDM-0013US", mock.DeviceState{
		"light:0": map[string]any{"output": false, "brightness": 0},
	})
	if err := run(t.Context(), sunriseOpts(tf, 20*time.Millisecond)); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	light, ok := demo.DeviceServer.GetState("bedroom")["light:0"].(map[string]any)
	if !ok {
		t.Fatal("light:0 state missing")
	}
	if light["output"] != true || light["brightness"] != 100 {
		t.Errorf("light:0 = %v, want output=true brightness=100", light)
	}
	if !strings.Contains(tf.OutString(), "full brightness") {
		t.Errorf("output should report full brightness, got: %s", tf.OutString())
	}
}

func TestRun_SunriseGen1RampsToFull(t *testing.T) {
	t.Parallel()

	tf, demo := startWakeMock(t, 1, "SHBDUO-1", mock.DeviceState{
		"lights": []any{map[string]any{"ison": false, "brightness": 0}},
	})
	if err := run(t.Context(), sunriseOpts(tf, 20*time.Millisecond)); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	lights, ok := demo.DeviceServer.GetState("bedroom")["lights"].([]any)
	if !ok || len(lights) != 1 {
		t.Fatal("lights state missing")
	}
	light, ok := lights[0].(map[string]any)
	if !ok {
		t.Fatal("lights[0] state has the wrong shape")
	}
	if light["ison"] != true || light["brightness"] != 100 {
		t.Errorf("lights[0] = %v, want ison=true brightness=100", light)
	}
}

// A device without a light component fails before the delay starts; the one
// hour delay would otherwise hang this test.
func TestRun_SunriseRejectsDeviceWithoutLight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		gen   int
		typ   string
		state mock.DeviceState
	}{
		{"gen2 switch", 2, "SNSW-001P16EU", mock.DeviceState{"switch:0": map[string]any{"output": false}}},
		{"gen1 relay", 1, "SHSW-1", mock.DeviceState{"relays": []any{map[string]any{"ison": false}}}},
		{"gen1 rgbw in color mode", 1, "SHRGBW2", mock.DeviceState{
			"mode":   "color",
			"lights": []any{map[string]any{"ison": false}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf, demo := startWakeMock(t, tt.gen, tt.typ, tt.state)
			opts := sunriseOpts(tf, time.Minute)
			opts.Delay = time.Hour

			err := run(t.Context(), opts)
			if err == nil || !strings.Contains(err.Error(), "no light component with brightness control") {
				t.Fatalf("error = %v, want the no-dimmable-light error", err)
			}
			if sw, ok := demo.DeviceServer.GetState("bedroom")["switch:0"].(map[string]any); ok && sw["output"] != false {
				t.Errorf("switch must not be turned on, got %v", sw)
			}
		})
	}
}

// Cancelling during the ramp (Ctrl+C) leaves the light at the level it reached
// and returns without an error.
func TestRun_SunriseStopsOnCancel(t *testing.T) {
	t.Parallel()

	tf, demo := startWakeMock(t, 2, "SNDM-0013US", mock.DeviceState{
		"light:0": map[string]any{"output": false, "brightness": 0},
	})
	ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
	defer cancel()

	// One hour over 99 steps puts the first step about 36s away, so the
	// timeout always lands after the 1% start and before any further step.
	if err := run(ctx, sunriseOpts(tf, time.Hour)); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	light, ok := demo.DeviceServer.GetState("bedroom")["light:0"].(map[string]any)
	if !ok {
		t.Fatal("light:0 state missing")
	}
	if light["output"] != true || light["brightness"] != 1 {
		t.Errorf("light:0 = %v, want output=true brightness=1", light)
	}
	if !strings.Contains(tf.OutString()+tf.ErrString(), "Sunrise stopped at 1%") {
		t.Errorf("expected a stop message, got out=%q err=%q", tf.OutString(), tf.ErrString())
	}
}
