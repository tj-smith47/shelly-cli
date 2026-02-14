package provision

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

const testSSID = "MyNetwork"

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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "provision" {
		t.Errorf("Use = %q, want %q", cmd.Use, "provision")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := map[string]bool{"prov": true, "setup": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias %q", alias)
		}
		delete(expectedAliases, alias)
	}
	for alias := range expectedAliases {
		t.Errorf("missing alias %q", alias)
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil — parent provision command should have RunE for auto-discover flow")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{"ssid", "ssid", ""},
		{"password", "password", ""},
		{"timeout", "timeout", "30s"},
		{"name", "name", ""},
		{"timezone", "timezone", ""},
		{"ble-only", "ble-only", "false"},
		{"ap-only", "ap-only", "false"},
		{"from-device", "from-device", ""},
		{"from-template", "from-template", ""},
		{"no-cloud", "no-cloud", "false"},
		{"yes", "yes", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_YesFlagShorthand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("yes flag not found")
	}
	if flag.Shorthand != "y" {
		t.Errorf("yes flag shorthand = %q, want %q", flag.Shorthand, "y")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"ssid", []string{"--ssid", testSSID}},
		{"password", []string{"--password", "secret"}},
		{"timeout", []string{"--timeout", "60s"}},
		{"name", []string{"--name", "my-device"}},
		{"timezone", []string{"--timezone", "America/Chicago"}},
		{"ble-only", []string{"--ble-only"}},
		{"ap-only", []string{"--ap-only"}},
		{"from-device", []string{"--from-device", "my-device"}},
		{"from-template", []string{"--from-template", "my-template"}},
		{"no-cloud", []string{"--no-cloud"}},
		{"yes long", []string{"--yes"}},
		{"yes short", []string{"-y"}},
		{"combined", []string{"--ssid", "Net", "--password", "pass", "--ble-only", "-y"}},
		{"from-device with ap-only", []string{"--from-device", "fl", "--ap-only", "-y"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("ParseFlags() error = %v", err)
			}
		})
	}
}

func TestNewCommand_HasSubcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	subcommands := cmd.Commands()
	if len(subcommands) < 3 {
		t.Errorf("expected at least 3 subcommands, got %d", len(subcommands))
	}

	subNames := make(map[string]bool)
	for _, sub := range subcommands {
		subNames[sub.Name()] = true
	}

	for _, name := range []string{"wifi", "ble", "bulk"} {
		if !subNames[name] {
			t.Errorf("%s subcommand not found", name)
		}
	}
}

func TestNewCommand_SubcommandsHaveCorrectParent(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	for _, sub := range cmd.Commands() {
		if sub.Parent() != cmd {
			t.Errorf("subcommand %q has incorrect parent", sub.Name())
		}
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	keywords := []string{"BLE", "WiFi AP", "Gen1", "Gen2"}
	for _, kw := range keywords {
		if !strings.Contains(cmd.Long, kw) {
			t.Errorf("Long should contain %q", kw)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	patterns := []string{"shelly provision", "--ssid", "--ble-only", "--from-device", "--from-template"}
	for _, p := range patterns {
		if !strings.Contains(cmd.Example, p) {
			t.Errorf("Example should contain %q", p)
		}
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use == "" {
		t.Error("Use is empty")
	}
	if cmd.Short == "" {
		t.Error("Short is empty")
	}
	if cmd.Long == "" {
		t.Error("Long is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}
	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestOptions_BuildOnboardOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		SSID:       testSSID,
		Password:   "secret",
		Timezone:   "UTC",
		DeviceName: "my-dev",
		Timeout:    45 * time.Second,
		BLEOnly:    true,
		NoCloud:    true,
	}

	result := opts.buildOnboardOptions()

	if result.Timezone != "UTC" {
		t.Errorf("Timezone = %q, want %q", result.Timezone, "UTC")
	}
	if result.DeviceName != "my-dev" {
		t.Errorf("DeviceName = %q, want %q", result.DeviceName, "my-dev")
	}
	if result.Timeout != 45*time.Second {
		t.Errorf("Timeout = %v, want %v", result.Timeout, 45*time.Second)
	}
	if !result.BLEOnly {
		t.Error("BLEOnly should be true")
	}
	if !result.NoCloud {
		t.Error("NoCloud should be true")
	}
	if result.WiFi == nil {
		t.Fatal("WiFi should not be nil")
	}
	if result.WiFi.SSID != testSSID {
		t.Errorf("WiFi.SSID = %q, want %q", result.WiFi.SSID, testSSID)
	}
}

func TestOptions_BuildOnboardOptions_NoSSID(t *testing.T) {
	t.Parallel()

	opts := &Options{}
	result := opts.buildOnboardOptions()

	if result.WiFi != nil {
		t.Error("WiFi should be nil when no SSID provided")
	}
}

func TestOptions_PromptWiFiCredentials_AlreadySet(t *testing.T) {
	t.Parallel()

	opts := &Options{Factory: cmdutil.NewFactory(), SSID: testSSID}
	err := opts.promptWiFiCredentials(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if opts.SSID != testSSID {
		t.Errorf("SSID = %q, want %q", opts.SSID, testSSID)
	}
}

func TestNewCommand_FromDeviceAndTemplateMutuallyExclusive(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	cmd.SetArgs([]string{"--from-device", "fl", "--from-template", "my-tpl"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for mutually exclusive flags, got nil")
	}
	if !strings.Contains(err.Error(), "from-device") || !strings.Contains(err.Error(), "from-template") {
		t.Errorf("error should mention both flags, got: %v", err)
	}
}

func TestNewCommand_FromDeviceFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{"--from-device", "fl"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}
	val, err := cmd.Flags().GetString("from-device")
	if err != nil {
		t.Fatalf("GetString error: %v", err)
	}
	if val != "fl" {
		t.Errorf("from-device = %q, want %q", val, "fl")
	}
}

func TestNewCommand_FromTemplateFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{"--from-template", "my-tpl"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}
	val, err := cmd.Flags().GetString("from-template")
	if err != nil {
		t.Fatalf("GetString error: %v", err)
	}
	if val != "my-tpl" {
		t.Errorf("from-template = %q, want %q", val, "my-tpl")
	}
}
