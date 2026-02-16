// Package enable provides the thermostat schedule enable command.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// NewCommand creates the thermostat schedule enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var scheduleID int

	cmd := factories.NewEnableDisableCommand(f, factories.EnableDisableOpts{
		Feature: "schedule",
		Enable:  true,
		Aliases: []string{"on"},
		Long:    `Enable a disabled schedule so it will run at its scheduled times.`,
		Example: `  # Enable schedule by ID
  shelly thermostat schedule enable gateway --id 1`,
		ServiceFunc: func(ctx context.Context, f *cmdutil.Factory, device string) error {
			svc := f.ShellyService()
			return svc.WithDevice(ctx, device, func(dev *shelly.DeviceClient) error {
				if dev.IsGen1() {
					return fmt.Errorf("thermostat component requires Gen2+ device")
				}
				params := map[string]any{
					"id":     scheduleID,
					"enable": true,
				}
				_, err := dev.Gen2().Call(ctx, "Schedule.Update", params)
				return err
			})
		},
	})

	cmd.Flags().IntVar(&scheduleID, "id", 0, "Schedule ID to enable (required)")
	utils.Must(cmd.MarkFlagRequired("id"))

	return cmd
}
