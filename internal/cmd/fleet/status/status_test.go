package status

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
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

	expectedAliases := []string{"st", "list", "ls"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Fatalf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check online flag
	onlineFlag := cmd.Flags().Lookup("online")
	if onlineFlag == nil {
		t.Fatal("online flag not found")
	}
	if onlineFlag.DefValue != "false" {
		t.Errorf("online default = %q, want %q", onlineFlag.DefValue, "false")
	}

	// Check offline flag
	offlineFlag := cmd.Flags().Lookup("offline")
	if offlineFlag == nil {
		t.Fatal("offline flag not found")
	}
	if offlineFlag.DefValue != "false" {
		t.Errorf("offline default = %q, want %q", offlineFlag.DefValue, "false")
	}
}

func TestNewCommand_OnlineFlagParsing(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting the online flag
	if err := cmd.Flags().Set("online", "true"); err != nil {
		t.Fatalf("failed to set online flag: %v", err)
	}

	val, err := cmd.Flags().GetBool("online")
	if err != nil {
		t.Fatalf("failed to get online value: %v", err)
	}

	if !val {
		t.Error("online = false, want true")
	}
}

func TestNewCommand_OfflineFlagParsing(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test setting the offline flag
	if err := cmd.Flags().Set("offline", "true"); err != nil {
		t.Fatalf("failed to set offline flag: %v", err)
	}

	val, err := cmd.Flags().GetBool("offline")
	if err != nil {
		t.Fatalf("failed to get offline value: %v", err)
	}

	if !val {
		t.Error("offline = false, want true")
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// status takes no args
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("expected no error for zero args, got: %v", err)
	}

	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("expected error for args provided, got nil")
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
	// The status command uses GetIntegratorCredentials which returns a different error message
	if !contains(errStr, "credentials") && !contains(errStr, "connect") {
		t.Errorf("error = %q, want to contain 'credentials' or 'connect'", errStr)
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
	if !contains(errStr, "authentication failed") && !contains(errStr, "failed") {
		t.Errorf("error = %q, want to contain 'authentication failed' or 'failed'", errStr)
	}
}

func TestRun_WithOnlineFlag(t *testing.T) {
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
	cmd.SetArgs([]string{"--online"})
	err := cmd.Execute()

	// Will fail auth, but flag parsing should work
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	// Verify flag was parsed by checking no parse error
	errStr := err.Error()
	if contains(errStr, "invalid") && contains(errStr, "online") {
		t.Errorf("online flag parsing failed: %v", err)
	}
}

func TestRun_WithOfflineFlag(t *testing.T) {
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
	cmd.SetArgs([]string{"--offline"})
	err := cmd.Execute()

	// Will fail auth, but flag parsing should work
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	// Verify flag was parsed by checking no parse error
	errStr := err.Error()
	if contains(errStr, "invalid") && contains(errStr, "offline") {
		t.Errorf("offline flag parsing failed: %v", err)
	}
}

func TestNewCommand_WithFactory(t *testing.T) {
	t.Parallel()

	// Test that factory is properly wired
	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"test-device": {Name: "test-device", Address: "10.0.0.1"},
	})

	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with factory")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	// Verify Options struct with defaults
	opts := &Options{}

	if opts.Online {
		t.Error("default Online = true, want false")
	}

	if opts.Offline {
		t.Error("default Offline = true, want false")
	}
}

func TestOptions_MutuallyExclusiveFlags(t *testing.T) {
	t.Parallel()

	// Options allows both flags to be set; filtering handles it
	opts := &Options{
		Online:  true,
		Offline: true,
	}

	// Both can be true at the struct level
	// The run function filters appropriately
	if !opts.Online || !opts.Offline {
		t.Error("expected both flags to be settable")
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRun_DemoMode_NoFixtures(t *testing.T) {
	t.Parallel()

	// Clear any existing demo state
	if err := os.Unsetenv("SHELLY_DEMO"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TAG"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}
	if err := os.Unsetenv("SHELLY_INTEGRATOR_TOKEN"); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	tf := factory.NewTestFactory(t)

	// Ensure config has no integrator credentials
	tf.Config.Integrator = config.IntegratorConfig{}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	// Without demo mode or credentials, this should fail
	if err == nil {
		t.Fatal("expected error for missing credentials without demo mode, got nil")
	}
}

func TestRun_BothOnlineAndOfflineFlags(t *testing.T) {
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
	cmd.SetArgs([]string{"--online", "--offline"})
	err := cmd.Execute()

	// Will fail auth, but flag parsing should work
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}

	// Verify no flag parse error (both flags can be set)
	errStr := err.Error()
	if contains(errStr, "invalid") && (contains(errStr, "online") || contains(errStr, "offline")) {
		t.Errorf("flag parsing failed: %v", err)
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		Online:  true,
		Offline: true,
	}

	if opts.Factory != f {
		t.Error("Factory not set correctly")
	}

	if !opts.Online {
		t.Error("Online should be true")
	}

	if !opts.Offline {
		t.Error("Offline should be true")
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse without setting flags
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	online, err := cmd.Flags().GetBool("online")
	if err != nil {
		t.Fatalf("get online flag: %v", err)
	}
	if online {
		t.Error("online should default to false")
	}

	offline, err := cmd.Flags().GetBool("offline")
	if err != nil {
		t.Fatalf("get offline flag: %v", err)
	}
	if offline {
		t.Error("offline should default to false")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}

	if !contains(cmd.Short, "status") {
		t.Error("Short should contain 'status'")
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}

	if !contains(cmd.Long, "fleet") {
		t.Error("Long should contain 'fleet'")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}

	if !contains(cmd.Example, "shelly fleet status") {
		t.Error("Example should contain 'shelly fleet status'")
	}

	if !contains(cmd.Example, "--online") {
		t.Error("Example should contain '--online'")
	}

	if !contains(cmd.Example, "--offline") {
		t.Error("Example should contain '--offline'")
	}

	if !contains(cmd.Example, "-o json") {
		t.Error("Example should contain '-o json'")
	}
}

func TestNewCommand_AllAliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, alias := range cmd.Aliases {
		aliasMap[alias] = true
	}

	expected := []string{"st", "list", "ls"}
	for _, e := range expected {
		if !aliasMap[e] {
			t.Errorf("missing alias %q", e)
		}
	}
}

func TestRun_ConfigLoadError(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	// Force config error by clearing credentials
	tf.Config.Integrator = config.IntegratorConfig{}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Should fail with credentials error
	errStr := err.Error()
	if !contains(errStr, "credentials") && !contains(errStr, "connect") {
		t.Errorf("expected credentials error, got: %s", errStr)
	}
}

// createFleetFixtures creates test fixtures with fleet devices.
func createFleetFixtures() *mock.Fixtures {
	return &mock.Fixtures{
		Version: "1",
		Fleet: mock.FleetFixture{
			Organization: "Test Org",
			Devices: []mock.FleetDeviceFixture{
				{ID: "device-1", Name: "Online Device 1", Model: "SNSW-001P16EU", Online: true, Firmware: "1.4.4"},
				{ID: "device-2", Name: "Online Device 2", Model: "SNSW-102P16EU", Online: true, Firmware: "1.4.3"},
				{ID: "device-3", Name: "Offline Device", Model: "SNSW-001P16EU", Online: false, Firmware: "1.4.2"},
			},
		},
	}
}

func TestRun_DemoMode_Success(t *testing.T) {
	// Set up demo mode environment
	t.Setenv("SHELLY_DEMO", "1")

	// Start demo mode with fleet fixtures
	fixtures := createFleetFixtures()
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Fleet Status") {
		t.Errorf("Output = %q, want to contain 'Fleet Status'", output)
	}
	if !strings.Contains(output, "3 devices") {
		t.Errorf("Output = %q, want to contain '3 devices'", output)
	}
}

func TestRun_DemoMode_ShowsOrganization(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")

	fixtures := createFleetFixtures()
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Test Org") {
		t.Errorf("Output = %q, want to contain organization 'Test Org'", output)
	}
}

func TestRun_DemoMode_OnlineFilter(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")

	fixtures := createFleetFixtures()
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--online"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Should only show 2 online devices
	if !strings.Contains(output, "2 devices") {
		t.Errorf("Output = %q, want to contain '2 devices' (online only)", output)
	}
}
func TestRun_DemoMode_OfflineFilter(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")

	fixtures := createFleetFixtures()
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--offline"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Should only show 1 offline device
	if !strings.Contains(output, "1 devices") {
		t.Errorf("Output = %q, want to contain '1 devices' (offline only)", output)
	}
}
func TestRun_DemoMode_BothFilters_NoDevices(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")

	fixtures := createFleetFixtures()
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	// Both filters together - should filter out everything
	cmd.SetArgs([]string{"--online", "--offline"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Warning goes to stderr
	errOutput := tf.ErrString()
	// When both filters are set, nothing matches (can't be both online and offline)
	if !strings.Contains(errOutput, "No devices found") {
		t.Errorf("ErrOutput = %q, want to contain 'No devices found'", errOutput)
	}
}
func TestRun_DemoMode_AllOnline_OfflineFilter(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")

	// Create fixtures with only online devices
	fixtures := &mock.Fixtures{
		Version: "1",
		Fleet: mock.FleetFixture{
			Organization: "All Online Org",
			Devices: []mock.FleetDeviceFixture{
				{ID: "device-1", Name: "Device 1", Model: "SNSW-001P16EU", Online: true},
				{ID: "device-2", Name: "Device 2", Model: "SNSW-001P16EU", Online: true},
			},
		},
	}
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	// Filter by offline - should find nothing since all are online
	cmd.SetArgs([]string{"--offline"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Warning goes to stderr
	errOutput := tf.ErrString()
	if !strings.Contains(errOutput, "No devices found") {
		t.Errorf("ErrOutput = %q, want to contain 'No devices found'", errOutput)
	}
}
func TestRun_DemoMode_NoOrganization(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")

	// Create fixtures without organization name
	fixtures := &mock.Fixtures{
		Version: "1",
		Fleet: mock.FleetFixture{
			Organization: "", // Empty organization
			Devices: []mock.FleetDeviceFixture{
				{ID: "device-1", Name: "Device 1", Model: "SNSW-001P16EU", Online: true},
			},
		},
	}
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Should not print "Organization:" when empty
	if strings.Contains(output, "Organization:") {
		t.Errorf("Output should not contain 'Organization:' when empty, got: %s", output)
	}
	// Should still show device count
	if !strings.Contains(output, "1 devices") {
		t.Errorf("Output = %q, want to contain '1 devices'", output)
	}
}

func TestRun_DemoMode_JSONOutput(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")
	viper.Set("output", "json")
	defer viper.Set("output", "")

	fixtures := createFleetFixtures()
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := tf.OutString()
	// JSON output should contain device data in JSON format
	if !strings.Contains(out, "Online Device 1") {
		t.Errorf("JSON output should contain 'Online Device 1', got: %s", out)
	}
	if !strings.Contains(out, `"online"`) {
		t.Errorf("JSON output should contain 'online' field, got: %s", out)
	}
}

func TestRun_DemoMode_JSONOutput_FilteredEmpty(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")
	viper.Set("output", "json")
	defer viper.Set("output", "")

	// Fleet with only offline devices, filter by online = empty result
	fixtures := &mock.Fixtures{
		Version: "1",
		Fleet: mock.FleetFixture{
			Organization: "Offline Only Org",
			Devices: []mock.FleetDeviceFixture{
				{ID: "device-1", Name: "Offline Device", Model: "SNSW-001P16EU", Online: false},
			},
		},
	}
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--online"}) // Filter by online, but all devices are offline

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := tf.OutString()
	// Empty JSON array after filtering
	if !strings.Contains(out, "[]") {
		t.Errorf("JSON output should be empty array '[]', got: %s", out)
	}
}

func TestRun_DemoMode_JSONOutput_WithFilter(t *testing.T) {
	t.Setenv("SHELLY_DEMO", "1")
	viper.Set("output", "json")
	defer viper.Set("output", "")

	fixtures := createFleetFixtures()
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures error: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--online"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := tf.OutString()
	// JSON output should contain online devices only
	if !strings.Contains(out, "Online Device 1") {
		t.Errorf("JSON output should contain 'Online Device 1' (online), got: %s", out)
	}
	// Should not contain offline device
	if strings.Contains(out, "Offline Device") {
		t.Errorf("JSON output should NOT contain 'Offline Device' (offline), got: %s", out)
	}
}
