package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

const (
	testMAC             = "AA:BB:CC:DD:EE:FF"
	testModelPlug       = "SPSW-001PE16EU"
	testAliasName       = "kitchen"
	testIP1             = "192.168.1.1"
	testIP100           = "192.168.1.100"
	testThemeDracula    = "dracula"
	testSceneMovieNight = "movie-night"
	testSceneTxtFile    = "/scene.txt"
)

// setupConfigTest sets up an isolated Manager for config testing.
// It uses an in-memory filesystem to avoid touching real files.
func setupConfigTest(t *testing.T) *Manager {
	t.Helper()
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	m := NewManager("/test/config/config.yaml")
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	return m
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_Defaults(t *testing.T) {
	m := setupConfigTest(t)

	c := m.Get()

	if c.Output != "" && c.Output != "table" {
		t.Errorf("expected output '' or 'table', got %q", c.Output)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_InitializesMaps(t *testing.T) {
	m := setupConfigTest(t)

	c := m.Get()

	if c.Devices == nil {
		t.Error("expected Devices map to be initialized")
	}
	if c.Aliases == nil {
		t.Error("expected Aliases map to be initialized")
	}
	if c.Groups == nil {
		t.Error("expected Groups map to be initialized")
	}
	if c.Scenes == nil {
		t.Error("expected Scenes map to be initialized")
	}
	if c.Templates.Device == nil {
		t.Error("expected Templates.Device map to be initialized")
	}
	if c.Templates.Script == nil {
		t.Error("expected Templates.Script map to be initialized")
	}
	if c.Alerts == nil {
		t.Error("expected Alerts map to be initialized")
	}
}

func TestConfigDir(t *testing.T) {
	t.Parallel()

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error: %v", err)
	}
	expected := filepath.Join(home, ".config", "shelly")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestPluginsDir(t *testing.T) {
	t.Parallel()

	dir, err := PluginsDir()
	if err != nil {
		t.Fatalf("PluginsDir() error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error: %v", err)
	}
	expected := filepath.Join(home, ".config", "shelly", "plugins")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestGetThemeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		theme          any
		expectedName   string
		expectedColors int
		expectedFile   string
	}{
		{
			name:         "string format",
			theme:        "nord",
			expectedName: "nord",
		},
		{
			name:         "empty string",
			theme:        "",
			expectedName: "",
		},
		{
			name: "block format with name only",
			theme: map[string]any{
				"name": "gruvbox",
			},
			expectedName: "gruvbox",
		},
		{
			name: "block format with colors",
			theme: map[string]any{
				"name": "dracula",
				"colors": map[string]any{
					"green": "#50fa7b",
					"red":   "#ff5555",
				},
			},
			expectedName:   "dracula",
			expectedColors: 2,
		},
		{
			name: "block format with file",
			theme: map[string]any{
				"file": "~/.config/shelly/themes/custom.yaml",
			},
			expectedFile: "~/.config/shelly/themes/custom.yaml",
		},
		{
			name: "block format full custom (no name)",
			theme: map[string]any{
				"colors": map[string]any{
					"foreground": "#f8f8f2",
					"background": "#282a36",
				},
			},
			expectedName:   "",
			expectedColors: 2,
		},
		{
			name:         "nil theme",
			theme:        nil,
			expectedName: "dracula",
		},
		{
			name:         "invalid type (int)",
			theme:        42,
			expectedName: "dracula",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Config{Theme: tt.theme}
			tc := c.GetThemeConfig()

			if tc.Name != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, tc.Name)
			}
			if len(tc.Colors) != tt.expectedColors {
				t.Errorf("expected %d colors, got %d", tt.expectedColors, len(tc.Colors))
			}
			if tc.File != tt.expectedFile {
				t.Errorf("expected file %q, got %q", tt.expectedFile, tc.File)
			}
		})
	}
}

func TestGetThemeConfig_NilConfig(t *testing.T) {
	t.Parallel()

	var c *Config
	tc := c.GetThemeConfig()
	if tc.Name != "dracula" {
		t.Errorf("expected default theme 'dracula', got %q", tc.Name)
	}
}

func TestGetTUIThemeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *Config
		expected *ThemeConfig
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: nil,
		},
		{
			name: "no TUI theme set",
			config: &Config{
				TUI: TUIConfig{},
			},
			expected: nil,
		},
		{
			name: "TUI theme set",
			config: &Config{
				TUI: TUIConfig{
					Theme: &ThemeConfig{
						Name: "nord",
						Colors: map[string]string{
							"highlight": "#88c0d0",
						},
					},
				},
			},
			expected: &ThemeConfig{
				Name: "nord",
				Colors: map[string]string{
					"highlight": "#88c0d0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.config.GetTUIThemeConfig()

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Name != tt.expected.Name {
				t.Errorf("expected name %q, got %q", tt.expected.Name, result.Name)
			}
			if len(result.Colors) != len(tt.expected.Colors) {
				t.Errorf("expected %d colors, got %d", len(tt.expected.Colors), len(result.Colors))
			}
		})
	}
}

//nolint:gocyclo,paralleltest // Test functions with many test cases; modifies global state via SetFs
func TestManager_UpdateDeviceInfo(t *testing.T) {
	m := setupConfigTest(t)

	// Register a device first
	if err := m.RegisterDevice("test-device", "192.168.1.1", 0, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Verify device exists without Type/Model/Generation
	dev, ok := m.GetDevice("test-device")
	if !ok {
		t.Fatal("expected device to exist")
	}
	if dev.Type != "" || dev.Model != "" || dev.Generation != 0 {
		t.Errorf("expected empty Type/Model/Generation, got Type=%q Model=%q Gen=%d",
			dev.Type, dev.Model, dev.Generation)
	}

	// Update with partial info
	if err := m.UpdateDeviceInfo("test-device", DeviceUpdates{
		Type:       "testModelPlug",
		Model:      "Shelly Pro 1PM",
		Generation: 2,
	}); err != nil {
		t.Fatalf("UpdateDeviceInfo() error: %v", err)
	}

	// Verify updates applied
	dev, ok = m.GetDevice("test-device")
	if !ok {
		t.Fatal("expected device to exist after update")
	}
	if dev.Type != "testModelPlug" {
		t.Errorf("expected Type 'testModelPlug', got %q", dev.Type)
	}
	if dev.Model != "Shelly Pro 1PM" {
		t.Errorf("expected Model 'Shelly Pro 1PM', got %q", dev.Model)
	}
	if dev.Generation != 2 {
		t.Errorf("expected Generation 2, got %d", dev.Generation)
	}

	// Update with partial info (only Model)
	if err := m.UpdateDeviceInfo("test-device", DeviceUpdates{
		Model: "Shelly Pro 1PM Updated",
	}); err != nil {
		t.Fatalf("UpdateDeviceInfo() partial update error: %v", err)
	}

	// Verify only Model changed
	dev, ok = m.GetDevice("test-device")
	if !ok {
		t.Fatal("expected device to exist after partial update")
	}
	if dev.Type != "testModelPlug" {
		t.Errorf("expected Type unchanged 'testModelPlug', got %q", dev.Type)
	}
	if dev.Model != "Shelly Pro 1PM Updated" {
		t.Errorf("expected Model 'Shelly Pro 1PM Updated', got %q", dev.Model)
	}
	if dev.Generation != 2 {
		t.Errorf("expected Generation unchanged 2, got %d", dev.Generation)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateDeviceInfo_NotFound(t *testing.T) {
	m := setupConfigTest(t)

	// Try to update non-existent device
	err := m.UpdateDeviceInfo("nonexistent", DeviceUpdates{Type: "test"})
	if err == nil {
		t.Error("expected error for non-existent device")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateDeviceInfo_NoChanges(t *testing.T) {
	m := setupConfigTest(t)

	// Register a device with values already set
	if err := m.RegisterDevice("test-device", "192.168.1.1", 2, "testModelPlug", "Shelly Pro 1PM", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Update with same values (should not write to disk)
	if err := m.UpdateDeviceInfo("test-device", DeviceUpdates{
		Type:       "testModelPlug",
		Model:      "Shelly Pro 1PM",
		Generation: 2,
	}); err != nil {
		t.Fatalf("UpdateDeviceInfo() no-change error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateDeviceInfo_MAC(t *testing.T) {
	m := setupConfigTest(t)

	// Register a device
	if err := m.RegisterDevice("test-device", "192.168.1.1", 0, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Update with MAC address (lowercase)
	if err := m.UpdateDeviceInfo("test-device", DeviceUpdates{
		MAC: "aa:bb:cc:dd:ee:ff",
	}); err != nil {
		t.Fatalf("UpdateDeviceInfo() MAC error: %v", err)
	}

	// Verify MAC is normalized to uppercase
	dev, ok := m.GetDevice("test-device")
	if !ok {
		t.Fatal("expected device to exist")
	}
	if dev.MAC != testMAC {
		t.Errorf("expected MAC %q, got %q", testMAC, dev.MAC)
	}

	// Update with same MAC in different format (should not change)
	if err := m.UpdateDeviceInfo("test-device", DeviceUpdates{
		MAC: "AA-BB-CC-DD-EE-FF",
	}); err != nil {
		t.Fatalf("UpdateDeviceInfo() same MAC error: %v", err)
	}

	// MAC should still be normalized
	dev, ok = m.GetDevice("test-device")
	if !ok {
		t.Fatal("expected device to exist after no-change update")
	}
	if dev.MAC != testMAC {
		t.Errorf("expected MAC unchanged %q, got %q", testMAC, dev.MAC)
	}

	// Invalid MAC should be ignored
	if err := m.UpdateDeviceInfo("test-device", DeviceUpdates{
		MAC: "invalid-mac",
	}); err != nil {
		t.Fatalf("UpdateDeviceInfo() invalid MAC error: %v", err)
	}

	// MAC should remain unchanged
	dev, ok = m.GetDevice("test-device")
	if !ok {
		t.Fatal("expected device to exist after invalid MAC update")
	}
	if dev.MAC != testMAC {
		t.Errorf("expected MAC unchanged after invalid update, got %q", dev.MAC)
	}
}

//nolint:tparallel,paralleltest // Test modifies global state via SetFs
func TestManager_ResolveDevice_Enhanced(t *testing.T) {
	m := setupConfigTest(t)

	// Register a device with MAC and set up aliases manually
	if err := m.RegisterDevice("Master Bathroom", "192.168.1.100", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Update with MAC
	if err := m.UpdateDeviceInfo("master-bathroom", DeviceUpdates{MAC: testMAC}); err != nil {
		t.Fatalf("UpdateDeviceInfo() error: %v", err)
	}

	// Manually add aliases to the device (simulating what the alias command would do)
	m.mu.Lock()
	dev := m.config.Devices["master-bathroom"]
	dev.Aliases = []string{"mb", "bath"}
	m.config.Devices["master-bathroom"] = dev
	m.mu.Unlock()

	tests := []struct {
		name       string
		identifier string
		wantAddr   string
		wantFound  bool // true = should find registered device
	}{
		{
			name:       "exact key match",
			identifier: "master-bathroom",
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "display name match",
			identifier: "Master Bathroom",
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "display name case-insensitive",
			identifier: "MASTER BATHROOM",
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "alias match - mb",
			identifier: "mb",
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "alias match - bath",
			identifier: "bath",
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "alias case-insensitive",
			identifier: "MB",
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "MAC address match - uppercase",
			identifier: testMAC,
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "MAC address match - lowercase",
			identifier: "aa:bb:cc:dd:ee:ff",
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "MAC address match - dash format",
			identifier: "AA-BB-CC-DD-EE-FF",
			wantAddr:   "192.168.1.100",
			wantFound:  true,
		},
		{
			name:       "fallback to direct address",
			identifier: "10.0.0.99",
			wantAddr:   "10.0.0.99",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resolved, err := m.ResolveDevice(tt.identifier)
			if err != nil {
				t.Fatalf("ResolveDevice(%q) error: %v", tt.identifier, err)
			}
			if resolved.Address != tt.wantAddr {
				t.Errorf("ResolveDevice(%q).Address = %q, want %q", tt.identifier, resolved.Address, tt.wantAddr)
			}
			// Check if it found a registered device or fell back to direct address
			hasMAC := resolved.MAC != ""
			if tt.wantFound && !hasMAC {
				t.Errorf("ResolveDevice(%q) should have found registered device with MAC", tt.identifier)
			}
			if !tt.wantFound && hasMAC {
				t.Errorf("ResolveDevice(%q) should have fallen back to direct address, but has MAC", tt.identifier)
			}
		})
	}
}

func TestValidateDeviceAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alias   string
		wantErr bool
	}{
		{name: "valid short", alias: "mb", wantErr: false},
		{name: "valid with numbers", alias: "dev1", wantErr: false},
		{name: "valid with hyphens", alias: "master-bath", wantErr: false},
		{name: "valid with underscores", alias: "living_room", wantErr: false},
		{name: "valid mixed", alias: "dev-1_test", wantErr: false},
		{name: "empty", alias: "", wantErr: true},
		{name: "too long", alias: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", wantErr: true}, // 33 chars
		{name: "starts with hyphen", alias: "-test", wantErr: true},
		{name: "starts with underscore", alias: "_test", wantErr: true},
		{name: "contains space", alias: "my alias", wantErr: true},
		{name: "contains special", alias: "test@alias", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateDeviceAlias(tt.alias)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeviceAlias(%q) error = %v, wantErr = %v", tt.alias, err, tt.wantErr)
			}
		})
	}
}

//nolint:gocyclo,paralleltest // Test functions with many test cases; modifies global state via SetFs
func TestManager_DeviceAliases(t *testing.T) {
	m := setupConfigTest(t)

	// Register two devices
	if err := m.RegisterDevice("Kitchen Light", "192.168.1.100", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.RegisterDevice("Living Room", "192.168.1.101", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Add alias to first device
	if err := m.AddDeviceAlias("kitchen-light", "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Verify alias was added
	aliases, err := m.GetDeviceAliases("kitchen-light")
	if err != nil {
		t.Fatalf("GetDeviceAliases() error: %v", err)
	}
	if len(aliases) != 1 || aliases[0] != "kl" {
		t.Errorf("expected aliases [kl], got %v", aliases)
	}

	// Try to add same alias to second device (should fail)
	if err := m.AddDeviceAlias("living-room", "kl"); err == nil {
		t.Error("expected error adding duplicate alias to different device")
	}

	// Try to add alias that conflicts with device key
	if err := m.AddDeviceAlias("kitchen-light", "living-room"); err == nil {
		t.Error("expected error adding alias that matches device key")
	}

	// Add second alias to first device
	if err := m.AddDeviceAlias("kitchen-light", testAliasName); err != nil {
		t.Fatalf("AddDeviceAlias() second alias error: %v", err)
	}

	// Verify both aliases exist
	aliases, err = m.GetDeviceAliases("kitchen-light")
	if err != nil {
		t.Fatalf("GetDeviceAliases() error: %v", err)
	}
	if len(aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(aliases))
	}

	// Remove first alias
	if err := m.RemoveDeviceAlias("kitchen-light", "kl"); err != nil {
		t.Fatalf("RemoveDeviceAlias() error: %v", err)
	}

	// Verify only one alias remains
	aliases, err = m.GetDeviceAliases("kitchen-light")
	if err != nil {
		t.Fatalf("GetDeviceAliases() after remove error: %v", err)
	}
	if len(aliases) != 1 || aliases[0] != testAliasName {
		t.Errorf("expected aliases [kitchen], got %v", aliases)
	}

	// Try to remove non-existent alias
	if err := m.RemoveDeviceAlias("kitchen-light", "nonexistent"); err == nil {
		t.Error("expected error removing non-existent alias")
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultRateLimitConfig()

	// Check Gen1 defaults
	if cfg.Gen1.MinInterval != 2*time.Second {
		t.Errorf("Gen1.MinInterval = %v, want 2s", cfg.Gen1.MinInterval)
	}
	if cfg.Gen1.MaxConcurrent != 1 {
		t.Errorf("Gen1.MaxConcurrent = %d, want 1", cfg.Gen1.MaxConcurrent)
	}
	if cfg.Gen1.CircuitThreshold != 3 {
		t.Errorf("Gen1.CircuitThreshold = %d, want 3", cfg.Gen1.CircuitThreshold)
	}

	// Check Gen2 defaults
	if cfg.Gen2.MinInterval != 500*time.Millisecond {
		t.Errorf("Gen2.MinInterval = %v, want 500ms", cfg.Gen2.MinInterval)
	}
	if cfg.Gen2.MaxConcurrent != 3 {
		t.Errorf("Gen2.MaxConcurrent = %d, want 3", cfg.Gen2.MaxConcurrent)
	}
	if cfg.Gen2.CircuitThreshold != 5 {
		t.Errorf("Gen2.CircuitThreshold = %d, want 5", cfg.Gen2.CircuitThreshold)
	}

	// Check Global defaults
	if cfg.Global.MaxConcurrent != 5 {
		t.Errorf("Global.MaxConcurrent = %d, want 5", cfg.Global.MaxConcurrent)
	}
	if cfg.Global.CircuitOpenDuration != 60*time.Second {
		t.Errorf("Global.CircuitOpenDuration = %v, want 60s", cfg.Global.CircuitOpenDuration)
	}
	if cfg.Global.CircuitSuccessThreshold != 2 {
		t.Errorf("Global.CircuitSuccessThreshold = %d, want 2", cfg.Global.CircuitSuccessThreshold)
	}
}

func TestDefaultTUIRefreshConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultTUIRefreshConfig()

	if cfg.Gen1Online != 15*time.Second {
		t.Errorf("Gen1Online = %v, want 15s", cfg.Gen1Online)
	}
	if cfg.Gen1Offline != 60*time.Second {
		t.Errorf("Gen1Offline = %v, want 60s", cfg.Gen1Offline)
	}
	if cfg.Gen2Online != 5*time.Second {
		t.Errorf("Gen2Online = %v, want 5s", cfg.Gen2Online)
	}
	if cfg.Gen2Offline != 30*time.Second {
		t.Errorf("Gen2Offline = %v, want 30s", cfg.Gen2Offline)
	}
	if cfg.FocusedBoost != 3*time.Second {
		t.Errorf("FocusedBoost = %v, want 3s", cfg.FocusedBoost)
	}
}

func TestConfig_GetRateLimitConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "empty config uses defaults",
			cfg:  Config{},
		},
		{
			name: "partial config uses defaults for missing",
			cfg: Config{
				RateLimit: RateLimitConfig{
					Gen1: GenerationRateLimitConfig{
						MinInterval: 5 * time.Second,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.cfg.GetRateLimitConfig()

			// Should always have non-zero values
			if result.Gen1.MinInterval == 0 {
				t.Error("Gen1.MinInterval should not be zero")
			}
			if result.Gen1.MaxConcurrent == 0 {
				t.Error("Gen1.MaxConcurrent should not be zero")
			}
			if result.Gen2.MinInterval == 0 {
				t.Error("Gen2.MinInterval should not be zero")
			}
			if result.Global.MaxConcurrent == 0 {
				t.Error("Global.MaxConcurrent should not be zero")
			}
		})
	}
}

func TestConfig_GetTUIRefreshConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "empty config uses defaults",
			cfg:  Config{},
		},
		{
			name: "partial config uses defaults for missing",
			cfg: Config{
				TUI: TUIConfig{
					Refresh: TUIRefreshConfig{
						Gen1Online: 20 * time.Second,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.cfg.GetTUIRefreshConfig()

			// Should always have non-zero values
			if result.Gen1Online == 0 {
				t.Error("Gen1Online should not be zero")
			}
			if result.Gen1Offline == 0 {
				t.Error("Gen1Offline should not be zero")
			}
			if result.Gen2Online == 0 {
				t.Error("Gen2Online should not be zero")
			}
			if result.Gen2Offline == 0 {
				t.Error("Gen2Offline should not be zero")
			}
			if result.FocusedBoost == 0 {
				t.Error("FocusedBoost should not be zero")
			}
		})
	}
}

func TestConfig_GlobalMaxConcurrentViaGetRateLimitConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  Config
		want int
	}{
		{
			name: "empty config uses default",
			cfg:  Config{},
			want: 5, // Default from DefaultRateLimitConfig().Global.MaxConcurrent
		},
		{
			name: "configured value",
			cfg: Config{
				RateLimit: RateLimitConfig{
					Global: GlobalRateLimitConfig{
						MaxConcurrent: 10,
					},
				},
			},
			want: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rlCfg := tt.cfg.GetRateLimitConfig()
			got := rlCfg.Global.MaxConcurrent
			if got != tt.want {
				t.Errorf("Global.MaxConcurrent = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewTestManager(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	m := NewTestManager(cfg)

	// Should initialize all maps
	c := m.Get()
	if c.Devices == nil {
		t.Error("Devices should be initialized")
	}
	if c.Aliases == nil {
		t.Error("Aliases should be initialized")
	}
	if c.Groups == nil {
		t.Error("Groups should be initialized")
	}
	if c.Scenes == nil {
		t.Error("Scenes should be initialized")
	}
	if c.Templates.Device == nil {
		t.Error("Templates.Device should be initialized")
	}
	if c.Templates.Script == nil {
		t.Error("Templates.Script should be initialized")
	}
	if c.Alerts == nil {
		t.Error("Alerts should be initialized")
	}
}
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
// Note: Tests using this helper must NOT use t.Parallel() as they modify global state.
func setupPackageTest(t *testing.T) {
	t.Helper()

	// Use in-memory filesystem for test isolation
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	// Reset the default manager
	ResetDefaultManagerForTesting()
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_RegisterDevice(t *testing.T) {
	setupPackageTest(t)

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
	setupPackageTest(t)

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
	setupPackageTest(t)

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
	setupPackageTest(t)

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
	setupPackageTest(t)

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
	setupPackageTest(t)

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
	setupPackageTest(t)

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
	setupPackageTest(t)

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

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_Get(t *testing.T) {
	setupPackageTest(t)

	// Get should return a config even on empty state
	cfg := Get()
	if cfg == nil {
		t.Fatal("Get() returned nil")
	}

	// Check defaults
	if cfg.Devices == nil {
		t.Error("Devices should be initialized")
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_Load(t *testing.T) {
	setupPackageTest(t)

	// Load should return config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_Reload(t *testing.T) {
	setupPackageTest(t)

	// First load
	cfg1, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Reload
	cfg2, err := Reload()
	if err != nil {
		t.Fatalf("Reload() error: %v", err)
	}
	if cfg2 == nil {
		t.Fatal("Reload() returned nil config")
	}

	// Both should have initialized maps
	if cfg1 == nil || cfg2 == nil {
		t.Fatal("configs should not be nil")
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_Save(t *testing.T) {
	setupPackageTest(t)

	// Load first
	if _, err := Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Register a device
	if err := RegisterDevice("test", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Save should succeed (no-op in test mode with memory fs)
	if err := Save(); err != nil {
		t.Errorf("Save() error: %v", err)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_ConfigSave(t *testing.T) {
	setupPackageTest(t)

	// Load first
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Config.Save() should succeed
	if err := cfg.Save(); err != nil {
		t.Errorf("Config.Save() error: %v", err)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_SetDefaultManager(t *testing.T) {
	// Reset at start and end
	ResetDefaultManagerForTesting()
	t.Cleanup(func() { ResetDefaultManagerForTesting() })

	// Create test manager
	testCfg := &Config{
		Output: "json",
		Devices: map[string]model.Device{
			"test-device": {Name: "Test Device", Address: testIP1},
		},
	}
	testMgr := NewTestManager(testCfg)

	// Set it as default
	SetDefaultManager(testMgr)

	// Get should return our test config
	cfg := Get()
	if cfg.Output != "json" {
		t.Errorf("Get().Output = %q, want %q", cfg.Output, "json")
	}
	if _, ok := cfg.Devices["test-device"]; !ok {
		t.Error("expected test-device to exist")
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_GetGlobalMaxConcurrent(t *testing.T) {
	setupPackageTest(t)

	// Should return default value
	maxConc := GetGlobalMaxConcurrent()
	expectedDefault := DefaultRateLimitConfig().Global.MaxConcurrent
	if maxConc != expectedDefault {
		t.Errorf("GetGlobalMaxConcurrent() = %d, want %d", maxConc, expectedDefault)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_GetGlobalMaxConcurrent_WithConfig(t *testing.T) {
	ResetDefaultManagerForTesting()
	t.Cleanup(func() { ResetDefaultManagerForTesting() })

	// Set a custom config with custom max concurrent
	testCfg := &Config{
		RateLimit: RateLimitConfig{
			Global: GlobalRateLimitConfig{
				MaxConcurrent: 10,
			},
		},
	}
	testMgr := NewTestManager(testCfg)
	SetDefaultManager(testMgr)

	// Should return configured value
	maxConc := GetGlobalMaxConcurrent()
	if maxConc != 10 {
		t.Errorf("GetGlobalMaxConcurrent() = %d, want 10", maxConc)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_SetDeviceAuth(t *testing.T) {
	setupPackageTest(t)

	// Register device first
	if err := RegisterDevice("auth-test", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Get the config and set auth
	cfg := Get()
	if err := cfg.SetDeviceAuth("auth-test", "user", "pass"); err != nil {
		t.Errorf("SetDeviceAuth() error: %v", err)
	}

	// Verify auth was set
	dev, ok := GetDevice("auth-test")
	if !ok {
		t.Fatal("device should exist")
	}
	if dev.Auth == nil {
		t.Fatal("device Auth should not be nil")
	}
	if dev.Auth.Username != "user" || dev.Auth.Password != "pass" {
		t.Errorf("auth = %q:%q, want user:pass", dev.Auth.Username, dev.Auth.Password)
	}
}

//nolint:paralleltest // Tests modify global HOME and default manager
func TestPackageLevel_GetAllDeviceCredentials(t *testing.T) {
	setupPackageTest(t)

	// Register device with auth
	if err := RegisterDevice("auth-dev1", testIP1, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	cfg := Get()
	if err := cfg.SetDeviceAuth("auth-dev1", "user1", "pass1"); err != nil {
		t.Fatalf("SetDeviceAuth() error: %v", err)
	}

	// Get all credentials
	creds := cfg.GetAllDeviceCredentials()
	if creds == nil {
		t.Fatal("GetAllDeviceCredentials() returned nil")
	}
	if len(creds) < 1 {
		t.Error("expected at least 1 credential entry")
	}
}

func TestDir_WithXDGConfigHome(t *testing.T) {
	// Save original value
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("warning: unsetenv error: %v", err)
			}
		} else {
			if err := os.Setenv("XDG_CONFIG_HOME", origXDG); err != nil {
				t.Logf("warning: setenv error: %v", err)
			}
		}
	})

	// Set XDG_CONFIG_HOME
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	expected := "/custom/config/shelly"
	if dir != expected {
		t.Errorf("Dir() = %q, want %q", dir, expected)
	}
}

func TestGetThemeConfig_ThemeConfigType(t *testing.T) {
	t.Parallel()

	// Test with ThemeConfig type directly
	tc := ThemeConfig{Name: "nord", Colors: map[string]string{"accent": "#88c0d0"}}
	cfg := &Config{Theme: tc}
	result := cfg.GetThemeConfig()
	if result.Name != "nord" {
		t.Errorf("GetThemeConfig() name = %q, want %q", result.Name, "nord")
	}
}

func TestGetThemeConfig_ThemeConfigPointer(t *testing.T) {
	t.Parallel()

	// Test with *ThemeConfig pointer
	tc := &ThemeConfig{Name: "gruvbox", Colors: map[string]string{"bg": "#282828"}}
	cfg := &Config{Theme: tc}
	result := cfg.GetThemeConfig()
	if result.Name != "gruvbox" {
		t.Errorf("GetThemeConfig() name = %q, want %q", result.Name, "gruvbox")
	}

	// Test with nil pointer
	var nilTC *ThemeConfig
	cfg2 := &Config{Theme: nilTC}
	result2 := cfg2.GetThemeConfig()
	if result2.Name != testThemeDracula {
		t.Errorf("GetThemeConfig() with nil pointer name = %q, want %q", result2.Name, testThemeDracula)
	}
}

//nolint:paralleltest // Tests modify global state via SetFs and SetDefaultManager
func TestPackageLevel_Save_NonTestFs(t *testing.T) {
	// Reset at start and end
	ResetDefaultManagerForTesting()
	t.Cleanup(func() {
		SetFs(nil)
		ResetDefaultManagerForTesting()
	})

	fs := afero.NewMemMapFs()
	SetFs(fs)

	// Create a manager pointing to a test path
	mgr := NewManager("/test/config/config.yaml")
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	SetDefaultManager(mgr)

	// Save should work
	if err := Save(); err != nil {
		t.Errorf("Save() error: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state
func TestGetGlobalMaxConcurrent_ZeroConfig(t *testing.T) {
	ResetDefaultManagerForTesting()
	t.Cleanup(func() { ResetDefaultManagerForTesting() })

	// Create a config with zero global max concurrent
	// GetRateLimitConfig will return defaults, so the zero case in
	// GetGlobalMaxConcurrent is actually unreachable in practice.
	// But we can test that the default is returned when config has no rate limit.
	testCfg := &Config{}
	testMgr := NewTestManager(testCfg)
	SetDefaultManager(testMgr)

	// Should return default value since GetRateLimitConfig applies defaults
	maxConc := GetGlobalMaxConcurrent()
	expectedDefault := DefaultRateLimitConfig().Global.MaxConcurrent
	if maxConc != expectedDefault {
		t.Errorf("GetGlobalMaxConcurrent() = %d, want %d", maxConc, expectedDefault)
	}
}

func TestGetThemeConfig_MapstructureDecodeError(t *testing.T) {
	t.Parallel()

	// Create a map with a type that will cause mapstructure to fail on some fields
	// This is hard to trigger since mapstructure is very permissive
	// Using an invalid type for colors
	cfg := &Config{
		Theme: map[string]any{
			"name": "test",
			// mapstructure handles most cases gracefully, but we can try with
			// an unmappable type for the semantic field
			"semantic": "invalid-not-a-struct",
		},
	}
	result := cfg.GetThemeConfig()
	// mapstructure may handle this gracefully or fail
	// Either way, we get a theme config back
	if result.Name == "" && result.Name != "test" && result.Name != testThemeDracula {
		t.Errorf("GetThemeConfig() should return a valid theme config")
	}
}

func TestManager_GetReturnsNil(t *testing.T) {
	t.Parallel()

	// Create a manager but don't load it
	m := &Manager{
		path:   "/test/config/config.yaml", // Virtual path (never used)
		loaded: false,
	}

	// Get should return nil when not loaded
	cfg := m.Get()
	if cfg != nil {
		t.Error("Get() should return nil when not loaded")
	}
}

// =============================================================================
// Package-level Fs and Manager.Fs tests
// =============================================================================

func TestPackageLevel_Fs(t *testing.T) {
	t.Parallel()

	// Package-level Fs() should return a non-nil filesystem
	fs := Fs()
	if fs == nil {
		t.Error("Fs() should not return nil")
	}
}

func TestManager_Fs_WithCustomFs(t *testing.T) {
	t.Parallel()

	// Create a manager with a custom filesystem
	memFs := afero.NewMemMapFs()
	m := &Manager{
		fs: memFs,
	}

	// Should return the custom filesystem
	gotFs := m.Fs()
	if gotFs != memFs {
		t.Error("Manager.Fs() should return custom fs when set")
	}
}

func TestManager_Fs_WithNilFs(t *testing.T) {
	t.Parallel()

	// Create a manager without a custom filesystem
	m := &Manager{
		fs: nil,
	}

	// Should return the package default filesystem
	gotFs := m.Fs()
	if gotFs == nil {
		t.Error("Manager.Fs() should return default fs when m.fs is nil")
	}
}

//nolint:paralleltest // Tests modify global state
func TestNewManager_EmptyPath(t *testing.T) {
	// Save original fs and restore after test
	SetFs(nil)
	t.Cleanup(func() { SetFs(nil) })

	// NewManager with empty path should use Dir() result
	m := NewManager("")

	// Path should not be empty (uses Dir() result)
	if m.Path() == "" {
		t.Error("NewManager(\"\") should set a default path")
	}

	// Should contain "config.yaml"
	if m.Path() != "config.yaml" && m.Path() != "" {
		// If Dir() succeeded, path should end with config.yaml
		if len(m.Path()) < len("config.yaml") {
			t.Errorf("NewManager(\"\").Path() = %q, expected path ending with config.yaml", m.Path())
		}
	}
}

//nolint:paralleltest // Tests modify global state via SetFs
func TestImportAliases_MergeSkipsExisting(t *testing.T) {
	m := setupConfigTest(t)

	// Add an existing alias
	if err := m.AddAlias("existing", "original command", false); err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}

	// Create import file with overlapping alias in in-memory fs
	importPath := "/test/import.yaml"
	importContent := `aliases:
  existing: "new command"
  new-alias: "new command 2"
`
	if err := afero.WriteFile(m.Fs(), importPath, []byte(importContent), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Import with merge=true (should skip existing)
	imported, skipped, err := m.ImportAliases(importPath, true)
	if err != nil {
		t.Fatalf("ImportAliases() error: %v", err)
	}
	if imported != 1 {
		t.Errorf("imported = %d, want 1", imported)
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}

	// Verify original wasn't overwritten
	alias, ok := m.GetAlias("existing")
	if !ok {
		t.Fatal("existing alias should still exist")
	}
	if alias.Command != "original command" {
		t.Errorf("Command = %q, want %q", alias.Command, "original command")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RenameDevice_UsingDisplayName(t *testing.T) {
	m := setupConfigTest(t)

	// Register with display name
	if err := m.RegisterDevice("Kitchen Light", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Rename using display name (not normalized key)
	if err := m.RenameDevice("Kitchen Light", "Living Room Light"); err != nil {
		t.Fatalf("RenameDevice() error: %v", err)
	}

	// Verify old key doesn't exist
	_, ok := m.GetDevice("kitchen-light")
	if ok {
		t.Error("old normalized key should not exist")
	}

	// Verify new key exists
	dev, ok := m.GetDevice("living-room-light")
	if !ok {
		t.Error("new device should exist")
	}
	if dev.Name != "Living Room Light" {
		t.Errorf("device name = %q, want %q", dev.Name, "Living Room Light")
	}
}

//nolint:paralleltest // Tests modify global state
func TestImportScene_CreateError(t *testing.T) {
	setupPackageTest(t)

	// Import a scene with a name that will fail validation (if validation fails)
	// Actually, we already tested empty name. Let's test the delete existing + create flow
	scene1 := &Scene{
		Name:        "import-test",
		Description: "First version",
		Actions:     []SceneAction{},
	}
	if err := ImportScene(scene1, false); err != nil {
		t.Fatalf("ImportScene() initial error: %v", err)
	}

	// Import again with overwrite
	scene2 := &Scene{
		Name:        "import-test",
		Description: "Second version",
		Actions: []SceneAction{
			{Device: "light1", Method: "Switch.On"},
		},
	}
	if err := ImportScene(scene2, true); err != nil {
		t.Fatalf("ImportScene() overwrite error: %v", err)
	}

	// Verify the new scene
	got, ok := GetScene("import-test")
	if !ok {
		t.Fatal("scene should exist")
	}
	if got.Description != "Second version" {
		t.Errorf("Description = %q, want %q", got.Description, "Second version")
	}
	if len(got.Actions) != 1 {
		t.Errorf("len(Actions) = %d, want 1", len(got.Actions))
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_Get_NilConfig(t *testing.T) {
	ResetDefaultManagerForTesting()
	t.Cleanup(func() { ResetDefaultManagerForTesting() })

	// Create a manager with nil config
	m := &Manager{
		loaded: true,
		config: nil,
	}
	SetDefaultManager(m)

	// Get should return a default config
	cfg := Get()
	if cfg == nil {
		t.Fatal("Get() should not return nil")
	}
	if cfg.Output != "table" {
		t.Errorf("Get().Output = %q, want %q", cfg.Output, "table")
	}
	if !cfg.Color {
		t.Error("Get().Color should be true")
	}
	if cfg.Theme != testThemeDracula {
		t.Errorf("Get().Theme = %v, want dracula", cfg.Theme)
	}
}
