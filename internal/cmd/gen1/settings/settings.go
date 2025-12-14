// Package settings provides the gen1 settings command.
package settings

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/httputil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the gen1 settings command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "settings <device>",
		Aliases: []string{"config", "cfg"},
		Short:   "Show Gen1 device settings",
		Long: `Show the full settings/configuration of a Gen1 Shelly device.

Retrieves settings from the /settings HTTP endpoint, which includes:
- Device name and configuration
- Relay/roller settings
- Light/color settings
- Network configuration
- Cloud settings
- Action URLs

Note: This command is for Gen1 devices only. For Gen2+ devices,
use 'shelly config get' instead.`,
		Example: `  # Show settings
  shelly gen1 settings living-room

  # Output as JSON
  shelly gen1 settings living-room --json`,
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

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	settings, err := httputil.FetchEndpoint(ctx, ios, opts.Device, "/settings")
	if err != nil {
		return err
	}

	if opts.JSON {
		output, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	// Display settings
	ios.Println(theme.Bold().Render("Gen1 Device Settings:"))
	ios.Println()

	displayDevice(ios, settings)
	displayName(ios, settings)
	displayFirmware(ios, settings)
	displayCloud(ios, settings)
	displayCoIoT(ios, settings)
	displayMQTT(ios, settings)
	displayRelaySettings(ios, settings)

	ios.Info("For full settings, use --json flag")

	return nil
}

func displayDevice(ios *iostreams.IOStreams, settings map[string]any) {
	device, ok := settings["device"].(map[string]any)
	if !ok {
		return
	}

	ios.Println("  " + theme.Highlight().Render("Device:"))
	if dtype, ok := device["type"].(string); ok {
		ios.Printf("    Type: %s\n", dtype)
	}
	if mac, ok := device["mac"].(string); ok {
		ios.Printf("    MAC: %s\n", mac)
	}
	if hostname, ok := device["hostname"].(string); ok {
		ios.Printf("    Hostname: %s\n", hostname)
	}
	ios.Println()
}

func displayName(ios *iostreams.IOStreams, settings map[string]any) {
	name, ok := settings["name"].(string)
	if !ok || name == "" {
		return
	}

	ios.Printf("  Name: %s\n", name)
	ios.Println()
}

func displayFirmware(ios *iostreams.IOStreams, settings map[string]any) {
	fw, ok := settings["fw"].(string)
	if !ok {
		return
	}

	ios.Println("  " + theme.Highlight().Render("Firmware:"))
	ios.Printf("    Current: %s\n", fw)
	if buildInfo, ok := settings["build_info"].(map[string]any); ok {
		if buildID, ok := buildInfo["build_id"].(string); ok {
			ios.Printf("    Build: %s\n", buildID)
		}
	}
	ios.Println()
}

func displayCloud(ios *iostreams.IOStreams, settings map[string]any) {
	cloud, ok := settings["cloud"].(map[string]any)
	if !ok {
		return
	}

	ios.Println("  " + theme.Highlight().Render("Cloud:"))
	enabled, hasEnabled := cloud["enabled"].(bool)
	enabledStr := theme.Dim().Render("Disabled")
	if hasEnabled && enabled {
		enabledStr = theme.StatusOK().Render("Enabled")
	}
	ios.Printf("    Enabled: %s\n", enabledStr)

	connected, hasConnected := cloud["connected"].(bool)
	connStr := theme.Dim().Render("Disconnected")
	if hasConnected && connected {
		connStr = theme.StatusOK().Render("Connected")
	}
	ios.Printf("    Connected: %s\n", connStr)
	ios.Println()
}

func displayCoIoT(ios *iostreams.IOStreams, settings map[string]any) {
	coiot, ok := settings["coiot"].(map[string]any)
	if !ok {
		return
	}

	ios.Println("  " + theme.Highlight().Render("CoIoT:"))
	enabled, hasEnabled := coiot["enabled"].(bool)
	enabledStr := theme.Dim().Render("Disabled")
	if hasEnabled && enabled {
		enabledStr = theme.StatusOK().Render("Enabled")
	}
	ios.Printf("    Enabled: %s\n", enabledStr)

	if updatePeriod, ok := coiot["update_period"].(float64); ok {
		ios.Printf("    Update Period: %.0fs\n", updatePeriod)
	}
	ios.Println()
}

func displayMQTT(ios *iostreams.IOStreams, settings map[string]any) {
	mqtt, ok := settings["mqtt"].(map[string]any)
	if !ok {
		return
	}

	ios.Println("  " + theme.Highlight().Render("MQTT:"))
	enabled, hasEnabled := mqtt["enable"].(bool)
	enabledStr := theme.Dim().Render("Disabled")
	if hasEnabled && enabled {
		enabledStr = theme.StatusOK().Render("Enabled")
	}
	ios.Printf("    Enabled: %s\n", enabledStr)

	if server, ok := mqtt["server"].(string); ok && server != "" {
		ios.Printf("    Server: %s\n", server)
	}
	ios.Println()
}

func displayRelaySettings(ios *iostreams.IOStreams, settings map[string]any) {
	relays, ok := settings["relays"].([]any)
	if !ok || len(relays) == 0 {
		return
	}

	ios.Println("  " + theme.Highlight().Render("Relay Settings:"))
	for i, r := range relays {
		relay, ok := r.(map[string]any)
		if !ok {
			continue
		}
		name := fmt.Sprintf("Relay %d", i)
		if n, ok := relay["name"].(string); ok && n != "" {
			name = n
		}
		ios.Printf("    %s:\n", name)
		if defState, ok := relay["default_state"].(string); ok {
			ios.Printf("      Default State: %s\n", defState)
		}
		if applyPower, ok := relay["appliance_type"].(string); ok && applyPower != "" {
			ios.Printf("      Appliance Type: %s\n", applyPower)
		}
	}
	ios.Println()
}
