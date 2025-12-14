// Package list provides the thermostat list command.
package list

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the thermostat list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List thermostats",
		Long: `List all thermostat components on a Shelly device.

Shows basic information about each thermostat including its ID
and current enabled state.`,
		Example: `  # List thermostats
  shelly thermostat list gateway

  # Output as JSON
  shelly thermostat list gateway --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// ThermostatInfo holds basic thermostat information.
type ThermostatInfo struct {
	ID      int     `json:"id"`
	Enabled bool    `json:"enabled"`
	Mode    string  `json:"mode,omitempty"`
	TargetC float64 `json:"target_c,omitempty"`
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	// Get full device status to find thermostats
	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return fmt.Errorf("failed to get device status: %w", err)
	}

	// Parse status to find thermostat components
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var fullStatus map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &fullStatus); err != nil {
		return fmt.Errorf("failed to parse status: %w", err)
	}

	thermostats := collectThermostats(fullStatus)

	if opts.JSON {
		output, err := json.MarshalIndent(thermostats, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	return displayThermostats(ios, thermostats, opts.Device)
}

func collectThermostats(status map[string]json.RawMessage) []ThermostatInfo {
	var thermostats []ThermostatInfo

	for key, raw := range status {
		if strings.HasPrefix(key, "thermostat:") {
			var t struct {
				ID       int      `json:"id"`
				Output   *bool    `json:"output"`
				TargetC  *float64 `json:"target_C"`
				CurrentC *float64 `json:"current_C"`
			}
			if err := json.Unmarshal(raw, &t); err == nil {
				info := ThermostatInfo{
					ID:      t.ID,
					Enabled: t.Output != nil && *t.Output,
				}
				if t.TargetC != nil {
					info.TargetC = *t.TargetC
				}
				thermostats = append(thermostats, info)
			}
		}
	}

	return thermostats
}

func displayThermostats(ios *iostreams.IOStreams, thermostats []ThermostatInfo, device string) error {
	if len(thermostats) == 0 {
		ios.Info("No thermostats found on %s", device)
		ios.Info("Thermostat support is available on Shelly BLU TRV via BLU Gateway.")
		return nil
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("Thermostats on %s:", device)))
	ios.Println()

	for _, t := range thermostats {
		status := theme.Dim().Render("Off")
		if t.Enabled {
			status = theme.StatusOK().Render("Active")
		}

		ios.Printf("  %s %d\n", theme.Highlight().Render("Thermostat"), t.ID)
		ios.Printf("    Status: %s\n", status)
		if t.TargetC > 0 {
			ios.Printf("    Target: %.1f°C\n", t.TargetC)
		}
		ios.Println()
	}

	ios.Success("Found %d thermostat(s)", len(thermostats))
	return nil
}
