// Package term provides terminal display functions for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayOnboardDevices shows a table of discovered devices for onboarding.
func DisplayOnboardDevices(ios *iostreams.IOStreams, devices []shelly.OnboardDevice) {
	if len(devices) == 0 {
		ios.NoResults("unprovisioned devices")
		return
	}

	builder := table.NewBuilder("#", "Device", "Model", "Gen", "Found via", "Status")

	for i, d := range devices {
		name := d.Name
		if name == "" {
			name = d.Address
		}

		genStr := fmt.Sprintf("Gen%d", d.Generation)
		if d.Generation == 0 {
			genStr = theme.Dim().Render("?")
		}

		status := formatOnboardStatus(d)

		builder.AddRow(
			fmt.Sprintf("%d", i+1),
			name,
			d.Model,
			genStr,
			string(d.Source),
			status,
		)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print onboard devices table", err)
	}
	ios.Println("")
	ios.Count("device", len(devices))
}

// formatOnboardStatus returns a themed status string for an onboard device.
func formatOnboardStatus(d shelly.OnboardDevice) string {
	switch {
	case d.Registered:
		return theme.StatusOK().Render("registered")
	case d.Provisioned:
		return theme.StatusWarn().Render("on network")
	default:
		return theme.StatusInfo().Render("new")
	}
}

// DisplayOnboardResults shows per-device provisioning results.
func DisplayOnboardResults(ios *iostreams.IOStreams, results []*shelly.OnboardResult) {
	for _, r := range results {
		name := r.Device.Name
		if name == "" {
			name = r.Device.Address
		}

		if r.Error != nil {
			ios.Error("  %s: %v", name, r.Error)
			continue
		}

		msg := fmt.Sprintf("%s (%s)", name, r.Method)
		if r.NewAddress != "" {
			msg += fmt.Sprintf(" → %s", r.NewAddress)
		}
		if r.Registered {
			msg += " [registered]"
		}
		ios.Success("  %s", msg)
	}
}

// DisplayOnboardSummary shows a final summary of onboarding results.
func DisplayOnboardSummary(ios *iostreams.IOStreams, results []*shelly.OnboardResult) {
	var succeeded, failed int
	for _, r := range results {
		if r.Error != nil {
			failed++
		} else {
			succeeded++
		}
	}

	ios.Println()
	if failed == 0 {
		ios.Success("All %d devices provisioned successfully", succeeded)
	} else {
		ios.Warning("%d of %d devices provisioned (%d failed)", succeeded, len(results), failed)
	}

	// Show address mapping for successful devices
	if succeeded > 0 {
		ios.Println()
		for _, r := range results {
			if r.Error != nil {
				continue
			}
			name := r.Device.Name
			if name == "" {
				name = r.Device.Address
			}
			if r.NewAddress != "" {
				ios.Printf("  %s → %s\n", name, r.NewAddress)
			}
		}
	}
}

// FormatOnboardDeviceOptions builds display strings for MultiSelect.
// Each option has the format: "DeviceName (Gen2, BLE)".
func FormatOnboardDeviceOptions(devices []shelly.OnboardDevice) []string {
	options := make([]string, len(devices))
	for i, d := range devices {
		name := d.Name
		if name == "" {
			name = d.Address
		}
		genStr := fmt.Sprintf("Gen%d", d.Generation)
		if d.Generation == 0 {
			genStr = "Gen?"
		}
		options[i] = fmt.Sprintf("%s (%s, %s)", name, genStr, d.Source)
	}
	return options
}

// SelectOnboardDevices presents an interactive multi-select for device selection.
// If autoConfirm is true, returns all devices without prompting.
func SelectOnboardDevices(ios *iostreams.IOStreams, devices []shelly.OnboardDevice, autoConfirm bool) ([]shelly.OnboardDevice, error) {
	if autoConfirm {
		return devices, nil
	}

	options := FormatOnboardDeviceOptions(devices)
	chosen, err := ios.MultiSelect("Select devices to provision:", options, options)
	if err != nil {
		return nil, fmt.Errorf("selection cancelled: %w", err)
	}

	chosenSet := make(map[string]struct{}, len(chosen))
	for _, c := range chosen {
		chosenSet[c] = struct{}{}
	}

	result := make([]shelly.OnboardDevice, 0, len(chosen))
	for i, opt := range options {
		if _, ok := chosenSet[opt]; ok {
			result = append(result, devices[i])
		}
	}
	return result, nil
}
