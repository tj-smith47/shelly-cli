// Package alert provides alert management commands.
package alert

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/alert/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alert/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alert/snooze"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alert/test"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alert/watch"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the alert command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "alert",
		Aliases: []string{"alerts", "notify"},
		Short:   "Manage monitoring alerts",
		Long: `Manage monitoring alerts for device conditions.

Alerts notify you when devices meet specified conditions like going offline,
exceeding power thresholds, or temperature changes.`,
		Example: `  # Create an offline alert
  shelly alert create kitchen-offline --device kitchen --condition offline

  # List all alerts
  shelly alert list

  # Test an alert
  shelly alert test kitchen-offline

  # Snooze an alert for 1 hour
  shelly alert snooze kitchen-offline --duration 1h

  # Start monitoring alerts
  shelly alert watch`,
	}

	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(test.NewCommand(f))
	cmd.AddCommand(snooze.NewCommand(f))
	cmd.AddCommand(watch.NewCommand(f))

	return cmd
}
