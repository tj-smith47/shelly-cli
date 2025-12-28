package shelly

import (
	"testing"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"
)

func TestSwitchInfo_Fields(t *testing.T) {
	t.Parallel()

	info := SwitchInfo{
		ID:     0,
		Name:   "Living Room Switch",
		Output: true,
		Power:  100.5,
	}

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "Living Room Switch" {
		t.Errorf("Name = %q, want %q", info.Name, "Living Room Switch")
	}
	if !info.Output {
		t.Error("Output = false, want true")
	}
	if info.Power != 100.5 {
		t.Errorf("Power = %f, want 100.5", info.Power)
	}
}

func TestSwitchInfo_ZeroValues(t *testing.T) {
	t.Parallel()

	var info SwitchInfo

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("Name = %q, want empty", info.Name)
	}
	if info.Output {
		t.Error("Output = true, want false")
	}
	if info.Power != 0 {
		t.Errorf("Power = %f, want 0", info.Power)
	}
}

func TestGen1RelayStatusToSwitch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		id     int
		status *gen1comp.RelayStatus
		wantOn bool
	}{
		{
			name: "relay on",
			id:   0,
			status: &gen1comp.RelayStatus{
				IsOn:      true,
				Source:    "web",
				Overpower: false,
			},
			wantOn: true,
		},
		{
			name: "relay off",
			id:   1,
			status: &gen1comp.RelayStatus{
				IsOn:      false,
				Source:    "switch",
				Overpower: false,
			},
			wantOn: false,
		},
		{
			name: "relay with overpower",
			id:   2,
			status: &gen1comp.RelayStatus{
				IsOn:      false,
				Source:    "overpower",
				Overpower: true,
			},
			wantOn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := gen1RelayStatusToSwitch(tt.id, tt.status)

			if result.ID != tt.id {
				t.Errorf("ID = %d, want %d", result.ID, tt.id)
			}
			if result.Output != tt.wantOn {
				t.Errorf("Output = %v, want %v", result.Output, tt.wantOn)
			}
			if result.Source != tt.status.Source {
				t.Errorf("Source = %q, want %q", result.Source, tt.status.Source)
			}
			if result.Overpower != tt.status.Overpower {
				t.Errorf("Overpower = %v, want %v", result.Overpower, tt.status.Overpower)
			}
		})
	}
}
