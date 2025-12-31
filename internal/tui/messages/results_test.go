package messages

import (
	"errors"
	"testing"
)

func TestNewSaveResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		componentID any
	}{
		{"int ID", 42},
		{"string ID", "my-key"},
		{"nil ID", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg := NewSaveResult(tt.componentID)

			if msg.ComponentID != tt.componentID {
				t.Errorf("expected ComponentID %v, got %v", tt.componentID, msg.ComponentID)
			}
			if !msg.Success {
				t.Error("expected Success to be true")
			}
			if msg.Err != nil {
				t.Errorf("expected Err to be nil, got %v", msg.Err)
			}
		})
	}
}

func TestNewSaveError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")

	tests := []struct {
		name        string
		componentID any
		err         error
	}{
		{"int ID with error", 42, testErr},
		{"string ID with error", "my-key", testErr},
		{"nil error", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg := NewSaveError(tt.componentID, tt.err)

			if msg.ComponentID != tt.componentID {
				t.Errorf("expected ComponentID %v, got %v", tt.componentID, msg.ComponentID)
			}
			if msg.Success {
				t.Error("expected Success to be false")
			}
			// Check error identity - both nil or both same error
			if (msg.Err == nil) != (tt.err == nil) {
				t.Errorf("expected Err %v, got %v", tt.err, msg.Err)
			}
			if tt.err != nil && !errors.Is(msg.Err, tt.err) {
				t.Errorf("expected Err %v, got %v", tt.err, msg.Err)
			}
		})
	}
}

func TestEditClosedMsg(t *testing.T) {
	t.Parallel()

	t.Run("saved true", func(t *testing.T) {
		t.Parallel()
		msg := EditClosedMsg{Saved: true}
		if !msg.Saved {
			t.Error("expected Saved to be true")
		}
	})

	t.Run("saved false", func(t *testing.T) {
		t.Parallel()
		msg := EditClosedMsg{Saved: false}
		if msg.Saved {
			t.Error("expected Saved to be false")
		}
	})
}

func TestEditOpenedMsg(t *testing.T) {
	t.Parallel()

	// Just ensure the type can be created
	_ = EditOpenedMsg{}
}
