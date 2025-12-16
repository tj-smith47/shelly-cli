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
