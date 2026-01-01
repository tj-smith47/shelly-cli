package create

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
			checkFunc: func(c *cobra.Command) bool { return c.Use == "create <device>" },
			wantOK:    true,
			errMsg:    "Use should be 'create <device>'",
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

	expectedAliases := []string{"add", "new"}
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

	// Test time flag
	timeFlag := cmd.Flags().Lookup("time")
	if timeFlag == nil {
		t.Fatal("--time flag not found")
	}
	if timeFlag.Shorthand != "t" {
		t.Errorf("time shorthand = %q, want t", timeFlag.Shorthand)
	}

	// Test target flag
	targetFlag := cmd.Flags().Lookup("target")
	if targetFlag == nil {
		t.Fatal("--target flag not found")
	}

	// Test mode flag
	modeFlag := cmd.Flags().Lookup("mode")
	if modeFlag == nil {
		t.Fatal("--mode flag not found")
	}

	// Test enable flag
	enableFlag := cmd.Flags().Lookup("enable")
	if enableFlag == nil {
		t.Fatal("--enable flag not found")
	}

	// Test disable flag
	disableFlag := cmd.Flags().Lookup("disable")
	if disableFlag == nil {
		t.Fatal("--disable flag not found")
	}

	// Test enabled flag
	enabledFlag := cmd.Flags().Lookup("enabled")
	if enabledFlag == nil {
		t.Fatal("--enabled flag not found")
	}

	// Test thermostat-id flag
	idFlag := cmd.Flags().Lookup("thermostat-id")
	if idFlag == nil {
		t.Fatal("--thermostat-id flag not found")
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
		"shelly thermostat schedule create",
		"--target",
		"--time",
		"--mode",
		"--disable",
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
		"schedule",
		"thermostat",
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

	if opts.ThermostatID != 0 {
		t.Errorf("Default ThermostatID = %d, want 0", opts.ThermostatID)
	}
	if opts.TargetC != 0 {
		t.Errorf("Default TargetC = %f, want 0", opts.TargetC)
	}
	if opts.TargetCSet {
		t.Error("Default TargetCSet should be false")
	}
	if opts.Mode != "" {
		t.Errorf("Default Mode = %q, want empty", opts.Mode)
	}
	if opts.Enable {
		t.Error("Default Enable should be false")
	}
	if opts.Disable {
		t.Error("Default Disable should be false")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "test-device",
		Timespec:   "0 0 8 * *",
		TargetCSet: true,
		TargetC:    22.0,
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
		Factory:    tf.Factory,
		Device:     "test-device",
		Timespec:   "0 0 8 * *",
		TargetCSet: true,
		TargetC:    22.0,
	}

	err := run(ctx, opts)

	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestRun_NoActionSpecified(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Timespec: "0 0 8 * *",
		// No target, mode, enable, or disable set
	}

	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error when no action specified")
	}
	if !strings.Contains(err.Error(), "at least one of") {
		t.Errorf("Expected error about missing action, got: %v", err)
	}
}

func TestRun_InvalidMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Timespec: "0 0 8 * *",
		Mode:     "invalid_mode",
	}

	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "mode") && !strings.Contains(err.Error(), "invalid") {
		t.Logf("Error: %v", err)
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
			name:    "time flag",
			args:    []string{"--time", "0 0 8 * *"},
			wantErr: false,
		},
		{
			name:    "time flag short",
			args:    []string{"-t", "0 0 8 * *"},
			wantErr: false,
		},
		{
			name:    "target flag",
			args:    []string{"--target", "22"},
			wantErr: false,
		},
		{
			name:    "mode flag",
			args:    []string{"--mode", "heat"},
			wantErr: false,
		},
		{
			name:    "enable flag",
			args:    []string{"--enable"},
			wantErr: false,
		},
		{
			name:    "disable flag",
			args:    []string{"--disable"},
			wantErr: false,
		},
		{
			name:    "thermostat-id flag",
			args:    []string{"--thermostat-id", "1"},
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
		Factory:    tf.Factory,
		Device:     "gen1-device",
		Timespec:   "0 0 8 * *",
		TargetCSet: true,
		TargetC:    22.0,
	}

	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error for Gen1 device")
	}

	if !strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("Expected error mentioning Gen2+, got: %v", err)
	}
}

func TestRun_WithMockGen2Device_CreateWithTarget(t *testing.T) {
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
		Factory:    tf.Factory,
		Device:     "thermostat-device",
		Timespec:   "0 0 8 * 1-5",
		TargetCSet: true,
		TargetC:    22.0,
		Enabled:    true,
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device schedule create: %v", err)
	}
}

func TestRun_WithMockGen2Device_CreateWithMode(t *testing.T) {
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
		Timespec: "@sunrise",
		Mode:     "heat",
		Enabled:  true,
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device schedule create with mode: %v", err)
	}
}

func TestRun_WithMockGen2Device_CreateWithEnable(t *testing.T) {
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
		Timespec: "0 0 6 * *",
		Enable:   true,
		Enabled:  true,
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device schedule create with enable: %v", err)
	}
}

func TestRun_WithMockGen2Device_CreateWithDisable(t *testing.T) {
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
		Timespec: "0 0 0 * *",
		Disable:  true,
		Enabled:  true,
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device schedule create with disable: %v", err)
	}
}

func TestRun_WithMockGen2Device_CreateDisabledSchedule(t *testing.T) {
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
		Factory:    tf.Factory,
		Device:     "thermostat-device",
		Timespec:   "0 0 9 * *",
		TargetCSet: true,
		TargetC:    20.0,
		Enabled:    false, // Disabled schedule
	}

	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("Unexpected error for Gen2 device create disabled schedule: %v", err)
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

func TestNewCommand_ExecuteWithTargetAndTime(t *testing.T) {
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
	cmd.SetArgs([]string{"thermostat-device", "--target", "22", "--time", "0 0 8 * *"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}

func TestNewCommand_ExecuteWithModeAndTime(t *testing.T) {
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
	cmd.SetArgs([]string{"thermostat-device", "--mode", "heat", "--time", "@sunrise"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}
