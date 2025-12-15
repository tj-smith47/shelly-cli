// Package watch provides the alert watch subcommand for monitoring alerts.
package watch

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const statusUnreachable = "unreachable"

// Options holds the command options.
type Options struct {
	Interval time.Duration
	Once     bool
}

// alertState tracks the state of an alert.
type alertState struct {
	lastTriggered time.Time
	lastValue     string
	triggered     bool
}

// NewCommand creates the alert watch command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "watch",
		Aliases: []string{"monitor", "daemon", "run"},
		Short:   "Monitor alerts in real-time",
		Long: `Monitor configured alerts and trigger actions when conditions are met.

This command runs continuously, polling device status at the specified interval
and executing alert actions when conditions are triggered.

Conditions supported:
  - offline: Device becomes unreachable
  - online: Device becomes reachable
  - power>N: Power consumption exceeds N watts
  - power<N: Power consumption below N watts
  - temperature>N: Temperature exceeds N degrees
  - temperature<N: Temperature below N degrees

Actions supported:
  - notify: Print to console (default)
  - webhook:URL: Send HTTP POST to URL with alert JSON
  - command:CMD: Execute shell command`,
		Example: `  # Monitor alerts every 30 seconds
  shelly alert watch

  # Monitor with custom interval
  shelly alert watch --interval 1m

  # Run once and exit (for cron)
  shelly alert watch --once`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Interval, "interval", "i", 30*time.Second, "Check interval")
	cmd.Flags().BoolVar(&opts.Once, "once", false, "Run once and exit (for cron/scheduled tasks)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if len(cfg.Alerts) == 0 {
		ios.Warning("No alerts configured")
		ios.Info("Create alerts with: shelly alert create <name> --device <device> --condition <condition>")
		return nil
	}

	enabledCount := 0
	for _, alert := range cfg.Alerts {
		if alert.Enabled && !isSnoozed(alert) {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		ios.Warning("No enabled alerts (all disabled or snoozed)")
		return nil
	}

	ios.Success("Alert monitor started")
	ios.Printf("  Monitoring %d alert(s) every %s\n", enabledCount, opts.Interval)
	ios.Printf("  Press Ctrl+C to stop\n")
	ios.Println("")

	// Track alert states
	states := make(map[string]*alertState)
	for name := range cfg.Alerts {
		states[name] = &alertState{}
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	// Run immediately on start
	checkAlerts(ctx, ios, svc, cfg, states)

	if opts.Once {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			ios.Println("")
			ios.Info("Alert monitor stopped")
			return nil
		case <-ticker.C:
			// Reload config to pick up changes
			cfg, err = f.Config()
			if err != nil {
				ios.DebugErr("reload config", err)
				continue
			}
			checkAlerts(ctx, ios, svc, cfg, states)
		}
	}
}

func checkAlerts(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, cfg *config.Config, states map[string]*alertState) {
	for name, alert := range cfg.Alerts {
		if !alert.Enabled || isSnoozed(alert) {
			continue
		}

		state, ok := states[name]
		if !ok {
			state = &alertState{}
			states[name] = state
		}

		triggered, value := evaluateCondition(ctx, svc, alert)

		// Only trigger on state change (edge detection)
		if triggered && !state.triggered {
			state.triggered = true
			state.lastTriggered = time.Now()
			state.lastValue = value
			executeAction(ctx, ios, alert, value)
		} else if !triggered && state.triggered {
			// Reset state when condition clears
			state.triggered = false
			ios.Info("[%s] Condition cleared: %s on %s", time.Now().Format("15:04:05"), alert.Condition, alert.Device)
		}
	}
}

func evaluateCondition(ctx context.Context, svc *shelly.Service, alert config.Alert) (triggered bool, value string) {
	condition := strings.ToLower(alert.Condition)

	// Check offline/online conditions
	if condition == "offline" || condition == "online" {
		return evaluateConnectivity(ctx, svc, alert.Device, condition)
	}

	// For threshold conditions, we need device status
	return evaluateThreshold(ctx, svc, alert.Device, condition)
}

func evaluateConnectivity(ctx context.Context, svc *shelly.Service, device, condition string) (triggered bool, value string) {
	conn, err := svc.Connect(ctx, device)
	if err != nil {
		return condition == "offline", statusUnreachable
	}
	defer iostreams.CloseWithDebug("closing alert watch connection", conn)

	// Device is reachable
	return condition == "online", "reachable"
}

func evaluateThreshold(ctx context.Context, svc *shelly.Service, device, condition string) (triggered bool, value string) {
	conn, err := svc.Connect(ctx, device)
	if err != nil {
		return false, statusUnreachable
	}
	defer iostreams.CloseWithDebug("closing alert watch connection", conn)

	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return false, "error"
	}

	status, ok := result.(map[string]any)
	if !ok {
		return false, "invalid"
	}

	return parseThresholdCondition(status, condition)
}

func parseThresholdCondition(status map[string]any, condition string) (triggered bool, value string) {
	// Parse condition like "power>100" or "temperature<30"
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

	return false, "unknown condition"
}

func checkThreshold(status map[string]any, condition, separator string, compare func(float64, float64) bool) (triggered bool, value string) {
	parts := strings.SplitN(condition, separator, 2)
	if len(parts) != 2 {
		return false, "invalid condition"
	}

	metric := parts[0]
	threshold, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return false, "invalid threshold"
	}

	currentValue := getMetricValue(status, metric)
	return compare(currentValue, threshold), fmt.Sprintf("%.1f", currentValue)
}

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

func executeAction(ctx context.Context, ios *iostreams.IOStreams, alert config.Alert, value string) {
	timestamp := time.Now().Format("15:04:05")

	switch {
	case alert.Action == "notify" || alert.Action == "":
		ios.Warning("[%s] ALERT: %s - %s on %s (value: %s)",
			timestamp, alert.Name, alert.Condition, alert.Device, value)

	case strings.HasPrefix(alert.Action, "webhook:"):
		executeWebhook(ctx, ios, alert, value, timestamp)

	case strings.HasPrefix(alert.Action, "command:"):
		executeCommand(ctx, ios, alert, timestamp)

	default:
		ios.Warning("[%s] Unknown action type: %s", timestamp, alert.Action)
	}
}

func executeWebhook(ctx context.Context, ios *iostreams.IOStreams, alert config.Alert, value, timestamp string) {
	url := strings.TrimPrefix(alert.Action, "webhook:")
	ios.Info("[%s] Triggering webhook for %s...", timestamp, alert.Name)

	payload := fmt.Sprintf(`{"alert":%q,"device":%q,"condition":%q,"value":%q,"timestamp":%q}`,
		alert.Name, alert.Device, alert.Condition, value, time.Now().Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		ios.Error("[%s] Webhook request creation failed: %v", timestamp, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		ios.Error("[%s] Webhook failed: %v", timestamp, err)
		return
	}
	if err := resp.Body.Close(); err != nil {
		ios.DebugErr("close response body", err)
	}
	ios.Success("[%s] Webhook sent (status: %d)", timestamp, resp.StatusCode)
}

func executeCommand(ctx context.Context, ios *iostreams.IOStreams, alert config.Alert, timestamp string) {
	cmdStr := strings.TrimPrefix(alert.Action, "command:")
	ios.Info("[%s] Executing command for %s...", timestamp, alert.Name)

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr) //nolint:gosec // User-provided command is intentional
	output, err := cmd.CombinedOutput()
	if err != nil {
		ios.Error("[%s] Command failed: %v", timestamp, err)
		return
	}
	ios.Success("[%s] Command executed", timestamp)
	if len(output) > 0 {
		ios.Printf("  Output: %s\n", strings.TrimSpace(string(output)))
	}
}

func isSnoozed(alert config.Alert) bool {
	if alert.SnoozedUntil == "" {
		return false
	}

	snoozedUntil, err := time.Parse(time.RFC3339, alert.SnoozedUntil)
	if err != nil {
		return false
	}

	return time.Now().Before(snoozedUntil)
}
