// Package status provides the gen1 status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/httputil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the gen1 status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "info"},
		Short:   "Show Gen1 device status",
		Long: `Show the full status of a Gen1 Shelly device.

Retrieves status from the /status HTTP endpoint, which includes:
- Relay states and power readings
- Roller positions
- Light/color settings
- Sensor values (temperature, humidity, etc.)
- WiFi connection info
- Device uptime and memory

Note: This command is for Gen1 devices only. For Gen2+ devices,
use 'shelly device info' or 'shelly status' instead.`,
		Example: `  # Show status
  shelly gen1 status living-room

  # Output as JSON
  shelly gen1 status living-room --json`,
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

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	status, err := httputil.FetchEndpoint(ctx, ios, opts.Device, "/status")
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

	// Display status
	ios.Println(theme.Bold().Render("Gen1 Device Status:"))
	ios.Println()

	displayWiFi(ios, status)
	displayRelays(ios, status)
	displayMeters(ios, status)
	displayRollers(ios, status)
	displayTemperature(ios, status)
	displaySystem(ios, status)

	return nil
}

func displayWiFi(ios *iostreams.IOStreams, status map[string]any) {
	wifi, ok := status["wifi_sta"].(map[string]any)
	if !ok {
		return
	}

	ios.Println("  " + theme.Highlight().Render("WiFi:"))
	connected, hasConnected := wifi["connected"].(bool)
	if !hasConnected || !connected {
		ios.Printf("    Connected: %s\n", theme.StatusError().Render("No"))
		ios.Println()
		return
	}

	ios.Printf("    Connected: %s\n", theme.StatusOK().Render("Yes"))
	if ssid, ok := wifi["ssid"].(string); ok {
		ios.Printf("    SSID: %s\n", ssid)
	}
	if ip, ok := wifi["ip"].(string); ok {
		ios.Printf("    IP: %s\n", ip)
	}
	if rssi, ok := wifi["rssi"].(float64); ok {
		ios.Printf("    RSSI: %.0f dBm\n", rssi)
	}
	ios.Println()
}

func displayRelays(ios *iostreams.IOStreams, status map[string]any) {
	relays, ok := status["relays"].([]any)
	if !ok || len(relays) == 0 {
		return
	}

	ios.Println("  " + theme.Highlight().Render("Relays:"))
	for i, r := range relays {
		relay, ok := r.(map[string]any)
		if !ok {
			continue
		}
		isOn, hasState := relay["ison"].(bool)
		state := theme.Dim().Render("OFF")
		if hasState && isOn {
			state = theme.StatusOK().Render("ON")
		}
		ios.Printf("    Relay %d: %s\n", i, state)
	}
	ios.Println()
}

func displayMeters(ios *iostreams.IOStreams, status map[string]any) {
	meters, ok := status["meters"].([]any)
	if !ok || len(meters) == 0 {
		return
	}

	ios.Println("  " + theme.Highlight().Render("Power Meters:"))
	for i, m := range meters {
		meter, ok := m.(map[string]any)
		if !ok {
			continue
		}
		if power, ok := meter["power"].(float64); ok {
			ios.Printf("    Meter %d: %.1f W\n", i, power)
		}
	}
	ios.Println()
}

func displayRollers(ios *iostreams.IOStreams, status map[string]any) {
	rollers, ok := status["rollers"].([]any)
	if !ok || len(rollers) == 0 {
		return
	}

	ios.Println("  " + theme.Highlight().Render("Rollers:"))
	for i, r := range rollers {
		roller, ok := r.(map[string]any)
		if !ok {
			continue
		}
		state := "unknown"
		if s, ok := roller["state"].(string); ok {
			state = s
		}
		ios.Printf("    Roller %d: %s\n", i, state)
		if pos, ok := roller["current_pos"].(float64); ok {
			ios.Printf("      Position: %.0f%%\n", pos)
		}
	}
	ios.Println()
}

func displayTemperature(ios *iostreams.IOStreams, status map[string]any) {
	if temp, ok := status["temperature"].(float64); ok {
		ios.Println("  " + theme.Highlight().Render("Device:"))
		ios.Printf("    Temperature: %.1f°C\n", temp)
	}

	tempStatus, ok := status["tmp"].(map[string]any)
	if !ok {
		return
	}

	ios.Println("  " + theme.Highlight().Render("Sensor:"))
	if tC, ok := tempStatus["tC"].(float64); ok {
		ios.Printf("    Temperature: %.1f°C\n", tC)
	}
	if isValid, ok := tempStatus["is_valid"].(bool); ok {
		ios.Printf("    Valid: %v\n", isValid)
	}
}

func displaySystem(ios *iostreams.IOStreams, status map[string]any) {
	uptime, ok := status["uptime"].(float64)
	if !ok {
		return
	}

	days := int(uptime) / 86400
	hours := (int(uptime) % 86400) / 3600
	mins := (int(uptime) % 3600) / 60
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("System:"))
	ios.Printf("    Uptime: %dd %dh %dm\n", days, hours, mins)

	if ram, ok := status["ram_free"].(float64); ok {
		ios.Printf("    Free RAM: %.0f bytes\n", ram)
	}
}
