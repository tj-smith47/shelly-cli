// Package model defines core domain types for the Shelly CLI.
package model

// AlarmSensor defines the interface for alarm-type sensors (smoke, flood).
// Used with Go generics for type-safe, DRY display functions.
type AlarmSensor interface {
	GetID() int
	IsAlarm() bool
	IsMuted() bool
	GetErrors() []string
}

// AlarmSensorReading represents a sensor with alarm and mute state (flood, smoke).
type AlarmSensorReading struct {
	ID     int      `json:"id"`
	Alarm  bool     `json:"alarm"`
	Mute   bool     `json:"mute"`
	Errors []string `json:"errors,omitempty"`
}

// GetID implements AlarmSensor.
func (a AlarmSensorReading) GetID() int { return a.ID }

// IsAlarm implements AlarmSensor.
func (a AlarmSensorReading) IsAlarm() bool { return a.Alarm }

// IsMuted implements AlarmSensor.
func (a AlarmSensorReading) IsMuted() bool { return a.Mute }

// GetErrors implements AlarmSensor.
func (a AlarmSensorReading) GetErrors() []string { return a.Errors }

// TemperatureReading represents a temperature sensor reading.
type TemperatureReading struct {
	ID     int      `json:"id"`
	TC     *float64 `json:"tC"`
	TF     *float64 `json:"tF"`
	Errors []string `json:"errors,omitempty"`
}

// HumidityReading represents a humidity sensor reading.
type HumidityReading struct {
	ID     int      `json:"id"`
	RH     *float64 `json:"rh"`
	Errors []string `json:"errors,omitempty"`
}

// IlluminanceReading represents an illuminance sensor reading.
type IlluminanceReading struct {
	ID     int      `json:"id"`
	Lux    *float64 `json:"lux"`
	Errors []string `json:"errors,omitempty"`
}

// VoltmeterReading represents a voltmeter sensor reading.
type VoltmeterReading struct {
	ID      int      `json:"id"`
	Voltage *float64 `json:"voltage"`
	Errors  []string `json:"errors,omitempty"`
}

// DevicePowerReading represents a device power (battery) sensor reading.
type DevicePowerReading struct {
	ID       int                       `json:"id"`
	Battery  DevicePowerBatteryStatus  `json:"battery"`
	External DevicePowerExternalStatus `json:"external"`
	Errors   []string                  `json:"errors,omitempty"`
}

// DevicePowerBatteryStatus represents battery information.
type DevicePowerBatteryStatus struct {
	V       float64 `json:"V"`
	Percent int     `json:"percent"`
}

// DevicePowerExternalStatus represents external power source status.
type DevicePowerExternalStatus struct {
	Present bool `json:"present"`
}

// SensorData holds aggregated sensor readings from a device.
type SensorData struct {
	Temperature []TemperatureReading `json:"temperature,omitempty"`
	Humidity    []HumidityReading    `json:"humidity,omitempty"`
	Flood       []AlarmSensorReading `json:"flood,omitempty"`
	Smoke       []AlarmSensorReading `json:"smoke,omitempty"`
	Illuminance []IlluminanceReading `json:"illuminance,omitempty"`
	Voltmeter   []VoltmeterReading   `json:"voltmeter,omitempty"`
}
