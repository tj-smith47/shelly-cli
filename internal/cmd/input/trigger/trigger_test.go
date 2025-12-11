// Package trigger provides the input trigger subcommand.
package trigger

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}

	if cmd.Use != "trigger <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "trigger <device>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	// Test id flag exists
	idFlag := cmd.Flags().Lookup("id")
	switch {
	case idFlag == nil:
		t.Error("id flag not found")
	case idFlag.Shorthand != "i":
		t.Errorf("id shorthand = %q, want %q", idFlag.Shorthand, "i")
	case idFlag.DefValue != "0":
		t.Errorf("id default = %q, want %q", idFlag.DefValue, "0")
	}

	// Test event flag exists
	eventFlag := cmd.Flags().Lookup("event")
	switch {
	case eventFlag == nil:
		t.Error("event flag not found")
	case eventFlag.Shorthand != "e":
		t.Errorf("event shorthand = %q, want %q", eventFlag.Shorthand, "e")
	case eventFlag.DefValue != EventSinglePush:
		t.Errorf("event default = %q, want %q", eventFlag.DefValue, EventSinglePush)
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	// The command should require exactly 1 argument
	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}

func TestEventTypes(t *testing.T) {
	t.Parallel()

	// Verify event type constants are defined correctly
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"EventSinglePush", EventSinglePush, "single_push"},
		{"EventDoublePush", EventDoublePush, "double_push"},
		{"EventLongPush", EventLongPush, "long_push"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.value != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.value, tt.want)
			}
		})
	}
}
