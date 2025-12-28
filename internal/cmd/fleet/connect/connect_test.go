package connect

import (
	"os"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "connect" {
		t.Errorf("Use = %q, want %q", cmd.Use, "connect")
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

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"login", "auth"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}

	for i, alias := range expectedAliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
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
		{
			name:    "no args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg",
			args:    []string{"arg1"},
			wantErr: true,
		},
		{
			name:    "multiple args",
			args:    []string{"arg1", "arg2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
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
		{
			name:     "host flag exists",
			flagName: "host",
			defValue: "",
		},
		{
			name:     "region flag exists",
			flagName: "region",
			defValue: "",
		},
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

func TestNewCommand_HostFlagParsing(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.Flags().Set("host", "shelly-13-eu.shelly.cloud"); err != nil {
		t.Fatalf("failed to set host flag: %v", err)
	}

	val, err := cmd.Flags().GetString("host")
	if err != nil {
		t.Fatalf("failed to get host value: %v", err)
	}

	if val != "shelly-13-eu.shelly.cloud" {
		t.Errorf("host = %q, want %q", val, "shelly-13-eu.shelly.cloud")
	}
}

func TestNewCommand_RegionFlagParsing(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.Flags().Set("region", "eu"); err != nil {
		t.Fatalf("failed to set region flag: %v", err)
	}

	val, err := cmd.Flags().GetString("region")
	if err != nil {
		t.Fatalf("failed to get region value: %v", err)
	}

	if val != "eu" {
		t.Errorf("region = %q, want %q", val, "eu")
	}
}

func TestNewCommand_RunEExists(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is not set")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "SHELLY_INTEGRATOR_TAG") {
		t.Error("Long should contain 'SHELLY_INTEGRATOR_TAG'")
	}

	if !strings.Contains(cmd.Long, "SHELLY_INTEGRATOR_TOKEN") {
		t.Error("Long should contain 'SHELLY_INTEGRATOR_TOKEN'")
	}

	if !strings.Contains(cmd.Long, "EU and US") {
		t.Error("Long should contain 'EU and US'")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "shelly fleet connect") {
		t.Error("Example should contain 'shelly fleet connect'")
	}

	if !strings.Contains(cmd.Example, "--host") {
		t.Error("Example should contain '--host'")
	}

	if !strings.Contains(cmd.Example, "--region") {
		t.Error("Example should contain '--region'")
	}
}

func TestCloudHosts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		region         string
		wantHosts      []string
		wantHostsExist bool
	}{
		{
			name:           "eu region has hosts",
			region:         "eu",
			wantHosts:      []string{"shelly-13-eu.shelly.cloud", "shelly-14-eu.shelly.cloud"},
			wantHostsExist: true,
		},
		{
			name:           "us region has hosts",
			region:         "us",
			wantHosts:      []string{"shelly-15-us.shelly.cloud", "shelly-16-us.shelly.cloud"},
			wantHostsExist: true,
		},
		{
			name:           "invalid region returns empty",
			region:         "invalid",
			wantHosts:      nil,
			wantHostsExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hosts, ok := cloudHosts[tt.region]
			if ok != tt.wantHostsExist {
				t.Errorf("cloudHosts[%q] exists = %v, want %v", tt.region, ok, tt.wantHostsExist)
			}
			if tt.wantHostsExist {
				if len(hosts) != len(tt.wantHosts) {
					t.Errorf("cloudHosts[%q] = %v, want %v", tt.region, hosts, tt.wantHosts)
					return
				}
				for i, host := range tt.wantHosts {
					if hosts[i] != host {
						t.Errorf("cloudHosts[%q][%d] = %q, want %q", tt.region, i, hosts[i], host)
					}
				}
			}
		})
	}
}

func TestCloudHosts_AllRegionsHaveHosts(t *testing.T) {
	t.Parallel()

	if len(cloudHosts) == 0 {
		t.Error("cloudHosts is empty")
	}

	for region, hosts := range cloudHosts {
		if len(hosts) == 0 {
			t.Errorf("region %q has no hosts", region)
		}
		for i, host := range hosts {
			if host == "" {
				t.Errorf("region %q host[%d] is empty", region, i)
			}
		}
	}
}

func TestCloudHosts_ExpectedRegions(t *testing.T) {
	t.Parallel()

	expectedRegions := []string{"eu", "us"}
	for _, region := range expectedRegions {
		if _, ok := cloudHosts[region]; !ok {
			t.Errorf("expected region %q not found in cloudHosts", region)
		}
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Host != "" {
		t.Errorf("Options.Host default = %q, want empty string", opts.Host)
	}

	if opts.Region != "" {
		t.Errorf("Options.Region default = %q, want empty string", opts.Region)
	}
}

func TestRun_MissingCredentials_NoEnv(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Ensure config has no integrator credentials
	tf.Config.Integrator = config.IntegratorConfig{}

	// Clear env vars
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}

	errStr := err.Error()
	// The connect command should return a credentials error
	if !strings.Contains(errStr, "credentials") && !strings.Contains(errStr, "integrator") {
		t.Errorf("error = %q, want to contain 'credentials' or 'integrator'", errStr)
	}
}

func TestRun_WithConfigCredentials_ButNoServer(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Set valid credentials in config
	tf.Config.Integrator = config.IntegratorConfig{
		Tag:   "test-tag",
		Token: "test-token",
	}

	// Clear env vars to test config fallback
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	// Should fail with authentication error (no real server)
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	errStr := err.Error()
	// This will fail authentication since credentials are fake
	if !strings.Contains(errStr, "authentication") && !strings.Contains(errStr, "failed") {
		t.Errorf("error = %q, want to contain 'authentication' or 'failed'", errStr)
	}
}

func TestRun_WithHostFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Set valid credentials in config
	tf.Config.Integrator = config.IntegratorConfig{
		Tag:   "test-tag",
		Token: "test-token",
	}

	// Clear env vars
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--host", "shelly-13-eu.shelly.cloud"})
	err := cmd.Execute()

	// Will fail auth, but flag parsing should work
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	// Verify flag was parsed by checking no parse error
	errStr := err.Error()
	if strings.Contains(errStr, "invalid") && strings.Contains(errStr, "host") {
		t.Errorf("host flag parsing failed: %v", err)
	}
}

func TestRun_WithRegionFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Set valid credentials in config
	tf.Config.Integrator = config.IntegratorConfig{
		Tag:   "test-tag",
		Token: "test-token",
	}

	// Clear env vars
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--region", "eu"})
	err := cmd.Execute()

	// Will fail auth, but flag parsing should work
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	// Verify flag was parsed by checking no parse error
	errStr := err.Error()
	if strings.Contains(errStr, "invalid") && strings.Contains(errStr, "region") {
		t.Errorf("region flag parsing failed: %v", err)
	}
}

func TestNewCommand_WithFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with factory")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_HasLocalFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Connect command should have --host and --region flags
	if !cmd.HasLocalFlags() {
		t.Error("connect command should have local flags")
	}

	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Error("host flag not found")
	}

	regionFlag := cmd.Flags().Lookup("region")
	if regionFlag == nil {
		t.Error("region flag not found")
	}
}

func TestOptions_HostField(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Host: "shelly-13-eu.shelly.cloud",
	}

	if opts.Host != "shelly-13-eu.shelly.cloud" {
		t.Errorf("Options.Host = %q, want %q", opts.Host, "shelly-13-eu.shelly.cloud")
	}
}

func TestOptions_RegionField(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Region: "eu",
	}

	if opts.Region != "eu" {
		t.Errorf("Options.Region = %q, want %q", opts.Region, "eu")
	}
}

func TestOptions_BothFields(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Host:   "custom.host.cloud",
		Region: "us",
	}

	if opts.Host != "custom.host.cloud" {
		t.Errorf("Options.Host = %q, want %q", opts.Host, "custom.host.cloud")
	}

	if opts.Region != "us" {
		t.Errorf("Options.Region = %q, want %q", opts.Region, "us")
	}
}

func TestCloudHosts_EUHasCorrectCount(t *testing.T) {
	t.Parallel()

	hosts := cloudHosts["eu"]
	if len(hosts) != 2 {
		t.Errorf("EU region has %d hosts, want 2", len(hosts))
	}
}

func TestCloudHosts_USHasCorrectCount(t *testing.T) {
	t.Parallel()

	hosts := cloudHosts["us"]
	if len(hosts) != 2 {
		t.Errorf("US region has %d hosts, want 2", len(hosts))
	}
}

func TestCloudHosts_TotalHostCount(t *testing.T) {
	t.Parallel()

	var total int
	for _, hosts := range cloudHosts {
		total += len(hosts)
	}

	if total != 4 {
		t.Errorf("total hosts = %d, want 4", total)
	}
}

func TestNewCommand_FlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Fatal("host flag not found")
	}
	if hostFlag.Usage == "" {
		t.Error("host flag usage is empty")
	}

	regionFlag := cmd.Flags().Lookup("region")
	if regionFlag == nil {
		t.Fatal("region flag not found")
	}
	if regionFlag.Usage == "" {
		t.Error("region flag usage is empty")
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short != "Connect to Shelly Cloud hosts" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Connect to Shelly Cloud hosts")
	}
}

func TestNewCommand_FlagParsingCombined(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse multiple flags at once
	if err := cmd.ParseFlags([]string{"--host", "test.cloud", "--region", "eu"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	host, err := cmd.Flags().GetString("host")
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}
	if host != "test.cloud" {
		t.Errorf("host = %q, want %q", host, "test.cloud")
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		t.Fatalf("failed to get region: %v", err)
	}
	if region != "eu" {
		t.Errorf("region = %q, want %q", region, "eu")
	}
}
