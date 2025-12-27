// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// StatusFetcher is a function that fetches component status from a device.
type StatusFetcher[T any] func(ctx context.Context, svc *shelly.Service, device string, id int) (T, error)

// StatusDisplay is a function that displays status in human-readable format.
type StatusDisplay[T any] func(ios *iostreams.IOStreams, status T)

// StatusOpts configures a status command.
type StatusOpts[T any] struct {
	// Component is the display name (e.g., "Switch", "Light", "Cover").
	Component string

	// Aliases are the command aliases (e.g., []string{"st", "s"}).
	Aliases []string

	// SpinnerMsg is the message shown while fetching status.
	// If empty, defaults to "Fetching {component} status...".
	SpinnerMsg string

	// Fetcher retrieves status from the device.
	Fetcher StatusFetcher[T]

	// Display renders the status in human-readable format.
	Display StatusDisplay[T]
}

// NewStatusCommand creates a component status command.
// This factory consolidates the common pattern across Switch, Light, Cover, etc. status commands.
func NewStatusCommand[T any](f *cmdutil.Factory, opts StatusOpts[T]) *cobra.Command {
	var componentID int

	componentLower := strings.ToLower(opts.Component)

	// Generate descriptions and examples
	short := fmt.Sprintf("Show %s status", componentLower)
	long := fmt.Sprintf("Show the current status of a %s component on the specified device.", componentLower)
	examples := fmt.Sprintf(`  # Show %s status
  shelly %s status <device>

  # Show status with JSON output
  shelly %s st <device> -o json`, componentLower, componentLower, componentLower)

	// Default spinner message
	spinnerMsg := opts.SpinnerMsg
	if spinnerMsg == "" {
		spinnerMsg = fmt.Sprintf("Fetching %s status...", componentLower)
	}

	cmd := &cobra.Command{
		Use:               "status <device>",
		Aliases:           opts.Aliases,
		Short:             short,
		Long:              long,
		Example:           examples,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd.Context(), f, opts, spinnerMsg, args[0], componentID)
		},
	}

	flags.AddComponentIDFlag(cmd, &componentID, opts.Component)

	return cmd
}

func runStatus[T any](
	ctx context.Context,
	f *cmdutil.Factory,
	opts StatusOpts[T],
	spinnerMsg string,
	device string,
	componentID int,
) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunStatus(ctx, ios, svc, device, componentID, spinnerMsg,
		cmdutil.StatusFetcher[T](opts.Fetcher),
		cmdutil.StatusDisplay[T](opts.Display))
}
