// Package list provides the light list subcommand.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the light list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <device>",
		Short:             "List light components",
		Long:              `List all light components on the specified device with their current status.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*shelly.DefaultTimeout/10)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunList(ctx, ios, svc, device,
		"Fetching light components...",
		"light components",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.LightInfo, error) {
			return svc.LightList(ctx, device)
		},
		displayList)
}

func displayList(ios *iostreams.IOStreams, lights []shelly.LightInfo) {
	t := output.NewTable("ID", "Name", "State", "Brightness", "Power")
	for _, lt := range lights {
		name := lt.Name
		if name == "" {
			name = fmt.Sprintf("light:%d", lt.ID)
		}

		state := theme.StatusError().Render("OFF")
		if lt.Output {
			state = theme.StatusOK().Render("ON")
		}

		brightness := "-"
		if lt.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", lt.Brightness)
		}

		power := "-"
		if lt.Power > 0 {
			power = fmt.Sprintf("%.1f W", lt.Power)
		}

		t.AddRow(fmt.Sprintf("%d", lt.ID), name, state, brightness, power)
	}
	t.Print()
}
