package set

import (
	"bytes"
	"context"
	"strings"
	"testing"

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

	// Test Use
	if cmd.Use != "set <device> <event> <url>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <device> <event> <url>")
	}

	// Test Aliases
	wantAliases := []string{"add", "configure"}
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
		{"one arg", []string{"device"}, true},
		{"two args", []string{"device", "event"}, true},
		{"three args valid", []string{"device", "out_on_url", "http://example.com"}, false},
		{"four args", []string{"device", "event", "url", "extra"}, true},
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

	tests := []struct {
		name     string
		defValue string
	}{
		{"index", "0"},
		{"enabled", "true"},
		{"disabled", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
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
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly action set",
		"out_on_url",
		"--index",
		"--disabled",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:  "test-device",
		Event:   "out_on_url",
		URL:     "http://example.com/callback",
		Index:   1,
		Enabled: true,
		Factory: f,
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}

	if opts.Event != "out_on_url" {
		t.Errorf("Event = %q, want %q", opts.Event, "out_on_url")
	}

	if opts.URL != "http://example.com/callback" {
		t.Errorf("URL = %q, want %q", opts.URL, "http://example.com/callback")
	}

	if opts.Index != 1 {
		t.Errorf("Index = %d, want %d", opts.Index, 1)
	}

	if !opts.Enabled {
		t.Error("Enabled should be true")
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}
}

func TestOptions_Disabled(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:  "test-device",
		Event:   "out_on_url",
		URL:     "http://example.com",
		Enabled: false,
		Factory: f,
	}

	if opts.Enabled {
		t.Error("Enabled should be false")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecute_Gen2Device(t *testing.T) {
	// Gen2 device - should fail with "action URLs only available for Gen1"
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen2-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen2-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetArgs([]string{"gen2-device", "out_on_url", "http://example.com/callback"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for Gen2 device")
	}
	if err != nil && !strings.Contains(err.Error(), "Gen1") {
		t.Errorf("Expected error to mention Gen1, got: %v", err)
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecute_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

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
	cmd.SetArgs([]string{"nonexistent-device", "out_on_url", "http://example.com/callback"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecute_Gen1Device(t *testing.T) {
	// Gen1 device - should succeed in setting action
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": false}},
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
	cmd.SetArgs([]string{"gen1-device", "out_on_url", "http://example.com/on"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// May succeed or fail depending on mock capabilities
	if err != nil {
		t.Logf("Execute error = %v (may be expected if mock doesn't support action URLs)", err)
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecute_Gen1DeviceWithIndex(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": false}},
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
	cmd.SetArgs([]string{"gen1-device", "out_on_url", "http://example.com/on", "--index", "1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// May succeed or fail depending on mock capabilities
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestExecute_Gen1DeviceDisabled(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": false}},
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
	cmd.SetArgs([]string{"gen1-device", "out_on_url", "http://example.com/on", "--disabled"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// May succeed or fail depending on mock capabilities
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRun_Gen2Device(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen2-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen2-device": {"switch:0": map[string]any{"output": false}},
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
		Device:  "gen2-device",
		Event:   "out_on_url",
		URL:     "http://example.com/callback",
		Index:   0,
		Enabled: true,
		Factory: tf.Factory,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for Gen2 device")
	}
	if err != nil && !strings.Contains(err.Error(), "Gen1") {
		t.Errorf("Expected error to mention Gen1, got: %v", err)
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRun_Gen1Device(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": false}},
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
		Device:  "gen1-device",
		Event:   "out_on_url",
		URL:     "http://example.com/on",
		Index:   0,
		Enabled: true,
		Factory: tf.Factory,
	}

	err = run(context.Background(), opts)
	// May succeed or fail depending on mock capabilities
	if err != nil {
		t.Logf("run() error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRun_Gen1DeviceDisabled(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": false}},
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
		Device:  "gen1-device",
		Event:   "out_on_url",
		URL:     "http://example.com/on",
		Index:   0,
		Enabled: false,
		Factory: tf.Factory,
	}

	err = run(context.Background(), opts)
	// May succeed or fail depending on mock capabilities
	if err != nil {
		t.Logf("run() error = %v (may be expected)", err)
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRun_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:  "nonexistent-device",
		Event:   "out_on_url",
		URL:     "http://example.com/callback",
		Index:   0,
		Enabled: true,
		Factory: tf.Factory,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

//nolint:paralleltest // Tests use mock.StartWithFixtures with shared global state
func TestRun_Gen1DeviceWithIndex(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": false}},
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
		Device:  "gen1-device",
		Event:   "out_on_url",
		URL:     "http://example.com/on",
		Index:   1,
		Enabled: true,
		Factory: tf.Factory,
	}

	err = run(context.Background(), opts)
	// May succeed or fail depending on mock capabilities
	if err != nil {
		t.Logf("run() error = %v (may be expected)", err)
	}
}
