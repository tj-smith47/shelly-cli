// Package snooze provides the alert snooze subcommand.
package snooze

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Name     string
	Duration string
	Clear    bool
}

// NewCommand creates the alert snooze command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "snooze <name>",
		Aliases: []string{"mute", "silence"},
		Short:   "Snooze an alert temporarily",
		Long: `Snooze an alert for a specified duration.

While snoozed, the alert will not trigger even if the condition is met.`,
		Example: `  # Snooze for 1 hour
  shelly alert snooze kitchen-offline --duration 1h

  # Snooze for 30 minutes
  shelly alert snooze kitchen-offline --duration 30m

  # Clear snooze
  shelly alert snooze kitchen-offline --clear`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Duration, "duration", "d", "1h", "Snooze duration (e.g., 30m, 1h, 2h)")
	cmd.Flags().BoolVar(&opts.Clear, "clear", false, "Clear existing snooze")

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	cfg, err := opts.Factory.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	alert, exists := cfg.Alerts[opts.Name]
	if !exists {
		return fmt.Errorf("alert %q not found", opts.Name)
	}

	if opts.Clear {
		alert.SnoozedUntil = ""
		cfg.Alerts[opts.Name] = alert
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		ios.Success("Cleared snooze for alert %q", opts.Name)
		return nil
	}

	duration, err := time.ParseDuration(opts.Duration)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", opts.Duration, err)
	}

	snoozedUntil := time.Now().Add(duration)
	alert.SnoozedUntil = snoozedUntil.Format(time.RFC3339)
	cfg.Alerts[opts.Name] = alert

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	ios.Success("Snoozed alert %q until %s", opts.Name, snoozedUntil.Format("15:04"))

	return nil
}
