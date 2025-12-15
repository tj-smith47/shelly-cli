// Package list provides the alert list subcommand.
package list

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the alert list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "show"},
		Short:   "List configured alerts",
		Long:    `List all configured monitoring alerts.`,
		Example: `  # List all alerts
  shelly alert list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f)
		},
	}

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory) error {
	ios := f.IOStreams()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if len(cfg.Alerts) == 0 {
		ios.Info("No alerts configured")
		ios.Println("")
		ios.Info("Create one with: shelly alert create <name> --device <device> --condition <condition>")
		return nil
	}

	ios.Success("Configured Alerts (%d)", len(cfg.Alerts))
	ios.Println("")

	for name, alert := range cfg.Alerts {
		status := "enabled"
		if !alert.Enabled {
			status = "disabled"
		}
		if alert.SnoozedUntil != "" {
			if snoozedUntil, err := time.Parse(time.RFC3339, alert.SnoozedUntil); err == nil {
				if time.Now().Before(snoozedUntil) {
					status = fmt.Sprintf("snoozed until %s", snoozedUntil.Format("15:04"))
				}
			}
		}

		ios.Printf("  %s [%s]\n", name, status)
		ios.Printf("    Device: %s\n", alert.Device)
		ios.Printf("    Condition: %s\n", alert.Condition)
		ios.Printf("    Action: %s\n", alert.Action)
		if alert.Description != "" {
			ios.Printf("    Description: %s\n", alert.Description)
		}
		ios.Println("")
	}

	return nil
}
