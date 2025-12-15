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
	Duration string
	Clear    bool
}

// NewCommand creates the alert snooze command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

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
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Duration, "duration", "d", "1h", "Snooze duration (e.g., 30m, 1h, 2h)")
	cmd.Flags().BoolVar(&opts.Clear, "clear", false, "Clear existing snooze")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, name string, opts *Options) error {
	ios := f.IOStreams()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	alert, exists := cfg.Alerts[name]
	if !exists {
		return fmt.Errorf("alert %q not found", name)
	}

	if opts.Clear {
		alert.SnoozedUntil = ""
		cfg.Alerts[name] = alert
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		ios.Success("Cleared snooze for alert %q", name)
		return nil
	}

	duration, err := time.ParseDuration(opts.Duration)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", opts.Duration, err)
	}

	snoozedUntil := time.Now().Add(duration)
	alert.SnoozedUntil = snoozedUntil.Format(time.RFC3339)
	cfg.Alerts[name] = alert

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	ios.Success("Snoozed alert %q until %s", name, snoozedUntil.Format("15:04"))

	return nil
}
