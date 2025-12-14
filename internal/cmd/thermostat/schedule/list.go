// Package schedule provides thermostat schedule management commands.
package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ListOptions holds list command options.
type ListOptions struct {
	Factory      *cmdutil.Factory
	Device       string
	ThermostatID int
	All          bool
	JSON         bool
}

func newListCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List thermostat schedules",
		Long: `List all schedules that control the thermostat.

By default, only shows schedules that target the thermostat component.
Use --all to show all device schedules.`,
		Example: `  # List thermostat schedules
  shelly thermostat schedule list gateway

  # List schedules for specific thermostat ID
  shelly thermostat schedule list gateway --thermostat-id 1

  # List all device schedules
  shelly thermostat schedule list gateway --all

  # Output as JSON
  shelly thermostat schedule list gateway --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runList(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ThermostatID, "thermostat-id", 0, "Filter by thermostat component ID")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Show all device schedules")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// ThermostatSchedule represents a schedule targeting a thermostat.
type ThermostatSchedule struct {
	ID           int      `json:"id"`
	Enabled      bool     `json:"enabled"`
	Timespec     string   `json:"timespec"`
	ThermostatID int      `json:"thermostat_id,omitempty"`
	TargetC      *float64 `json:"target_c,omitempty"`
	Mode         string   `json:"mode,omitempty"`
	Enable       *bool    `json:"enable,omitempty"`
}

func runList(ctx context.Context, opts *ListOptions) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	// Get all schedules using raw RPC call
	ios.StartProgress("Getting schedules...")
	result, err := conn.Call(ctx, "Schedule.List", nil)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to list schedules: %w", err)
	}

	// Parse the response
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var scheduleResp components.ScheduleListResponse
	if err := json.Unmarshal(jsonBytes, &scheduleResp); err != nil {
		return fmt.Errorf("failed to parse schedules: %w", err)
	}

	// Filter for thermostat schedules
	thermostatSchedules := filterThermostatSchedules(scheduleResp.Jobs, opts.ThermostatID, opts.All)

	if opts.JSON {
		output, err := json.MarshalIndent(thermostatSchedules, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	displaySchedules(ios, thermostatSchedules, opts.Device, opts.All)
	return nil
}

// filterThermostatSchedules filters schedule jobs to only those targeting thermostats.
// If showAll is true, returns all schedules without filtering.
// If thermostatID > 0, only returns schedules targeting that specific thermostat.
func filterThermostatSchedules(jobs []components.ScheduleJob, thermostatID int, showAll bool) []ThermostatSchedule {
	var schedules []ThermostatSchedule

	for _, job := range jobs {
		if showAll {
			// Include all schedules without parsing their calls
			schedules = append(schedules, ThermostatSchedule{
				ID:       job.ID,
				Enabled:  job.Enable,
				Timespec: job.Timespec,
			})
			continue
		}

		// Only include if the job calls a Thermostat.* method
		if sched, ok := extractThermostatSchedule(job, thermostatID); ok {
			schedules = append(schedules, sched)
		}
	}

	return schedules
}

// extractThermostatSchedule checks if a schedule job targets a thermostat.
// Returns the extracted schedule and true if the job contains a Thermostat.* call.
// If filterID > 0, only matches schedules targeting that thermostat ID.
func extractThermostatSchedule(job components.ScheduleJob, filterID int) (ThermostatSchedule, bool) {
	for _, call := range job.Calls {
		// Only interested in Thermostat.* method calls (e.g., SetConfig, SetTarget)
		if !strings.HasPrefix(call.Method, "Thermostat.") {
			continue
		}

		sched := ThermostatSchedule{
			ID:       job.ID,
			Enabled:  job.Enable,
			Timespec: job.Timespec,
		}

		// Extract thermostat-specific parameters from the call
		params, ok := call.Params.(map[string]any)
		if !ok {
			return sched, true
		}

		extractThermostatParams(&sched, params)

		// Skip if filtering by thermostat ID and doesn't match
		if filterID > 0 && sched.ThermostatID != filterID {
			continue
		}

		return sched, true
	}
	return ThermostatSchedule{}, false
}

// extractThermostatParams extracts thermostat-related parameters from a schedule call.
// Handles both SetConfig (nested config object) and SetTarget (direct params) formats.
func extractThermostatParams(sched *ThermostatSchedule, params map[string]any) {
	// Get the thermostat component ID
	if id, ok := params["id"].(float64); ok {
		sched.ThermostatID = int(id)
	}

	// SetConfig format: params.config.{target_C, thermostat_mode, enable}
	if config, ok := params["config"].(map[string]any); ok {
		extractConfigParams(sched, config)
	}

	// SetTarget format: params.target_C directly
	if targetC, ok := params["target_C"].(float64); ok {
		sched.TargetC = &targetC
	}
}

// extractConfigParams extracts target temperature, mode, and enable state from config object.
func extractConfigParams(sched *ThermostatSchedule, config map[string]any) {
	if targetC, ok := config["target_C"].(float64); ok {
		sched.TargetC = &targetC
	}
	if mode, ok := config["thermostat_mode"].(string); ok {
		sched.Mode = mode
	}
	if enable, ok := config["enable"].(bool); ok {
		sched.Enable = &enable
	}
}

func displaySchedules(ios *iostreams.IOStreams, schedules []ThermostatSchedule, device string, showAll bool) {
	if len(schedules) == 0 {
		if showAll {
			ios.Info("No schedules found on %s", device)
		} else {
			ios.Info("No thermostat schedules found on %s", device)
			ios.Info("Use --all to see all device schedules")
		}
		return
	}

	title := "Thermostat Schedules"
	if showAll {
		title = "All Schedules"
	}
	ios.Println(theme.Bold().Render(fmt.Sprintf("%s on %s:", title, device)))
	ios.Println()

	for _, sched := range schedules {
		status := theme.Dim().Render("Disabled")
		if sched.Enabled {
			status = theme.StatusOK().Render("Enabled")
		}

		ios.Printf("  %s %d\n", theme.Highlight().Render("Schedule"), sched.ID)
		ios.Printf("    Status:   %s\n", status)
		ios.Printf("    Timespec: %s\n", sched.Timespec)

		if sched.ThermostatID > 0 {
			ios.Printf("    Thermostat: %d\n", sched.ThermostatID)
		}
		if sched.TargetC != nil {
			ios.Printf("    Target: %.1f°C\n", *sched.TargetC)
		}
		if sched.Mode != "" {
			ios.Printf("    Mode: %s\n", sched.Mode)
		}
		if sched.Enable != nil {
			enableStr := "disable"
			if *sched.Enable {
				enableStr = "enable"
			}
			ios.Printf("    Action: %s thermostat\n", enableStr)
		}
		ios.Println()
	}

	ios.Success("Found %d schedule(s)", len(schedules))
}
