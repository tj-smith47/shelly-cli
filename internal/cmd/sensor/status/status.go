// Package status provides the sensor status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/model"
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
		jsonOut, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(jsonOut))
		return nil
	}

	cmdutil.DisplayAllSensorData(ios, data, opts.Device)
	return nil
}

func collectSensorData(status map[string]json.RawMessage) *model.SensorData {
	data := &model.SensorData{}

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

func collectTemperature(raw json.RawMessage, data *model.SensorData) {
	var temp struct {
		ID int      `json:"id"`
		TC *float64 `json:"tC"`
		TF *float64 `json:"tF"`
	}
	if err := json.Unmarshal(raw, &temp); err == nil {
		data.Temperature = append(data.Temperature, model.TemperatureReading{
			ID: temp.ID,
			TC: temp.TC,
			TF: temp.TF,
		})
	}
}

func collectHumidity(raw json.RawMessage, data *model.SensorData) {
	var humid struct {
		ID int      `json:"id"`
		RH *float64 `json:"rh"`
	}
	if err := json.Unmarshal(raw, &humid); err == nil {
		data.Humidity = append(data.Humidity, model.HumidityReading{
			ID: humid.ID,
			RH: humid.RH,
		})
	}
}

func collectFlood(raw json.RawMessage, data *model.SensorData) {
	var flood struct {
		ID    int  `json:"id"`
		Alarm bool `json:"alarm"`
		Mute  bool `json:"mute"`
	}
	if err := json.Unmarshal(raw, &flood); err == nil {
		data.Flood = append(data.Flood, model.AlarmSensorReading{
			ID:    flood.ID,
			Alarm: flood.Alarm,
			Mute:  flood.Mute,
		})
	}
}

func collectSmoke(raw json.RawMessage, data *model.SensorData) {
	var smoke struct {
		ID    int  `json:"id"`
		Alarm bool `json:"alarm"`
		Mute  bool `json:"mute"`
	}
	if err := json.Unmarshal(raw, &smoke); err == nil {
		data.Smoke = append(data.Smoke, model.AlarmSensorReading{
			ID:    smoke.ID,
			Alarm: smoke.Alarm,
			Mute:  smoke.Mute,
		})
	}
}

func collectIlluminance(raw json.RawMessage, data *model.SensorData) {
	var illum struct {
		ID  int      `json:"id"`
		Lux *float64 `json:"lux"`
	}
	if err := json.Unmarshal(raw, &illum); err == nil {
		data.Illuminance = append(data.Illuminance, model.IlluminanceReading{
			ID:  illum.ID,
			Lux: illum.Lux,
		})
	}
}

func collectVoltmeter(raw json.RawMessage, data *model.SensorData) {
	var volt struct {
		ID      int      `json:"id"`
		Voltage *float64 `json:"voltage"`
	}
	if err := json.Unmarshal(raw, &volt); err == nil {
		data.Voltmeter = append(data.Voltmeter, model.VoltmeterReading{
			ID:      volt.ID,
			Voltage: volt.Voltage,
		})
	}
}
