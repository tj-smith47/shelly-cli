// Package cmdutil provides display helpers that print directly to IOStreams.
package cmdutil

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// PrintPowerMetrics outputs power, voltage, and current with units.
// Nil values are skipped.
func PrintPowerMetrics(ios *iostreams.IOStreams, power, voltage, current *float64) {
	if power != nil {
		ios.Printf("  Power:   %.1f W\n", *power)
	}
	if voltage != nil {
		ios.Printf("  Voltage: %.1f V\n", *voltage)
	}
	if current != nil {
		ios.Printf("  Current: %.3f A\n", *current)
	}
}

// PrintPowerMetricsWide outputs power metrics with wider alignment for cover status.
func PrintPowerMetricsWide(ios *iostreams.IOStreams, power, voltage, current *float64) {
	if power != nil {
		ios.Printf("  Power:    %.1f W\n", *power)
	}
	if voltage != nil {
		ios.Printf("  Voltage:  %.1f V\n", *voltage)
	}
	if current != nil {
		ios.Printf("  Current:  %.3f A\n", *current)
	}
}
