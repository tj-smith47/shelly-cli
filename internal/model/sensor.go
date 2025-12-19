// Package model defines core domain types for the Shelly CLI.
package model

// Sensor is the base interface for all sensor types.
// Used with Go generics for type-safe, DRY display functions.
type Sensor interface {
	GetID() int
	GetErrors() []string
}

// AlarmSensor defines the interface for alarm-type sensors (smoke, flood).
// Used with Go generics for type-safe, DRY display functions.
type AlarmSensor interface {
	Sensor
	IsAlarm() bool
	IsMuted() bool
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

// GetID implements Sensor.
func (t TemperatureReading) GetID() int { return t.ID }

// GetErrors implements Sensor.
func (t TemperatureReading) GetErrors() []string { return t.Errors }

// HumidityReading represents a humidity sensor reading.
type HumidityReading struct {
	ID     int      `json:"id"`
	RH     *float64 `json:"rh"`
	Errors []string `json:"errors,omitempty"`
}

// GetID implements Sensor.
func (h HumidityReading) GetID() int { return h.ID }

// GetErrors implements Sensor.
func (h HumidityReading) GetErrors() []string { return h.Errors }

// IlluminanceReading represents an illuminance sensor reading.
type IlluminanceReading struct {
	ID     int      `json:"id"`
	Lux    *float64 `json:"lux"`
	Errors []string `json:"errors,omitempty"`
}

// GetID implements Sensor.
func (i IlluminanceReading) GetID() int { return i.ID }

// GetErrors implements Sensor.
func (i IlluminanceReading) GetErrors() []string { return i.Errors }

// VoltmeterReading represents a voltmeter sensor reading.
type VoltmeterReading struct {
	ID      int      `json:"id"`
	Voltage *float64 `json:"voltage"`
	Errors  []string `json:"errors,omitempty"`
}

// GetID implements Sensor.
func (v VoltmeterReading) GetID() int { return v.ID }

// GetErrors implements Sensor.
func (v VoltmeterReading) GetErrors() []string { return v.Errors }

// DevicePowerReading represents a device power (battery) sensor reading.
type DevicePowerReading struct {
	ID       int                       `json:"id"`
	Battery  DevicePowerBatteryStatus  `json:"battery"`
	External DevicePowerExternalStatus `json:"external"`
	Errors   []string                  `json:"errors,omitempty"`
}

// GetID implements Sensor.
func (d DevicePowerReading) GetID() int { return d.ID }

// GetErrors implements Sensor.
func (d DevicePowerReading) GetErrors() []string { return d.Errors }

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
