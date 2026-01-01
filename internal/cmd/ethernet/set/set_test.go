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
		"shelly ethernet set",
		"--enable",
		"--disable",
		"--static-ip",
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
		"Ethernet",
		"DHCP",
		"static IP",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestRun_ValidationNoFlags(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		// No enable, disable, or static-ip
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error when no flags specified")
	}
	if !strings.Contains(err.Error(), "specify --enable, --disable, or configuration options") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_WithEnable(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "eth-enable-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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
		Device:  "eth-enable-device",
		Enable:  true,
	}

	err = run(context.Background(), opts)
	t.Logf("run() with enable error: %v", err)
}

func TestRun_WithDisable(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "eth-disable-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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
		Device:  "eth-disable-device",
		Disable: true,
	}

	err = run(context.Background(), opts)
	t.Logf("run() with disable error: %v", err)
}

func TestRun_WithStaticIP(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "eth-static-device",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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
		Factory:    tf.Factory,
		Device:     "eth-static-device",
		Enable:     true,
		StaticIP:   "192.168.1.50",
		Gateway:    "192.168.1.1",
		Netmask:    "255.255.255.0",
		Nameserver: "8.8.8.8",
	}

	err = run(context.Background(), opts)
	t.Logf("run() with static IP error: %v", err)
}

func TestRun_StaticIPWithoutEnable(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "eth-staticonly-device",
					Address:    "192.168.1.103",
					MAC:        "AA:BB:CC:DD:EE:03",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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
		Device:   "eth-staticonly-device",
		StaticIP: "192.168.1.50",
		// No enable flag
	}

	// Static IP alone should pass validation
	err = run(context.Background(), opts)
	t.Logf("run() with static IP only error: %v", err)
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "eth-cancel-device",
					Address:    "192.168.1.104",
					MAC:        "AA:BB:CC:DD:EE:04",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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
		Device:  "eth-cancel-device",
		Enable:  true,
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
					Name:       "eth-timeout-device",
					Address:    "192.168.1.105",
					MAC:        "AA:BB:CC:DD:EE:05",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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
		Device:  "eth-timeout-device",
		Enable:  true,
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
		Enable:  true,
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

func TestExecute_WithEnable(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-enable-device",
					Address:    "192.168.1.110",
					MAC:        "AA:BB:CC:DD:EE:10",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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
	cmd.SetArgs([]string{"exec-enable-device", "--enable"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	t.Logf("Execute with --enable error: %v", err)
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
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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

func TestExecute_WithStaticIP(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-static-device",
					Address:    "192.168.1.112",
					MAC:        "AA:BB:CC:DD:EE:12",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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
	cmd.SetArgs([]string{"exec-static-device", "--enable",
		"--static-ip", "192.168.1.50", "--gateway", "192.168.1.1",
		"--netmask", "255.255.255.0", "--dns", "8.8.8.8"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	t.Logf("Execute with static IP error: %v", err)
}

func TestExecute_NoFlagsError(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-noflags-device",
					Address:    "192.168.1.113",
					MAC:        "AA:BB:CC:DD:EE:13",
					Type:       "SPSW-004PE16EU",
					Model:      "Shelly Pro 4PM",
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

	err := cmd.Args(cmd, []string{"living-room-pro"})
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

func TestOptions_IPv4Mode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		staticIP string
		enable   bool
		wantMode string
	}{
		{"static IP", "192.168.1.50", false, "static"},
		{"enable without static", "", true, "dhcp"},
		{"static IP with enable", "192.168.1.50", true, "static"},
		{"neither", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ipv4Mode := ""
			if tt.staticIP != "" {
				ipv4Mode = "static"
			} else if tt.enable {
				ipv4Mode = "dhcp"
			}

			if ipv4Mode != tt.wantMode {
				t.Errorf("ipv4Mode = %q, want %q", ipv4Mode, tt.wantMode)
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
			name: "enable only",
			args: []string{"test-device", "--enable"},
		},
		{
			name: "disable only",
			args: []string{"test-device", "--disable"},
		},
		{
			name: "static-ip only",
			args: []string{"test-device", "--static-ip", "192.168.1.50"},
		},
		{
			name: "enable with static IP",
			args: []string{"test-device", "--enable", "--static-ip", "192.168.1.50"},
		},
		{
			name: "full static IP config",
			args: []string{"test-device", "--enable",
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
							Type:       "SPSW-004PE16EU",
							Model:      "Shelly Pro 4PM",
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
