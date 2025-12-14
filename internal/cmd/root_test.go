// Package cmd provides the root command and command wiring for the CLI.
package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestRootCommandStructure(t *testing.T) {
	t.Parallel()

	// Test that the root command is properly initialized
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "shelly" {
		t.Errorf("Use = %q, want %q", rootCmd.Use, "shelly")
	}

	if rootCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if rootCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"output", "output", "o", "table"},
		{"template", "template", "", ""},
		{"verbose", "verbose", "v", "false"},
		{"quiet", "quiet", "q", "false"},
		{"config", "config", "", ""},
		{"no-color", "no-color", "", "false"},
	}

	for _, tt := range tests {
		flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
		if flag == nil {
			t.Errorf("%s flag not found", tt.flagName)
			continue
		}
		if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
			t.Errorf("%s shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
		}
		if flag.DefValue != tt.defValue {
			t.Errorf("%s default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
		}
	}
}

func TestRootCommand_Subcommands(t *testing.T) {
	t.Parallel()

	subcommands := rootCmd.Commands()
	expectedSubcommands := map[string]bool{
		"discover": false,
		"switch":   false,
		"cover":    false,
		"light":    false,
		"rgb":      false,
		"input":    false,
		"version":  false,
	}

	for _, sub := range subcommands {
		if _, ok := expectedSubcommands[sub.Use]; ok {
			expectedSubcommands[sub.Use] = true
		}
	}

	for name, found := range expectedSubcommands {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestMust_NilError(t *testing.T) {
	t.Parallel()

	// Should not panic with nil error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("must(nil) panicked: %v", r)
		}
	}()

	must(nil)
}

func TestShouldDisableColor(t *testing.T) {
	// Cannot run in parallel due to env var and viper manipulation

	tests := []struct {
		name       string
		noColorEnv string
		shellyEnv  string
		viperVal   bool
		want       bool
	}{
		{"default", "", "", false, false},
		{"flag set", "", "", true, true},
		{"NO_COLOR set", "1", "", false, true},
		{"SHELLY_NO_COLOR set", "", "1", false, true},
		{"NO_COLOR empty string", "set", "", false, true},
		{"all set", "1", "1", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up env vars using t.Setenv (auto-restores after test)
			// First unset, then use t.Setenv to track
			if err := os.Unsetenv("NO_COLOR"); err != nil {
				t.Logf("warning: failed to unset NO_COLOR: %v", err)
			}
			if err := os.Unsetenv("SHELLY_NO_COLOR"); err != nil {
				t.Logf("warning: failed to unset SHELLY_NO_COLOR: %v", err)
			}
			viper.Set("no-color", false)

			// Set test conditions
			if tt.noColorEnv != "" {
				t.Setenv("NO_COLOR", tt.noColorEnv)
			}
			if tt.shellyEnv != "" {
				t.Setenv("SHELLY_NO_COLOR", tt.shellyEnv)
			}
			viper.Set("no-color", tt.viperVal)

			got := shouldDisableColor()
			if got != tt.want {
				t.Errorf("shouldDisableColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
