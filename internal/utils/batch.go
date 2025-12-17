// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

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
