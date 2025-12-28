package shelly

import (
	"testing"
)

func TestDeviceTemplate_Fields(t *testing.T) {
	t.Parallel()

	template := DeviceTemplate{
		Model:      "SNSW-001P16EU",
		App:        "Plus1PM",
		Generation: 2,
		Config: map[string]any{
			"switch:0": map[string]any{
				"name": "Test Switch",
			},
		},
	}

	if template.Model != "SNSW-001P16EU" {
		t.Errorf("expected Model 'SNSW-001P16EU', got %q", template.Model)
	}
	if template.App != "Plus1PM" {
		t.Errorf("expected App 'Plus1PM', got %q", template.App)
	}
	if template.Generation != 2 {
		t.Errorf("expected Generation 2, got %d", template.Generation)
	}
	if template.Config == nil {
		t.Error("expected Config to be set")
	}
}

func TestSanitizeConfig(t *testing.T) { //nolint:gocyclo // comprehensive test coverage
	t.Parallel()

	t.Run("removes WiFi passwords", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]any{
			"wifi": map[string]any{
				"sta": map[string]any{
					"ssid": "MyNetwork",
					"pass": "secret123",
				},
				"sta1": map[string]any{
					"ssid": "BackupNetwork",
					"pass": "secret456",
				},
				"ap": map[string]any{
					"ssid": "Shelly-AP",
					"pass": "appass",
				},
			},
		}

		sanitizeConfig(cfg)

		wifi, ok := cfg["wifi"].(map[string]any)
		if !ok {
			t.Fatal("wifi should be map[string]any")
		}
		sta, ok := wifi["sta"].(map[string]any)
		if !ok {
			t.Fatal("sta should be map[string]any")
		}
		if _, exists := sta["pass"]; exists {
			t.Error("sta password should be removed")
		}
		if sta["ssid"] != "MyNetwork" {
			t.Error("sta ssid should be preserved")
		}

		sta1, ok := wifi["sta1"].(map[string]any)
		if !ok {
			t.Fatal("sta1 should be map[string]any")
		}
		if _, exists := sta1["pass"]; exists {
			t.Error("sta1 password should be removed")
		}

		ap, ok := wifi["ap"].(map[string]any)
		if !ok {
			t.Fatal("ap should be map[string]any")
		}
		if _, exists := ap["pass"]; exists {
			t.Error("ap password should be removed")
		}
	})

	t.Run("removes auth password", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]any{
			"auth": map[string]any{
				"user":    "admin",
				"pass":    "secretpass",
				"enabled": true,
			},
		}

		sanitizeConfig(cfg)

		auth, ok := cfg["auth"].(map[string]any)
		if !ok {
			t.Fatal("auth should be map[string]any")
		}
		if _, exists := auth["pass"]; exists {
			t.Error("auth password should be removed")
		}
		if auth["user"] != "admin" {
			t.Error("auth user should be preserved")
		}
	})

	t.Run("removes cloud server", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]any{
			"cloud": map[string]any{
				"enabled": true,
				"server":  "shelly-13-eu.shelly.cloud:6012/jrpc",
			},
		}

		sanitizeConfig(cfg)

		cloud, ok := cfg["cloud"].(map[string]any)
		if !ok {
			t.Fatal("cloud should be map[string]any")
		}
		if _, exists := cloud["server"]; exists {
			t.Error("cloud server should be removed")
		}
		if cloud["enabled"] != true {
			t.Error("cloud enabled should be preserved")
		}
	})

	t.Run("handles missing sections", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]any{
			"switch:0": map[string]any{
				"name": "Test",
			},
		}

		// Should not panic
		sanitizeConfig(cfg)

		if cfg["switch:0"] == nil {
			t.Error("other config sections should be preserved")
		}
	})

	t.Run("handles nil values", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]any{
			"wifi": nil,
			"auth": nil,
		}

		// Should not panic
		sanitizeConfig(cfg)
	})
}

func TestCompareForApply(t *testing.T) {
	t.Parallel()

	t.Run("detects additions", func(t *testing.T) {
		t.Parallel()

		current := map[string]any{}
		template := map[string]any{
			"new_key": "value",
		}

		changes := compareForApply(current, template)

		if len(changes) != 1 {
			t.Errorf("expected 1 change, got %d", len(changes))
		}
		if len(changes) > 0 && changes[0][0] != '+' {
			t.Error("expected addition indicator '+'")
		}
	})

	t.Run("detects modifications", func(t *testing.T) {
		t.Parallel()

		current := map[string]any{
			"key": "old_value",
		}
		template := map[string]any{
			"key": "new_value",
		}

		changes := compareForApply(current, template)

		if len(changes) != 1 {
			t.Errorf("expected 1 change, got %d", len(changes))
		}
		if len(changes) > 0 && changes[0][0] != '~' {
			t.Error("expected modification indicator '~'")
		}
	})

	t.Run("ignores identical values", func(t *testing.T) {
		t.Parallel()

		current := map[string]any{
			"key": "same_value",
		}
		template := map[string]any{
			"key": "same_value",
		}

		changes := compareForApply(current, template)

		if len(changes) != 0 {
			t.Errorf("expected 0 changes, got %d", len(changes))
		}
	})

	t.Run("empty configs", func(t *testing.T) {
		t.Parallel()

		current := map[string]any{}
		template := map[string]any{}

		changes := compareForApply(current, template)

		if len(changes) != 0 {
			t.Errorf("expected 0 changes, got %d", len(changes))
		}
	})
}

func TestSummarizeValue(t *testing.T) {
	t.Parallel()

	t.Run("map value", func(t *testing.T) {
		t.Parallel()

		v := map[string]any{
			"key1": "val1",
			"key2": "val2",
		}

		result := summarizeValue(v)

		if result != "{...2 keys}" {
			t.Errorf("expected '{...2 keys}', got %q", result)
		}
	})

	t.Run("slice value", func(t *testing.T) {
		t.Parallel()

		v := []any{"a", "b", "c"}

		result := summarizeValue(v)

		if result != "[...3 items]" {
			t.Errorf("expected '[...3 items]', got %q", result)
		}
	})

	t.Run("short string", func(t *testing.T) {
		t.Parallel()

		v := "short"

		result := summarizeValue(v)

		if result != `"short"` {
			t.Errorf("expected '\"short\"', got %q", result)
		}
	})

	t.Run("long string truncated", func(t *testing.T) {
		t.Parallel()

		v := "this is a very long string that should be truncated"

		result := summarizeValue(v)

		if len(result) > 25 {
			t.Errorf("expected truncated string, got %q", result)
		}
	})

	t.Run("number value", func(t *testing.T) {
		t.Parallel()

		v := 42.5

		result := summarizeValue(v)

		if result != "42.5" {
			t.Errorf("expected '42.5', got %q", result)
		}
	})

	t.Run("boolean value", func(t *testing.T) {
		t.Parallel()

		v := true

		result := summarizeValue(v)

		if result != "true" {
			t.Errorf("expected 'true', got %q", result)
		}
	})
}
