package install

import (
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

	if cmd.Use != "install <device> <template>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "install <device> <template>")
	}

	wantAliases := []string{"add", "deploy"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}
	for i, want := range wantAliases {
		if cmd.Aliases[i] != want {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
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
		{"one arg", []string{"device"}, true},
		{"two args valid", []string{"device", "motion-light"}, false},
		{"three args", []string{"device", "template", "extra"}, true},
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

	// Check --configure flag
	configureFlag := cmd.Flags().Lookup("configure")
	if configureFlag == nil {
		t.Fatal("--configure flag not found")
	}
	if configureFlag.DefValue != "false" {
		t.Errorf("--configure default = %q, want %q", configureFlag.DefValue, "false")
	}

	// Check --enable flag
	enableFlag := cmd.Flags().Lookup("enable")
	if enableFlag == nil {
		t.Fatal("--enable flag not found")
	}
	if enableFlag.DefValue != "false" {
		t.Errorf("--enable default = %q, want %q", enableFlag.DefValue, "false")
	}

	// Check --name flag
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatal("--name flag not found")
	}
	if nameFlag.DefValue != "" {
		t.Errorf("--name default = %q, want empty", nameFlag.DefValue)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly script template install",
		"motion-light",
		"--configure",
		"--enable",
		"--name",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestRun_TemplateNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Device:   "test-device",
		Template: "nonexistent-template",
		Factory:  tf.Factory,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want to contain 'not found'", err)
	}
}

func TestRun_InstallsBuiltInTemplate(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {
				"switch:0": map[string]any{
					"output": false,
				},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:   "test-device",
		Template: "motion-light",
		Factory:  tf.Factory,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Installed template") {
		t.Errorf("output = %q, want to contain 'Installed template'", output)
	}
	if !strings.Contains(output, "motion-light") {
		t.Errorf("output = %q, want to contain 'motion-light'", output)
	}
}

func TestRun_InstallWithCustomName(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "my-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"my-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:   "my-device",
		Template: "power-monitor",
		Name:     "Custom Power Script",
		Factory:  tf.Factory,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Installed template") {
		t.Errorf("output = %q, want to contain 'Installed template'", output)
	}
}

func TestRun_InstallWithEnable(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "enable-device",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"enable-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:   "enable-device",
		Template: "toggle-sync",
		Enable:   true,
		Factory:  tf.Factory,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Installed template") {
		t.Errorf("output = %q, want to contain 'Installed template'", output)
	}
	if !strings.Contains(output, "Script enabled") {
		t.Errorf("output = %q, want to contain 'Script enabled'", output)
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{}, // No devices
		},
		DeviceStates: map[string]mock.DeviceState{},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:   "nonexistent-device",
		Template: "motion-light",
		Factory:  tf.Factory,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

func TestRun_AllBuiltInTemplates(t *testing.T) {
	t.Parallel()

	templates := []string{
		"motion-light",
		"power-monitor",
		"schedule-helper",
		"toggle-sync",
		"energy-logger",
	}

	for _, tplName := range templates {
		tt := tplName
		t.Run(tt, func(t *testing.T) {
			t.Parallel()

			fixtures := &mock.Fixtures{
				Config: mock.ConfigFixture{
					Devices: []mock.DeviceFixture{
						{
							Name:       "template-test-device",
							Address:    "192.168.1.200",
							MAC:        "AA:BB:CC:DD:EE:FF",
							Model:      "SNSW-001P16EU",
							Type:       "Plus1PM",
							Generation: 2,
						},
					},
				},
				DeviceStates: map[string]mock.DeviceState{
					"template-test-device": {},
				},
			}

			demo, err := mock.StartWithFixtures(fixtures)
			if err != nil {
				t.Fatalf("failed to start demo: %v", err)
			}
			defer demo.Cleanup()

			tf := factory.NewTestFactory(t)
			demo.InjectIntoFactory(tf.Factory)

			opts := &Options{
				Device:   "template-test-device",
				Template: tt,
				Factory:  tf.Factory,
			}

			err = run(context.Background(), opts)
			if err != nil {
				t.Fatalf("run() error = %v", err)
			}

			output := tf.OutString()
			if !strings.Contains(output, "Installed template") {
				t.Errorf("output = %q, want to contain 'Installed template'", output)
			}
			if !strings.Contains(output, tt) {
				t.Errorf("output = %q, want to contain %q", output, tt)
			}
		})
	}
}
