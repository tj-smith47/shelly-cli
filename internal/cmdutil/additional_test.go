package cmdutil_test

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// =============================================================================
// GetScene Tests
// =============================================================================

func TestFactory_GetScene(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set config with a scene
	cfg := &config.Config{
		Scenes: map[string]config.Scene{
			"movie-night": {
				Actions: []config.SceneAction{
					{Device: "dev1", Method: "Switch.Set"},
				},
			},
		},
	}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should find existing scene
	scene := f.GetScene("movie-night")
	if scene == nil {
		t.Fatal("GetScene should return scene")
	}
	if len(scene.Actions) != 1 {
		t.Errorf("GetScene actions = %d, want 1", len(scene.Actions))
	}

	// Should return nil for non-existent scene
	if f.GetScene("non-existent") != nil {
		t.Error("GetScene should return nil for non-existent scene")
	}
}

// =============================================================================
// GetDeviceTemplate Tests
// =============================================================================

func TestFactory_GetDeviceTemplate(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Set config with a device template
	cfg := &config.Config{
		Templates: config.TemplatesConfig{
			Device: map[string]config.DeviceTemplate{
				"motion-sensor": {
					Name: "Motion Sensor",
				},
			},
		},
	}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should find existing template
	tmpl := f.GetDeviceTemplate("motion-sensor")
	if tmpl == nil {
		t.Fatal("GetDeviceTemplate should return template")
	}
	if tmpl.Name != "Motion Sensor" {
		t.Errorf("GetDeviceTemplate name = %q, want %q", tmpl.Name, "Motion Sensor")
	}

	// Should return nil for non-existent template
	if f.GetDeviceTemplate("non-existent") != nil {
		t.Error("GetDeviceTemplate should return nil for non-existent template")
	}
}

// =============================================================================
// Service Lazy Init Tests
// =============================================================================

func TestFactory_KVSService_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// First call should initialize
	svc1 := f.KVSService()
	if svc1 == nil {
		t.Fatal("KVSService() returned nil")
	}

	// Second call should return same instance
	svc2 := f.KVSService()
	if svc1 != svc2 {
		t.Error("KVSService() should return cached instance")
	}
}

func TestFactory_AutomationService_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// First call should initialize
	svc1 := f.AutomationService()
	if svc1 == nil {
		t.Fatal("AutomationService() returned nil")
	}

	// Second call should return same instance
	svc2 := f.AutomationService()
	if svc1 != svc2 {
		t.Error("AutomationService() should return cached instance")
	}
}

func TestFactory_ModbusService_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// First call should initialize
	svc1 := f.ModbusService()
	if svc1 == nil {
		t.Fatal("ModbusService() returned nil")
	}

	// Second call should return same instance
	svc2 := f.ModbusService()
	if svc1 != svc2 {
		t.Error("ModbusService() should return cached instance")
	}
}

func TestFactory_SensorAddonService_LazyInit(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// First call should initialize
	svc1 := f.SensorAddonService()
	if svc1 == nil {
		t.Fatal("SensorAddonService() returned nil")
	}

	// Second call should return same instance
	svc2 := f.SensorAddonService()
	if svc1 != svc2 {
		t.Error("SensorAddonService() should return cached instance")
	}
}

// =============================================================================
// ConfirmAction Tests
// =============================================================================

func TestFactory_ConfirmAction_YesTrue(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// When yes is true, should return true immediately without prompting
	confirmed, err := f.ConfirmAction("Are you sure?", true)
	if err != nil {
		t.Fatalf("ConfirmAction error = %v", err)
	}
	if !confirmed {
		t.Error("ConfirmAction should return true when yes is true")
	}
}

func TestFactory_ConfirmAction_NonTTY(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// In non-TTY mode (Test creates non-TTY), default confirmation should be false
	confirmed, err := f.ConfirmAction("Are you sure?", false)
	if err != nil {
		t.Fatalf("ConfirmAction error = %v", err)
	}
	if confirmed {
		t.Error("ConfirmAction should return false in non-TTY mode without --yes")
	}
}

// =============================================================================
// MustConfigManager Tests
// =============================================================================

func TestFactory_MustConfigManager(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{}
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// MustConfigManager should return manager when no error
	gotMgr := f.MustConfigManager()
	if gotMgr != mgr {
		t.Error("MustConfigManager should return the config manager")
	}
}

// =============================================================================
// CapConcurrency Tests
// =============================================================================

func TestCapConcurrency_WithinLimit(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()

	// Request concurrency within the limit should not be capped
	result := cmdutil.CapConcurrency(ios, 3)

	// Default global max concurrent is 10 (from config package)
	if result != 3 {
		t.Errorf("CapConcurrency(3) = %d, want 3", result)
	}
	if errOut.Len() > 0 {
		t.Error("CapConcurrency should not warn when within limit")
	}
}

func TestCapConcurrency_ExceedsLimit(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()

	// Request concurrency exceeding the limit should be capped
	// Default global max concurrent is 10, so requesting 20 should be capped
	result := cmdutil.CapConcurrency(ios, 100)

	// Should be capped to global max (10 by default)
	if result == 100 {
		t.Errorf("CapConcurrency(100) should be capped, got %d", result)
	}
	if errOut.Len() == 0 {
		t.Error("CapConcurrency should warn when exceeding limit")
	}
}

// =============================================================================
// GetDevice/GetGroup/GetAlias with empty config
// =============================================================================

func TestFactory_GetDevice_NotFound(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{} // empty config
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should return nil when device not in config
	dev := f.GetDevice("nonexistent")
	if dev != nil {
		t.Error("GetDevice should return nil for non-existent device")
	}
}

func TestFactory_GetGroup_NotFound(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{} // empty config
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should return nil when group not in config
	grp := f.GetGroup("nonexistent")
	if grp != nil {
		t.Error("GetGroup should return nil for non-existent group")
	}
}

func TestFactory_GetAlias_NotFound(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{} // empty config
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should return nil when alias not in config
	alias := f.GetAlias("nonexistent")
	if alias != nil {
		t.Error("GetAlias should return nil for non-existent alias")
	}
}

func TestFactory_GetScene_NotFound(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{} // empty config
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should return nil when scene not in config
	scene := f.GetScene("nonexistent")
	if scene != nil {
		t.Error("GetScene should return nil for non-existent scene")
	}
}

func TestFactory_GetDeviceTemplate_NotFound(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{} // empty config
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should return nil when template not in config
	tmpl := f.GetDeviceTemplate("nonexistent")
	if tmpl != nil {
		t.Error("GetDeviceTemplate should return nil for non-existent template")
	}
}

// =============================================================================
// ResolveDevice with empty config
// =============================================================================

func TestFactory_ResolveDevice_NotFound(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{} // empty config
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should return false when device not found
	_, ok := f.ResolveDevice("nonexistent")
	if ok {
		t.Error("ResolveDevice should return false for non-existent device")
	}
}

// =============================================================================
// ExpandTargets with empty config
// =============================================================================

func TestFactory_ExpandTargets_NoDevices(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cfg := &config.Config{} // empty config
	mgr := config.NewTestManager(cfg)
	f.SetConfigManager(mgr)

	// Should return error when --all is used but no devices configured
	_, err := f.ExpandTargets(nil, "", true)
	if err == nil {
		t.Error("ExpandTargets with --all should error when no devices configured")
	}
}
