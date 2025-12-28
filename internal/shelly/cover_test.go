package shelly

import (
	"testing"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"
)

func TestCoverInfo_Fields(t *testing.T) {
	t.Parallel()

	info := CoverInfo{
		ID:       0,
		Name:     "Living Room Blinds",
		State:    "stopped",
		Position: 50,
		Power:    10.5,
	}

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "Living Room Blinds" {
		t.Errorf("Name = %q, want %q", info.Name, "Living Room Blinds")
	}
	if info.State != "stopped" {
		t.Errorf("State = %q, want %q", info.State, "stopped")
	}
	if info.Position != 50 {
		t.Errorf("Position = %d, want 50", info.Position)
	}
	if info.Power != 10.5 {
		t.Errorf("Power = %f, want 10.5", info.Power)
	}
}

func TestCoverInfo_ZeroValues(t *testing.T) {
	t.Parallel()

	var info CoverInfo

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("Name = %q, want empty", info.Name)
	}
	if info.State != "" {
		t.Errorf("State = %q, want empty", info.State)
	}
	if info.Position != 0 {
		t.Errorf("Position = %d, want 0", info.Position)
	}
	if info.Power != 0 {
		t.Errorf("Power = %f, want 0", info.Power)
	}
}

func TestCoverInfo_States(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    string
		expected string
	}{
		{"stopped state", "stopped", "stopped"},
		{"open state", "open", "open"},
		{"closed state", "closed", "closed"},
		{"opening state", "opening", "opening"},
		{"closing state", "closing", "closing"},
		{"calibrating state", "calibrating", "calibrating"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info := CoverInfo{State: tt.state}
			if info.State != tt.expected {
				t.Errorf("State = %q, want %q", info.State, tt.expected)
			}
		})
	}
}

func TestGen1RollerStatusToCover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		id           int
		status       *gen1comp.RollerStatus
		wantState    string
		wantPosition *int
	}{
		{
			name: "stopped at position 50",
			id:   0,
			status: &gen1comp.RollerStatus{
				State:      "stop",
				CurrentPos: 50,
				IsValid:    true,
				Power:      0,
			},
			wantState:    "stop",
			wantPosition: intPtr(50),
		},
		{
			name: "opening",
			id:   0,
			status: &gen1comp.RollerStatus{
				State:      "open",
				CurrentPos: 75,
				IsValid:    true,
				Power:      15.0,
			},
			wantState:    "open",
			wantPosition: intPtr(75),
		},
		{
			name: "invalid position",
			id:   1,
			status: &gen1comp.RollerStatus{
				State:      "stop",
				CurrentPos: -1,
				IsValid:    false,
			},
			wantState:    "stop",
			wantPosition: nil,
		},
		{
			name: "calibrating",
			id:   0,
			status: &gen1comp.RollerStatus{
				State:       "stop",
				Calibrating: true,
				CurrentPos:  0,
				IsValid:     true,
			},
			wantState:    "stop",
			wantPosition: nil, // Position -1 invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := gen1RollerStatusToCover(tt.id, tt.status)

			if result.ID != tt.id {
				t.Errorf("ID = %d, want %d", result.ID, tt.id)
			}
			if result.State != tt.wantState {
				t.Errorf("State = %q, want %q", result.State, tt.wantState)
			}
			if result.Calibrating != tt.status.Calibrating {
				t.Errorf("Calibrating = %v, want %v", result.Calibrating, tt.status.Calibrating)
			}
		})
	}
}

// intPtr is a helper to create int pointers in tests.
func intPtr(i int) *int {
	return &i
}
