// Package list provides the switch list subcommand.
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

// NewCommand creates the switch list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List switch components",
		Long:  `List all switch components on the specified device with their current status.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*shelly.DefaultTimeout/10) // 15s
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunList(ctx, ios, svc, device,
		"Fetching switch components...",
		"switch components",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.SwitchInfo, error) {
			return svc.SwitchList(ctx, device)
		},
		displayList)
}

func displayList(ios *iostreams.IOStreams, switches []shelly.SwitchInfo) {
	t := output.NewTable("ID", "Name", "State", "Power")
	for _, sw := range switches {
		name := sw.Name
		if name == "" {
			name = fmt.Sprintf("switch:%d", sw.ID)
		}

		state := theme.StatusError().Render("OFF")
		if sw.Output {
			state = theme.StatusOK().Render("ON")
		}

		power := "-"
		if sw.Power > 0 {
			power = fmt.Sprintf("%.1f W", sw.Power)
		}

		t.AddRow(fmt.Sprintf("%d", sw.ID), name, state, power)
	}
	t.Print()
}
