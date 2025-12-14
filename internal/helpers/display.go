// Package helpers provides common functionality shared across CLI commands.
package helpers

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// FormatAuth returns a styled string for authentication status.
func FormatAuth(hasAuth bool) string {
	if hasAuth {
		return theme.StatusWarn().Render("Yes")
	}
	return theme.StatusOK().Render("No")
}

// FormatGeneration returns a styled string for device generation.
func FormatGeneration(gen int) string {
	if gen == 0 {
		return theme.Dim().Render("unknown")
	}
	return theme.Bold().Render(fmt.Sprintf("Gen%d", gen))
}

// FormatOnOff returns a styled string for on/off state.
func FormatOnOff(on bool) string {
	if on {
		return theme.SwitchOn()
	}
	return theme.SwitchOff()
}

// FormatState returns a styled string for component state.
func FormatState(state string) string {
	switch state {
	case "open", "opening":
		return theme.StatusOK().Render(state)
	case "closed", "closing":
		return theme.StatusInfo().Render(state)
	case "stopped", "idle":
		return theme.Dim().Render(state)
	case "calibrating":
		return theme.StatusWarn().Render(state)
	default:
		return state
	}
}

// DisplayDiscoveredDevices prints a table of discovered devices with themed formatting.
func DisplayDiscoveredDevices(devices []discovery.DiscoveredDevice) {
	if len(devices) == 0 {
		iostreams.NoResults("devices")
		return
	}

	table := output.NewTable("ID", "Address", "Model", "Generation", "Protocol", "Auth")

	for _, d := range devices {
		gen := fmt.Sprintf("Gen%d", d.Generation)

		// Theme the auth status
		auth := theme.StatusOK().Render("No")
		if d.AuthRequired {
			auth = theme.StatusWarn().Render("Yes")
		}

		table.AddRow(
			d.ID,
			d.Address.String(),
			d.Model,
			gen,
			string(d.Protocol),
			auth,
		)
	}

	table.Print()
	iostreams.Count("device", len(devices))
}

// DisplayDeviceSummary prints a summary line for a single device with themed formatting.
func DisplayDeviceSummary(d discovery.DiscoveredDevice) {
	auth := ""
	if d.AuthRequired {
		auth = theme.StatusWarn().Render(" (auth required)")
	}

	id := theme.Bold().Render(d.ID)
	address := theme.Highlight().Render(d.Address.String())
	model := theme.Dim().Render(d.Model)

	iostreams.Plain("  %s @ %s - %s Gen%d%s", id, address, model, d.Generation, auth)
}

// ResolveBatchTargets resolves batch operation targets from flags and arguments.
// Returns a list of device identifiers (names or addresses).
//
// Priority: explicit args > stdin > group > all
// Stdin is read when no args provided and stdin is not a TTY (piped input).
func ResolveBatchTargets(groupName string, all bool, args []string) ([]string, error) {
	// Priority: explicit devices > stdin > group > all
	if len(args) > 0 {
		return args, nil
	}

	// Check if stdin has piped input (not a TTY)
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		targets, err := readDevicesFromStdin()
		if err != nil {
			return nil, err
		}
		if len(targets) > 0 {
			return targets, nil
		}
	}

	if groupName != "" {
		devices, err := config.GetGroupDevices(groupName)
		if err != nil {
			return nil, fmt.Errorf("failed to get group devices: %w", err)
		}
		if len(devices) == 0 {
			return nil, fmt.Errorf("group %q has no devices", groupName)
		}
		targets := make([]string, len(devices))
		for i, d := range devices {
			targets[i] = d.Name
		}
		return targets, nil
	}

	if all {
		devices := config.ListDevices()
		if len(devices) == 0 {
			return nil, fmt.Errorf("no devices registered")
		}
		targets := make([]string, 0, len(devices))
		for name := range devices {
			targets = append(targets, name)
		}
		return targets, nil
	}

	return nil, fmt.Errorf("specify devices, --group, --all, or pipe device names via stdin")
}

// readDevicesFromStdin reads device names from stdin (one per line or space-separated).
func readDevicesFromStdin() ([]string, error) {
	var targets []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		// Support both newline-separated and space-separated
		fields := strings.Fields(line)
		targets = append(targets, fields...)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}
	return targets, nil
}
