// Package status provides the zigbee status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the zigbee status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "info"},
		Short:   "Show Zigbee network status",
		Long: `Show Zigbee network status for a Shelly device.

Displays the current Zigbee state including:
- Whether Zigbee is enabled
- Network state (not_configured, ready, steering, joined)
- EUI64 address (device's Zigbee identifier)
- PAN ID and channel when joined to a network
- Coordinator information`,
		Example: `  # Show Zigbee status
  shelly zigbee status living-room

  # Output as JSON
  shelly zigbee status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// ZigbeeStatus represents the full Zigbee status.
type ZigbeeStatus struct {
	Enabled          bool   `json:"enabled"`
	NetworkState     string `json:"network_state"`
	EUI64            string `json:"eui64,omitempty"`
	PANID            uint16 `json:"pan_id,omitempty"`
	Channel          int    `json:"channel,omitempty"`
	CoordinatorEUI64 string `json:"coordinator_eui64,omitempty"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	status, err := fetchZigbeeStatus(ctx, svc, opts.Device, ios)
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputStatusJSON(ios, status)
	}

	displayZigbeeStatus(ios, status)
	return nil
}

func fetchZigbeeStatus(ctx context.Context, svc *shelly.Service, device string, ios *iostreams.IOStreams) (ZigbeeStatus, error) {
	var status ZigbeeStatus

	// Get config
	cfg, err := svc.ZigbeeGetConfig(ctx, device)
	if err != nil {
		return status, fmt.Errorf("zigbee not available on this device: %w", err)
	}
	if enable, ok := cfg["enable"].(bool); ok {
		status.Enabled = enable
	}

	// Get status
	st, err := svc.ZigbeeGetStatus(ctx, device)
	if err != nil {
		ios.Debug("Zigbee.GetStatus failed: %v", err)
		return status, nil // Config succeeded, return partial info
	}

	if networkState, ok := st["network_state"].(string); ok {
		status.NetworkState = networkState
	}
	if eui64, ok := st["eui64"].(string); ok {
		status.EUI64 = eui64
	}
	if panID, ok := st["pan_id"].(float64); ok {
		status.PANID = uint16(panID)
	}
	if channel, ok := st["channel"].(float64); ok {
		status.Channel = int(channel)
	}
	if coordinatorEUI64, ok := st["coordinator_eui64"].(string); ok {
		status.CoordinatorEUI64 = coordinatorEUI64
	}

	return status, nil
}

func outputStatusJSON(ios *iostreams.IOStreams, status ZigbeeStatus) error {
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(output))
	return nil
}

func displayZigbeeStatus(ios *iostreams.IOStreams, status ZigbeeStatus) {
	ios.Println(theme.Bold().Render("Zigbee Status:"))
	ios.Println()

	displayEnabledState(ios, status.Enabled)
	displayNetworkState(ios, status.NetworkState)

	if status.EUI64 != "" {
		ios.Printf("  EUI64: %s\n", status.EUI64)
	}

	if status.NetworkState == "joined" {
		displayNetworkInfo(ios, status)
	}
}

func displayEnabledState(ios *iostreams.IOStreams, enabled bool) {
	enabledStr := theme.Dim().Render("Disabled")
	if enabled {
		enabledStr = theme.StatusOK().Render("Enabled")
	}
	ios.Printf("  Enabled: %s\n", enabledStr)
}

func displayNetworkState(ios *iostreams.IOStreams, state string) {
	if state == "" {
		return
	}

	stateStyle := theme.Dim()
	switch state {
	case "joined":
		stateStyle = theme.StatusOK()
	case "steering":
		stateStyle = theme.StatusWarn()
	}
	ios.Printf("  Network State: %s\n", stateStyle.Render(state))
}

func displayNetworkInfo(ios *iostreams.IOStreams, status ZigbeeStatus) {
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Network Info:"))
	ios.Printf("    PAN ID: 0x%04X\n", status.PANID)
	ios.Printf("    Channel: %d\n", status.Channel)
	if status.CoordinatorEUI64 != "" {
		ios.Printf("    Coordinator: %s\n", status.CoordinatorEUI64)
	}
}
