// Package term provides composed terminal presentation for the CLI.
package term

import (
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayLoRaStatus displays the LoRa add-on status.
func DisplayLoRaStatus(ios *iostreams.IOStreams, full model.LoRaFullStatus) {
	ios.Println(theme.Bold().Render("LoRa Add-on Status:"))
	ios.Println()

	if full.Config != nil {
		ios.Println("  " + theme.Highlight().Render("Configuration:"))
		ios.Printf("    Component ID: %d\n", full.Config.ID)
		ios.Printf("    Frequency: %d Hz (%.3f MHz)\n", full.Config.Freq, float64(full.Config.Freq)/1e6)
		ios.Printf("    Bandwidth: %d\n", full.Config.BW)
		ios.Printf("    Data Rate (SF): %d\n", full.Config.DR)
		ios.Printf("    TX Power: %d dBm\n", full.Config.TxP)
	}

	if full.Status != nil {
		ios.Println()
		ios.Println("  " + theme.Highlight().Render("Last Packet:"))
		ios.Printf("    RSSI: %d dBm\n", full.Status.RSSI)
		ios.Printf("    SNR: %.1f dB\n", full.Status.SNR)
	}
}

// OutputLoRaStatusJSON outputs LoRa status as JSON.
func OutputLoRaStatusJSON(ios *iostreams.IOStreams, full model.LoRaFullStatus) error {
	output, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(output))
	return nil
}
