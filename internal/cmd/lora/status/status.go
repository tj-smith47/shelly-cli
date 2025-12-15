// Package status provides the lora status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	JSON    bool
}

// NewCommand creates the lora status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "info"},
		Short:   "Show LoRa add-on status",
		Long: `Show LoRa add-on status for a Shelly device.

Displays the current LoRa configuration and signal quality
information from the last received packet.`,
		Example: `  # Show LoRa status
  shelly lora status living-room

  # Specify component ID (default: 100)
  shelly lora status living-room --id 100

  # Output as JSON
  shelly lora status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 100, "LoRa component ID")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// LoRaFullStatus combines config and status.
type LoRaFullStatus struct {
	Config *LoRaConfig `json:"config,omitempty"`
	Status *LoRaStatus `json:"status,omitempty"`
}

// LoRaConfig represents LoRa configuration.
type LoRaConfig struct {
	ID   int   `json:"id"`
	Freq int64 `json:"freq"`
	BW   int   `json:"bw"`
	DR   int   `json:"dr"`
	TxP  int   `json:"txp"`
}

// LoRaStatus represents LoRa status.
type LoRaStatus struct {
	ID   int     `json:"id"`
	RSSI int     `json:"rssi"`
	SNR  float64 `json:"snr"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var full LoRaFullStatus

	// Get config
	cfgMap, err := svc.LoRaGetConfig(ctx, opts.Device, opts.ID)
	if err != nil {
		ios.Debug("LoRa.GetConfig failed: %v", err)
		return fmt.Errorf("LoRa not available on this device: %w", err)
	}

	cfg := &LoRaConfig{}
	if id, ok := cfgMap["id"].(float64); ok {
		cfg.ID = int(id)
	}
	if freq, ok := cfgMap["freq"].(float64); ok {
		cfg.Freq = int64(freq)
	}
	if bw, ok := cfgMap["bw"].(float64); ok {
		cfg.BW = int(bw)
	}
	if dr, ok := cfgMap["dr"].(float64); ok {
		cfg.DR = int(dr)
	}
	if txp, ok := cfgMap["txp"].(float64); ok {
		cfg.TxP = int(txp)
	}
	full.Config = cfg

	// Get status
	statusMap, err := svc.LoRaGetStatus(ctx, opts.Device, opts.ID)
	if err != nil {
		ios.Debug("LoRa.GetStatus failed: %v", err)
		// Config succeeded, status failed - still show partial info
	}
	if err == nil {
		full.Status = parseLoRaStatus(statusMap)
	}

	if opts.JSON {
		output, err := json.MarshalIndent(full, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

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

	return nil
}

func parseLoRaStatus(statusMap map[string]any) *LoRaStatus {
	st := &LoRaStatus{}
	if id, ok := statusMap["id"].(float64); ok {
		st.ID = int(id)
	}
	if rssi, ok := statusMap["rssi"].(float64); ok {
		st.RSSI = int(rssi)
	}
	if snr, ok := statusMap["snr"].(float64); ok {
		st.SNR = snr
	}
	return st
}
