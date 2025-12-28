package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-go/gen1"
)

func TestIsButtonEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event gen1.ActionEvent
		want  bool
	}{
		{"longpush", gen1.ActionLongpush, true},
		{"shortpush", gen1.ActionShortpush, true},
		{"doublepush", gen1.ActionDoublepush, true},
		{"triplepush", gen1.ActionTriplepush, true},
		{"btn1_on", gen1.ActionBtn1On, true},
		{"btn1_off", gen1.ActionBtn1Off, true},
		{"btn2_on", gen1.ActionBtn2On, true},
		{"btn2_off", gen1.ActionBtn2Off, true},
		{"roller_open", gen1.ActionRollerOpen, false},
		{"sensor_open", gen1.ActionSensorOpen, false},
		{"output_on", gen1.ActionOutputOn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsButtonEvent(tt.event)
			if got != tt.want {
				t.Errorf("IsButtonEvent(%v) = %v, want %v", tt.event, got, tt.want)
			}
		})
	}
}

func TestIsRollerEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event gen1.ActionEvent
		want  bool
	}{
		{"roller_open", gen1.ActionRollerOpen, true},
		{"roller_close", gen1.ActionRollerClose, true},
		{"roller_stop", gen1.ActionRollerStop, true},
		{"roller_open_url", gen1.ActionRollerOpenUrl, true},
		{"roller_close_url", gen1.ActionRollerCloseUrl, true},
		{"roller_stop_url", gen1.ActionRollerStopUrl, true},
		{"button_push", gen1.ActionShortpush, false},
		{"sensor_open", gen1.ActionSensorOpen, false},
		{"output_on", gen1.ActionOutputOn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsRollerEvent(tt.event)
			if got != tt.want {
				t.Errorf("IsRollerEvent(%v) = %v, want %v", tt.event, got, tt.want)
			}
		})
	}
}

func TestIsSensorEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event gen1.ActionEvent
		want  bool
	}{
		{"sensor_open", gen1.ActionSensorOpen, true},
		{"sensor_close", gen1.ActionSensorClose, true},
		{"sensor_motion", gen1.ActionSensorMotion, true},
		{"sensor_no_motion", gen1.ActionSensorNoMotion, true},
		{"sensor_flood", gen1.ActionSensorFlood, true},
		{"sensor_no_flood", gen1.ActionSensorNoFlood, true},
		{"sensor_smoke", gen1.ActionSensorSmoke, true},
		{"sensor_no_smoke", gen1.ActionSensorNoSmoke, true},
		{"sensor_gas", gen1.ActionSensorGas, true},
		{"sensor_no_gas", gen1.ActionSensorNoGas, true},
		{"sensor_vibration", gen1.ActionSensorVibration, true},
		{"sensor_temp", gen1.ActionSensorTemp, true},
		{"sensor_temp_under", gen1.ActionSensorTempUnder, true},
		{"button_push", gen1.ActionShortpush, false},
		{"roller_open", gen1.ActionRollerOpen, false},
		{"output_on", gen1.ActionOutputOn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsSensorEvent(tt.event)
			if got != tt.want {
				t.Errorf("IsSensorEvent(%v) = %v, want %v", tt.event, got, tt.want)
			}
		})
	}
}
