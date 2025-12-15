// Package test provides the alert test subcommand.
package test

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the alert test command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "test <name>",
		Aliases: []string{"trigger", "fire"},
		Short:   "Test an alert by triggering it",
		Long: `Test an alert by manually triggering its action.

This simulates the alert condition being met and executes the configured action.`,
		Example: `  # Test an alert
  shelly alert test kitchen-offline`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, name string) error {
	ios := f.IOStreams()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	alert, exists := cfg.Alerts[name]
	if !exists {
		return fmt.Errorf("alert %q not found", name)
	}

	ios.Info("Testing alert %q...", name)
	ios.Printf("  Device: %s\n", alert.Device)
	ios.Printf("  Condition: %s\n", alert.Condition)
	ios.Printf("  Action: %s\n", alert.Action)
	ios.Println("")

	// Execute the action
	switch {
	case alert.Action == "notify":
		// Desktop notification - just print for now
		ios.Success("[TEST] Alert triggered: %s on %s", alert.Condition, alert.Device)
	case len(alert.Action) > 8 && alert.Action[:8] == "webhook:":
		url := alert.Action[8:]
		ios.Info("[TEST] Would send webhook to: %s", url)
	case len(alert.Action) > 8 && alert.Action[:8] == "command:":
		cmd := alert.Action[8:]
		ios.Info("[TEST] Would execute command: %s", cmd)
	default:
		ios.Warning("Unknown action type: %s", alert.Action)
	}

	ios.Println("")
	ios.Success("Alert test completed")

	return nil
}
