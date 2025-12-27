package term

import (
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayScriptList prints a table of scripts.
func DisplayScriptList(ios *iostreams.IOStreams, scripts []automation.ScriptInfo) {
	table := output.NewTable("ID", "Name", "Enabled", "Running")
	for _, s := range scripts {
		name := s.Name
		if name == "" {
			name = output.FormatPlaceholder("(unnamed)")
		}
		table.AddRow(fmt.Sprintf("%d", s.ID), name, output.RenderYesNo(s.Enable, output.CaseLower, theme.FalseDim), output.RenderRunningState(s.Running))
	}
	printTable(ios, table)
}

// DisplayScheduleList prints a table of schedules.
func DisplayScheduleList(ios *iostreams.IOStreams, schedules []automation.ScheduleJob) {
	table := output.NewTable("ID", "Enabled", "Timespec", "Calls")
	for _, s := range schedules {
		callsSummary := formatScheduleCallsSummary(s.Calls)
		table.AddRow(fmt.Sprintf("%d", s.ID), output.RenderYesNo(s.Enable, output.CaseLower, theme.FalseDim), s.Timespec, callsSummary)
	}
	printTable(ios, table)
}

func formatScheduleCallsSummary(calls []automation.ScheduleCall) string {
	if len(calls) == 0 {
		return output.FormatPlaceholder("(none)")
	}

	if len(calls) == 1 {
		call := calls[0]
		if len(call.Params) == 0 {
			return call.Method
		}
		params, err := json.Marshal(call.Params)
		if err != nil {
			return call.Method
		}
		return fmt.Sprintf("%s %s", call.Method, string(params))
	}

	return fmt.Sprintf("%d calls", len(calls))
}

// DisplayWebhookList prints a table of webhooks.
func DisplayWebhookList(ios *iostreams.IOStreams, webhooks []shelly.WebhookInfo) {
	ios.Title("Webhooks")
	ios.Println()

	table := output.NewTable("ID", "Event", "URLs", "Enabled")
	for _, w := range webhooks {
		urls := joinStrings(w.URLs, ", ")
		if len(urls) > 40 {
			urls = urls[:37] + "..."
		}
		table.AddRow(fmt.Sprintf("%d", w.ID), w.Event, urls, output.RenderYesNo(w.Enable, output.CaseTitle, theme.FalseError))
	}
	printTable(ios, table)

	ios.Printf("\n%d webhook(s) configured\n", len(webhooks))
}

// DisplayThermostatSchedules displays thermostat schedules with optional details.
func DisplayThermostatSchedules(ios *iostreams.IOStreams, schedules []shelly.ThermostatSchedule, device string, showAll bool) {
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
		ios.Printf("  %s %d\n", theme.Highlight().Render("Schedule"), sched.ID)
		ios.Printf("    Status:   %s\n", output.RenderEnabledState(sched.Enabled))
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

// ThermostatScheduleCreateDisplay contains display parameters for schedule creation success.
type ThermostatScheduleCreateDisplay struct {
	Device     string
	ScheduleID int
	Timespec   string
	TargetC    *float64
	Mode       string
	Enable     bool
	Disable    bool
	Enabled    bool
}

// DisplayThermostatScheduleCreate displays the result of creating a thermostat schedule.
func DisplayThermostatScheduleCreate(ios *iostreams.IOStreams, d ThermostatScheduleCreateDisplay) {
	ios.Success("Created schedule %d", d.ScheduleID)
	ios.Printf("  Timespec: %s\n", d.Timespec)

	if d.TargetC != nil {
		ios.Printf("  Target: %.1f°C\n", *d.TargetC)
	}
	if d.Mode != "" {
		ios.Printf("  Mode: %s\n", d.Mode)
	}
	if d.Enable {
		ios.Printf("  Action: enable thermostat\n")
	}
	if d.Disable {
		ios.Printf("  Action: disable thermostat\n")
	}

	if !d.Enabled {
		ios.Info("Schedule is disabled. Enable with: shelly thermostat schedule enable %s --id %d", d.Device, d.ScheduleID)
	}
}
