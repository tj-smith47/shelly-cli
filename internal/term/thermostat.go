// Package term provides composed terminal presentation for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayThermostatStatus displays the full status of a thermostat.
func DisplayThermostatStatus(ios *iostreams.IOStreams, status *components.ThermostatStatus, id int) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Thermostat %d Status:", id)))
	ios.Println()

	displayThermostatTemperature(ios, status)
	displayThermostatValve(ios, status)
	displayThermostatHumidity(ios, status)
	displayThermostatModes(ios, status)
	displayThermostatFlagsAndErrors(ios, status)
}

func displayThermostatTemperature(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
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

func displayThermostatValve(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
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

func displayThermostatHumidity(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
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

func displayThermostatModes(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
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

func displayThermostatFlagsAndErrors(ios *iostreams.IOStreams, status *components.ThermostatStatus) {
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

// DisplayThermostats displays a list of thermostats.
func DisplayThermostats(ios *iostreams.IOStreams, thermostats []model.ThermostatInfo, device string) {
	if len(thermostats) == 0 {
		ios.Info("No thermostats found on %s", device)
		ios.Info("Thermostat support is available on Shelly BLU TRV via BLU Gateway.")
		return
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("Thermostats on %s:", device)))
	ios.Println()

	for _, t := range thermostats {
		ios.Printf("  %s %d\n", theme.Highlight().Render("Thermostat"), t.ID)
		ios.Printf("    Status: %s\n", output.RenderActive(t.Enabled, output.CaseTitle, theme.FalseDim))
		if t.TargetC > 0 {
			ios.Printf("    Target: %.1f°C\n", t.TargetC)
		}
		ios.Println()
	}

	ios.Success("Found %d thermostat(s)", len(thermostats))
}
