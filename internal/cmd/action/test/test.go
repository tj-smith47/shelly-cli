// Package test provides the action test command.
package test

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen1"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Device  string
	Event   string
	Index   int
	Factory *cmdutil.Factory
}

// NewCommand creates the action test command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "test <device> <event>",
		Aliases: []string{"trigger", "fire"},
		Short:   "Test/trigger an action on a Gen1 device",
		Long: `Test (trigger) an action on a Gen1 Shelly device.

This simulates the event that would trigger the action URL, causing
the device to make the configured HTTP request.

Gen1 devices trigger actions based on actual state changes. This command
will temporarily change the device state to trigger the action callback.

For output actions (out_on_url, out_off_url), the device relay will be toggled.
For button actions, the physical button press must be used.

Gen2+ devices use webhooks. See 'shelly webhook test'.`,
		Example: `  # Test output on action (turns relay on, triggering out_on_url)
  shelly action test living-room out_on_url

  # Test output off action
  shelly action test living-room out_off_url

  # Test action on specific relay
  shelly action test relay out_on_url --index 1`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Event = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Index, "index", 0, "Action index (for multi-channel devices)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Check if device is Gen1
	isGen1, _, err := svc.IsGen1Device(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", opts.Device, err)
	}

	if !isGen1 {
		ios.Warning("Device %s is not a Gen1 device", opts.Device)
		ios.Info("Gen2+ devices use webhooks. Try: shelly webhook test %s", opts.Device)
		return fmt.Errorf("action test only available for Gen1 devices")
	}

	event := gen1.ActionEvent(opts.Event)

	// Trigger the action by changing device state
	var actionTaken string
	err = svc.WithGen1Connection(ctx, opts.Device, func(conn *client.Gen1Client) error {
		relay, relayErr := conn.Relay(opts.Index)
		if relayErr != nil {
			return relayErr
		}

		switch event {
		case gen1.ActionOutputOnUrl, gen1.ActionOutputOn:
			actionTaken = "turning relay on"
			return relay.TurnOn(ctx)
		case gen1.ActionOutputOffUrl, gen1.ActionOutputOff:
			actionTaken = "turning relay off"
			return relay.TurnOff(ctx)
		default:
			// For other action types, explain how to trigger
			ios.Warning("Action %s cannot be triggered programmatically", opts.Event)
			ios.Println()
			ios.Info("This action type requires physical interaction:")
			switch {
			case isButtonEvent(event):
				ios.Info("  Press the physical button on the device")
			case isRollerEvent(event):
				ios.Info("  Use 'shelly cover open/close/stop %s'", opts.Device)
			case isSensorEvent(event):
				ios.Info("  Trigger the sensor condition")
			default:
				ios.Info("  Trigger the corresponding device event")
			}
			return nil
		}
	})
	if err != nil {
		return fmt.Errorf("failed to trigger action: %w", err)
	}

	if actionTaken != "" {
		ios.Success("Action %s triggered by %s", opts.Event, actionTaken)
		ios.Info("The configured URL callback should have been executed.")
	}

	return nil
}

func isButtonEvent(e gen1.ActionEvent) bool {
	switch e { //nolint:exhaustive // it be that way sometimes
	case gen1.ActionLongpush, gen1.ActionShortpush, gen1.ActionDoublepush, gen1.ActionTriplepush,
		gen1.ActionBtn1On, gen1.ActionBtn1Off, gen1.ActionBtn2On, gen1.ActionBtn2Off:
		return true
	}
	return false
}

func isRollerEvent(e gen1.ActionEvent) bool {
	switch e { //nolint:exhaustive // it be that way sometimes
	case gen1.ActionRollerOpen, gen1.ActionRollerClose, gen1.ActionRollerStop,
		gen1.ActionRollerOpenUrl, gen1.ActionRollerCloseUrl, gen1.ActionRollerStopUrl:
		return true
	}
	return false
}

func isSensorEvent(e gen1.ActionEvent) bool {
	switch e { //nolint:exhaustive // it be that way sometimes
	case gen1.ActionSensorOpen, gen1.ActionSensorClose, gen1.ActionSensorMotion,
		gen1.ActionSensorNoMotion, gen1.ActionSensorFlood, gen1.ActionSensorNoFlood,
		gen1.ActionSensorSmoke, gen1.ActionSensorNoSmoke, gen1.ActionSensorGas,
		gen1.ActionSensorNoGas, gen1.ActionSensorVibration, gen1.ActionSensorTemp,
		gen1.ActionSensorTempUnder:
		return true
	}
	return false
}
