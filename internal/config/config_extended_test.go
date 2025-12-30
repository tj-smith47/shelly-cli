package config

import (
	"os"
	"testing"
	"time"
)

const (
	testIP1   = "192.168.1.1"
	testIP100 = "192.168.1.100"
)

func TestRateLimitConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  RateLimitConfig
		wantErr bool
	}{
		{
			name:    "default config is valid",
			config:  DefaultRateLimitConfig(),
			wantErr: false,
		},
		{
			name: "gen1 max_concurrent too high",
			config: RateLimitConfig{
				Gen1:   GenerationRateLimitConfig{MaxConcurrent: 3}, // Max is 2
				Gen2:   GenerationRateLimitConfig{MaxConcurrent: 3},
				Global: GlobalRateLimitConfig{MaxConcurrent: 5},
			},
			wantErr: true,
		},
		{
			name: "gen2 max_concurrent too high",
			config: RateLimitConfig{
				Gen1:   GenerationRateLimitConfig{MaxConcurrent: 1},
				Gen2:   GenerationRateLimitConfig{MaxConcurrent: 6}, // Max is 5
				Global: GlobalRateLimitConfig{MaxConcurrent: 5},
			},
			wantErr: true,
		},
		{
			name: "gen1 negative max_concurrent",
			config: RateLimitConfig{
				Gen1:   GenerationRateLimitConfig{MaxConcurrent: -1},
				Gen2:   GenerationRateLimitConfig{MaxConcurrent: 3},
				Global: GlobalRateLimitConfig{MaxConcurrent: 5},
			},
			wantErr: true,
		},
		{
			name: "gen1 negative min_interval",
			config: RateLimitConfig{
				Gen1:   GenerationRateLimitConfig{MinInterval: -1 * time.Second},
				Gen2:   GenerationRateLimitConfig{},
				Global: GlobalRateLimitConfig{MaxConcurrent: 5},
			},
			wantErr: true,
		},
		{
			name: "gen2 negative min_interval",
			config: RateLimitConfig{
				Gen1:   GenerationRateLimitConfig{},
				Gen2:   GenerationRateLimitConfig{MinInterval: -1 * time.Second},
				Global: GlobalRateLimitConfig{MaxConcurrent: 5},
			},
			wantErr: true,
		},
		{
			name: "global max_concurrent zero",
			config: RateLimitConfig{
				Gen1:   GenerationRateLimitConfig{},
				Gen2:   GenerationRateLimitConfig{},
				Global: GlobalRateLimitConfig{MaxConcurrent: 0},
			},
			wantErr: true,
		},
		{
			name: "all zeroes except valid global",
			config: RateLimitConfig{
				Gen1:   GenerationRateLimitConfig{},
				Gen2:   GenerationRateLimitConfig{},
				Global: GlobalRateLimitConfig{MaxConcurrent: 1},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRateLimitConfig_IsZero(t *testing.T) {
	t.Parallel()

	t.Run("empty config is zero", func(t *testing.T) {
		t.Parallel()
		cfg := RateLimitConfig{}
		if !cfg.IsZero() {
			t.Error("empty RateLimitConfig should be zero")
		}
	})

	t.Run("default config is not zero", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultRateLimitConfig()
		if cfg.IsZero() {
			t.Error("DefaultRateLimitConfig should not be zero")
		}
	})

	t.Run("partial config is not zero", func(t *testing.T) {
		t.Parallel()
		cfg := RateLimitConfig{
			Gen1: GenerationRateLimitConfig{MinInterval: time.Second},
		}
		if cfg.IsZero() {
			t.Error("config with one field set should not be zero")
		}
	})
}

func TestConfig_GetEditor(t *testing.T) {
	t.Parallel()

	t.Run("nil config returns empty", func(t *testing.T) {
		t.Parallel()
		var cfg *Config
		if got := cfg.GetEditor(); got != "" {
			t.Errorf("GetEditor() = %q, want empty", got)
		}
	})

	t.Run("empty editor returns empty", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{}
		if got := cfg.GetEditor(); got != "" {
			t.Errorf("GetEditor() = %q, want empty", got)
		}
	})

	t.Run("editor set returns it", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Editor: "vim"}
		if got := cfg.GetEditor(); got != "vim" {
			t.Errorf("GetEditor() = %q, want %q", got, "vim")
		}
	})
}

func TestConfig_GetIntegratorCredentials(t *testing.T) {
	t.Parallel()

	t.Run("config with credentials succeeds", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Integrator: IntegratorConfig{
				Tag:   "test-tag",
				Token: "test-token",
			},
		}
		tag, token, err := cfg.GetIntegratorCredentials()
		if err != nil {
			t.Fatalf("GetIntegratorCredentials() error: %v", err)
		}
		if tag != "test-tag" {
			t.Errorf("tag = %q, want %q", tag, "test-tag")
		}
		if token != "test-token" {
			t.Errorf("token = %q, want %q", token, "test-token")
		}
	})

	t.Run("missing tag fails", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Integrator: IntegratorConfig{
				Token: "test-token",
			},
		}
		_, _, err := cfg.GetIntegratorCredentials()
		if err == nil {
			t.Error("expected error with missing tag")
		}
	})

	t.Run("missing token fails", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Integrator: IntegratorConfig{
				Tag: "test-tag",
			},
		}
		_, _, err := cfg.GetIntegratorCredentials()
		if err == nil {
			t.Error("expected error with missing token")
		}
	})
}

func TestCacheDir(t *testing.T) {
	t.Parallel()

	cacheDir, err := CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() error: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error: %v", err)
	}

	// Should be ~/.config/shelly/cache
	if cacheDir == "" {
		t.Error("CacheDir() returned empty string")
	}
	if len(cacheDir) <= len(homeDir) {
		t.Errorf("CacheDir() = %q, expected longer path", cacheDir)
	}
}

// =============================================================================
// Package-level function tests (test functions that delegate to default manager)
// =============================================================================

// setupPackageTest sets up an isolated environment for package-level function tests.
// Returns a cleanup function that MUST be deferred.
func setupPackageTest(t *testing.T) func() {
	t.Helper()

	// Save original values
	originalHome := os.Getenv("HOME")
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")

	// Reset the default manager
	ResetDefaultManagerForTesting()

	// Create temp directory and set as HOME
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	// Also set XDG_CONFIG_HOME to temp directory to ensure config isolation
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}

	// Create config directory and minimal config (XDG_CONFIG_HOME takes precedence)
	configDir := tmpDir + "/shelly"
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := configDir + "/config.yaml"
	configContent := "devices: {}\ngroups: {}\naliases: {}\nscenes: {}\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return func() {
		ResetDefaultManagerForTesting()
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		if originalXDGConfig != "" {
			if err := os.Setenv("XDG_CONFIG_HOME", originalXDGConfig); err != nil {
				t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("warning: failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_RegisterDevice(t *testing.T) {
	cleanup := setupPackageTest(t)
	defer cleanup()

	// Test RegisterDevice
	err := RegisterDevice("test-device", testIP100, 2, "shelly-plus-1pm", "SHPLUG1-S", nil)
	if err != nil {
		t.Errorf("RegisterDevice() error = %v", err)
	}

	// Verify device was registered
	dev, ok := GetDevice("test-device")
	if !ok {
		t.Fatal("GetDevice() returned false")
	}
	if dev.Address != testIP100 {
		t.Errorf("device address = %q, want %q", dev.Address, testIP100)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_ListDevices(t *testing.T) {
	cleanup := setupPackageTest(t)
	defer cleanup()

	// Register a device
	if err := RegisterDevice("dev1", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	// Test ListDevices
	devices := ListDevices()
	if len(devices) != 1 {
		t.Errorf("ListDevices() returned %d devices, want 1", len(devices))
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_GroupOperations(t *testing.T) {
	cleanup := setupPackageTest(t)
	defer cleanup()

	// Create group
	err := CreateGroup("lights")
	if err != nil {
		t.Errorf("CreateGroup() error = %v", err)
	}

	// List groups
	groups := ListGroups()
	if len(groups) != 1 {
		t.Errorf("ListGroups() returned %d groups, want 1", len(groups))
	}

	// Get group
	grp, ok := GetGroup("lights")
	if !ok {
		t.Fatal("GetGroup() returned false")
	}
	if len(grp.Devices) != 0 {
		t.Errorf("group devices = %d, want 0", len(grp.Devices))
	}

	// Register device and add to group
	if err := RegisterDevice("kitchen", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}
	if err := AddDeviceToGroup("lights", "kitchen"); err != nil {
		t.Errorf("AddDeviceToGroup() error = %v", err)
	}

	// Get group devices
	devs, err := GetGroupDevices("lights")
	if err != nil {
		t.Errorf("GetGroupDevices() error = %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("GetGroupDevices() returned %d devices, want 1", len(devs))
	}

	// Remove from group
	if err := RemoveDeviceFromGroup("lights", "kitchen"); err != nil {
		t.Errorf("RemoveDeviceFromGroup() error = %v", err)
	}

	// Delete group
	if err := DeleteGroup("lights"); err != nil {
		t.Errorf("DeleteGroup() error = %v", err)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_ResolveDevice(t *testing.T) {
	cleanup := setupPackageTest(t)
	defer cleanup()

	// Register device
	if err := RegisterDevice("kitchen", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	// Resolve by name
	dev, err := ResolveDevice("kitchen")
	if err != nil {
		t.Fatalf("ResolveDevice() error = %v", err)
	}
	if dev.Address != testIP1 {
		t.Errorf("ResolveDevice() address = %q, want %q", dev.Address, testIP1)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_DeviceAlias(t *testing.T) {
	cleanup := setupPackageTest(t)
	defer cleanup()

	// Register device
	if err := RegisterDevice("kitchen-light", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	// Add alias
	if err := AddDeviceAlias("kitchen-light", "kl"); err != nil {
		t.Errorf("AddDeviceAlias() error = %v", err)
	}

	// Check conflict
	if err := CheckAliasConflict("kl", ""); err == nil {
		t.Error("CheckAliasConflict() should return error for existing alias")
	}

	// Get aliases
	aliases, err := GetDeviceAliases("kitchen-light")
	if err != nil {
		t.Errorf("GetDeviceAliases() error = %v", err)
	}
	if len(aliases) != 1 || aliases[0] != "kl" {
		t.Errorf("GetDeviceAliases() = %v, want [kl]", aliases)
	}

	// Remove alias
	if err := RemoveDeviceAlias("kitchen-light", "kl"); err != nil {
		t.Errorf("RemoveDeviceAlias() error = %v", err)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_RenameDevice(t *testing.T) {
	cleanup := setupPackageTest(t)
	defer cleanup()

	// Register device
	if err := RegisterDevice("old-name", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	// Rename
	if err := RenameDevice("old-name", "new-name"); err != nil {
		t.Errorf("RenameDevice() error = %v", err)
	}

	// Verify
	_, ok := GetDevice("old-name")
	if ok {
		t.Error("old device name should not exist")
	}
	_, ok = GetDevice("new-name")
	if !ok {
		t.Error("new device name should exist")
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_UpdateDeviceAddress(t *testing.T) {
	cleanup := setupPackageTest(t)
	defer cleanup()

	// Register device
	if err := RegisterDevice("kitchen", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	// Update address
	if err := UpdateDeviceAddress("kitchen", testIP100); err != nil {
		t.Errorf("UpdateDeviceAddress() error = %v", err)
	}

	// Verify
	dev, _ := GetDevice("kitchen")
	if dev.Address != testIP100 {
		t.Errorf("device address = %q, want %q", dev.Address, testIP100)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_UnregisterDevice(t *testing.T) {
	cleanup := setupPackageTest(t)
	defer cleanup()

	// Register device
	if err := RegisterDevice("kitchen", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	// Unregister
	if err := UnregisterDevice("kitchen"); err != nil {
		t.Errorf("UnregisterDevice() error = %v", err)
	}

	// Verify
	_, ok := GetDevice("kitchen")
	if ok {
		t.Error("device should not exist after unregister")
	}
}
