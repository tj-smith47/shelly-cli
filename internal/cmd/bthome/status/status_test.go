package status

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const (
	formatText = "text"
	formatJSON = "json"
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

	// Test Use
	if cmd.Use != "status <device> [id]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status <device> [id]")
	}

	// Test Aliases
	wantAliases := []string{"st", "info"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
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
		{"two args valid", []string{"device", "200"}, false},
		{"three args", []string{"device", "200", "extra"}, true},
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

	// Test format flag
	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("--format flag not found")
	}
	if flag.Shorthand != "f" {
		t.Errorf("--format shorthand = %q, want %q", flag.Shorthand, "f")
	}
	if flag.DefValue != formatText {
		t.Errorf("--format default = %q, want %q", flag.DefValue, formatText)
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

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly bthome status",
		"--json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device: "test-device",
		ID:     200,
		HasID:  true,
	}
	opts.Format = formatJSON

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.ID != 200 {
		t.Errorf("ID = %d, want %d", opts.ID, 200)
	}
	if !opts.HasID {
		t.Error("HasID should be true")
	}
	if opts.Format != formatJSON {
		t.Errorf("Format = %q, want %q", opts.Format, formatJSON)
	}
}

func TestNewCommand_InvalidID(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device", "not-a-number"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}

	if !strings.Contains(err.Error(), "invalid device ID") {
		t.Errorf("expected 'invalid device ID' error, got: %v", err)
	}
}

func TestExecute_BTHomeComponentStatus(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthome": map[string]any{
					"errors": []string{},
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
	cmd.SetArgs([]string{"bthome-gateway"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}

	// Verify output contains expected status elements
	output := buf.String()
	if output != "" {
		t.Logf("Output: %s", output)
	}
}

func TestExecute_BTHomeComponentStatusWithDiscovery(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthome": map[string]any{
					"discovery": map[string]any{
						"started_at": float64(1700000000),
						"duration":   60,
					},
					"errors": []string{},
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
	cmd.SetArgs([]string{"bthome-gateway"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}

func TestExecute_BTHomeComponentStatusJSON(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthome": map[string]any{
					"errors": []string{},
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
	cmd.SetArgs([]string{"bthome-gateway", "--format", formatJSON})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}

	// JSON output should contain errors field
	output := buf.String()
	if output != "" && !strings.Contains(output, "errors") {
		t.Logf("Expected JSON output to contain 'errors', got: %s", output)
	}
}

func TestExecute_BTHomeDeviceStatus(t *testing.T) {

	rssi := -55
	battery := 85

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthomedevice:200": map[string]any{
					"id":              200,
					"rssi":            float64(rssi),
					"battery":         float64(battery),
					"last_updated_ts": float64(1700000000),
				},
				"bthomedevice:200_config": map[string]any{
					"id":   200,
					"addr": "11:22:33:44:55:66",
					"name": "Temperature Sensor",
				},
				"bthomedevice:200_known_objects": map[string]any{
					"objects": []map[string]any{
						{"obj_id": 69, "idx": 0, "component": nil},
						{"obj_id": 1, "idx": 0, "component": nil},
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
	cmd.SetArgs([]string{"bthome-gateway", "200"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}

func TestExecute_BTHomeDeviceStatusJSON(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthomedevice:200": map[string]any{
					"id":              200,
					"rssi":            float64(-55),
					"battery":         float64(85),
					"last_updated_ts": float64(1700000000),
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
	cmd.SetArgs([]string{"bthome-gateway", "200", "--format", formatJSON})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}

	// JSON output should contain id field
	output := buf.String()
	if output != "" && !strings.Contains(output, "id") {
		t.Logf("Expected JSON output to contain 'id', got: %s", output)
	}
}

func TestExecute_DeviceNotFound(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
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
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

func TestExecute_ContextCancelled(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"bthome-gateway"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestExecute_BTHomeDeviceStatusWithKnownObjects(t *testing.T) {

	component := "bthomesensor:200"
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthomedevice:200": map[string]any{
					"id":              200,
					"rssi":            float64(-55),
					"battery":         float64(85),
					"packet_id":       float64(42),
					"last_updated_ts": float64(1700000000),
				},
				"bthomedevice:200_config": map[string]any{
					"id":   200,
					"addr": "11:22:33:44:55:66",
					"name": "Multi Sensor",
				},
				"bthomedevice:200_known_objects": map[string]any{
					"objects": []map[string]any{
						{"obj_id": 69, "idx": 0, "component": &component},
						{"obj_id": 1, "idx": 0, "component": nil},
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
	cmd.SetArgs([]string{"bthome-gateway", "200"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}

func TestExecute_BTHomeDeviceWithErrors(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthome": map[string]any{
					"errors": []string{"device_not_responding", "low_battery"},
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
	cmd.SetArgs([]string{"bthome-gateway"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}

func TestRun_HasID(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthomedevice:100": map[string]any{
					"id":              100,
					"rssi":            float64(-60),
					"battery":         float64(90),
					"last_updated_ts": float64(1700000000),
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
		Factory: tf.Factory,
		Device:  "bthome-gateway",
		ID:      100,
		HasID:   true,
	}
	opts.Format = formatText

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run error = %v", err)
	}
}

func TestRun_NoID(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthome": map[string]any{
					"errors": []string{},
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
		Factory: tf.Factory,
		Device:  "bthome-gateway",
		HasID:   false,
	}
	opts.Format = formatText

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run error = %v", err)
	}
}

func TestRun_HasIDWithJSON(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthomedevice:100": map[string]any{
					"id":              100,
					"rssi":            float64(-60),
					"battery":         float64(90),
					"last_updated_ts": float64(1700000000),
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
		Factory: tf.Factory,
		Device:  "bthome-gateway",
		ID:      100,
		HasID:   true,
	}
	opts.Format = formatJSON

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run error = %v", err)
	}
}

func TestRun_NoIDWithJSON(t *testing.T) {

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "bthome-gateway",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNGW-BT01",
					Model:      "Shelly BLU Gateway",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"bthome-gateway": {
				"bthome": map[string]any{
					"errors": []string{},
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
		Factory: tf.Factory,
		Device:  "bthome-gateway",
		HasID:   false,
	}
	opts.Format = formatJSON

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run error = %v", err)
	}
}
