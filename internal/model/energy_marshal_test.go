package model

import (
	"encoding/json"
	"testing"
)

func TestPMStatus_MarshalJSON_OmitsUnmeasuredVoltageAndCurrent(t *testing.T) {
	t.Parallel()

	gen1, err := json.Marshal(PMStatus{ID: 0, APower: 14.2})
	if err != nil {
		t.Fatal(err)
	}
	if string(gen1) != `{"id":0,"apower":14.2}` {
		t.Errorf("no voltage: %s", gen1)
	}

	idle, err := json.Marshal(&PMStatus{ID: 0, Voltage: 120.1})
	if err != nil {
		t.Fatal(err)
	}
	if string(idle) != `{"id":0,"voltage":120.1,"current":0,"apower":0}` {
		t.Errorf("idle Gen2 meter: %s", idle)
	}
}
