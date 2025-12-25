package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/examples/plugins/shelly-tasmota/tasmota"
)

// Control action constants.
const (
	actionOn     = "on"
	actionOff    = "off"
	actionToggle = "toggle"
)

// ControlResult is the output format for the control hook.
type ControlResult struct {
	Success bool   `json:"success"`
	State   string `json:"state,omitempty"`
	Error   string `json:"error,omitempty"`
}

var controlFlags struct {
	address   string
	authUser  string
	authPass  string
	timeout   time.Duration
	action    string
	component string
	id        int
}

// NewControlCmd creates the control command.
func NewControlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "control",
		Short: "Control a device (on/off/toggle)",
		Long: `Execute control commands on a Tasmota device.

Supported actions:
  - on:     Turn the relay/switch on
  - off:    Turn the relay/switch off
  - toggle: Toggle the relay/switch state

Supported components:
  - switch: Controls relay outputs (default)
  - light:  Also controls relay outputs (Tasmota treats lights as switches)

Output is JSON in the ControlResult format expected by shelly-cli plugin hooks.`,
		Example: `  # Turn on the first relay
  shelly-tasmota control --address=192.168.1.50 --action=on

  # Turn off relay 2 (id=1, 0-indexed)
  shelly-tasmota control --address=192.168.1.50 --action=off --id=1

  # Toggle relay with authentication
  shelly-tasmota control --address=192.168.1.50 --action=toggle --auth-user=admin --auth-pass=secret`,
		RunE: runControl,
	}

	cmd.Flags().StringVar(&controlFlags.address, "address", "", "Device IP address (required)")
	cmd.Flags().StringVar(&controlFlags.authUser, "auth-user", "", "HTTP auth username")
	cmd.Flags().StringVar(&controlFlags.authPass, "auth-pass", "", "HTTP auth password")
	cmd.Flags().DurationVar(&controlFlags.timeout, "timeout", 5*time.Second, "Request timeout")
	cmd.Flags().StringVar(&controlFlags.action, "action", "", "Action to perform: on, off, toggle (required)")
	cmd.Flags().StringVar(&controlFlags.component, "component", "switch", "Component type: switch, light")
	cmd.Flags().IntVar(&controlFlags.id, "id", 0, "Component ID (0-indexed)")

	if err := cmd.MarkFlagRequired("address"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to mark flag required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("action"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to mark flag required: %v\n", err)
	}

	return cmd
}

func runControl(cmd *cobra.Command, _ []string) error {
	// Validate action
	action := strings.ToLower(controlFlags.action)
	var tasmotaAction string
	switch action {
	case actionOn:
		tasmotaAction = "ON"
	case actionOff:
		tasmotaAction = "OFF"
	case actionToggle:
		tasmotaAction = "TOGGLE"
	default:
		result := ControlResult{
			Success: false,
			Error:   fmt.Sprintf("invalid action: %q (must be on, off, or toggle)", controlFlags.action),
		}
		return outputJSON(result)
	}

	// Validate component - for Tasmota, both switch and light map to Power commands
	component := strings.ToLower(controlFlags.component)
	if component != "switch" && component != "light" {
		result := ControlResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported component: %q (must be switch or light)", controlFlags.component),
		}
		return outputJSON(result)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), controlFlags.timeout)
	defer cancel()

	client := tasmota.NewClient(controlFlags.address, controlFlags.authUser, controlFlags.authPass)

	// Execute the power command
	resp, err := client.Power(ctx, controlFlags.id, tasmotaAction)
	if err != nil {
		result := ControlResult{
			Success: false,
			Error:   fmt.Sprintf("failed to execute command: %v", err),
		}
		return outputJSON(result)
	}

	// Determine the resulting state from the response
	state := extractPowerState(resp, controlFlags.id)

	result := ControlResult{
		Success: true,
		State:   state,
	}
	return outputJSON(result)
}

// extractPowerState extracts the state for the given relay ID from the response.
func extractPowerState(resp *tasmota.PowerResponse, id int) string {
	var rawState string

	switch id {
	case 0:
		// Single relay returns POWER, multi-relay returns POWER1
		if resp.Power != "" {
			rawState = resp.Power
		} else {
			rawState = resp.Power1
		}
	case 1:
		rawState = resp.Power2
	case 2:
		rawState = resp.Power3
	case 3:
		rawState = resp.Power4
	}

	// Normalize to lowercase
	switch strings.ToUpper(rawState) {
	case "ON":
		return actionOn
	case "OFF":
		return actionOff
	default:
		return rawState
	}
}
