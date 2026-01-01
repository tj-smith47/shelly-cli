package term

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayWiFiStatus prints WiFi status information.
func DisplayWiFiStatus(ios *iostreams.IOStreams, status *shelly.WiFiStatus) {
	ios.Title("WiFi Status")
	ios.Println()

	ios.Printf("  Status:      %s\n", status.Status)
	ios.Printf("  SSID:        %s\n", valueOrEmpty(status.SSID))
	ios.Printf("  IP Address:  %s\n", valueOrEmpty(status.StaIP))
	ios.Printf("  Signal:      %d dBm\n", status.RSSI)
	if status.APCount > 0 {
		ios.Printf("  AP Clients:  %d\n", status.APCount)
	}
}

// DisplayWiFiAPClients prints a table of connected WiFi AP clients.
func DisplayWiFiAPClients(ios *iostreams.IOStreams, clients []shelly.WiFiAPClient) {
	ios.Title("Connected Clients")
	ios.Println()

	builder := table.NewBuilder("MAC Address", "IP Address")
	for _, c := range clients {
		ip := c.IP
		if ip == "" {
			ip = "<no IP>"
		}
		builder.AddRow(c.MAC, ip)
	}
	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}

	ios.Printf("\n%d client(s) connected\n", len(clients))
}

// DisplayWiFiScanResults prints a table of WiFi scan results.
func DisplayWiFiScanResults(ios *iostreams.IOStreams, results []shelly.WiFiScanResult) {
	ios.Title("Available WiFi Networks")
	ios.Println()

	builder := table.NewBuilder("SSID", "Signal", "Channel", "Security")
	for _, r := range results {
		ssid := r.SSID
		if ssid == "" {
			ssid = "<hidden>"
		}
		signal := formatWiFiSignal(r.RSSI)
		builder.AddRow(ssid, signal, fmt.Sprintf("%d", r.Channel), r.Auth)
	}
	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}

	ios.Printf("\n%d network(s) found\n", len(results))
}

func formatWiFiSignal(rssi int) string {
	bars := "▁▃▅▇"
	var strength int
	switch {
	case rssi >= -50:
		strength = 4
	case rssi >= -60:
		strength = 3
	case rssi >= -70:
		strength = 2
	default:
		strength = 1
	}
	return fmt.Sprintf("%s %d dBm", bars[:strength], rssi)
}

// DisplayEthernetStatus prints Ethernet status information.
func DisplayEthernetStatus(ios *iostreams.IOStreams, status *shelly.EthernetStatus) {
	ios.Title("Ethernet Status")
	ios.Println()

	if status.IP != "" {
		ios.Printf("  Status:     Connected\n")
		ios.Printf("  IP Address: %s\n", status.IP)
	} else {
		ios.Printf("  Status:     Not connected\n")
	}
}

// DisplayMQTTStatus prints MQTT status information.
func DisplayMQTTStatus(ios *iostreams.IOStreams, status *shelly.MQTTStatus) {
	ios.Title("MQTT Status")
	ios.Println()

	ios.Printf("  Status: %s\n", output.RenderBoolState(status.Connected, "Connected", "Disconnected"))
}

// DisplayCloudConnectionStatus prints cloud connection status.
func DisplayCloudConnectionStatus(ios *iostreams.IOStreams, status *shelly.CloudStatus) {
	ios.Title("Cloud Status")
	ios.Println()

	ios.Printf("  Status: %s\n", output.RenderBoolState(status.Connected, "Connected", "Disconnected"))
}

// DisplayCloudDevice prints cloud device details.
func DisplayCloudDevice(ios *iostreams.IOStreams, device *network.CloudDevice, showStatus bool) {
	ios.Title("Cloud Device")
	ios.Println()

	ios.Printf("  ID:     %s\n", device.ID)

	if device.Model != "" {
		ios.Printf("  Model:  %s\n", device.Model)
	}

	if device.Generation > 0 {
		ios.Printf("  Gen:    %d\n", device.Generation)
	}

	if device.MAC != "" {
		ios.Printf("  MAC:    %s\n", device.MAC)
	}

	if device.FirmwareVersion != "" {
		ios.Printf("  FW:     %s\n", device.FirmwareVersion)
	}

	ios.Printf("  Status: %s\n", output.RenderOnline(device.Online, output.CaseTitle))

	// Show status JSON if requested and available
	if showStatus && len(device.Status) > 0 {
		ios.Println()
		ios.Title("Device Status")
		ios.Println()
		printCloudJSON(ios, device.Status)
	}

	// Show settings if available
	if showStatus && len(device.Settings) > 0 {
		ios.Println()
		ios.Title("Device Settings")
		ios.Println()
		printCloudJSON(ios, device.Settings)
	}
}

func printCloudJSON(ios *iostreams.IOStreams, data json.RawMessage) {
	var prettyJSON map[string]any
	if err := json.Unmarshal(data, &prettyJSON); err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	formatted, err := json.MarshalIndent(prettyJSON, "  ", "  ")
	if err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	ios.Printf("  %s\n", string(formatted))
}

// DisplayCloudDevices prints a table of cloud devices.
func DisplayCloudDevices(ios *iostreams.IOStreams, devices []network.CloudDevice) {
	if len(devices) == 0 {
		ios.Info("No devices found in your Shelly Cloud account")
		return
	}

	// Sort by ID for consistent display
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].ID < devices[j].ID
	})

	builder := table.NewBuilder("ID", "Model", "Gen", "Online")

	for _, d := range devices {
		devModel := d.Model
		if devModel == "" {
			devModel = output.FormatPlaceholder("unknown")
		}

		gen := output.FormatPlaceholder("-")
		if d.Generation > 0 {
			gen = fmt.Sprintf("%d", d.Generation)
		}

		builder.AddRow(d.ID, devModel, gen, output.RenderYesNo(d.Online, output.CaseLower, theme.FalseError))
	}

	ios.Printf("Found %d device(s):\n\n", len(devices))
	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print cloud devices table", err)
	}
}

// TokenStatusInfo holds display info for cloud token status.
type TokenStatusInfo struct {
	Display string
	Warning string
}

// GetTokenStatus checks token validity and returns display info.
func GetTokenStatus(token string) TokenStatusInfo {
	if err := network.ValidateToken(token); err != nil {
		return TokenStatusInfo{
			Display: output.RenderTokenValidity(false, false),
			Warning: "Token is invalid. Please run 'shelly cloud login' to re-authenticate.",
		}
	}

	if network.IsTokenExpired(token) {
		return TokenStatusInfo{
			Display: output.RenderTokenValidity(true, true),
			Warning: "Token has expired. Please run 'shelly cloud login' to re-authenticate.",
		}
	}

	return TokenStatusInfo{
		Display: output.RenderTokenValidity(true, false),
	}
}

// DisplayTLSConfig displays the TLS configuration and returns whether a custom CA is configured.
func DisplayTLSConfig(ios *iostreams.IOStreams, config map[string]any) bool {
	hasCustomCA := false

	// Check MQTT TLS settings
	if mqtt, ok := config["mqtt"].(map[string]any); ok {
		ios.Printf("MQTT:\n")
		if server, ok := mqtt["server"].(string); ok {
			ios.Printf("  Server: %s\n", server)
		}
		if sslCA, ok := mqtt["ssl_ca"].(string); ok && sslCA != "" {
			ios.Printf("  SSL CA: %s\n", sslCA)
			hasCustomCA = true
		} else {
			ios.Printf("  SSL CA: (not configured)\n")
		}
	}

	// Check cloud settings
	if cloud, ok := config["cloud"].(map[string]any); ok {
		ios.Printf("Cloud:\n")
		if enabled, ok := cloud["enable"].(bool); ok {
			ios.Printf("  Enabled: %t\n", enabled)
		}
	}

	// Check WS (WebSocket) settings
	if ws, ok := config["ws"].(map[string]any); ok {
		ios.Printf("WebSocket:\n")
		if sslCA, ok := ws["ssl_ca"].(string); ok && sslCA != "" {
			ios.Printf("  SSL CA: %s\n", sslCA)
			hasCustomCA = true
		}
	}

	return hasCustomCA
}
