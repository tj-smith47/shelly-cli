package plugins

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

const testPlatformTasmota = "tasmota"

func TestNewHookExecutor(t *testing.T) {
	t.Parallel()

	plugin := &Plugin{Name: "test"}
	executor := NewHookExecutor(plugin)

	if executor == nil {
		t.Fatal("NewHookExecutor() returned nil")
	}
	if executor.plugin != plugin {
		t.Error("HookExecutor has wrong plugin reference")
	}
}

func TestHookExecutor_ExecuteDetect_NoHook(t *testing.T) {
	t.Parallel()

	// Plugin without hooks
	plugin := &Plugin{Name: "test"}
	executor := NewHookExecutor(plugin)

	_, err := executor.ExecuteDetect(context.Background(), "192.168.1.100", nil)
	if err == nil {
		t.Error("ExecuteDetect() should fail when plugin has no hooks")
	}
}

func TestHookExecutor_ExecuteDetect_NoManifest(t *testing.T) {
	t.Parallel()

	// Plugin with manifest but no hooks
	plugin := &Plugin{
		Name: "test",
		Manifest: &Manifest{
			Name: "test",
		},
	}
	executor := NewHookExecutor(plugin)

	_, err := executor.ExecuteDetect(context.Background(), "192.168.1.100", nil)
	if err == nil {
		t.Error("ExecuteDetect() should fail when manifest has no hooks")
	}
}

func TestHookExecutor_ExecuteStatus_NoHook(t *testing.T) {
	t.Parallel()

	plugin := &Plugin{Name: "test"}
	executor := NewHookExecutor(plugin)

	_, err := executor.ExecuteStatus(context.Background(), "192.168.1.100", nil)
	if err == nil {
		t.Error("ExecuteStatus() should fail when plugin has no hooks")
	}
}

func TestHookExecutor_ExecuteControl_NoHook(t *testing.T) {
	t.Parallel()

	plugin := &Plugin{Name: "test"}
	executor := NewHookExecutor(plugin)

	_, err := executor.ExecuteControl(context.Background(), "192.168.1.100", nil, "on", "switch", 0)
	if err == nil {
		t.Error("ExecuteControl() should fail when plugin has no hooks")
	}
}

func TestHookExecutor_ExecuteCheckUpdates_NoHook(t *testing.T) {
	t.Parallel()

	plugin := &Plugin{Name: "test"}
	executor := NewHookExecutor(plugin)

	_, err := executor.ExecuteCheckUpdates(context.Background(), "192.168.1.100", nil)
	if err == nil {
		t.Error("ExecuteCheckUpdates() should fail when plugin has no hooks")
	}
}

func TestHookExecutor_ExecuteApplyUpdate_NoHook(t *testing.T) {
	t.Parallel()

	plugin := &Plugin{Name: "test"}
	executor := NewHookExecutor(plugin)

	_, err := executor.ExecuteApplyUpdate(context.Background(), "192.168.1.100", nil, "stable", "")
	if err == nil {
		t.Error("ExecuteApplyUpdate() should fail when plugin has no hooks")
	}
}

func TestHookExecutor_ExecuteDetect_WithMockHook(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a mock detect hook script
	hookScript := `#!/bin/bash
echo '{"detected": true, "platform": "` + testPlatformTasmota + `", "device_id": "test123", "model": "Sonoff Basic"}'
`
	hookPath := filepath.Join(tmpDir, "detect")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				Detect: "./detect",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	result, err := executor.ExecuteDetect(context.Background(), "192.168.1.100", nil)
	if err != nil {
		t.Fatalf("ExecuteDetect() error: %v", err)
	}

	if !result.Detected {
		t.Error("expected Detected to be true")
	}
	if result.Platform != testPlatformTasmota {
		t.Errorf("expected Platform %q, got %q", testPlatformTasmota, result.Platform)
	}
	if result.DeviceID != "test123" {
		t.Errorf("expected DeviceID 'test123', got %q", result.DeviceID)
	}
}

func TestHookExecutor_ExecuteDetect_WithAuth(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a mock detect hook that outputs the arguments
	hookScript := `#!/bin/bash
# Verify auth arguments are passed
for arg in "$@"; do
    if [ "$arg" == "--auth-user" ]; then
        found_user=1
    fi
done
echo '{"detected": true, "platform": "test"}'
`
	hookPath := filepath.Join(tmpDir, "detect")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				Detect: "./detect",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	auth := &model.Auth{
		Username: "admin",
		Password: "secret",
	}

	result, err := executor.ExecuteDetect(context.Background(), "192.168.1.100", auth)
	if err != nil {
		t.Fatalf("ExecuteDetect() with auth error: %v", err)
	}

	if !result.Detected {
		t.Error("expected Detected to be true")
	}
}

func TestHookExecutor_ExecuteControl_WithMockHook(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a mock control hook script
	hookScript := `#!/bin/bash
echo '{"success": true, "state": "on"}'
`
	hookPath := filepath.Join(tmpDir, "control")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				Control: "./control",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	result, err := executor.ExecuteControl(context.Background(), "192.168.1.100", nil, "on", "switch", 0)
	if err != nil {
		t.Fatalf("ExecuteControl() error: %v", err)
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.State != "on" {
		t.Errorf("expected State 'on', got %q", result.State)
	}
}

func TestHookExecutor_ExecuteHook_Timeout(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a hook that sleeps longer than our timeout
	hookScript := `#!/bin/bash
sleep 10
echo '{"detected": true}'
`
	hookPath := filepath.Join(tmpDir, "detect")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				Detect: "./detect",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	// Use a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = executor.ExecuteDetect(ctx, "192.168.1.100", nil)
	if err == nil {
		t.Error("ExecuteDetect() should fail with timeout")
	}
}

func TestHookExecutor_ExecuteHook_FailedExecution(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a hook that fails with stderr
	hookScript := `#!/bin/bash
echo "Error: device not found" >&2
exit 1
`
	hookPath := filepath.Join(tmpDir, "detect")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				Detect: "./detect",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	_, err = executor.ExecuteDetect(context.Background(), "192.168.1.100", nil)
	if err == nil {
		t.Error("ExecuteDetect() should fail when hook exits with error")
	}
	// Check that stderr is included in error message
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestHookExecutor_ExecuteHook_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a hook that outputs invalid JSON
	hookScript := `#!/bin/bash
echo 'not valid json'
`
	hookPath := filepath.Join(tmpDir, "detect")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				Detect: "./detect",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	_, err = executor.ExecuteDetect(context.Background(), "192.168.1.100", nil)
	if err == nil {
		t.Error("ExecuteDetect() should fail when hook returns invalid JSON")
	}
}

func TestHookExecutor_ExecuteStatus_WithMockHook(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a mock status hook script
	hookScript := `#!/bin/bash
echo '{"online": true, "energy": {"power": 42.5, "voltage": 230.0}}'
`
	hookPath := filepath.Join(tmpDir, "status")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				Status: "./status",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	result, err := executor.ExecuteStatus(context.Background(), "192.168.1.100", nil)
	if err != nil {
		t.Fatalf("ExecuteStatus() error: %v", err)
	}

	if !result.Online {
		t.Error("expected Online to be true")
	}
	if result.Energy == nil {
		t.Fatal("expected Energy to be non-nil")
	}
	if result.Energy.Power != 42.5 {
		t.Errorf("expected Power 42.5, got %v", result.Energy.Power)
	}
}

func TestHookExecutor_ExecuteCheckUpdates_WithMockHook(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a mock check_updates hook script
	hookScript := `#!/bin/bash
echo '{"current_version": "14.3.0", "latest_stable": "15.0.0", "has_update": true}'
`
	hookPath := filepath.Join(tmpDir, "check-updates")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				CheckUpdates: "./check-updates",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	result, err := executor.ExecuteCheckUpdates(context.Background(), "192.168.1.100", nil)
	if err != nil {
		t.Fatalf("ExecuteCheckUpdates() error: %v", err)
	}

	if result.CurrentVersion != "14.3.0" {
		t.Errorf("expected CurrentVersion '14.3.0', got %q", result.CurrentVersion)
	}
	if result.LatestStable != "15.0.0" {
		t.Errorf("expected LatestStable '15.0.0', got %q", result.LatestStable)
	}
	if !result.HasUpdate {
		t.Error("expected HasUpdate to be true")
	}
}

func TestHookExecutor_ExecuteApplyUpdate_WithMockHook(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a mock apply_update hook script
	hookScript := `#!/bin/bash
echo '{"success": true, "message": "Update initiated", "rebooting": true}'
`
	hookPath := filepath.Join(tmpDir, "apply-update")
	testutil.WriteTestScript(t, hookPath, hookScript)

	plugin := &Plugin{
		Name: "test",
		Dir:  tmpDir,
		Manifest: &Manifest{
			Name: "test",
			Hooks: &Hooks{
				ApplyUpdate: "./apply-update",
			},
		},
	}
	executor := NewHookExecutor(plugin)

	result, err := executor.ExecuteApplyUpdate(context.Background(), "192.168.1.100", nil, "stable", "")
	if err != nil {
		t.Fatalf("ExecuteApplyUpdate() error: %v", err)
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}
	if !result.Rebooting {
		t.Error("expected Rebooting to be true")
	}
}
