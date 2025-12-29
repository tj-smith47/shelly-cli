package cmdutil_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const testDeviceAddress = "testDeviceAddress"

func TestNew(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	if f == nil {
		t.Fatal("New() returned nil")
	}

	// Verify functions are set
	if f.IOStreams == nil {
		t.Error("IOStreams function is nil")
	}
	if f.Config == nil {
		t.Error("Config function is nil")
	}
	if f.ShellyService == nil {
		t.Error("ShellyService function is nil")
	}
}

func TestFactory_IOStreams_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// First call should initialize
	ios1 := f.IOStreams()
	if ios1 == nil {
		t.Fatal("IOStreams() returned nil")
	}

	// Second call should return same instance
	ios2 := f.IOStreams()
	if ios1 != ios2 {
		t.Error("IOStreams() should return cached instance")
	}
}

func TestFactory_ShellyService_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// First call should initialize
	svc1 := f.ShellyService()
	if svc1 == nil {
		t.Fatal("ShellyService() returned nil")
	}

	// Second call should return same instance
	svc2 := f.ShellyService()
	if svc1 != svc2 {
		t.Error("ShellyService() should return cached instance")
	}
}

func TestNewWithIOStreams(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewWithIOStreams(ios)
	if f == nil {
		t.Fatal("NewWithIOStreams() returned nil")
	}

	// Should return the custom IOStreams
	gotIOS := f.IOStreams()
	if gotIOS != ios {
		t.Error("NewWithIOStreams should use provided IOStreams")
	}
}

func TestFactory_SetIOStreams(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Create custom IOStreams
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	// Set custom IOStreams
	result := f.SetIOStreams(ios)
	if result != f {
		t.Error("SetIOStreams should return factory for chaining")
	}

	// Should return the custom IOStreams
	gotIOS := f.IOStreams()
	if gotIOS != ios {
		t.Error("SetIOStreams should set the IOStreams")
	}
}

func TestFactory_SetConfigManager(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Create custom config manager
	cfg := &config.Config{}
	mgr := config.NewTestManager(cfg)

	// Set custom config manager
	result := f.SetConfigManager(mgr)
	if result != f {
		t.Error("SetConfigManager should return factory for chaining")
	}

	// Should return the custom config
	gotCfg, err := f.Config()
	if err != nil {
		t.Fatalf("Config() returned error: %v", err)
	}
	if gotCfg != cfg {
		t.Error("SetConfigManager should set the config")
	}
}

func TestFactory_SetShellyService(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Create custom service
	svc := shelly.NewService()

	// Set custom service
	result := f.SetShellyService(svc)
	if result != f {
		t.Error("SetShellyService should return factory for chaining")
	}

	// Should return the custom service
	gotSvc := f.ShellyService()
	if gotSvc != svc {
		t.Error("SetShellyService should set the service")
	}
}

func TestFactory_Chaining(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	cfg := &config.Config{}
	mgr := config.NewTestManager(cfg)
	svc := shelly.NewService()

	f := cmdutil.NewFactory().
		SetIOStreams(ios).
		SetConfigManager(mgr).
		SetShellyService(svc)

	if f.IOStreams() != ios {
		t.Error("Chained SetIOStreams failed")
	}
	gotCfg, err := f.Config()
	if err != nil {
		t.Fatalf("Config() error: %v", err)
	}
	if gotCfg != cfg {
		t.Error("Chained SetConfigManager failed")
	}
	if f.ShellyService() != svc {
		t.Error("Chained SetShellyService failed")
	}
}

func TestFactory_MustConfig(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// MustConfig should return config when no error
	gotCfg := f.MustConfig()
	if gotCfg != cfg {
		t.Error("MustConfig should return the config")
	}
}

func TestFactory_SetIOStreams_OverridesOriginal(t *testing.T) {
	t.Parallel()

	// Create factory with initial IOStreams
	in1 := &bytes.Buffer{}
	out1 := &bytes.Buffer{}
	errOut1 := &bytes.Buffer{}
	ios1 := iostreams.Test(in1, out1, errOut1)
	f := cmdutil.NewWithIOStreams(ios1)

	// Override with new IOStreams
	in2 := &bytes.Buffer{}
	out2 := &bytes.Buffer{}
	errOut2 := &bytes.Buffer{}
	ios2 := iostreams.Test(in2, out2, errOut2)
	f.SetIOStreams(ios2)

	// Should return the new IOStreams
	gotIOS := f.IOStreams()
	if gotIOS != ios2 {
		t.Error("SetIOStreams should override previous IOStreams")
	}
}

func TestFactory_SetConfigManager_OverridesOriginal(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set first config
	cfg1 := &config.Config{}
	mgr1 := config.NewTestManager(cfg1)
	f.SetConfigManager(mgr1)

	// Override with second config
	cfg2 := &config.Config{}
	mgr2 := config.NewTestManager(cfg2)
	f.SetConfigManager(mgr2)

	// Should return the new config
	gotCfg, err := f.Config()
	if err != nil {
		t.Fatalf("Config() error: %v", err)
	}
	if gotCfg != cfg2 {
		t.Error("SetConfigManager should override previous config")
	}
}

func TestFactory_SetShellyService_OverridesOriginal(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set first service
	svc1 := shelly.NewService()
	f.SetShellyService(svc1)

	// Override with second service
	svc2 := shelly.NewService()
	f.SetShellyService(svc2)

	// Should return the new service
	gotSvc := f.ShellyService()
	if gotSvc != svc2 {
		t.Error("SetShellyService should override previous service")
	}
}

func TestFactory_Browser_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// First call should initialize
	br1 := f.Browser()
	if br1 == nil {
		t.Fatal("Browser() returned nil")
	}

	// Second call should return same instance
	br2 := f.Browser()
	if br1 != br2 {
		t.Error("Browser() should return cached instance")
	}
}

func TestFactory_SetBrowser(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set custom browser (using nil as mock since we're just testing the setter)
	result := f.SetBrowser(nil)
	if result != f {
		t.Error("SetBrowser should return factory for chaining")
	}
}

func TestFactory_WithTimeout(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	ctx := t.Context()

	// Create context with timeout
	timeoutCtx, cancel := f.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	if timeoutCtx == nil {
		t.Fatal("WithTimeout returned nil context")
	}

	// Verify deadline is set
	deadline, ok := timeoutCtx.Deadline()
	if !ok {
		t.Error("WithTimeout should set a deadline")
	}
	if deadline.IsZero() {
		t.Error("WithTimeout deadline should not be zero")
	}
}

func TestFactory_WithDefaultTimeout(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	ctx := t.Context()

	// Create context with default timeout
	timeoutCtx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	if timeoutCtx == nil {
		t.Fatal("WithDefaultTimeout returned nil context")
	}

	// Verify deadline is set
	_, ok := timeoutCtx.Deadline()
	if !ok {
		t.Error("WithDefaultTimeout should set a deadline")
	}
}

func TestFactory_GetDevice(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set config with a device
	cfg := &config.Config{
		Devices: map[string]model.Device{
			"test-device": {Address: testDeviceAddress},
		},
	}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should find existing device
	dev := f.GetDevice("test-device")
	if dev == nil {
		t.Fatal("GetDevice should return device")
	}
	if dev.Address != testDeviceAddress {
		t.Errorf("GetDevice address = %q, want %q", dev.Address, testDeviceAddress)
	}

	// Should return nil for non-existent device
	if f.GetDevice("non-existent") != nil {
		t.Error("GetDevice should return nil for non-existent device")
	}
}

func TestFactory_GetGroup(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set config with a group
	cfg := &config.Config{
		Groups: map[string]config.Group{
			"test-group": {Devices: []string{"dev1", "dev2"}},
		},
	}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should find existing group
	grp := f.GetGroup("test-group")
	if grp == nil {
		t.Fatal("GetGroup should return group")
	}
	if len(grp.Devices) != 2 {
		t.Errorf("GetGroup devices = %d, want 2", len(grp.Devices))
	}

	// Should return nil for non-existent group
	if f.GetGroup("non-existent") != nil {
		t.Error("GetGroup should return nil for non-existent group")
	}
}

func TestFactory_GetAlias(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set config with an alias
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"test-alias": {Command: "device list"},
		},
	}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should find existing alias
	alias := f.GetAlias("test-alias")
	if alias == nil {
		t.Fatal("GetAlias should return alias")
	}

	// Should return nil for non-existent alias
	if f.GetAlias("non-existent") != nil {
		t.Error("GetAlias should return nil for non-existent alias")
	}
}

func TestFactory_ResolveAddress(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set config with a device
	cfg := &config.Config{
		Devices: map[string]model.Device{
			"test-device": {Address: testDeviceAddress},
		},
	}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should resolve device name to address
	addr := f.ResolveAddress("test-device")
	if addr != testDeviceAddress {
		t.Errorf("ResolveAddress = %q, want %q", addr, testDeviceAddress)
	}

	// Should return identifier as-is if not found
	addr = f.ResolveAddress("192.168.1.200")
	if addr != "192.168.1.200" {
		t.Errorf("ResolveAddress = %q, want %q", addr, "192.168.1.200")
	}
}

func TestFactory_ResolveDevice(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set config with a device
	cfg := &config.Config{
		Devices: map[string]model.Device{
			"test-device": {Address: testDeviceAddress},
		},
	}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should resolve device name
	dev, ok := f.ResolveDevice("test-device")
	if !ok {
		t.Fatal("ResolveDevice should return true for existing device")
	}
	if dev.Address != testDeviceAddress {
		t.Errorf("ResolveDevice address = %q, want %q", dev.Address, testDeviceAddress)
	}

	// Should return false for non-existent device
	_, ok = f.ResolveDevice("non-existent")
	if ok {
		t.Error("ResolveDevice should return false for non-existent device")
	}
}

func TestFactory_ExpandTargets(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set config with devices and groups
	cfg := &config.Config{
		Devices: map[string]model.Device{
			"dev1": {Address: "192.168.1.1"},
			"dev2": {Address: "192.168.1.2"},
			"dev3": {Address: "192.168.1.3"},
		},
		Groups: map[string]config.Group{
			"test-group": {Devices: []string{"dev1", "dev2"}},
		},
	}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Test with args
	targets, err := f.ExpandTargets([]string{"dev1"}, "", false)
	if err != nil {
		t.Fatalf("ExpandTargets error: %v", err)
	}
	if len(targets) != 1 || targets[0] != "192.168.1.1" {
		t.Errorf("ExpandTargets args = %v, want [192.168.1.1]", targets)
	}

	// Test with group
	targets, err = f.ExpandTargets(nil, "test-group", false)
	if err != nil {
		t.Fatalf("ExpandTargets group error: %v", err)
	}
	if len(targets) != 2 {
		t.Errorf("ExpandTargets group len = %d, want 2", len(targets))
	}

	// Test with all
	targets, err = f.ExpandTargets(nil, "", true)
	if err != nil {
		t.Fatalf("ExpandTargets all error: %v", err)
	}
	if len(targets) != 3 {
		t.Errorf("ExpandTargets all len = %d, want 3", len(targets))
	}

	// Test with non-existent group
	_, err = f.ExpandTargets(nil, "non-existent", false)
	if err == nil {
		t.Error("ExpandTargets should error for non-existent group")
	}

	// Test with no targets
	_, err = f.ExpandTargets(nil, "", false)
	if err == nil {
		t.Error("ExpandTargets should error for no targets")
	}
}

func TestFactory_OutputFormat(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Default output format is empty
	// Note: We can't easily test viper-based output format without setting up viper
	// This test just ensures the method doesn't panic
	_ = f.OutputFormat()
	_ = f.IsJSONOutput()
	_ = f.IsYAMLOutput()
	_ = f.IsStructuredOutput()
}

func TestFactory_Logger(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Logger should not return nil
	logger := f.Logger()
	if logger == nil {
		t.Error("Logger should not return nil")
	}
}

func TestFactory_MustConfig_Panic(t *testing.T) {
	t.Parallel()

	// Use a factory with an erroring ConfigManager
	f := cmdutil.NewFactory()
	f.ConfigManager = func() (*config.Manager, error) {
		return nil, errors.New("config load failed")
	}
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("config load failed")
	}

	// MustConfig should panic when config fails
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustConfig should panic when config fails to load")
		}
	}()
	f.MustConfig()
}

func TestFactory_MustConfigManager_Panic(t *testing.T) {
	t.Parallel()

	// Use a factory with an erroring ConfigManager
	f := cmdutil.NewFactory()
	f.ConfigManager = func() (*config.Manager, error) {
		return nil, errors.New("config manager load failed")
	}

	// MustConfigManager should panic when config manager fails
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustConfigManager should panic when config manager fails to load")
		}
	}()
	f.MustConfigManager()
}

func TestFactory_GetDevice_ConfigError(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("config error")
	}

	// Should return nil when config fails
	if f.GetDevice("any") != nil {
		t.Error("GetDevice should return nil when config fails")
	}
}

func TestFactory_GetGroup_ConfigError(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("config error")
	}

	// Should return nil when config fails
	if f.GetGroup("any") != nil {
		t.Error("GetGroup should return nil when config fails")
	}
}

func TestFactory_GetAlias_ConfigError(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("config error")
	}

	// Should return nil when config fails
	if f.GetAlias("any") != nil {
		t.Error("GetAlias should return nil when config fails")
	}
}

func TestFactory_GetScene_ConfigError(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("config error")
	}

	// Should return nil when config fails
	if f.GetScene("any") != nil {
		t.Error("GetScene should return nil when config fails")
	}
}

func TestFactory_GetDeviceTemplate_ConfigError(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("config error")
	}

	// Should return nil when config fails
	if f.GetDeviceTemplate("any") != nil {
		t.Error("GetDeviceTemplate should return nil when config fails")
	}
}

func TestFactory_ResolveDevice_ConfigError(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("config error")
	}

	// Should return false when config fails
	_, ok := f.ResolveDevice("any")
	if ok {
		t.Error("ResolveDevice should return false when config fails")
	}
}

func TestFactory_ExpandTargets_ConfigError(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	f.Config = func() (*config.Config, error) {
		return nil, errors.New("config error")
	}

	// Should return error when config fails
	_, err := f.ExpandTargets([]string{"device1"}, "", false)
	if err == nil {
		t.Error("ExpandTargets should return error when config fails")
	}
}

func TestFactory_SetBrowser_NilThenGet(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Get browser first to initialize
	br1 := f.Browser()
	if br1 == nil {
		t.Fatal("Browser() returned nil")
	}

	// Clear cached browser instance by setting nil
	f.SetBrowser(nil)

	// Get browser again - should fall back to original
	br2 := f.Browser()
	if br2 == nil {
		t.Error("Browser should not be nil after fallback")
	}
}

func TestFactory_SetIOStreams_NilThenGet(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Get IOStreams first to initialize
	ios1 := f.IOStreams()
	if ios1 == nil {
		t.Fatal("IOStreams() returned nil")
	}

	// Set to nil and create new factory method that clears cache
	f.SetIOStreams(nil)

	// Get IOStreams again - should fall back to original
	ios2 := f.IOStreams()
	if ios2 == nil {
		t.Error("IOStreams should not be nil after fallback")
	}
}

func TestFactory_SetConfigManager_NilThenGet(t *testing.T) {
	t.Parallel()

	// Create factory and set a config manager first
	f := cmdutil.NewFactory()
	cfg := &config.Config{}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Verify it works
	gotMgr, err := f.ConfigManager()
	if err != nil {
		t.Fatalf("ConfigManager() error: %v", err)
	}
	if gotMgr != mgr {
		t.Error("SetConfigManager should set the config manager")
	}

	// Now set to nil
	f.SetConfigManager(nil)

	// Get config manager again - should fall back to original
	gotMgr2, err := f.ConfigManager()
	if err != nil {
		t.Fatalf("ConfigManager() after nil error: %v", err)
	}
	if gotMgr2 == nil {
		t.Error("ConfigManager should not be nil after fallback")
	}
}

func TestFactory_SetShellyService_NilThenGet(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Get service first to initialize
	svc1 := f.ShellyService()
	if svc1 == nil {
		t.Fatal("ShellyService() returned nil")
	}

	// Set to nil
	f.SetShellyService(nil)

	// Get service again - should fall back to original
	svc2 := f.ShellyService()
	if svc2 == nil {
		t.Error("ShellyService should not be nil after fallback")
	}
}
