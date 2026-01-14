// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/component"
)

// CoverAction represents a cover operation type.
type CoverAction string

// Cover action constants.
const (
	CoverActionOpen  CoverAction = "open"
	CoverActionClose CoverAction = "close"
	CoverActionStop  CoverAction = "stop"
)

// CoverServiceFunc is a function that performs a cover action.
// duration is nil for stop action or when --duration not specified.
type CoverServiceFunc func(ctx context.Context, svc *shelly.Service, device string, id int, duration *int) error

// CoverOpts configures a cover command.
type CoverOpts struct {
	// Action is the operation type: CoverActionOpen, CoverActionClose, or CoverActionStop.
	Action CoverAction

	// ServiceFunc is called to perform the action.
	ServiceFunc CoverServiceFunc
}

// NewCoverCommand creates a cover open/close/stop command.
func NewCoverCommand(f *cmdutil.Factory, opts CoverOpts) *cobra.Command {
	var (
		coverFlags flags.ComponentNameFlags
		duration   int
	)

	actionStr := string(opts.Action)
	hasDuration := opts.Action != CoverActionStop

	// Build command metadata based on action type
	var (
		use      string
		aliases  []string
		short    string
		long     string
		examples string
	)

	switch opts.Action {
	case CoverActionOpen:
		use = "open <device>"
		aliases = []string{"up", "raise"}
		short = "Open cover"
		long = "Open a cover/roller component on the specified device."
		examples = `  # Open cover fully
  shelly cover open bedroom

  # Open cover for 5 seconds
  shelly cover up bedroom --duration 5

  # Open cover by name
  shelly cover open bedroom --name "Living Room Blinds"`

	case CoverActionClose:
		use = "close <device>"
		aliases = []string{"down", "lower"}
		short = "Close cover"
		long = "Close a cover/roller component on the specified device."
		examples = `  # Close cover fully
  shelly cover close bedroom

  # Close cover for 5 seconds
  shelly cover down bedroom --duration 5

  # Close cover by name
  shelly cover close bedroom --name "Living Room Blinds"`

	case CoverActionStop:
		use = "stop <device>"
		aliases = []string{"halt", "pause"}
		short = "Stop cover"
		long = "Stop a cover/roller component on the specified device."
		examples = `  # Stop cover movement
  shelly cover stop bedroom

  # Stop specific cover by ID
  shelly cover halt bedroom --id 1

  # Stop cover by name
  shelly cover stop bedroom --name "Living Room Blinds"`
	}

	cmd := &cobra.Command{
		Use:               use,
		Aliases:           aliases,
		Short:             short,
		Long:              long,
		Example:           examples,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCover(cmd.Context(), f, opts, args[0], &coverFlags, duration, hasDuration)
		},
	}

	flags.AddComponentNameFlags(cmd, &coverFlags, "Cover")
	if hasDuration {
		cmd.Flags().IntVarP(&duration, "duration", "d", 0, fmt.Sprintf("Duration in seconds (0 = full %s)", actionStr))
	}

	return cmd
}

func runCover(ctx context.Context, f *cmdutil.Factory, opts CoverOpts, device string, coverFlags *flags.ComponentNameFlags, duration int, hasDuration bool) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Resolve cover name to ID if provided
	coverID := coverFlags.ID
	if coverFlags.HasName() {
		configFetcher := makeConfigFetcher(svc)
		resolvedID, err := component.ResolveIDWithGen(ctx, configFetcher, device, component.TypeCover, coverFlags.ID, coverFlags.Name)
		if err != nil {
			return fmt.Errorf("failed to resolve cover name %q: %w", coverFlags.Name, err)
		}
		coverID = resolvedID
	}

	actionStr := string(opts.Action)
	var spinnerMsg string
	switch opts.Action {
	case CoverActionOpen:
		spinnerMsg = "Opening cover..."
	case CoverActionClose:
		spinnerMsg = "Closing cover..."
	case CoverActionStop:
		spinnerMsg = "Stopping cover..."
	}

	return cmdutil.RunWithSpinner(ctx, ios, spinnerMsg, func(ctx context.Context) error {
		var dur *int
		if hasDuration && duration > 0 {
			dur = &duration
		}

		if err := opts.ServiceFunc(ctx, svc, device, coverID, dur); err != nil {
			return fmt.Errorf("failed to %s cover: %w", actionStr, err)
		}

		coverDisplayName := formatComponentDisplayName("Cover", coverID, coverFlags.Name)
		switch opts.Action {
		case CoverActionStop:
			ios.Success("%s stopped", coverDisplayName)
		default:
			ios.Success("%s %sing", coverDisplayName, actionStr)
		}
		return nil
	})
}
