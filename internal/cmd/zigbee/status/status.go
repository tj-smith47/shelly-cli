// Package status provides the zigbee status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
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

	var status ZigbeeStatus

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		// Get Zigbee config
		cfgResult, err := conn.Call(ctx, "Zigbee.GetConfig", nil)
		if err != nil {
			ios.Debug("Zigbee.GetConfig failed: %v", err)
			return fmt.Errorf("zigbee not available on this device: %w", err)
		}

		var cfg struct {
			Enable bool `json:"enable"`
		}
		cfgBytes, err := json.Marshal(cfgResult)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
		status.Enabled = cfg.Enable

		// Get Zigbee status
		statusResult, err := conn.Call(ctx, "Zigbee.GetStatus", nil)
		if err != nil {
			ios.Debug("Zigbee.GetStatus failed: %v", err)
			return nil // Config succeeded, status failed - still show partial info
		}

		var st struct {
			NetworkState     string `json:"network_state"`
			EUI64            string `json:"eui64"`
			PANID            uint16 `json:"pan_id"`
			Channel          int    `json:"channel"`
			CoordinatorEUI64 string `json:"coordinator_eui64"`
		}
		statusBytes, statusMarshalErr := json.Marshal(statusResult)
		if statusMarshalErr != nil {
			ios.Debug("failed to marshal status: %v", statusMarshalErr)
			return nil
		}
		if err := json.Unmarshal(statusBytes, &st); err != nil {
			ios.Debug("failed to parse status: %v", err)
			return nil
		}

		status.NetworkState = st.NetworkState
		status.EUI64 = st.EUI64
		status.PANID = st.PANID
		status.Channel = st.Channel
		status.CoordinatorEUI64 = st.CoordinatorEUI64

		return nil
	})
	if err != nil {
		return err
	}

	if opts.JSON {
		output, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	ios.Println(theme.Bold().Render("Zigbee Status:"))
	ios.Println()

	enabledStr := theme.Dim().Render("Disabled")
	if status.Enabled {
		enabledStr = theme.StatusOK().Render("Enabled")
	}
	ios.Printf("  Enabled: %s\n", enabledStr)

	if status.NetworkState != "" {
		stateStyle := theme.Dim()
		switch status.NetworkState {
		case "joined":
			stateStyle = theme.StatusOK()
		case "steering":
			stateStyle = theme.StatusWarn()
		}
		ios.Printf("  Network State: %s\n", stateStyle.Render(status.NetworkState))
	}

	if status.EUI64 != "" {
		ios.Printf("  EUI64: %s\n", status.EUI64)
	}

	if status.NetworkState == "joined" {
		ios.Println()
		ios.Println("  " + theme.Highlight().Render("Network Info:"))
		ios.Printf("    PAN ID: 0x%04X\n", status.PANID)
		ios.Printf("    Channel: %d\n", status.Channel)
		if status.CoordinatorEUI64 != "" {
			ios.Printf("    Coordinator: %s\n", status.CoordinatorEUI64)
		}
	}

	return nil
}
