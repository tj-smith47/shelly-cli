// Package list provides the rgb list subcommand.
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

// NewCommand creates the rgb list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List RGB components",
		Long:  `List all RGB light components on the specified device with their current status.`,
		Args:  cobra.ExactArgs(1),
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
		"Fetching RGB components...",
		"RGB components",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.RGBInfo, error) {
			return svc.RGBList(ctx, device)
		},
		displayList)
}

func displayList(ios *iostreams.IOStreams, rgbs []shelly.RGBInfo) {
	t := output.NewTable("ID", "Name", "State", "Color", "Brightness", "Power")
	for _, rgb := range rgbs {
		name := rgb.Name
		if name == "" {
			name = fmt.Sprintf("rgb:%d", rgb.ID)
		}

		state := theme.StatusError().Render("OFF")
		if rgb.Output {
			state = theme.StatusOK().Render("ON")
		}

		color := fmt.Sprintf("R:%d G:%d B:%d", rgb.Red, rgb.Green, rgb.Blue)

		brightness := "-"
		if rgb.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", rgb.Brightness)
		}

		power := "-"
		if rgb.Power > 0 {
			power = fmt.Sprintf("%.1f W", rgb.Power)
		}

		t.AddRow(fmt.Sprintf("%d", rgb.ID), name, state, color, brightness, power)
	}
	t.Print()
}
