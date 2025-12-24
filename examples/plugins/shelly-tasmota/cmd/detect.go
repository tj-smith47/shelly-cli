// Package cmd implements the plugin commands for the Tasmota plugin.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/examples/plugins/shelly-tasmota/tasmota"
)

// ComponentInfo describes a device component (matches shelly-cli types).
type ComponentInfo struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

// DetectionResult is the output format for the detect hook.
type DetectionResult struct {
	Detected   bool            `json:"detected"`
	Platform   string          `json:"platform"`
	DeviceID   string          `json:"device_id,omitempty"`
	DeviceName string          `json:"device_name,omitempty"`
	Model      string          `json:"model,omitempty"`
	Firmware   string          `json:"firmware,omitempty"`
	MAC        string          `json:"mac,omitempty"`
	Components []ComponentInfo `json:"components,omitempty"`
}

var detectFlags struct {
	address  string
	authUser string
	authPass string
	timeout  time.Duration
}

// NewDetectCmd creates the detect command.
func NewDetectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect if a device is running Tasmota",
		Long: `Probe an IP address to determine if the device is running Tasmota firmware.

Output is JSON in the DeviceDetectionResult format expected by shelly-cli plugin hooks.
Exit code 0 means detection was attempted (check "detected" field for result).
Exit code 1 means an error occurred during detection.`,
		Example: `  shelly-tasmota detect --address=192.168.1.50
  shelly-tasmota detect --address=192.168.1.50 --auth-user=admin --auth-pass=secret`,
		RunE: runDetect,
	}

	cmd.Flags().StringVar(&detectFlags.address, "address", "", "Device IP address (required)")
	cmd.Flags().StringVar(&detectFlags.authUser, "auth-user", "", "HTTP auth username")
	cmd.Flags().StringVar(&detectFlags.authPass, "auth-pass", "", "HTTP auth password")
	cmd.Flags().DurationVar(&detectFlags.timeout, "timeout", 5*time.Second, "Request timeout")

	if err := cmd.MarkFlagRequired("address"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to mark flag required: %v\n", err)
	}

	return cmd
}

func runDetect(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), detectFlags.timeout)
	defer cancel()

	client := tasmota.NewClient(detectFlags.address, detectFlags.authUser, detectFlags.authPass)

	// Try to get status - this tells us if it's Tasmota and gives us device info
	status, err := client.GetStatus(ctx)
	if err != nil {
		// Not a Tasmota device or unreachable - return detected=false
		result := DetectionResult{
			Detected: false,
			Platform: "tasmota",
		}
		return outputJSON(result)
	}

	// It's a Tasmota device! Build the detection result
	result := DetectionResult{
		Detected: true,
		Platform: "tasmota",
	}

	// Extract device name
	if len(status.Status.FriendlyName) > 0 {
		result.DeviceName = status.Status.FriendlyName[0]
	}
	if result.DeviceName == "" && status.Status.DeviceName != "" {
		result.DeviceName = status.Status.DeviceName
	}

	// Device ID from hostname or topic
	if status.StatusNET.Hostname != "" {
		result.DeviceID = status.StatusNET.Hostname
	} else if status.Status.Topic != "" {
		result.DeviceID = status.Status.Topic
	}

	// MAC address
	result.MAC = status.StatusNET.Mac

	// Firmware version
	result.Firmware = status.StatusFWR.Version

	// Model from hardware info
	if status.StatusFWR.Hardware != "" {
		result.Model = status.StatusFWR.Hardware
	}

	// Discover components - count relays and add sensor if present
	result.Components = discoverComponents(status)

	return outputJSON(result)
}

// discoverComponents detects available components on the device.
func discoverComponents(status *tasmota.StatusAll) []ComponentInfo {
	// Count relays from power state
	relayCount := countRelaysFromStatus(status)

	// Pre-allocate for relays plus up to 4 possible sensors (energy, temp sensors)
	components := make([]ComponentInfo, 0, relayCount+4)
	for i := range relayCount {
		name := ""
		if i < len(status.Status.FriendlyName) {
			name = status.Status.FriendlyName[i]
		}
		components = append(components, ComponentInfo{
			Type: "switch",
			ID:   i,
			Name: name,
		})
	}

	// Check for energy monitoring
	if status.StatusSNS.Energy != nil {
		components = append(components, ComponentInfo{
			Type: "energy",
			ID:   0,
			Name: "Energy Meter",
		})
	}

	// Check for temperature sensors
	if status.StatusSNS.DS18B20 != nil {
		components = append(components, ComponentInfo{
			Type: "sensor",
			ID:   0,
			Name: "Temperature",
		})
	}
	if status.StatusSNS.AM2301 != nil {
		components = append(components, ComponentInfo{
			Type: "sensor",
			ID:   1,
			Name: "Temperature/Humidity",
		})
	}
	if status.StatusSNS.BME280 != nil {
		components = append(components, ComponentInfo{
			Type: "sensor",
			ID:   2,
			Name: "Environmental",
		})
	}

	return components
}

// countRelaysFromStatus counts relays from the Status response.
func countRelaysFromStatus(status *tasmota.StatusAll) int {
	// Try StatusSTS first
	if status.StatusSTS.Power != "" || status.StatusSTS.Power1 != "" {
		count := 1
		if status.StatusSTS.Power2 != "" {
			count = 2
		}
		if status.StatusSTS.Power3 != "" {
			count = 3
		}
		if status.StatusSTS.Power4 != "" {
			count = 4
		}
		return count
	}

	// Fall back to FriendlyName count
	if len(status.Status.FriendlyName) > 0 {
		return len(status.Status.FriendlyName)
	}

	// Default to 1 if we got this far (it's a Tasmota device)
	return 1
}

// outputJSON writes the result as JSON to stdout.
func outputJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
