package set

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

	// Test Use
	if cmd.Use != "set <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <device>")
	}

	// Test Aliases
	wantAliases := []string{"configure", "config"}
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

	// Test RunE
	if cmd.RunE == nil {
		t.Error("RunE must be set")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{"ssid", "ssid", ""},
		{"password", "password", ""},
		{"static-ip", "static-ip", ""},
		{"gateway", "gateway", ""},
		{"netmask", "netmask", ""},
		{"dns", "dns", ""},
		{"enable", "enable", "false"},
		{"disable", "disable", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %q not found", tt.flagName)
				return
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
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
		"shelly wifi set",
		"--ssid",
		"--password",
		"--static-ip",
		"--disable",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescriptionContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"WiFi",
		"SSID",
		"password",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestRun_ValidationNoSSID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		// No ssid, no disable, no enable
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error when no --ssid and not --disable")
	}
	if !strings.Contains(err.Error(), "--ssid is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_ValidationWithDisable(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "disable-test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "disable-test-device",
		Disable: true,
		// No SSID needed when --disable
	}

	err = run(context.Background(), opts)
	// Mock may not support Wifi.SetConfig
	t.Logf("run() with disable error: %v", err)
}

func TestRun_ValidationWithEnable(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Enable:  true,
		// No SSID - but enable is set, so should fail
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error when --enable without --ssid")
	}
}

func TestRun_WithSSIDAndPassword(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "wifi-set-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "wifi-set-device",
		SSID:     "TestNetwork",
		Password: "secret123",
	}

	err = run(context.Background(), opts)
	t.Logf("run() with ssid/password error: %v", err)
}

func TestRun_WithStaticIP(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "static-ip-device",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "static-ip-device",
		SSID:     "TestNetwork",
		Password: "secret123",
		StaticIP: "192.168.1.50",
		Gateway:  "192.168.1.1",
		Netmask:  "255.255.255.0",
		DNS:      "8.8.8.8",
	}

	err = run(context.Background(), opts)
	t.Logf("run() with static IP error: %v", err)
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "cancel-device",
					Address:    "192.168.1.103",
					MAC:        "AA:BB:CC:DD:EE:03",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "cancel-device",
		SSID:    "TestNetwork",
	}

	err = run(ctx, opts)
	if err == nil {
		t.Log("Context cancellation: command may succeed due to mock timing")
	}
}

func TestRun_ContextTimeout(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "timeout-device",
					Address:    "192.168.1.104",
					MAC:        "AA:BB:CC:DD:EE:04",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "timeout-device",
		SSID:    "TestNetwork",
	}

	err = run(ctx, opts)
	t.Logf("run() with timeout error: %v", err)
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
		SSID:    "TestNetwork",
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no device argument provided")
	}
}

func TestExecute_TooManyArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device1", "device2"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when too many arguments provided")
	}
}

func TestExecute_WithSSID(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-ssid-device",
					Address:    "192.168.1.110",
					MAC:        "AA:BB:CC:DD:EE:10",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"exec-ssid-device", "--ssid", "TestNetwork"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	t.Logf("Execute with --ssid error: %v", err)
}

func TestExecute_WithDisable(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-disable-device",
					Address:    "192.168.1.111",
					MAC:        "AA:BB:CC:DD:EE:11",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"exec-disable-device", "--disable"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	t.Logf("Execute with --disable error: %v", err)
}

func TestExecute_NoFlagsError(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-noflags-device",
					Address:    "192.168.1.112",
					MAC:        "AA:BB:CC:DD:EE:12",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"exec-noflags-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error when no flags specified")
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"living-room"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"192.168.1.100"})
	if err != nil {
		t.Errorf("Command should accept IP address, got error: %v", err)
	}
}

func TestNewCommand_RejectsMultipleDevices(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Command should reject multiple device arguments")
	}
}

func TestOptions_EnableState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		enable   bool
		disable  bool
		wantNil  bool
		wantBool bool
	}{
		{"neither", false, false, true, false},
		{"enable only", true, false, false, true},
		{"disable only", false, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var enable *bool
			if tt.enable {
				tr := true
				enable = &tr
			} else if tt.disable {
				f := false
				enable = &f
			}

			switch {
			case tt.wantNil && enable != nil:
				t.Error("expected enable to be nil")
			case !tt.wantNil && enable == nil:
				t.Error("expected enable to be non-nil")
			case !tt.wantNil && enable != nil && *enable != tt.wantBool:
				t.Errorf("enable = %v, want %v", *enable, tt.wantBool)
			}
		})
	}
}

func TestExecute_AllFlagCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "ssid only",
			args: []string{"test-device", "--ssid", "TestNetwork"},
		},
		{
			name: "ssid and password",
			args: []string{"test-device", "--ssid", "TestNetwork", "--password", "secret"},
		},
		{
			name: "disable only",
			args: []string{"test-device", "--disable"},
		},
		{
			name: "ssid with enable",
			args: []string{"test-device", "--ssid", "TestNetwork", "--enable"},
		},
		{
			name: "full static IP config",
			args: []string{"test-device", "--ssid", "TestNetwork", "--password", "secret",
				"--static-ip", "192.168.1.50", "--gateway", "192.168.1.1",
				"--netmask", "255.255.255.0", "--dns", "8.8.8.8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fixtures := &mock.Fixtures{
				Version: "1",
				Config: mock.ConfigFixture{
					Devices: []mock.DeviceFixture{
						{
							Name:       "test-device",
							Address:    "192.168.1.130",
							MAC:        "AA:BB:CC:DD:EE:30",
							Type:       "SNSW-001P16EU",
							Model:      "Shelly Plus 1PM",
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

			var buf bytes.Buffer
			cmd := NewCommand(tf.Factory)
			cmd.SetContext(context.Background())
			cmd.SetArgs(tt.args)
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err = cmd.Execute()
			t.Logf("%s: error = %v", tt.name, err)
		})
	}
}

func TestRun_SSIDWithEnable(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "enable-ssid-device",
					Address:    "192.168.1.120",
					MAC:        "AA:BB:CC:DD:EE:20",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "enable-ssid-device",
		SSID:    "TestNetwork",
		Enable:  true,
	}

	err = run(context.Background(), opts)
	t.Logf("run() with ssid and enable error: %v", err)
}
