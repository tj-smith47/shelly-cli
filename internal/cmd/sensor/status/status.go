// Package status provides the sensor status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

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

// NewCommand creates the sensor status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "all", "readings"},
		Short:   "Show all sensor readings",
		Long: `Show all sensor readings from a Shelly device.

Displays a combined view of all available sensors including:
- Temperature (°C/°F)
- Humidity (%)
- Flood detection status
- Smoke detection status
- Illuminance (lux)
- Voltage readings

Only sensors present on the device will be shown.`,
		Example: `  # Show all sensor readings
  shelly sensor status living-room

  # Output as JSON
  shelly sensor status living-room --json`,
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

// SensorData holds all sensor readings.
type SensorData struct {
	Temperature []TempReading  `json:"temperature,omitempty"`
	Humidity    []HumidReading `json:"humidity,omitempty"`
	Flood       []FloodReading `json:"flood,omitempty"`
	Smoke       []SmokeReading `json:"smoke,omitempty"`
	Illuminance []LuxReading   `json:"illuminance,omitempty"`
	Voltmeter   []VoltReading  `json:"voltmeter,omitempty"`
}

// TempReading represents a temperature reading.
type TempReading struct {
	ID    int      `json:"id"`
	TempC *float64 `json:"temp_c,omitempty"`
	TempF *float64 `json:"temp_f,omitempty"`
}

// HumidReading represents a humidity reading.
type HumidReading struct {
	ID       int      `json:"id"`
	Humidity *float64 `json:"humidity,omitempty"`
}

// FloodReading represents a flood sensor reading.
type FloodReading struct {
	ID    int  `json:"id"`
	Alarm bool `json:"alarm"`
	Mute  bool `json:"mute"`
}

// SmokeReading represents a smoke sensor reading.
type SmokeReading struct {
	ID    int  `json:"id"`
	Alarm bool `json:"alarm"`
	Mute  bool `json:"mute"`
}

// LuxReading represents an illuminance reading.
type LuxReading struct {
	ID  int      `json:"id"`
	Lux *float64 `json:"lux,omitempty"`
}

// VoltReading represents a voltage reading.
type VoltReading struct {
	ID      int      `json:"id"`
	Voltage *float64 `json:"voltage,omitempty"`
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	// Get full device status
	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return fmt.Errorf("failed to get device status: %w", err)
	}

	// Marshal result to JSON for parsing
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var fullStatus map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &fullStatus); err != nil {
		return fmt.Errorf("failed to parse status: %w", err)
	}

	// Collect sensor data
	data := collectSensorData(fullStatus)

	if opts.JSON {
		output, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	return displaySensorData(ios, data, opts.Device)
}

func collectSensorData(status map[string]json.RawMessage) *SensorData {
	data := &SensorData{}

	for key, raw := range status {
		switch {
		case strings.HasPrefix(key, "temperature:"):
			collectTemperature(raw, data)
		case strings.HasPrefix(key, "humidity:"):
			collectHumidity(raw, data)
		case strings.HasPrefix(key, "flood:"):
			collectFlood(raw, data)
		case strings.HasPrefix(key, "smoke:"):
			collectSmoke(raw, data)
		case strings.HasPrefix(key, "illuminance:"):
			collectIlluminance(raw, data)
		case strings.HasPrefix(key, "voltmeter:"):
			collectVoltmeter(raw, data)
		}
	}

	return data
}

func collectTemperature(raw json.RawMessage, data *SensorData) {
	var temp struct {
		ID int      `json:"id"`
		TC *float64 `json:"tC"`
		TF *float64 `json:"tF"`
	}
	if err := json.Unmarshal(raw, &temp); err == nil {
		data.Temperature = append(data.Temperature, TempReading{
			ID:    temp.ID,
			TempC: temp.TC,
			TempF: temp.TF,
		})
	}
}

func collectHumidity(raw json.RawMessage, data *SensorData) {
	var humid struct {
		ID int      `json:"id"`
		RH *float64 `json:"rh"`
	}
	if err := json.Unmarshal(raw, &humid); err == nil {
		data.Humidity = append(data.Humidity, HumidReading{
			ID:       humid.ID,
			Humidity: humid.RH,
		})
	}
}

func collectFlood(raw json.RawMessage, data *SensorData) {
	var flood struct {
		ID    int  `json:"id"`
		Alarm bool `json:"alarm"`
		Mute  bool `json:"mute"`
	}
	if err := json.Unmarshal(raw, &flood); err == nil {
		data.Flood = append(data.Flood, FloodReading{
			ID:    flood.ID,
			Alarm: flood.Alarm,
			Mute:  flood.Mute,
		})
	}
}

func collectSmoke(raw json.RawMessage, data *SensorData) {
	var smoke struct {
		ID    int  `json:"id"`
		Alarm bool `json:"alarm"`
		Mute  bool `json:"mute"`
	}
	if err := json.Unmarshal(raw, &smoke); err == nil {
		data.Smoke = append(data.Smoke, SmokeReading{
			ID:    smoke.ID,
			Alarm: smoke.Alarm,
			Mute:  smoke.Mute,
		})
	}
}

func collectIlluminance(raw json.RawMessage, data *SensorData) {
	var illum struct {
		ID  int      `json:"id"`
		Lux *float64 `json:"lux"`
	}
	if err := json.Unmarshal(raw, &illum); err == nil {
		data.Illuminance = append(data.Illuminance, LuxReading{
			ID:  illum.ID,
			Lux: illum.Lux,
		})
	}
}

func collectVoltmeter(raw json.RawMessage, data *SensorData) {
	var volt struct {
		ID      int      `json:"id"`
		Voltage *float64 `json:"voltage"`
	}
	if err := json.Unmarshal(raw, &volt); err == nil {
		data.Voltmeter = append(data.Voltmeter, VoltReading{
			ID:      volt.ID,
			Voltage: volt.Voltage,
		})
	}
}

func displaySensorData(ios *iostreams.IOStreams, data *SensorData, device string) error {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Sensor Readings for %s:", device)))
	ios.Println()

	hasData := displayTemperature(ios, data.Temperature) ||
		displayHumidity(ios, data.Humidity) ||
		displayFlood(ios, data.Flood) ||
		displaySmoke(ios, data.Smoke) ||
		displayIlluminance(ios, data.Illuminance) ||
		displayVoltmeter(ios, data.Voltmeter)

	if !hasData {
		ios.Info("No sensors found on this device.")
	}

	return nil
}

func displayTemperature(ios *iostreams.IOStreams, temps []TempReading) bool {
	if len(temps) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Temperature:"))
	for _, t := range temps {
		if t.TempC != nil {
			ios.Printf("    Sensor %d: %.1f°C", t.ID, *t.TempC)
			if t.TempF != nil {
				ios.Printf(" (%.1f°F)", *t.TempF)
			}
			ios.Println()
		}
	}
	ios.Println()
	return true
}

func displayHumidity(ios *iostreams.IOStreams, humids []HumidReading) bool {
	if len(humids) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Humidity:"))
	for _, h := range humids {
		if h.Humidity != nil {
			ios.Printf("    Sensor %d: %.1f%%\n", h.ID, *h.Humidity)
		}
	}
	ios.Println()
	return true
}

func displayFlood(ios *iostreams.IOStreams, floods []FloodReading) bool {
	if len(floods) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Flood Detection:"))
	for _, f := range floods {
		status := theme.StatusOK().Render("Clear")
		if f.Alarm {
			status = theme.StatusError().Render("WATER DETECTED!")
		}
		muteStr := ""
		if f.Mute {
			muteStr = " " + theme.Dim().Render("(muted)")
		}
		ios.Printf("    Sensor %d: %s%s\n", f.ID, status, muteStr)
	}
	ios.Println()
	return true
}

func displaySmoke(ios *iostreams.IOStreams, smokes []SmokeReading) bool {
	if len(smokes) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Smoke Detection:"))
	for _, s := range smokes {
		status := theme.StatusOK().Render("Clear")
		if s.Alarm {
			status = theme.StatusError().Render("SMOKE DETECTED!")
		}
		muteStr := ""
		if s.Mute {
			muteStr = " " + theme.Dim().Render("(muted)")
		}
		ios.Printf("    Sensor %d: %s%s\n", s.ID, status, muteStr)
	}
	ios.Println()
	return true
}

func displayIlluminance(ios *iostreams.IOStreams, luxes []LuxReading) bool {
	if len(luxes) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Illuminance:"))
	for _, l := range luxes {
		if l.Lux != nil {
			ios.Printf("    Sensor %d: %.0f lux\n", l.ID, *l.Lux)
		}
	}
	ios.Println()
	return true
}

func displayVoltmeter(ios *iostreams.IOStreams, volts []VoltReading) bool {
	if len(volts) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Voltage:"))
	for _, v := range volts {
		if v.Voltage != nil {
			ios.Printf("    Sensor %d: %.2f V\n", v.ID, *v.Voltage)
		}
	}
	ios.Println()
	return true
}
