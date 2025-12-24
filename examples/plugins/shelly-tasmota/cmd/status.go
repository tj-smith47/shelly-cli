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

// StatusResult is the output format for the status hook.
type StatusResult struct {
	Online     bool           `json:"online"`
	Components map[string]any `json:"components,omitempty"`
	Sensors    map[string]any `json:"sensors,omitempty"`
	Energy     *EnergyStatus  `json:"energy,omitempty"`
}

// EnergyStatus contains power/energy metrics.
type EnergyStatus struct {
	Power         float64 `json:"power,omitempty"`
	Voltage       float64 `json:"voltage,omitempty"`
	Current       float64 `json:"current,omitempty"`
	Total         float64 `json:"total,omitempty"`
	ApparentPower float64 `json:"apparent_power,omitempty"`
	ReactivePower float64 `json:"reactive_power,omitempty"`
	PowerFactor   float64 `json:"power_factor,omitempty"`
}

var statusFlags struct {
	address  string
	authUser string
	authPass string
	timeout  time.Duration
}

// NewStatusCmd creates the status command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get device status",
		Long: `Retrieve the current status of a Tasmota device.

Output is JSON in the DeviceStatusResult format expected by shelly-cli plugin hooks.
Includes power states, sensor readings, and energy metrics when available.`,
		Example: `  shelly-tasmota status --address=192.168.1.50
  shelly-tasmota status --address=192.168.1.50 --auth-user=admin --auth-pass=secret`,
		RunE: runStatus,
	}

	cmd.Flags().StringVar(&statusFlags.address, "address", "", "Device IP address (required)")
	cmd.Flags().StringVar(&statusFlags.authUser, "auth-user", "", "HTTP auth username")
	cmd.Flags().StringVar(&statusFlags.authPass, "auth-pass", "", "HTTP auth password")
	cmd.Flags().DurationVar(&statusFlags.timeout, "timeout", 5*time.Second, "Request timeout")

	if err := cmd.MarkFlagRequired("address"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to mark flag required: %v\n", err)
	}

	return cmd
}

func runStatus(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), statusFlags.timeout)
	defer cancel()

	client := tasmota.NewClient(statusFlags.address, statusFlags.authUser, statusFlags.authPass)

	// Try to get full status
	status, err := client.GetStatus(ctx)
	if err != nil {
		// Device offline or unreachable
		result := StatusResult{
			Online: false,
		}
		return outputJSON(result)
	}

	// Build the status result
	result := StatusResult{
		Online:     true,
		Components: make(map[string]any),
		Sensors:    make(map[string]any),
	}

	// Add switch/relay states
	addSwitchStates(&result, status)

	// Add sensor readings
	addSensorReadings(&result, status)

	// Add energy metrics
	addEnergyMetrics(&result, status)

	// Add WiFi signal info as a sensor
	if status.StatusSTS.Wifi.RSSI > 0 {
		result.Sensors["wifi"] = map[string]any{
			"rssi":     status.StatusSTS.Wifi.RSSI,
			"signal":   status.StatusSTS.Wifi.Signal,
			"ssid":     status.StatusSTS.Wifi.SSId,
			"channel":  status.StatusSTS.Wifi.Channel,
			"downtime": status.StatusSTS.Wifi.Downtime,
		}
	}

	return outputJSON(result)
}

// addSwitchStates adds relay/switch states to the result.
func addSwitchStates(result *StatusResult, status *tasmota.StatusAll) {
	// Parse power states from StatusSTS
	states := make(map[int]bool)

	// Single relay uses POWER, multi-relay uses POWER1, POWER2, etc.
	if status.StatusSTS.Power != "" {
		states[0] = strings.EqualFold(status.StatusSTS.Power, "ON")
	} else if status.StatusSTS.Power1 != "" {
		states[0] = strings.EqualFold(status.StatusSTS.Power1, "ON")
	}
	if status.StatusSTS.Power2 != "" {
		states[1] = strings.EqualFold(status.StatusSTS.Power2, "ON")
	}
	if status.StatusSTS.Power3 != "" {
		states[2] = strings.EqualFold(status.StatusSTS.Power3, "ON")
	}
	if status.StatusSTS.Power4 != "" {
		states[3] = strings.EqualFold(status.StatusSTS.Power4, "ON")
	}

	// Add each switch as component "switch:N"
	for id, isOn := range states {
		state := "off"
		if isOn {
			state = "on"
		}
		name := ""
		if id < len(status.Status.FriendlyName) {
			name = status.Status.FriendlyName[id]
		}

		key := fmt.Sprintf("switch:%d", id)
		result.Components[key] = map[string]any{
			"output": isOn,
			"state":  state,
			"name":   name,
		}
	}
}

// addSensorReadings adds sensor data to the result.
func addSensorReadings(result *StatusResult, status *tasmota.StatusAll) {
	sns := status.StatusSNS

	if sns.DS18B20 != nil {
		result.Sensors["temperature"] = map[string]any{
			"value": sns.DS18B20.Temperature,
			"unit":  "C",
			"id":    sns.DS18B20.ID,
		}
	}

	if sns.AM2301 != nil {
		result.Sensors["temperature"] = map[string]any{
			"value": sns.AM2301.Temperature,
			"unit":  "C",
		}
		result.Sensors["humidity"] = map[string]any{
			"value": sns.AM2301.Humidity,
			"unit":  "%",
		}
		result.Sensors["dewpoint"] = map[string]any{
			"value": sns.AM2301.DewPoint,
			"unit":  "C",
		}
	}

	if sns.BME280 != nil {
		result.Sensors["temperature"] = map[string]any{
			"value": sns.BME280.Temperature,
			"unit":  "C",
		}
		result.Sensors["humidity"] = map[string]any{
			"value": sns.BME280.Humidity,
			"unit":  "%",
		}
		result.Sensors["pressure"] = map[string]any{
			"value": sns.BME280.Pressure,
			"unit":  "hPa",
		}
		result.Sensors["dewpoint"] = map[string]any{
			"value": sns.BME280.DewPoint,
			"unit":  "C",
		}
	}
}

// addEnergyMetrics adds energy monitoring data to the result.
func addEnergyMetrics(result *StatusResult, status *tasmota.StatusAll) {
	energy := status.StatusSNS.Energy
	if energy == nil {
		return
	}

	result.Energy = &EnergyStatus{
		Power:         energy.Power,
		Voltage:       energy.Voltage,
		Current:       energy.Current,
		Total:         energy.Total,
		ApparentPower: energy.ApparentPower,
		ReactivePower: energy.ReactivePower,
		PowerFactor:   energy.Factor,
	}
}
