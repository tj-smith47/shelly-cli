package factoryreset

import (
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"testing"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "factory-reset <device>" {
		t.Errorf("Use = %q, want 'factory-reset <device>'", cmd.Use)
	}
	if len(cmd.Aliases) < 3 {
		t.Errorf("Expected at least 3 aliases, got %d", len(cmd.Aliases))
	}
	// Check all expected aliases
	expectedAliases := []string{"fr", "reset", "wipe"}
	for _, expected := range expectedAliases {
		found := false
		for _, alias := range cmd.Aliases {
			if alias == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected alias %q not found in %v", expected, cmd.Aliases)
		}
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("yes flag not found")
	}
	if yesFlag.Shorthand != "y" {
		t.Errorf("yes shorthand = %q, want y", yesFlag.Shorthand)
	}

	confirmFlag := cmd.Flags().Lookup("confirm")
	if confirmFlag == nil {
		t.Fatal("confirm flag not found")
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}
}
