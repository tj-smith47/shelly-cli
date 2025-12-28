// Package term provides terminal display functions.
package term

import (
	"strings"

	"github.com/tj-smith47/shelly-go/profiles"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayProfile prints a device profile to the terminal.
func DisplayProfile(ios *iostreams.IOStreams, p *profiles.Profile) {
	ios.Title("Device Profile")
	ios.Println()

	ios.Printf("  %s: %s\n", theme.Dim().Render("Model"), theme.Highlight().Render(p.Model))
	ios.Printf("  %s: %s\n", theme.Dim().Render("Name"), p.Name)
	if p.App != "" {
		ios.Printf("  %s: %s\n", theme.Dim().Render("App"), p.App)
	}
	ios.Printf("  %s: %s\n", theme.Dim().Render("Generation"), p.Generation.String())
	ios.Printf("  %s: %s\n", theme.Dim().Render("Series"), string(p.Series))
	ios.Printf("  %s: %s\n", theme.Dim().Render("Form Factor"), string(p.FormFactor))
	ios.Printf("  %s: %s\n", theme.Dim().Render("Power Source"), string(p.PowerSource))

	// Components
	ios.Println()
	ios.Printf("  %s:\n", theme.Dim().Render("Components"))
	displayProfileComponents(ios, p)

	// Protocols
	ios.Println()
	ios.Printf("  %s:\n", theme.Dim().Render("Protocols"))
	protocols := profileProtocolList(p)
	ios.Printf("    %s\n", strings.Join(protocols, ", "))

	// Capabilities
	ios.Println()
	ios.Printf("  %s:\n", theme.Dim().Render("Capabilities"))
	caps := profileCapabilityList(p)
	for _, cap := range caps {
		ios.Printf("    • %s\n", cap)
	}

	// Limits
	if profileHasLimits(p) {
		ios.Println()
		ios.Printf("  %s:\n", theme.Dim().Render("Limits"))
		displayProfileLimits(ios, p)
	}

	ios.Println()
}

func displayProfileComponents(ios *iostreams.IOStreams, p *profiles.Profile) {
	if p.Components.Switches > 0 {
		ios.Printf("    Switches: %d\n", p.Components.Switches)
	}
	if p.Components.Covers > 0 {
		ios.Printf("    Covers: %d\n", p.Components.Covers)
	}
	if p.Components.Lights > 0 {
		ios.Printf("    Lights: %d\n", p.Components.Lights)
	}
	if p.Components.Inputs > 0 {
		ios.Printf("    Inputs: %d\n", p.Components.Inputs)
	}
	if p.Components.PowerMeters > 0 {
		ios.Printf("    Power Meters: %d\n", p.Components.PowerMeters)
	}
	if p.Components.EnergyMeters > 0 {
		ios.Printf("    Energy Meters: %d\n", p.Components.EnergyMeters)
	}
	if p.Components.TemperatureSensors > 0 {
		ios.Printf("    Temperature Sensors: %d\n", p.Components.TemperatureSensors)
	}
	if p.Components.HumiditySensors > 0 {
		ios.Printf("    Humidity Sensors: %d\n", p.Components.HumiditySensors)
	}
	if p.Components.RGBChannels > 0 {
		ios.Printf("    RGB Channels: %d\n", p.Components.RGBChannels)
	}
	if p.Components.Thermostat {
		ios.Printf("    Thermostat: %s\n", boolYesNo(true))
	}
	if p.Components.Display {
		ios.Printf("    Display: %s\n", boolYesNo(true))
	}
}

func displayProfileLimits(ios *iostreams.IOStreams, p *profiles.Profile) {
	if p.Limits.MaxScripts > 0 {
		ios.Printf("    Max Scripts: %d\n", p.Limits.MaxScripts)
	}
	if p.Limits.MaxSchedules > 0 {
		ios.Printf("    Max Schedules: %d\n", p.Limits.MaxSchedules)
	}
	if p.Limits.MaxWebhooks > 0 {
		ios.Printf("    Max Webhooks: %d\n", p.Limits.MaxWebhooks)
	}
	if p.Limits.MaxKVSEntries > 0 {
		ios.Printf("    Max KVS Entries: %d\n", p.Limits.MaxKVSEntries)
	}
	if p.Limits.MaxPower > 0 {
		ios.Printf("    Max Power: %.0fW\n", p.Limits.MaxPower)
	}
	if p.Limits.MaxInputCurrent > 0 {
		ios.Printf("    Max Input Current: %.1fA\n", p.Limits.MaxInputCurrent)
	}
}

func boolYesNo(b bool) string {
	if b {
		return theme.StatusOK().Render("Yes")
	}
	return theme.StatusError().Render("No")
}

func profileProtocolList(p *profiles.Profile) []string {
	var protocols []string
	if p.Protocols.HTTP {
		protocols = append(protocols, "HTTP")
	}
	if p.Protocols.WebSocket {
		protocols = append(protocols, "WebSocket")
	}
	if p.Protocols.MQTT {
		protocols = append(protocols, "MQTT")
	}
	if p.Protocols.CoIoT {
		protocols = append(protocols, "CoIoT")
	}
	if p.Protocols.BLE {
		protocols = append(protocols, "BLE")
	}
	if p.Protocols.Matter {
		protocols = append(protocols, "Matter")
	}
	if p.Protocols.Zigbee {
		protocols = append(protocols, "Zigbee")
	}
	if p.Protocols.ZWave {
		protocols = append(protocols, "Z-Wave")
	}
	if p.Protocols.Ethernet {
		protocols = append(protocols, "Ethernet")
	}
	return protocols
}

func profileCapabilityList(p *profiles.Profile) []string {
	// Map capabilities to their display names
	capabilityMap := []struct {
		enabled bool
		name    string
	}{
		{p.Capabilities.PowerMetering, "Power Metering"},
		{p.Capabilities.EnergyMetering, "Energy Metering"},
		{p.Capabilities.CoverSupport, "Cover Control"},
		{p.Capabilities.DimmingSupport, "Dimming"},
		{p.Capabilities.ColorSupport, "RGB Color"},
		{p.Capabilities.ColorTemperature, "Color Temperature"},
		{p.Capabilities.Scripting, "Scripting"},
		{p.Capabilities.Schedules, "Schedules"},
		{p.Capabilities.Webhooks, "Webhooks"},
		{p.Capabilities.KVS, "KVS Storage"},
		{p.Capabilities.VirtualComponents, "Virtual Components"},
		{p.Capabilities.SensorAddon, "Sensor Add-on"},
		{p.Capabilities.InputEvents, "Input Events"},
		{p.Capabilities.Effects, "Lighting Effects"},
		{p.Capabilities.NoNeutral, "No Neutral Required"},
		{p.Capabilities.BidirectionalMetering, "Bidirectional Metering"},
		{p.Capabilities.ThreePhase, "3-Phase Support"},
	}

	var caps []string
	for _, c := range capabilityMap {
		if c.enabled {
			caps = append(caps, c.name)
		}
	}

	if len(caps) == 0 {
		caps = append(caps, "Basic functionality")
	}
	return caps
}

func profileHasLimits(p *profiles.Profile) bool {
	return p.Limits.MaxScripts > 0 ||
		p.Limits.MaxSchedules > 0 ||
		p.Limits.MaxWebhooks > 0 ||
		p.Limits.MaxKVSEntries > 0 ||
		p.Limits.MaxPower > 0 ||
		p.Limits.MaxInputCurrent > 0
}

// FilterProfilesByCapability filters profiles by capability.
func FilterProfilesByCapability(items []*profiles.Profile, capability string) []*profiles.Profile {
	var result []*profiles.Profile
	byCapability := profiles.ListByCapability(capability)
	capMap := make(map[string]bool)
	for _, p := range byCapability {
		capMap[p.Model] = true
	}
	for _, p := range items {
		if capMap[p.Model] {
			result = append(result, p)
		}
	}
	return result
}

// FilterProfilesByProtocol filters profiles by protocol.
func FilterProfilesByProtocol(items []*profiles.Profile, protocol string) []*profiles.Profile {
	var result []*profiles.Profile
	byProtocol := profiles.ListByProtocol(protocol)
	protoMap := make(map[string]bool)
	for _, p := range byProtocol {
		protoMap[p.Model] = true
	}
	for _, p := range items {
		if protoMap[p.Model] {
			result = append(result, p)
		}
	}
	return result
}

// ParseProfileGeneration parses a generation string.
func ParseProfileGeneration(s string) types.Generation {
	switch s {
	case "1", "gen1", "Gen1":
		return types.Gen1
	case "2", "gen2", "Gen2":
		return types.Gen2
	case "3", "gen3", "Gen3":
		return types.Gen3
	case "4", "gen4", "Gen4":
		return types.Gen4
	default:
		return types.GenerationUnknown
	}
}

// ParseProfileSeries parses a series string.
func ParseProfileSeries(s string) profiles.Series {
	switch s {
	case "classic":
		return profiles.SeriesClassic
	case "plus":
		return profiles.SeriesPlus
	case "pro":
		return profiles.SeriesPro
	case "mini":
		return profiles.SeriesMini
	case "blu":
		return profiles.SeriesBLU
	case "wave":
		return profiles.SeriesWave
	case "wave_pro":
		return profiles.SeriesWavePro
	case "standard":
		return profiles.SeriesStandard
	default:
		return ""
	}
}
