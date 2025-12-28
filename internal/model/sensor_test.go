package model

import "testing"

func TestAlarmSensorReading_GetID(t *testing.T) {
	t.Parallel()

	s := AlarmSensorReading{ID: 42}
	if got := s.GetID(); got != 42 {
		t.Errorf("GetID() = %d, want 42", got)
	}
}

func TestAlarmSensorReading_IsAlarm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		alarm bool
		want  bool
	}{
		{"alarm active", true, true},
		{"alarm inactive", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := AlarmSensorReading{Alarm: tt.alarm}
			if got := s.IsAlarm(); got != tt.want {
				t.Errorf("IsAlarm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlarmSensorReading_IsMuted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mute bool
		want bool
	}{
		{"muted", true, true},
		{"not muted", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := AlarmSensorReading{Mute: tt.mute}
			if got := s.IsMuted(); got != tt.want {
				t.Errorf("IsMuted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlarmSensorReading_GetErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		errors []string
		want   int
	}{
		{"no errors", nil, 0},
		{"empty errors", []string{}, 0},
		{"one error", []string{"sensor_fault"}, 1},
		{"multiple errors", []string{"sensor_fault", "low_battery"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := AlarmSensorReading{Errors: tt.errors}
			got := s.GetErrors()
			if len(got) != tt.want {
				t.Errorf("GetErrors() len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestTemperatureReading_GetID(t *testing.T) {
	t.Parallel()

	s := TemperatureReading{ID: 0}
	if got := s.GetID(); got != 0 {
		t.Errorf("GetID() = %d, want 0", got)
	}
}

func TestTemperatureReading_GetErrors(t *testing.T) {
	t.Parallel()

	s := TemperatureReading{Errors: []string{"disconnected"}}
	errors := s.GetErrors()
	if len(errors) != 1 {
		t.Errorf("GetErrors() len = %d, want 1", len(errors))
	}
	if errors[0] != "disconnected" {
		t.Errorf("GetErrors()[0] = %q, want %q", errors[0], "disconnected")
	}
}

func TestHumidityReading_GetID(t *testing.T) {
	t.Parallel()

	s := HumidityReading{ID: 1}
	if got := s.GetID(); got != 1 {
		t.Errorf("GetID() = %d, want 1", got)
	}
}

func TestHumidityReading_GetErrors(t *testing.T) {
	t.Parallel()

	s := HumidityReading{Errors: nil}
	if got := s.GetErrors(); got != nil {
		t.Errorf("GetErrors() = %v, want nil", got)
	}
}

func TestIlluminanceReading_GetID(t *testing.T) {
	t.Parallel()

	s := IlluminanceReading{ID: 2}
	if got := s.GetID(); got != 2 {
		t.Errorf("GetID() = %d, want 2", got)
	}
}

func TestIlluminanceReading_GetErrors(t *testing.T) {
	t.Parallel()

	s := IlluminanceReading{Errors: []string{"overload"}}
	errors := s.GetErrors()
	if len(errors) != 1 || errors[0] != "overload" {
		t.Errorf("GetErrors() = %v, want [overload]", errors)
	}
}

func TestVoltmeterReading_GetID(t *testing.T) {
	t.Parallel()

	s := VoltmeterReading{ID: 3}
	if got := s.GetID(); got != 3 {
		t.Errorf("GetID() = %d, want 3", got)
	}
}

func TestVoltmeterReading_GetErrors(t *testing.T) {
	t.Parallel()

	s := VoltmeterReading{Errors: []string{}}
	if got := s.GetErrors(); len(got) != 0 {
		t.Errorf("GetErrors() len = %d, want 0", len(got))
	}
}

func TestDevicePowerReading_GetID(t *testing.T) {
	t.Parallel()

	s := DevicePowerReading{ID: 0}
	if got := s.GetID(); got != 0 {
		t.Errorf("GetID() = %d, want 0", got)
	}
}

func TestDevicePowerReading_GetErrors(t *testing.T) {
	t.Parallel()

	s := DevicePowerReading{Errors: []string{"battery_low", "external_fault"}}
	errors := s.GetErrors()
	if len(errors) != 2 {
		t.Errorf("GetErrors() len = %d, want 2", len(errors))
	}
}

func TestDevicePowerBatteryStatus(t *testing.T) {
	t.Parallel()

	status := DevicePowerBatteryStatus{
		V:       3.7,
		Percent: 85,
	}

	if status.V != 3.7 {
		t.Errorf("V = %f, want 3.7", status.V)
	}
	if status.Percent != 85 {
		t.Errorf("Percent = %d, want 85", status.Percent)
	}
}

func TestDevicePowerExternalStatus(t *testing.T) {
	t.Parallel()

	status := DevicePowerExternalStatus{Present: true}
	if !status.Present {
		t.Error("Present = false, want true")
	}
}

func TestSensorData(t *testing.T) {
	t.Parallel()

	data := SensorData{
		Temperature: []TemperatureReading{{ID: 0}},
		Humidity:    []HumidityReading{{ID: 0}},
		Flood:       []AlarmSensorReading{{ID: 0, Alarm: true}},
	}

	if len(data.Temperature) != 1 {
		t.Errorf("Temperature len = %d, want 1", len(data.Temperature))
	}
	if len(data.Humidity) != 1 {
		t.Errorf("Humidity len = %d, want 1", len(data.Humidity))
	}
	if len(data.Flood) != 1 {
		t.Errorf("Flood len = %d, want 1", len(data.Flood))
	}
	if len(data.Smoke) != 0 {
		t.Errorf("Smoke len = %d, want 0", len(data.Smoke))
	}
}
