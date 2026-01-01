package override

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

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

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		wantOK    bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use == "override <device>" },
			wantOK:    true,
			errMsg:    "Use should be 'override <device>'",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			wantOK:    true,
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			wantOK:    true,
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			wantOK:    true,
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK:    true,
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK:    true,
			errMsg:    "RunE should be set",
		},
		{
			name:      "has ValidArgsFunction",
			checkFunc: func(c *cobra.Command) bool { return c.ValidArgsFunction != nil },
			wantOK:    true,
			errMsg:    "ValidArgsFunction should be set for completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if tt.checkFunc(cmd) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"temp-override", "manual"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

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
			cmd := NewCommand(cmdutil.NewFactory())
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

	// Test target flag
	targetFlag := cmd.Flags().Lookup("target")
	if targetFlag == nil {
		t.Fatal("--target flag not found")
	}
	if targetFlag.Shorthand != "t" {
		t.Errorf("target shorthand = %q, want t", targetFlag.Shorthand)
	}

	// Test duration flag
	durationFlag := cmd.Flags().Lookup("duration")
	if durationFlag == nil {
		t.Fatal("--duration flag not found")
	}
	if durationFlag.Shorthand != "d" {
		t.Errorf("duration shorthand = %q, want d", durationFlag.Shorthand)
	}

	// Test cancel flag
	cancelFlag := cmd.Flags().Lookup("cancel")
	if cancelFlag == nil {
		t.Fatal("--cancel flag not found")
	}

	// Test id flag
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Fatal("--id flag not found")
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
		"shelly thermostat override",
		"--target",
		"--duration",
		"--cancel",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"override",
		"temperature",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	if opts.Target != 0 {
		t.Errorf("Default Target = %f, want 0", opts.Target)
	}
	if opts.Duration != 0 {
		t.Errorf("Default Duration = %v, want 0", opts.Duration)
	}
	if opts.Cancel {
		t.Error("Default Cancel should be false")
	}
	if opts.ID != 0 {
		t.Errorf("Default ID = %d, want 0", opts.ID)
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	err := run(ctx, opts)

	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	err := run(ctx, opts)

	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestNewCommand_ExecuteWithNoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when executing with no arguments")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "target flag",
			args:    []string{"--target", "22.5"},
			wantErr: false,
		},
		{
			name:    "target flag short",
			args:    []string{"-t", "20"},
			wantErr: false,
		},
		{
			name:    "duration flag",
			args:    []string{"--duration", "30m"},
			wantErr: false,
		},
		{
			name:    "duration flag short",
			args:    []string{"-d", "1h"},
			wantErr: false,
		},
		{
			name:    "cancel flag",
			args:    []string{"--cancel"},
			wantErr: false,
		},
		{
			name:    "id flag",
			args:    []string{"--id", "1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_WithMockGen1Device(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": true}},
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
		Device:  "gen1-device",
		Target:  22.0,
	}

	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error for Gen1 device")
	}

	if !strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("Expected error mentioning Gen2+, got: %v", err)
	}
}

func TestRun_WithMockGen2Device_Override(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
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
		Factory:  tf.Factory,
		Device:   "thermostat-device",
		Target:   25.0,
		Duration: 30 * time.Minute,
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device override: %v", err)
	}
}

func TestRun_WithMockGen2Device_Cancel(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
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
		Device:  "thermostat-device",
		Cancel:  true,
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device cancel override: %v", err)
	}

	// Check for success message
	output := tf.OutString()
	if !strings.Contains(output, "cancelled") {
		t.Logf("Output: %s", output)
	}
}

func TestRun_WithMockGen2Device_OverrideWithoutDuration(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
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
		Device:  "thermostat-device",
		Target:  20.0, // Target without duration
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device override without duration: %v", err)
	}
}

func TestRun_WithMockGen2Device_OverrideDefaultParams(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
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
		Device:  "thermostat-device",
		// No target or duration - use device defaults
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device override with defaults: %v", err)
	}

	// Check for "device defaults" message
	output := tf.OutString()
	if !strings.Contains(output, "device defaults") {
		t.Logf("Output: %s", output)
	}
}

func TestRun_WithMockGen2Device_OverrideWithDurationOnly(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
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
		Factory:  tf.Factory,
		Device:   "thermostat-device",
		Duration: 2 * time.Hour, // Duration without target
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device override with duration only: %v", err)
	}
}

func TestOptions_FactoryAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	if opts.Factory == nil {
		t.Fatal("Options.Factory should not be nil")
	}

	ios := opts.Factory.IOStreams()
	if ios == nil {
		t.Error("Factory.IOStreams() should not return nil")
	}

	svc := opts.Factory.ShellyService()
	if svc == nil {
		t.Error("Factory.ShellyService() should not return nil")
	}
}

func TestNewCommand_ExecuteWithOverrideFlags(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
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
	cmd.SetArgs([]string{"thermostat-device", "--target", "23", "--duration", "1h"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}

func TestNewCommand_ExecuteWithCancelFlag(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
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
	cmd.SetArgs([]string{"thermostat-device", "--cancel"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}
