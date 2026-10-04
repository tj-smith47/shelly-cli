// Package list provides the alert list subcommand.
package list

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the alert list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "show"},
		Short:   "List configured alerts",
		Long:    `List all configured monitoring alerts.`,
		Example: `  # List all alerts
  shelly alert list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	cfg, err := opts.Factory.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	alerts := make([]config.Alert, 0, len(cfg.Alerts))
	for name, alert := range cfg.Alerts {
		if alert.Name == "" {
			alert.Name = name
		}
		alerts = append(alerts, alert)
	}
	sort.Slice(alerts, func(i, j int) bool { return alerts[i].Name < alerts[j].Name })

	return cmdutil.PrintList(ios, alerts, func(ios *iostreams.IOStreams, alerts []config.Alert) {
		ios.Success("Configured Alerts (%d)", len(alerts))
		ios.Println("")

		for _, alert := range alerts {
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

			ios.Printf("  %s [%s]\n", alert.Name, status)
			ios.Printf("    Device: %s\n", alert.Device)
			ios.Printf("    Condition: %s\n", alert.Condition)
			ios.Printf("    Action: %s\n", alert.Action)
			if alert.Description != "" {
				ios.Printf("    Description: %s\n", alert.Description)
			}
			ios.Println("")
		}
	}, func() {
		ios.Info("No alerts configured")
		ios.Println("")
		ios.Info("Create one with: shelly alert create <name> --device <device> --condition <condition>")
	})
}
