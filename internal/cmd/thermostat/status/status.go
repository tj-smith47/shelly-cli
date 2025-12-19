// Package status provides the thermostat status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	JSON    bool
}

// NewCommand creates the thermostat status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "show", "info"},
		Short:   "Show thermostat status",
		Long: `Show detailed status of a thermostat component.

Displays:
- Current temperature
- Target temperature
- Valve position (0-100%)
- Operating mode
- Boost/Override status
- Errors and flags`,
		Example: `  # Show thermostat status (default ID 0)
  shelly thermostat status gateway

  # Show specific thermostat
  shelly thermostat status gateway --id 1

  # Output as JSON
  shelly thermostat status gateway --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Thermostat component ID")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	thermostat := conn.Thermostat(opts.ID)

	var status *components.ThermostatStatus
	err = cmdutil.RunWithSpinner(ctx, ios, "Getting thermostat status...", func(ctx context.Context) error {
		var statusErr error
		status, statusErr = thermostat.GetStatus(ctx)
		return statusErr
	})
	if err != nil {
		return fmt.Errorf("failed to get thermostat status: %w", err)
	}

	if opts.JSON {
		jsonBytes, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(jsonBytes))
		return nil
	}

	displayStatus(ios, status, opts.ID)
	return nil
}

func displayStatus(ios *iostreams.IOStreams, status *components.ThermostatStatus, id int) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Thermostat %d Status:", id)))
	ios.Println()

	displayTemperature(ios, status)
	displayValve(ios, status)
	displayHumidity(ios, status)
	displayModes(ios, status)
	displayFlagsAndErrors(ios, status)
}

func displayTemperature(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
	ios.Println("  " + theme.Highlight().Render("Temperature:"))
	if status.CurrentC != nil {
		tempStr := fmt.Sprintf("%.1f°C", *status.CurrentC)
		if status.CurrentF != nil {
			tempStr += fmt.Sprintf(" (%.1f°F)", *status.CurrentF)
		}
		ios.Printf("    Current:  %s\n", tempStr)
	}
	if status.TargetC != nil {
		targetStr := fmt.Sprintf("%.1f°C", *status.TargetC)
		if status.TargetF != nil {
			targetStr += fmt.Sprintf(" (%.1f°F)", *status.TargetF)
		}
		ios.Printf("    Target:   %s\n", targetStr)
	}
	ios.Println()
}

func displayValve(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
	ios.Println("  " + theme.Highlight().Render("Valve:"))
	if status.Pos != nil {
		posBar := output.RenderProgressBar(*status.Pos, 100)
		ios.Printf("    Position: %s %d%%\n", posBar, *status.Pos)
	}
	if status.Output != nil {
		ios.Printf("    Output:   %s\n", output.RenderValveState(*status.Output))
	}
	ios.Println()
}

func displayHumidity(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
	if status.CurrentHumidity == nil {
		return
	}
	ios.Println("  " + theme.Highlight().Render("Humidity:"))
	ios.Printf("    Current: %.1f%%\n", *status.CurrentHumidity)
	if status.TargetHumidity != nil {
		ios.Printf("    Target:  %.1f%%\n", *status.TargetHumidity)
	}
	ios.Println()
}

func displayModes(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
	if status.Boost != nil && status.Boost.StartedAt > 0 {
		ios.Println("  " + theme.StatusWarn().Render("Boost Mode Active"))
		ios.Printf("    Duration: %d seconds\n", status.Boost.Duration)
		ios.Println()
	}
	if status.Override != nil && status.Override.StartedAt > 0 {
		ios.Println("  " + theme.Highlight().Render("Override Active"))
		ios.Printf("    Duration: %d seconds\n", status.Override.Duration)
		ios.Println()
	}
}

func displayFlagsAndErrors(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
	if len(status.Flags) > 0 {
		ios.Println("  " + theme.Highlight().Render("Flags:"))
		for _, flag := range status.Flags {
			ios.Printf("    - %s\n", flag)
		}
		ios.Println()
	}
	if len(status.Errors) > 0 {
		ios.Println("  " + theme.StatusError().Render("Errors:"))
		for _, e := range status.Errors {
			ios.Printf("    - %s\n", e)
		}
		ios.Println()
	}
}
