// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"github.com/tj-smith47/shelly-go/gen1"
)

// IsButtonEvent returns true if the action event is a button-related event.
// Button events require physical button interaction and cannot be triggered
// programmatically via the API.
func IsButtonEvent(e gen1.ActionEvent) bool {
	switch e { //nolint:exhaustive // only checking button events
	case gen1.ActionLongpush, gen1.ActionShortpush, gen1.ActionDoublepush, gen1.ActionTriplepush,
		gen1.ActionBtn1On, gen1.ActionBtn1Off, gen1.ActionBtn2On, gen1.ActionBtn2Off:
		return true
	}
	return false
}

// IsRollerEvent returns true if the action event is a roller/cover-related event.
// Roller events can be triggered via the cover commands.
func IsRollerEvent(e gen1.ActionEvent) bool {
	switch e { //nolint:exhaustive // only checking roller events
	case gen1.ActionRollerOpen, gen1.ActionRollerClose, gen1.ActionRollerStop,
		gen1.ActionRollerOpenUrl, gen1.ActionRollerCloseUrl, gen1.ActionRollerStopUrl:
		return true
	}
	return false
}

// IsSensorEvent returns true if the action event is a sensor-related event.
// Sensor events are triggered by environmental conditions.
func IsSensorEvent(e gen1.ActionEvent) bool {
	switch e { //nolint:exhaustive // only checking sensor events
	case gen1.ActionSensorOpen, gen1.ActionSensorClose, gen1.ActionSensorMotion,
		gen1.ActionSensorNoMotion, gen1.ActionSensorFlood, gen1.ActionSensorNoFlood,
		gen1.ActionSensorSmoke, gen1.ActionSensorNoSmoke, gen1.ActionSensorGas,
		gen1.ActionSensorNoGas, gen1.ActionSensorVibration, gen1.ActionSensorTemp,
		gen1.ActionSensorTempUnder:
		return true
	}
	return false
}
