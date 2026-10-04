// Package wait provides the wait command, which blocks until a device responds
// or its output reaches a given state.
package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// probeTimeout bounds one reachability check so a device that drops packets
// cannot use up the whole wait in a single attempt.
const probeTimeout = 5 * time.Second

const (
	stateOn  = "on"
	stateOff = "off"
)

// Options holds the command options.
type Options struct {
	flags.QuickComponentFlags

	Factory  *cmdutil.Factory
	Device   string
	Online   bool
	State    string
	Timeout  time.Duration
	Interval time.Duration
}

// NewCommand creates the wait command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:  f,
		Online:   true,
		Timeout:  2 * time.Minute,
		Interval: 2 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "wait <device>",
		Aliases: []string{"wait-for", "await"},
		Short:   "Wait until a device is online or its output is on or off",
		Long: `Wait until a device responds, or until its output is on or off, then exit.

The device is checked every --interval until the condition holds or --timeout
passes. The exit code is zero once the condition holds and non-zero on
timeout, so the command can gate the next step of a script.

Without --state the command waits for the device to answer. That is useful
after anything that takes a device off the network for a while:
  - A firmware update
  - A reboot or factory reset
  - A power cycle or a WiFi change

With --state on or --state off the command waits until every switch, light
and RGB output of the device is in that state (or only the one given by --id).
That is useful after a timer, a schedule, a scene or a physical button is
expected to change the output.`,
		Example: `  # Continue once the device is back after a firmware update
  shelly firmware update kitchen --yes && shelly wait kitchen

  # Give a slow device up to five minutes
  shelly wait garage --timeout 5m

  # Check more often
  shelly wait 192.168.1.100 --interval 500ms

  # Use in a script without output
  if shelly wait kitchen --online --timeout 120s -q; then
    shelly on kitchen
  fi

  # Continue once the light has switched off
  shelly wait hallway --state off --timeout 60s

  # Watch one output of a multi-channel device
  shelly wait bathroom --state on --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Online, "online", true, "Wait until the device responds")
	cmd.Flags().StringVar(&opts.State, "state", "", "Wait until the device output is in this state: on, off")
	cmd.Flags().IntVar(&opts.ID, "id", -1, "Component ID to watch with --state (omit to watch all)")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", opts.Timeout, "How long to wait before giving up")
	cmd.Flags().DurationVar(&opts.Interval, "interval", opts.Interval, "Time between checks")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if !opts.Online {
		return fmt.Errorf("--online=false is not supported: use --state to wait for an output state")
	}
	if opts.State != "" && opts.State != stateOn && opts.State != stateOff {
		return fmt.Errorf("invalid --state %q: must be %s or %s", opts.State, stateOn, stateOff)
	}

	goal := "come online"
	reached := "is online"
	if opts.State != "" {
		goal = "turn " + opts.State
		reached = "is " + opts.State
	}
	ios.Info("Waiting up to %v for %s to %s...", opts.Timeout, opts.Device, goal)

	check := func(probeCtx context.Context) error {
		if opts.State == "" {
			_, err := svc.DevicePing(probeCtx, opts.Device)
			return err
		}
		states, err := svc.OutputStates(probeCtx, opts.Device, opts.ComponentIDPointer())
		if err != nil {
			return err
		}
		for _, on := range states {
			if on != (opts.State == stateOn) {
				return fmt.Errorf("an output is still %s", map[bool]string{true: stateOn, false: stateOff}[on])
			}
		}
		return nil
	}

	start := time.Now()
	if err := cmdutil.PollUntil(ctx, opts.Timeout, opts.Interval, probeTimeout, check); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("device %s did not %s within %v: %w", opts.Device, goal, opts.Timeout, err)
	}
	ios.Success("%s %s after %v", opts.Device, reached, time.Since(start).Round(time.Millisecond))
	return nil
}
