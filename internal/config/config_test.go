package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_Defaults(t *testing.T) {
	t.Parallel()

	// Create temp dir for isolated config
	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	c := m.Get()

	if c.Output != "" && c.Output != "table" {
		t.Errorf("expected output '' or 'table', got %q", c.Output)
	}
}

func TestManager_InitializesMaps(t *testing.T) {
	t.Parallel()

	// Create temp dir for isolated config
	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

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
	if c.Templates == nil {
		t.Error("expected Templates map to be initialized")
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
				TUI: TUIConfig{
					RefreshInterval: 5,
				},
			},
			expected: nil,
		},
		{
			name: "TUI theme set",
			config: &Config{
				TUI: TUIConfig{
					RefreshInterval: 5,
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

func TestManager_UpdateDeviceInfo(t *testing.T) {
	t.Parallel()

	// Create temp dir for isolated config
	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

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
		Type:       "SPSW-001PE16EU",
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
	if dev.Type != "SPSW-001PE16EU" {
		t.Errorf("expected Type 'SPSW-001PE16EU', got %q", dev.Type)
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
	if dev.Type != "SPSW-001PE16EU" {
		t.Errorf("expected Type unchanged 'SPSW-001PE16EU', got %q", dev.Type)
	}
	if dev.Model != "Shelly Pro 1PM Updated" {
		t.Errorf("expected Model 'Shelly Pro 1PM Updated', got %q", dev.Model)
	}
	if dev.Generation != 2 {
		t.Errorf("expected Generation unchanged 2, got %d", dev.Generation)
	}
}

func TestManager_UpdateDeviceInfo_NotFound(t *testing.T) {
	t.Parallel()

	// Create temp dir for isolated config
	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Try to update non-existent device
	err := m.UpdateDeviceInfo("nonexistent", DeviceUpdates{Type: "test"})
	if err == nil {
		t.Error("expected error for non-existent device")
	}
}

func TestManager_UpdateDeviceInfo_NoChanges(t *testing.T) {
	t.Parallel()

	// Create temp dir for isolated config
	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Register a device with values already set
	if err := m.RegisterDevice("test-device", "192.168.1.1", 2, "SPSW-001PE16EU", "Shelly Pro 1PM", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Update with same values (should not write to disk)
	if err := m.UpdateDeviceInfo("test-device", DeviceUpdates{
		Type:       "SPSW-001PE16EU",
		Model:      "Shelly Pro 1PM",
		Generation: 2,
	}); err != nil {
		t.Fatalf("UpdateDeviceInfo() no-change error: %v", err)
	}
}

func TestManager_UpdateDeviceInfo_MAC(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

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
	if dev.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected MAC 'AA:BB:CC:DD:EE:FF', got %q", dev.MAC)
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
	if dev.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected MAC unchanged 'AA:BB:CC:DD:EE:FF', got %q", dev.MAC)
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
	if dev.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected MAC unchanged after invalid update, got %q", dev.MAC)
	}
}

func TestManager_ResolveDevice_Enhanced(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Register a device with MAC and set up aliases manually
	if err := m.RegisterDevice("Master Bathroom", "192.168.1.100", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Update with MAC
	if err := m.UpdateDeviceInfo("master-bathroom", DeviceUpdates{MAC: "AA:BB:CC:DD:EE:FF"}); err != nil {
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
			identifier: "AA:BB:CC:DD:EE:FF",
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

func TestManager_DeviceAliases(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

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
	if err := m.AddDeviceAlias("kitchen-light", "kitchen"); err != nil {
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
	if len(aliases) != 1 || aliases[0] != "kitchen" {
		t.Errorf("expected aliases [kitchen], got %v", aliases)
	}

	// Try to remove non-existent alias
	if err := m.RemoveDeviceAlias("kitchen-light", "nonexistent"); err == nil {
		t.Error("expected error removing non-existent alias")
	}
}
