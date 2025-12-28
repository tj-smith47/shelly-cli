package term

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const testAlertDevice = "device1"

func TestDisplayAlertTriggered(t *testing.T) {
	t.Parallel()

	t.Run("notify action", func(t *testing.T) {
		t.Parallel()

		ios, _, errOut := testIOStreams()
		result := shelly.AlertCheckResult{
			Name:      "High Temp",
			Device:    testAlertDevice,
			Condition: "temperature > 30",
			Value:     "35",
		}
		actionResult := shelly.ActionResult{
			Type: shelly.ActionTypeNotify,
		}

		DisplayAlertTriggered(ios, result, actionResult)

		// Warning goes to errOut
		errOutput := errOut.String()
		if !strings.Contains(errOutput, "ALERT") {
			t.Errorf("output should contain 'ALERT', got %q", errOutput)
		}
		if !strings.Contains(errOutput, "High Temp") {
			t.Errorf("output should contain alert name, got %q", errOutput)
		}
	})

	t.Run("webhook success", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		result := shelly.AlertCheckResult{
			Name:   "Test Alert",
			Device: testAlertDevice,
		}
		actionResult := shelly.ActionResult{
			Type:       shelly.ActionTypeWebhook,
			StatusCode: 200,
		}

		DisplayAlertTriggered(ios, result, actionResult)

		output := out.String()
		if !strings.Contains(output, "Webhook sent") {
			t.Errorf("output should contain 'Webhook sent', got %q", output)
		}
		if !strings.Contains(output, "200") {
			t.Errorf("output should contain status code, got %q", output)
		}
	})

	t.Run("webhook failure", func(t *testing.T) {
		t.Parallel()

		ios, _, errOut := testIOStreams()
		result := shelly.AlertCheckResult{
			Name:   "Test Alert",
			Device: testAlertDevice,
		}
		actionResult := shelly.ActionResult{
			Type:  shelly.ActionTypeWebhook,
			Error: errors.New("connection refused"),
		}

		DisplayAlertTriggered(ios, result, actionResult)

		errOutput := errOut.String()
		if !strings.Contains(errOutput, "Webhook failed") {
			t.Errorf("output should contain 'Webhook failed', got %q", errOutput)
		}
	})

	t.Run("command success", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		result := shelly.AlertCheckResult{
			Name:   "Test Alert",
			Device: testAlertDevice,
		}
		actionResult := shelly.ActionResult{
			Type:   shelly.ActionTypeCommand,
			Output: []byte("command output"),
		}

		DisplayAlertTriggered(ios, result, actionResult)

		output := out.String()
		if !strings.Contains(output, "Command executed") {
			t.Errorf("output should contain 'Command executed', got %q", output)
		}
		if !strings.Contains(output, "command output") {
			t.Errorf("output should contain command output, got %q", output)
		}
	})

	t.Run("command failure", func(t *testing.T) {
		t.Parallel()

		ios, _, errOut := testIOStreams()
		result := shelly.AlertCheckResult{
			Name:   "Test Alert",
			Device: testAlertDevice,
		}
		actionResult := shelly.ActionResult{
			Type:  shelly.ActionTypeCommand,
			Error: errors.New("command not found"),
		}

		DisplayAlertTriggered(ios, result, actionResult)

		errOutput := errOut.String()
		if !strings.Contains(errOutput, "Command failed") {
			t.Errorf("output should contain 'Command failed', got %q", errOutput)
		}
	})

	t.Run("unknown action with error", func(t *testing.T) {
		t.Parallel()

		ios, _, errOut := testIOStreams()
		result := shelly.AlertCheckResult{
			Name:   "Test Alert",
			Device: testAlertDevice,
		}
		actionResult := shelly.ActionResult{
			Type:  "unknown",
			Error: errors.New("some error"),
		}

		DisplayAlertTriggered(ios, result, actionResult)

		errOutput := errOut.String()
		if !strings.Contains(errOutput, "some error") {
			t.Errorf("output should contain error message, got %q", errOutput)
		}
	})
}

func TestDisplayAlertCleared(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	result := shelly.AlertCheckResult{
		Condition: "temperature > 30",
		Device:    testAlertDevice,
	}

	DisplayAlertCleared(ios, result)

	// Info may go to either stream
	allOutput := out.String() + errOut.String()
	if !strings.Contains(allOutput, "Condition cleared") {
		t.Errorf("output should contain 'Condition cleared', got %q", allOutput)
	}
	if !strings.Contains(allOutput, testAlertDevice) {
		t.Errorf("output should contain device name, got %q", allOutput)
	}
}

func TestDisplayAlertActionStarting(t *testing.T) {
	t.Parallel()

	t.Run("webhook action", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		DisplayAlertActionStarting(ios, shelly.ActionTypeWebhook, "High Temp Alert")

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "webhook") {
			t.Errorf("output should contain 'webhook', got %q", allOutput)
		}
		if !strings.Contains(allOutput, "High Temp Alert") {
			t.Errorf("output should contain alert name, got %q", allOutput)
		}
	})

	t.Run("command action", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		DisplayAlertActionStarting(ios, shelly.ActionTypeCommand, "Power Alert")

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "command") {
			t.Errorf("output should contain 'command', got %q", allOutput)
		}
		if !strings.Contains(allOutput, "Power Alert") {
			t.Errorf("output should contain alert name, got %q", allOutput)
		}
	})

	t.Run("unknown action", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		DisplayAlertActionStarting(ios, "unknown", "Test Alert")

		allOutput := out.String() + errOut.String()
		// Unknown action should not produce output
		if allOutput != "" {
			t.Errorf("unknown action should not produce output, got %q", allOutput)
		}
	})
}

func TestAlertCheckResult_Fields(t *testing.T) {
	t.Parallel()

	result := shelly.AlertCheckResult{
		Name:      "High Temp",
		Device:    testAlertDevice,
		Condition: "temperature > 30",
		Value:     "35.5",
		Action:    shelly.AlertActionTriggered,
	}

	if result.Name != "High Temp" {
		t.Errorf("got Name=%q, want High Temp", result.Name)
	}
	if result.Device != testAlertDevice {
		t.Errorf("got Device=%q, want %q", result.Device, testAlertDevice)
	}
	if result.Condition != "temperature > 30" {
		t.Errorf("got Condition=%q, want temperature > 30", result.Condition)
	}
	if result.Value != "35.5" {
		t.Errorf("got Value=%q, want 35.5", result.Value)
	}
	if result.Action != shelly.AlertActionTriggered {
		t.Errorf("got Action=%q, want %q", result.Action, shelly.AlertActionTriggered)
	}
}

func TestActionResult_Fields(t *testing.T) {
	t.Parallel()

	t.Run("success result", func(t *testing.T) {
		t.Parallel()

		result := shelly.ActionResult{
			Type:       shelly.ActionTypeWebhook,
			StatusCode: 200,
			Output:     []byte("OK"),
		}

		if result.Type != shelly.ActionTypeWebhook {
			t.Errorf("got Type=%q, want %q", result.Type, shelly.ActionTypeWebhook)
		}
		if result.StatusCode != http.StatusOK {
			t.Errorf("got StatusCode=%d, want %d", result.StatusCode, http.StatusOK)
		}
		if string(result.Output) != "OK" {
			t.Errorf("got Output=%q, want OK", string(result.Output))
		}
	})

	t.Run("error result", func(t *testing.T) {
		t.Parallel()

		testErr := errors.New("test error")
		result := shelly.ActionResult{
			Type:  shelly.ActionTypeCommand,
			Error: testErr,
		}

		if !errors.Is(result.Error, testErr) {
			t.Errorf("got Error=%v, want %v", result.Error, testErr)
		}
	})
}
