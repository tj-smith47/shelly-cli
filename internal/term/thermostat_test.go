package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

//nolint:gocyclo // comprehensive test coverage
func TestDisplayThermostatStatus(t *testing.T) {
	t.Parallel()

	t.Run("basic status", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		currentC := 22.5
		currentF := 72.5
		targetC := 21.0
		targetF := 69.8
		status := &components.ThermostatStatus{
			CurrentC: &currentC,
			CurrentF: &currentF,
			TargetC:  &targetC,
			TargetF:  &targetF,
		}

		DisplayThermostatStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "Thermostat") {
			t.Error("output should contain 'Thermostat'")
		}
		if !strings.Contains(output, "22.5") {
			t.Error("output should contain current temperature")
		}
		if !strings.Contains(output, "21.0") {
			t.Error("output should contain target temperature")
		}
	})

	t.Run("with valve position", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		pos := 50
		outputBool := true
		status := &components.ThermostatStatus{
			Pos:    &pos,
			Output: &outputBool,
		}

		DisplayThermostatStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "Valve") {
			t.Error("output should contain 'Valve'")
		}
		if !strings.Contains(output, "50%") {
			t.Error("output should contain valve position")
		}
	})

	t.Run("with humidity", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		currentHumidity := 60.0
		targetHumidity := 50.0
		status := &components.ThermostatStatus{
			CurrentHumidity: &currentHumidity,
			TargetHumidity:  &targetHumidity,
		}

		DisplayThermostatStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "Humidity") {
			t.Error("output should contain 'Humidity'")
		}
		if !strings.Contains(output, "60.0") {
			t.Error("output should contain current humidity")
		}
	})

	t.Run("with boost mode", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := &components.ThermostatStatus{
			Boost: &components.ThermostatModeInfo{
				StartedAt: 1700000000,
				Duration:  300,
			},
		}

		DisplayThermostatStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "Boost") {
			t.Error("output should contain 'Boost'")
		}
		if !strings.Contains(output, "300") {
			t.Error("output should contain boost duration")
		}
	})

	t.Run("with override mode", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := &components.ThermostatStatus{
			Override: &components.ThermostatModeInfo{
				StartedAt: 1700000000,
				Duration:  600,
			},
		}

		DisplayThermostatStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "Override") {
			t.Error("output should contain 'Override'")
		}
		if !strings.Contains(output, "600") {
			t.Error("output should contain override duration")
		}
	})

	t.Run("with flags", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := &components.ThermostatStatus{
			Flags: []string{"calibrating", "window_open"},
		}

		DisplayThermostatStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "Flags") {
			t.Error("output should contain 'Flags'")
		}
		if !strings.Contains(output, "calibrating") {
			t.Error("output should contain flag value")
		}
	})

	t.Run("with errors", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := &components.ThermostatStatus{
			Errors: []string{"motor_error", "sensor_error"},
		}

		DisplayThermostatStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "Errors") {
			t.Error("output should contain 'Errors'")
		}
		if !strings.Contains(output, "motor_error") {
			t.Error("output should contain error value")
		}
	})
}

func TestDisplayThermostats(t *testing.T) {
	t.Parallel()

	t.Run("with thermostats", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		thermostats := []model.ThermostatInfo{
			{ID: 0, Enabled: true, TargetC: 22.0},
			{ID: 1, Enabled: false, TargetC: 0},
		}

		DisplayThermostats(ios, thermostats, "gateway-device")

		output := out.String()
		if !strings.Contains(output, "gateway-device") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "Thermostat") {
			t.Error("output should contain 'Thermostat'")
		}
		if !strings.Contains(output, "22.0") {
			t.Error("output should contain target temperature")
		}
		if !strings.Contains(output, "2 thermostat") {
			t.Error("output should contain count")
		}
	})

	t.Run("empty thermostats", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()

		DisplayThermostats(ios, []model.ThermostatInfo{}, "device1")

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No thermostats found") {
			t.Errorf("output should contain 'No thermostats found', got %q", allOutput)
		}
		if !strings.Contains(allOutput, "BLU TRV") {
			t.Error("output should contain hint about BLU TRV")
		}
	})
}

func TestDisplayThermostatTemperature(t *testing.T) {
	t.Parallel()

	t.Run("celsius only", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		currentC := 20.0
		status := &components.ThermostatStatus{
			CurrentC: &currentC,
		}

		displayThermostatTemperature(ios, status)

		output := out.String()
		if !strings.Contains(output, "20.0") {
			t.Error("output should contain temperature")
		}
	})

	t.Run("with fahrenheit", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		currentC := 20.0
		currentF := 68.0
		status := &components.ThermostatStatus{
			CurrentC: &currentC,
			CurrentF: &currentF,
		}

		displayThermostatTemperature(ios, status)

		output := out.String()
		if !strings.Contains(output, "68.0") {
			t.Error("output should contain fahrenheit temperature")
		}
	})
}
