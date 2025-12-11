package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/viper"
)

func resetConfig() {
	cfgMu.Lock()
	cfg = nil
	cfgOnce = sync.Once{}
	cfgMu.Unlock()
	viper.Reset()
}

//nolint:paralleltest // Test modifies global config state via resetConfig
func TestGet_ReturnsDefaults(t *testing.T) {
	resetConfig()

	c := Get()

	if c.Output != "table" {
		t.Errorf("expected output 'table', got %q", c.Output)
	}
	if !c.Color {
		t.Error("expected color to be true")
	}
	if c.Theme != "dracula" {
		t.Errorf("expected theme 'dracula', got %q", c.Theme)
	}
	if c.APIMode != "local" {
		t.Errorf("expected api_mode 'local', got %q", c.APIMode)
	}
}

//nolint:paralleltest // Test modifies global config state via resetConfig
func TestLoad_InitializesMaps(t *testing.T) {
	resetConfig()

	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if c.Devices == nil {
		t.Error("expected Devices map to be initialized")
	}
	if c.Aliases == nil {
		t.Error("expected Aliases map to be initialized")
	}
	if c.Groups == nil {
		t.Error("expected Groups map to be initialized")
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
