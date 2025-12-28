package shelly

import (
	"encoding/json"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestCollectSensorsByPrefix(t *testing.T) {
	t.Parallel()

	t.Run("collects temperature sensors", func(t *testing.T) {
		t.Parallel()

		status := map[string]json.RawMessage{
			"temperature:0": json.RawMessage(`{"id":0,"tC":22.5,"tF":72.5}`),
			"temperature:1": json.RawMessage(`{"id":1,"tC":23.0,"tF":73.4}`),
			"switch:0":      json.RawMessage(`{"id":0,"output":true}`),
		}

		result := CollectSensorsByPrefix[model.TemperatureReading](status, "temperature:", nil)

		if len(result) != 2 {
			t.Errorf("expected 2 sensors, got %d", len(result))
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		t.Parallel()

		status := map[string]json.RawMessage{
			"switch:0": json.RawMessage(`{"id":0,"output":true}`),
		}

		result := CollectSensorsByPrefix[model.TemperatureReading](status, "temperature:", nil)

		if len(result) != 0 {
			t.Errorf("expected 0 sensors, got %d", len(result))
		}
	})

	t.Run("skips invalid JSON", func(t *testing.T) {
		t.Parallel()

		status := map[string]json.RawMessage{
			"temperature:0": json.RawMessage(`{"id":0,"tC":22.5}`),
			"temperature:1": json.RawMessage(`{invalid json}`),
		}

		result := CollectSensorsByPrefix[model.TemperatureReading](status, "temperature:", nil)

		if len(result) != 1 {
			t.Errorf("expected 1 sensor (skip invalid), got %d", len(result))
		}
	})
}

func TestCollectSensorsByPrefixSilent(t *testing.T) {
	t.Parallel()

	status := map[string]json.RawMessage{
		"humidity:0": json.RawMessage(`{"id":0,"rh":55.5}`),
	}

	result := CollectSensorsByPrefixSilent[model.HumidityReading](status, "humidity:")

	if len(result) != 1 {
		t.Errorf("expected 1 sensor, got %d", len(result))
	}
}

func TestCollectSensorData(t *testing.T) {
	t.Parallel()

	t.Run("collects all sensor types", func(t *testing.T) {
		t.Parallel()

		status := map[string]json.RawMessage{
			"temperature:0": json.RawMessage(`{"id":0,"tC":22.5}`),
			"humidity:0":    json.RawMessage(`{"id":0,"rh":55.0}`),
			"flood:0":       json.RawMessage(`{"id":0,"alarm":false,"mute":false}`),
			"smoke:0":       json.RawMessage(`{"id":0,"alarm":false,"mute":false}`),
			"illuminance:0": json.RawMessage(`{"id":0,"lux":500}`),
			"voltmeter:0":   json.RawMessage(`{"id":0,"voltage":12.5}`),
		}

		data := CollectSensorData(status)

		if len(data.Temperature) != 1 {
			t.Error("expected 1 temperature sensor")
		}
		if len(data.Humidity) != 1 {
			t.Error("expected 1 humidity sensor")
		}
		if len(data.Flood) != 1 {
			t.Error("expected 1 flood sensor")
		}
		if len(data.Smoke) != 1 {
			t.Error("expected 1 smoke sensor")
		}
		if len(data.Illuminance) != 1 {
			t.Error("expected 1 illuminance sensor")
		}
		if len(data.Voltmeter) != 1 {
			t.Error("expected 1 voltmeter sensor")
		}
	})

	t.Run("handles empty status", func(t *testing.T) {
		t.Parallel()

		status := map[string]json.RawMessage{}

		data := CollectSensorData(status)

		if data == nil {
			t.Fatal("expected non-nil SensorData")
		}
		if len(data.Temperature) != 0 {
			t.Error("expected empty temperature list")
		}
	})
}
