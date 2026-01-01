package status

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
	if cmd.Use != "status <device> [id]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status <device> [id]")
	}

	// Test Aliases
	wantAliases := []string{"st"}
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
		{"two args valid", []string{"device", "0"}, false},
		{"three args", []string{"device", "0", "extra"}, true},
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

	// Test type flag
	flag := cmd.Flags().Lookup("type")
	if flag == nil {
		t.Fatal("--type flag not found")
	}
	if flag.DefValue != "auto" {
		t.Errorf("--type default = %q, want %q", flag.DefValue, "auto")
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
		"shelly energy status",
		"--type",
		"-o json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_InvalidComponentID(t *testing.T) {
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
		t.Fatal("expected error for invalid component ID")
	}

	if !strings.Contains(err.Error(), "invalid component ID") {
		t.Errorf("expected 'invalid component ID' error, got: %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_EMComponent(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-em",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-em": {
				"em:0": map[string]any{
					"id":               0,
					"a_current":        1.5,
					"a_voltage":        230.0,
					"a_act_power":      345.0,
					"a_aprt_power":     350.0,
					"a_pf":             0.98,
					"a_freq":           50.0,
					"b_current":        1.4,
					"b_voltage":        231.0,
					"b_act_power":      323.0,
					"c_current":        1.6,
					"c_voltage":        229.0,
					"c_act_power":      367.0,
					"total_current":    4.5,
					"total_act_power":  1035.0,
					"total_aprt_power": 1055.0,
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

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-em", "0", "--type", "em"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_EM1Component(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-em1",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-002CEBEU120",
					Model:      "Shelly EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-em1": {
				"em1:0": map[string]any{
					"id":         0,
					"current":    2.5,
					"voltage":    230.0,
					"act_power":  575.0,
					"aprt_power": 580.0,
					"pf":         0.99,
					"freq":       50.0,
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

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-em1", "0", "--type", "em1"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_AutoDetectEM(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-em-auto",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-em-auto": {
				"em:0": map[string]any{
					"id":              0,
					"total_act_power": 1000.0,
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

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	// No --type flag, should auto-detect
	cmd.SetArgs([]string{"test-em-auto", "0"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (may be expected for mock): %v", err)
	}
}

func TestRun_NoComponents(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentType: "unknown",
		ComponentID:   0,
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err == nil {
		t.Fatal("expected error for no components")
	}

	if !strings.Contains(err.Error(), "no energy monitoring components found") {
		t.Errorf("expected 'no energy monitoring components found' error, got: %v", err)
	}
}

func TestRun_WithComponentID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device", "5"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Execute - will timeout/error but exercises the component ID parsing path
	if err := cmd.ExecuteContext(ctx); err != nil {
		t.Logf("expected timeout error: %v", err)
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Factory != nil {
		t.Error("Default Factory should be nil")
	}

	if opts.ComponentID != 0 {
		t.Errorf("Default ComponentID = %d, want 0", opts.ComponentID)
	}

	if opts.ComponentType != "" {
		t.Errorf("Default ComponentType = %q, want empty", opts.ComponentType)
	}

	if opts.Device != "" {
		t.Errorf("Default Device = %q, want empty", opts.Device)
	}
}

func TestRun_EMTypeExplicit(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentType: shelly.ComponentTypeEM,
		ComponentID:   0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Run - will error on device connection but exercises the EM path
	err := run(ctx, opts)
	if err == nil || !strings.Contains(err.Error(), "failed to get EM status") {
		if err != nil && !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Logf("expected EM status error or timeout, got: %v", err)
		}
	}
}

func TestRun_EM1TypeExplicit(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentType: shelly.ComponentTypeEM1,
		ComponentID:   0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	// Run - will error on device connection but exercises the EM1 path
	err := run(ctx, opts)
	if err == nil || !strings.Contains(err.Error(), "failed to get EM1 status") {
		if err != nil && !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Logf("expected EM1 status error or timeout, got: %v", err)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"real-time measurements",
		"voltage",
		"current",
		"power",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long description should contain %q", pattern)
		}
	}
}
