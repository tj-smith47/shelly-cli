// Package wait provides the wait command for pausing execution.
package wait

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the wait command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "wait <duration>",
		Aliases: []string{"delay", "pause"},
		Short:   "Wait for a duration",
		Long: `Wait for a specified duration before continuing.

Useful for:
  - Adding delays between commands in scripts
  - Waiting for devices to initialize
  - Creating sequenced automation

The duration can be specified in common formats:
  - Seconds: 30s, 45s
  - Minutes: 5m, 10m
  - Hours: 1h, 2h
  - Combined: 1h30m, 5m30s

Press Ctrl+C to cancel the wait early.`,
		Example: `  # Wait 5 seconds
  shelly wait 5s

  # Wait 2 minutes
  shelly wait 2m

  # Wait 1 hour
  shelly wait 1h

  # Use in a script
  shelly on kitchen && shelly wait 5s && shelly off kitchen`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, durationStr string) error {
	ios := f.IOStreams()

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return err
	}

	ios.Info("Waiting %v...", duration)

	select {
	case <-ctx.Done():
		ios.Println("")
		ios.Warning("Wait cancelled")
		return nil
	case <-time.After(duration):
		ios.Success("Done")
		return nil
	}
}
