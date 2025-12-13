package history

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "history <device> [id]" {
		t.Errorf("Use = %q, want 'history <device> [id]'", cmd.Use)
	}

	expectedAliases := []string{"hist", "events"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
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

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "type", shorthand: "", defValue: shelly.ComponentTypeAuto},
		{name: "period", shorthand: "p", defValue: ""},
		{name: "from", shorthand: "", defValue: ""},
		{name: "to", shorthand: "", defValue: ""},
		{name: "limit", shorthand: "", defValue: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept 1 or 2 arguments
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "0"})
	if err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "0", "extra"})
	if err == nil {
		t.Error("expected error with 3 args")
	}
}

// Note: Tests for CalculateTimeRange and ParseTime are now in internal/shelly/energy_test.go
// since these functions were extracted to the service layer for DRY compliance.

func TestCalculateTotalEnergy(t *testing.T) {
	t.Parallel()

	blocks := []components.EMDataBlock{
		{
			TS:     time.Now().Unix(),
			Period: 60, // 1 minute
			Values: []components.EMDataValues{
				{TotalActivePower: 1000}, // 1000W for 60 seconds
				{TotalActivePower: 2000}, // 2000W for 60 seconds
			},
		},
	}

	// Expected: (1000W * 60s + 2000W * 60s) / 3600 / 1000 = 0.05 kWh
	expected := 0.05
	result := calculateTotalEnergy(blocks)

	if result != expected {
		t.Errorf("calculateTotalEnergy() = %.3f kWh, want %.3f kWh", result, expected)
	}
}

func TestCalculateTotalEnergyEM1(t *testing.T) {
	t.Parallel()

	blocks := []components.EM1DataBlock{
		{
			TS:     time.Now().Unix(),
			Period: 60, // 1 minute
			Values: []components.EM1DataValues{
				{ActivePower: 500},  // 500W for 60 seconds
				{ActivePower: 1500}, // 1500W for 60 seconds
			},
		},
	}

	// Expected: (500W * 60s + 1500W * 60s) / 3600 / 1000 = 0.0333... kWh
	expected := 0.03333333333333333
	result := calculateTotalEnergyEM1(blocks)

	if result != expected {
		t.Errorf("calculateTotalEnergyEM1() = %.10f kWh, want %.10f kWh", result, expected)
	}
}
