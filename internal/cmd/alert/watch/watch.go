// Package watch provides the alert watch subcommand for monitoring alerts.
package watch

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Interval time.Duration
	Once     bool
}

// NewCommand creates the alert watch command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "watch",
		Aliases: []string{"monitor", "daemon", "run"},
		Short:   "Monitor alerts in real-time",
		Long: `Monitor configured alerts and trigger actions when conditions are met.

This command runs continuously, polling device status at the specified interval
and executing alert actions when conditions are triggered.

Conditions supported:
  - offline: Device becomes unreachable
  - online: Device becomes reachable
  - power>N: Power consumption exceeds N watts
  - power<N: Power consumption below N watts
  - temperature>N: Temperature exceeds N degrees
  - temperature<N: Temperature below N degrees

Actions supported:
  - notify: Print to console (default)
  - webhook:URL: Send HTTP POST to URL with alert JSON
  - command:CMD: Execute shell command`,
		Example: `  # Monitor alerts every 30 seconds
  shelly alert watch

  # Monitor with custom interval
  shelly alert watch --interval 1m

  # Run once and exit (for cron)
  shelly alert watch --once`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Interval, "interval", "i", 30*time.Second, "Check interval")
	cmd.Flags().BoolVar(&opts.Once, "once", false, "Run once and exit (for cron/scheduled tasks)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if len(cfg.Alerts) == 0 {
		ios.Warning("No alerts configured")
		ios.Info("Create alerts with: shelly alert create <name> --device <device> --condition <condition>")
		return nil
	}

	enabledCount := 0
	for _, alert := range cfg.Alerts {
		if alert.Enabled && !alert.IsSnoozed() {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		ios.Warning("No enabled alerts (all disabled or snoozed)")
		return nil
	}

	ios.Success("Alert monitor started")
	ios.Printf("  Monitoring %d alert(s) every %s\n", enabledCount, opts.Interval)
	ios.Printf("  Press Ctrl+C to stop\n")
	ios.Println("")

	states := make(map[string]*shelly.AlertState)
	for name := range cfg.Alerts {
		states[name] = &shelly.AlertState{}
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	// Run immediately on start
	for _, result := range svc.CheckAlerts(ctx, cfg.Alerts, states) {
		alert := cfg.Alerts[result.Name]
		term.DisplayAlertResult(ctx, ios, alert, result)
	}

	if opts.Once {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			ios.Println("")
			ios.Info("Alert monitor stopped")
			return nil
		case <-ticker.C:
			cfg, err = f.Config()
			if err != nil {
				ios.DebugErr("reload config", err)
				continue
			}
			for _, result := range svc.CheckAlerts(ctx, cfg.Alerts, states) {
				alert := cfg.Alerts[result.Name]
				term.DisplayAlertResult(ctx, ios, alert, result)
			}
		}
	}
}
