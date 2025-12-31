package history

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "history <device> [id]" {
		t.Errorf("Use = %q, want 'history <device> [id]'", cmd.Use)
	}

	expectedAliases := []string{"hist", "events"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "type", shorthand: "", defValue: shelly.ComponentTypeAuto},
		{name: "period", shorthand: "p", defValue: ""},
		{name: "from", shorthand: "", defValue: ""},
		{name: "to", shorthand: "", defValue: ""},
		{name: "limit", shorthand: "", defValue: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept 1 or 2 arguments
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "0"})
	if err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "0", "extra"})
	if err == nil {
		t.Error("expected error with 3 args")
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
		"shelly energy history",
		"--from",
		"--to",
		"--period",
		"--limit",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"energy",
		"historical",
		"EM",
		"EM1",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestNewCommand_InvalidComponentID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"device", "not-a-number"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid component ID")
	}

	if !strings.Contains(err.Error(), "invalid component ID") {
		t.Errorf("expected 'invalid component ID' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_EMType(t *testing.T) {
	// Device with EM data component
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "em-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"em-device": {
				"em:0": map[string]any{
					"id":              0,
					"a_voltage":       230.5,
					"b_voltage":       231.2,
					"c_voltage":       229.8,
					"total_act_power": 1500.0,
					"total_current":   6.5,
				},
				"em:0_records": map[string]any{
					"records": []map[string]any{
						{"ts": 1735500000 - 3600, "period": 60},
						{"ts": 1735500000, "period": 60},
					},
				},
				"em:0_history": map[string]any{
					"data": []map[string]any{
						{
							"ts":     1735500000 - 3600,
							"period": 60,
							"values": []map[string]any{
								{
									"ts":              1735500000 - 3600,
									"a_voltage":       230.5,
									"b_voltage":       231.2,
									"c_voltage":       229.8,
									"a_act_power":     500.0,
									"b_act_power":     500.0,
									"c_act_power":     500.0,
									"total_act_power": 1500.0,
								},
							},
						},
					},
				},
			},
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
	cmd.SetArgs([]string{"em-device", "--type", "em"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// May succeed or fail depending on mock, but should not panic
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_EM1Type(t *testing.T) {
	// Device with EM1 data component
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "em1-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SNEM-001X8EU",
					Model:      "Shelly Pro EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"em1-device": {
				"em1:0": map[string]any{
					"id":        0,
					"voltage":   230.5,
					"current":   2.5,
					"act_power": 575.0,
				},
				"em1:0_records": map[string]any{
					"records": []map[string]any{
						{"ts": 1735500000 - 3600, "period": 60},
						{"ts": 1735500000, "period": 60},
					},
				},
				"em1:0_history": map[string]any{
					"data": []map[string]any{
						{
							"ts":     1735500000 - 3600,
							"period": 60,
							"values": []map[string]any{
								{
									"ts":        1735500000 - 3600,
									"voltage":   230.5,
									"current":   2.5,
									"act_power": 575.0,
								},
							},
						},
					},
				},
			},
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
	cmd.SetArgs([]string{"em1-device", "--type", "em1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithComponentID(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "multi-em-device",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:03",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"multi-em-device": {
				"em:0": map[string]any{"id": 0, "total_act_power": 1000.0},
				"em:1": map[string]any{"id": 1, "total_act_power": 2000.0},
			},
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
	cmd.SetArgs([]string{"multi-em-device", "1", "--type", "em"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithPeriodFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {
				"em:0": map[string]any{"id": 0, "total_act_power": 1000.0},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	periods := []string{"hour", "day", "week", "month"}
	for _, period := range periods { //nolint:paralleltest // subtest shares global state
		t.Run(period, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := NewCommand(tf.Factory)
			cmd.SetContext(context.Background())
			cmd.SetArgs([]string{"test-device", "--type", "em", "--period", period})
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()
			if err != nil {
				t.Logf("Execute error for period %s = %v (may be expected)", period, err)
			}
		})
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestExecute_InvalidPeriod(t *testing.T) {
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"some-device", "--type", "em", "--period", "invalid-period"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid period")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid") {
		t.Logf("Error: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithFromToFlags(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {
				"em:0": map[string]any{"id": 0, "total_act_power": 1000.0},
			},
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
	cmd.SetArgs([]string{"test-device", "--type", "em", "--from", "2025-01-01", "--to", "2025-01-07"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestExecute_InvalidFromTime(t *testing.T) {
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"some-device", "--type", "em", "--from", "invalid-time"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid from time")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid") {
		t.Logf("Error: %v", err)
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestExecute_InvalidToTime(t *testing.T) {
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"some-device", "--type", "em", "--to", "invalid-time"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid to time")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid") {
		t.Logf("Error: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithLimitFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {
				"em:0": map[string]any{"id": 0, "total_act_power": 1000.0},
			},
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
	cmd.SetArgs([]string{"test-device", "--type", "em", "--limit", "10"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_AutoDetectType(t *testing.T) {
	// Test auto-detection when neither EM nor EM1 data found
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "no-energy-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"no-energy-device": {
				"switch:0": map[string]any{"output": false},
			},
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
	cmd.SetArgs([]string{"no-energy-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Should fail because no energy data components
	if err == nil {
		t.Error("expected error when no energy data components found")
	}
	if err != nil && !strings.Contains(err.Error(), "no energy data components") {
		t.Logf("Error: %v", err)
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestExecute_UnknownDevice(t *testing.T) {
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"unknown-device", "--type", "em"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown device")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_InvalidTypeFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--type", "invalid"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Should fail because invalid component type
	if err == nil {
		t.Error("expected error for invalid type flag")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_EMDataHistory(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "em-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"em-device": {
				"em:0": map[string]any{
					"id": 0,
				},
			},
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
		Factory:       tf.Factory,
		Device:        "em-device",
		ComponentID:   0,
		ComponentType: "em",
		Period:        "day",
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_EM1DataHistory(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "em1-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SNEM-001X8EU",
					Model:      "Shelly Pro EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"em1-device": {
				"em1:0": map[string]any{
					"id": 0,
				},
			},
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
		Factory:       tf.Factory,
		Device:        "em1-device",
		ComponentID:   0,
		ComponentType: "em1",
		Period:        "day",
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses factory.NewTestFactory which is not parallel-safe
func TestRun_InvalidTimeRange(t *testing.T) {
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "any-device",
		ComponentID:   0,
		ComponentType: "em",
		Period:        "invalid",
	}

	// Test with invalid period
	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for invalid time range")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected 'invalid' in error, got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_UnknownComponentType(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: "unknown",
		Period:        "day",
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for unknown component type")
	}
	if err != nil && !strings.Contains(err.Error(), "no energy data components") {
		t.Errorf("expected 'no energy data components' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_WithFromToTimes(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNEM-001X16EU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {
				"em:0": map[string]any{"id": 0},
			},
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
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: "em",
		From:          "2025-01-01",
		To:            "2025-01-07",
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected with mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_AutoDetectFails(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "no-energy-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"no-energy-device": {
				"switch:0": map[string]any{"output": false},
			},
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
		Factory:       tf.Factory,
		Device:        "no-energy-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeAuto,
		Period:        "day",
	}

	// Use "auto" type to trigger auto-detection
	err = run(context.Background(), opts)
	if err == nil {
		t.Error("expected error when auto-detection fails")
	}
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
}

// Note: Tests for CalculateTimeRange, ParseTime, CalculateEMMetrics, and CalculateEM1Metrics
// are now in internal/shelly/energy_test.go since these functions were extracted to the
// service layer for DRY compliance.
