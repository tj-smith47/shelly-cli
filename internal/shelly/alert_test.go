package shelly

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

func TestAlertConditionResult_Fields(t *testing.T) {
	t.Parallel()

	result := AlertConditionResult{
		Triggered: true,
		Value:     "42.5",
	}

	if !result.Triggered {
		t.Error("expected Triggered to be true")
	}
	if result.Value != "42.5" {
		t.Errorf("expected Value '42.5', got %q", result.Value)
	}
}

func TestAlertState_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	state := AlertState{
		LastTriggered: now,
		LastValue:     "100",
		Triggered:     true,
	}

	if !state.Triggered {
		t.Error("expected Triggered to be true")
	}
	if state.LastValue != "100" {
		t.Errorf("expected LastValue '100', got %q", state.LastValue)
	}
	if state.LastTriggered != now {
		t.Error("expected LastTriggered to match")
	}
}

func TestAlertCheckResult_Fields(t *testing.T) {
	t.Parallel()

	result := AlertCheckResult{
		Name:      "power-alert",
		Condition: "power>100",
		Device:    "kitchen-switch",
		Value:     "150.5",
		Action:    AlertActionTriggered,
	}

	if result.Name != "power-alert" {
		t.Errorf("expected Name 'power-alert', got %q", result.Name)
	}
	if result.Condition != "power>100" {
		t.Errorf("expected Condition 'power>100', got %q", result.Condition)
	}
	if result.Device != "kitchen-switch" {
		t.Errorf("expected Device 'kitchen-switch', got %q", result.Device)
	}
	if result.Action != AlertActionTriggered {
		t.Error("expected Action AlertActionTriggered")
	}
}

func TestAlertAction_Constants(t *testing.T) {
	t.Parallel()

	if AlertActionNone != 0 {
		t.Error("AlertActionNone should be 0")
	}
	if AlertActionTriggered != 1 {
		t.Error("AlertActionTriggered should be 1")
	}
	if AlertActionCleared != 2 {
		t.Error("AlertActionCleared should be 2")
	}
}

func TestActionResult_Fields(t *testing.T) {
	t.Parallel()

	result := ActionResult{
		Type:       ActionTypeWebhook,
		StatusCode: 200,
		Output:     []byte("success"),
		Error:      nil,
	}

	if result.Type != ActionTypeWebhook {
		t.Errorf("expected Type 'webhook', got %q", result.Type)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode %d, got %d", http.StatusOK, result.StatusCode)
	}
	if string(result.Output) != "success" {
		t.Errorf("expected Output 'success', got %q", string(result.Output))
	}
}

func TestActionTypeConstants(t *testing.T) {
	t.Parallel()

	if ActionTypeNotify != "notify" {
		t.Error("ActionTypeNotify should be 'notify'")
	}
	if ActionTypeWebhook != "webhook" {
		t.Error("ActionTypeWebhook should be 'webhook'")
	}
	if ActionTypeCommand != "command" {
		t.Error("ActionTypeCommand should be 'command'")
	}
	if ActionTypeUnknown != "unknown" {
		t.Error("ActionTypeUnknown should be 'unknown'")
	}
}

func TestParseThresholdCondition(t *testing.T) {
	t.Parallel()

	t.Run("power greater than", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{
			"switch:0": map[string]any{
				"apower": 150.0,
			},
		}

		result := parseThresholdCondition(status, "power>100")

		if !result.Triggered {
			t.Error("expected condition to be triggered")
		}
		if result.Value != "150.0" {
			t.Errorf("expected Value '150.0', got %q", result.Value)
		}
	})

	t.Run("power less than", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{
			"switch:0": map[string]any{
				"apower": 50.0,
			},
		}

		result := parseThresholdCondition(status, "power<100")

		if !result.Triggered {
			t.Error("expected condition to be triggered")
		}
	})

	t.Run("power not triggered", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{
			"switch:0": map[string]any{
				"apower": 50.0,
			},
		}

		result := parseThresholdCondition(status, "power>100")

		if result.Triggered {
			t.Error("expected condition NOT to be triggered")
		}
	})

	t.Run("unknown condition", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{}

		result := parseThresholdCondition(status, "power=100")

		if result.Triggered {
			t.Error("expected unknown condition to not trigger")
		}
		if result.Value != "unknown condition" {
			t.Errorf("expected Value 'unknown condition', got %q", result.Value)
		}
	})
}

func TestCheckThreshold(t *testing.T) {
	t.Parallel()

	t.Run("valid threshold comparison", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{
			"switch:0": map[string]any{
				"apower": 200.0,
			},
		}

		result := checkThreshold(status, "power>100", ">", func(current, threshold float64) bool {
			return current > threshold
		})

		if !result.Triggered {
			t.Error("expected condition to be triggered")
		}
	})

	t.Run("invalid condition format", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{}

		result := checkThreshold(status, "power", ">", func(current, threshold float64) bool {
			return current > threshold
		})

		if result.Triggered {
			t.Error("expected invalid condition to not trigger")
		}
		if result.Value != "invalid condition" {
			t.Errorf("expected Value 'invalid condition', got %q", result.Value)
		}
	})

	t.Run("invalid threshold value", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{}

		result := checkThreshold(status, "power>abc", ">", func(current, threshold float64) bool {
			return current > threshold
		})

		if result.Triggered {
			t.Error("expected invalid threshold to not trigger")
		}
		if result.Value != "invalid threshold" {
			t.Errorf("expected Value 'invalid threshold', got %q", result.Value)
		}
	})
}

func TestGetMetricValue(t *testing.T) {
	t.Parallel()

	t.Run("extracts power from switch:0", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{
			"switch:0": map[string]any{
				"apower": 100.5,
			},
		}

		value := getMetricValue(status, "power")

		if value != 100.5 {
			t.Errorf("expected 100.5, got %f", value)
		}
	})

	t.Run("extracts temperature from temperature:0", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{
			"temperature:0": map[string]any{
				"tC": 25.5,
			},
		}

		value := getMetricValue(status, "temperature")

		if value != 25.5 {
			t.Errorf("expected 25.5, got %f", value)
		}
	})

	t.Run("extracts voltage", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{
			"pm1:0": map[string]any{
				"voltage": 230.5,
			},
		}

		value := getMetricValue(status, "voltage")

		if value != 230.5 {
			t.Errorf("expected 230.5, got %f", value)
		}
	})

	t.Run("extracts current", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{
			"em:0": map[string]any{
				"current": 5.2,
			},
		}

		value := getMetricValue(status, "current")

		if value != 5.2 {
			t.Errorf("expected 5.2, got %f", value)
		}
	})

	t.Run("returns 0 for missing metric", func(t *testing.T) {
		t.Parallel()

		status := map[string]any{}

		value := getMetricValue(status, "power")

		if value != 0 {
			t.Errorf("expected 0, got %f", value)
		}
	})
}

func TestExtractMetric(t *testing.T) {
	t.Parallel()

	t.Run("extracts apower", func(t *testing.T) {
		t.Parallel()

		comp := map[string]any{
			"apower": 150.0,
		}

		value := extractMetric(comp, "power")
		if value != 150.0 {
			t.Errorf("expected 150.0, got %f", value)
		}
	})

	t.Run("extracts temperature with nested structure", func(t *testing.T) {
		t.Parallel()

		comp := map[string]any{
			"temperature": map[string]any{
				"tC": 22.5,
			},
		}

		value := extractMetric(comp, "temperature")
		if value != 22.5 {
			t.Errorf("expected 22.5, got %f", value)
		}
	})

	t.Run("returns 0 for unknown metric", func(t *testing.T) {
		t.Parallel()

		comp := map[string]any{
			"apower": 100.0,
		}

		value := extractMetric(comp, "unknown")
		if value != 0 {
			t.Errorf("expected 0, got %f", value)
		}
	})
}

func TestExtractTemperature(t *testing.T) {
	t.Parallel()

	t.Run("from nested temperature map", func(t *testing.T) {
		t.Parallel()

		comp := map[string]any{
			"temperature": map[string]any{
				"tC": 30.5,
			},
		}

		value := extractTemperature(comp)
		if value != 30.5 {
			t.Errorf("expected 30.5, got %f", value)
		}
	})

	t.Run("from direct tC field", func(t *testing.T) {
		t.Parallel()

		comp := map[string]any{
			"tC": 28.0,
		}

		value := extractTemperature(comp)
		if value != 28.0 {
			t.Errorf("expected 28.0, got %f", value)
		}
	})

	t.Run("returns 0 when no temperature", func(t *testing.T) {
		t.Parallel()

		comp := map[string]any{
			"power": 100.0,
		}

		value := extractTemperature(comp)
		if value != 0 {
			t.Errorf("expected 0, got %f", value)
		}
	})
}

func TestWebhookResult_Fields(t *testing.T) {
	t.Parallel()

	result := WebhookResult{
		StatusCode: http.StatusCreated,
		Error:      nil,
	}

	if result.StatusCode != http.StatusCreated {
		t.Errorf("expected StatusCode %d, got %d", http.StatusCreated, result.StatusCode)
	}
	if result.Error != nil {
		t.Error("expected Error to be nil")
	}
}

func TestCommandResult_Fields(t *testing.T) {
	t.Parallel()

	result := CommandResult{
		Output: []byte("command output"),
		Error:  nil,
	}

	if string(result.Output) != "command output" {
		t.Errorf("expected Output 'command output', got %q", string(result.Output))
	}
	if result.Error != nil {
		t.Error("expected Error to be nil")
	}
}

func TestExecuteCommand(t *testing.T) {
	t.Parallel()

	t.Run("successful command", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		result := ExecuteCommand(ctx, "echo hello")

		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}
		if string(result.Output) != "hello\n" {
			t.Errorf("expected Output 'hello\\n', got %q", string(result.Output))
		}
	})

	t.Run("failed command", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		result := ExecuteCommand(ctx, "exit 1")

		if result.Error == nil {
			t.Error("expected error for failed command")
		}
	})
}

func TestExecuteAlertAction(t *testing.T) {
	t.Parallel()

	t.Run("notify action", func(t *testing.T) {
		t.Parallel()

		alert := config.Alert{
			Name:   "test-alert",
			Action: "notify",
		}

		result := ExecuteAlertAction(context.Background(), alert, "100")

		if result.Type != ActionTypeNotify {
			t.Errorf("expected Type 'notify', got %q", result.Type)
		}
	})

	t.Run("empty action defaults to notify", func(t *testing.T) {
		t.Parallel()

		alert := config.Alert{
			Name:   "test-alert",
			Action: "",
		}

		result := ExecuteAlertAction(context.Background(), alert, "100")

		if result.Type != ActionTypeNotify {
			t.Errorf("expected Type 'notify', got %q", result.Type)
		}
	})

	t.Run("command action", func(t *testing.T) {
		t.Parallel()

		alert := config.Alert{
			Name:   "test-alert",
			Action: "command:echo test",
		}

		result := ExecuteAlertAction(context.Background(), alert, "100")

		if result.Type != ActionTypeCommand {
			t.Errorf("expected Type 'command', got %q", result.Type)
		}
		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}
	})

	t.Run("unknown action", func(t *testing.T) {
		t.Parallel()

		alert := config.Alert{
			Name:   "test-alert",
			Action: "invalid-action",
		}

		result := ExecuteAlertAction(context.Background(), alert, "100")

		if result.Type != ActionTypeUnknown {
			t.Errorf("expected Type 'unknown', got %q", result.Type)
		}
		if result.Error == nil {
			t.Error("expected error for unknown action")
		}
	})
}
