package config

import (
	"os"
	"testing"
	"time"
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
