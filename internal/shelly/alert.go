// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

const statusUnreachable = "unreachable"

// Action type constants.
const (
	ActionTypeNotify  = "notify"
	ActionTypeWebhook = "webhook"
	ActionTypeCommand = "command"
	ActionTypeUnknown = "unknown"
)

// AlertConditionResult holds the result of evaluating an alert condition.
type AlertConditionResult struct {
	Triggered bool
	Value     string
}

// AlertState tracks the state of an alert for edge detection.
type AlertState struct {
	LastTriggered time.Time
	LastValue     string
	Triggered     bool
}

// AlertCheckResult represents what happened when checking an alert.
type AlertCheckResult struct {
	Name      string
	Condition string
	Device    string
	Value     string
	Action    AlertAction
}

// AlertAction indicates what action occurred.
type AlertAction int

const (
	// AlertActionNone means no state change.
	AlertActionNone AlertAction = iota
	// AlertActionTriggered means the alert just triggered.
	AlertActionTriggered
	// AlertActionCleared means the alert condition cleared.
	AlertActionCleared
)

// CheckAlert evaluates an alert and updates state, returning what action to take.
func (s *Service) CheckAlert(ctx context.Context, alert config.Alert, state *AlertState) AlertCheckResult {
	result := s.EvaluateAlertCondition(ctx, alert)

	checkResult := AlertCheckResult{
		Name:      alert.Name,
		Condition: alert.Condition,
		Device:    alert.Device,
		Value:     result.Value,
	}

	// Edge detection: only trigger on state change
	if result.Triggered && !state.Triggered {
		state.Triggered = true
		state.LastTriggered = time.Now()
		state.LastValue = result.Value
		checkResult.Action = AlertActionTriggered
	} else if !result.Triggered && state.Triggered {
		state.Triggered = false
		checkResult.Action = AlertActionCleared
	}

	return checkResult
}

// ActionResult holds the result of executing an alert action.
type ActionResult struct {
	Type       string // "notify", "webhook", "command", "unknown"
	StatusCode int    // For webhook
	Output     []byte // For command
	Error      error
}

// ExecuteAlertAction executes the action for a triggered alert.
func ExecuteAlertAction(ctx context.Context, alert config.Alert, value string) ActionResult {
	switch {
	case alert.Action == ActionTypeNotify || alert.Action == "":
		return ActionResult{Type: ActionTypeNotify}

	case strings.HasPrefix(alert.Action, ActionTypeWebhook+":"):
		url := strings.TrimPrefix(alert.Action, ActionTypeWebhook+":")
		result := ExecuteWebhook(ctx, url, alert, value)
		return ActionResult{
			Type:       ActionTypeWebhook,
			StatusCode: result.StatusCode,
			Error:      result.Error,
		}

	case strings.HasPrefix(alert.Action, ActionTypeCommand+":"):
		cmdStr := strings.TrimPrefix(alert.Action, ActionTypeCommand+":")
		result := ExecuteCommand(ctx, cmdStr)
		return ActionResult{
			Type:   ActionTypeCommand,
			Output: result.Output,
			Error:  result.Error,
		}

	default:
		return ActionResult{Type: ActionTypeUnknown, Error: fmt.Errorf("unknown action: %s", alert.Action)}
	}
}

// EvaluateAlertCondition checks if an alert's condition is met.
func (s *Service) EvaluateAlertCondition(ctx context.Context, alert config.Alert) AlertConditionResult {
	condition := strings.ToLower(alert.Condition)

	// Check offline/online conditions
	if condition == "offline" || condition == "online" {
		return s.evaluateConnectivity(ctx, alert.Device, condition)
	}

	// For threshold conditions, we need device status
	return s.evaluateThreshold(ctx, alert.Device, condition)
}

// evaluateConnectivity checks online/offline conditions.
func (s *Service) evaluateConnectivity(ctx context.Context, device, condition string) AlertConditionResult {
	conn, err := s.Connect(ctx, device)
	if err != nil {
		return AlertConditionResult{
			Triggered: condition == "offline",
			Value:     statusUnreachable,
		}
	}
	defer iostreams.CloseWithDebug("closing alert connectivity check", conn)

	// Device is reachable
	return AlertConditionResult{
		Triggered: condition == "online",
		Value:     "reachable",
	}
}

// evaluateThreshold checks threshold-based conditions (power, temperature, etc.).
func (s *Service) evaluateThreshold(ctx context.Context, device, condition string) AlertConditionResult {
	conn, err := s.Connect(ctx, device)
	if err != nil {
		return AlertConditionResult{Triggered: false, Value: statusUnreachable}
	}
	defer iostreams.CloseWithDebug("closing alert threshold check", conn)

	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return AlertConditionResult{Triggered: false, Value: "error"}
	}

	status, ok := result.(map[string]any)
	if !ok {
		return AlertConditionResult{Triggered: false, Value: "invalid"}
	}

	return parseThresholdCondition(status, condition)
}

// parseThresholdCondition parses and evaluates conditions like "power>100" or "temperature<30".
func parseThresholdCondition(status map[string]any, condition string) AlertConditionResult {
	if strings.Contains(condition, ">") {
		return checkThreshold(status, condition, ">", func(current, threshold float64) bool {
			return current > threshold
		})
	}

	if strings.Contains(condition, "<") {
		return checkThreshold(status, condition, "<", func(current, threshold float64) bool {
			return current < threshold
		})
	}

	return AlertConditionResult{Triggered: false, Value: "unknown condition"}
}

// checkThreshold evaluates a threshold condition.
func checkThreshold(status map[string]any, condition, separator string, compare func(float64, float64) bool) AlertConditionResult {
	parts := strings.SplitN(condition, separator, 2)
	if len(parts) != 2 {
		return AlertConditionResult{Triggered: false, Value: "invalid condition"}
	}

	metric := parts[0]
	threshold, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return AlertConditionResult{Triggered: false, Value: "invalid threshold"}
	}

	currentValue := getMetricValue(status, metric)
	return AlertConditionResult{
		Triggered: compare(currentValue, threshold),
		Value:     fmt.Sprintf("%.1f", currentValue),
	}
}

// getMetricValue extracts a metric value from device status.
func getMetricValue(status map[string]any, metric string) float64 {
	metric = strings.ToLower(metric)
	componentPrefixes := []string{"switch:0", "pm1:0", "em:0", "temperature:0", "cover:0"}

	for _, prefix := range componentPrefixes {
		comp, ok := status[prefix].(map[string]any)
		if !ok {
			continue
		}
		if v := extractMetric(comp, metric); v != 0 {
			return v
		}
	}
	return 0
}

// extractMetric extracts a specific metric from a component.
func extractMetric(comp map[string]any, metric string) float64 {
	switch metric {
	case "power", "apower":
		if v, ok := comp["apower"].(float64); ok {
			return v
		}
	case "temperature", "temp":
		return extractTemperature(comp)
	case "voltage":
		if v, ok := comp["voltage"].(float64); ok {
			return v
		}
	case "current":
		if v, ok := comp["current"].(float64); ok {
			return v
		}
	}
	return 0
}

// extractTemperature extracts temperature from a component.
func extractTemperature(comp map[string]any) float64 {
	if v, ok := comp["temperature"].(map[string]any); ok {
		if tC, ok := v["tC"].(float64); ok {
			return tC
		}
	}
	if v, ok := comp["tC"].(float64); ok {
		return v
	}
	return 0
}

// WebhookResult holds the result of executing a webhook.
type WebhookResult struct {
	StatusCode int
	Error      error
}

// ExecuteWebhook sends an HTTP POST to the specified URL with alert data.
func ExecuteWebhook(ctx context.Context, url string, alert config.Alert, value string) WebhookResult {
	payload := fmt.Sprintf(`{"alert":%q,"device":%q,"condition":%q,"value":%q,"timestamp":%q}`,
		alert.Name, alert.Device, alert.Condition, value, time.Now().Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return WebhookResult{Error: fmt.Errorf("create request: %w", err)}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return WebhookResult{Error: fmt.Errorf("send request: %w", err)}
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			iostreams.DebugErr("closing webhook response body", cerr)
		}
	}()

	return WebhookResult{StatusCode: resp.StatusCode}
}

// CommandResult holds the result of executing a shell command.
type CommandResult struct {
	Output []byte
	Error  error
}

// CheckAlerts checks all enabled alerts and executes actions for triggered ones.
// It returns action results for display by the caller.
func (s *Service) CheckAlerts(ctx context.Context, alerts map[string]config.Alert, states map[string]*AlertState) []AlertCheckResult {
	var results []AlertCheckResult

	for name, alert := range alerts {
		if !alert.Enabled || alert.IsSnoozed() {
			continue
		}

		state, ok := states[name]
		if !ok {
			state = &AlertState{}
			states[name] = state
		}

		result := s.CheckAlert(ctx, alert, state)
		if result.Action != AlertActionNone {
			results = append(results, result)
		}
	}

	return results
}

// ExecuteCommand runs a shell command.
func ExecuteCommand(ctx context.Context, command string) CommandResult {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return CommandResult{
			Output: stderr.Bytes(),
			Error:  err,
		}
	}

	return CommandResult{Output: stdout.Bytes()}
}
