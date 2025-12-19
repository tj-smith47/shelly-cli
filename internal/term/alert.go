// Package term provides terminal display functions.
package term

import (
	"context"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// DisplayAlertTriggered displays an alert that was triggered.
func DisplayAlertTriggered(ios *iostreams.IOStreams, result shelly.AlertCheckResult, actionResult shelly.ActionResult) {
	timestamp := time.Now().Format("15:04:05")

	switch actionResult.Type {
	case shelly.ActionTypeNotify:
		ios.Warning("[%s] ALERT: %s - %s on %s (value: %s)",
			timestamp, result.Name, result.Condition, result.Device, result.Value)

	case shelly.ActionTypeWebhook:
		if actionResult.Error != nil {
			ios.Error("[%s] Webhook failed: %v", timestamp, actionResult.Error)
		} else {
			ios.Success("[%s] Webhook sent (status: %d)", timestamp, actionResult.StatusCode)
		}

	case shelly.ActionTypeCommand:
		if actionResult.Error != nil {
			ios.Error("[%s] Command failed: %v", timestamp, actionResult.Error)
		} else {
			ios.Success("[%s] Command executed", timestamp)
			if len(actionResult.Output) > 0 {
				ios.Printf("  Output: %s\n", strings.TrimSpace(string(actionResult.Output)))
			}
		}

	default:
		if actionResult.Error != nil {
			ios.Warning("[%s] %v", timestamp, actionResult.Error)
		}
	}
}

// DisplayAlertCleared displays when an alert condition is cleared.
func DisplayAlertCleared(ios *iostreams.IOStreams, result shelly.AlertCheckResult) {
	timestamp := time.Now().Format("15:04:05")
	ios.Info("[%s] Condition cleared: %s on %s", timestamp, result.Condition, result.Device)
}

// DisplayAlertActionStarting displays when an action is about to execute.
func DisplayAlertActionStarting(ios *iostreams.IOStreams, actionType, alertName string) {
	timestamp := time.Now().Format("15:04:05")
	switch actionType {
	case shelly.ActionTypeWebhook:
		ios.Info("[%s] Triggering webhook for %s...", timestamp, alertName)
	case shelly.ActionTypeCommand:
		ios.Info("[%s] Executing command for %s...", timestamp, alertName)
	}
}

// DisplayAlertResult displays the result of checking an alert.
func DisplayAlertResult(ctx context.Context, ios *iostreams.IOStreams, alert config.Alert, result shelly.AlertCheckResult) {
	switch result.Action {
	case shelly.AlertActionTriggered:
		DisplayAlertActionStarting(ios, alert.Action, alert.Name)
		actionResult := shelly.ExecuteAlertAction(ctx, alert, result.Value)
		DisplayAlertTriggered(ios, result, actionResult)

	case shelly.AlertActionCleared:
		DisplayAlertCleared(ios, result)

	case shelly.AlertActionNone:
		iostreams.Debug("alert %s: no state change (value: %s)", result.Name, result.Value)
	}
}
