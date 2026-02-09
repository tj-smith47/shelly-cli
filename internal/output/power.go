// Package output provides formatters for CLI output.
package output

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// PowerValueOptions configures how a power value is formatted.
type PowerValueOptions struct {
	Colored           bool     // Apply color based on power threshold (>=1000W red, >=100W yellow, else green)
	ShowChange        bool     // Show change indicator arrows (↑/↓)
	PrevPower         *float64 // Previous power value for change indicator
	ZeroAsPlaceholder bool     // Show placeholder instead of "0.0 W"
	Placeholder       string   // Custom placeholder (default "-")
}

// FormatPowerValue formats a power value with the given options.
// This is the unified power value formatter.
func FormatPowerValue(watts float64, opts PowerValueOptions) string {
	// Handle zero/negative as placeholder
	if opts.ZeroAsPlaceholder && watts <= 0 {
		return getPlaceholder(opts.Placeholder)
	}

	// Format base power string
	s := formatPowerBase(watts)

	// Apply change indicator if previous value available
	if opts.ShowChange && opts.PrevPower != nil {
		return formatWithChange(s, watts, *opts.PrevPower, opts.Colored)
	}

	// Apply coloring if requested
	if opts.Colored {
		return applyPowerColor(watts, s)
	}

	return s
}

// getPlaceholder returns the placeholder string, defaulting to "-".
func getPlaceholder(placeholder string) string {
	if placeholder == "" {
		return "-"
	}
	return placeholder
}

// formatWithChange formats power with a change indicator arrow.
func formatWithChange(s string, current, previous float64, colored bool) string {
	if current > previous {
		s += " ↑"
		if colored {
			return theme.StatusWarn().Render(s)
		}
	} else if current < previous {
		s += " ↓"
		if colored {
			return theme.StatusOK().Render(s)
		}
	}
	// No change - apply regular coloring if requested
	if colored {
		return applyPowerColor(current, s)
	}
	return s
}

// formatPowerBase formats power as W or kW without any styling.
// Handles both positive and negative values (e.g., solar return energy).
func formatPowerBase(watts float64) string {
	absVal := watts
	if absVal < 0 {
		absVal = -absVal
	}
	if absVal >= 1000 {
		return fmt.Sprintf("%.2f kW", watts/1000)
	}
	return fmt.Sprintf("%.1f W", watts)
}

// applyPowerColor applies threshold-based coloring to a power string.
// Red for >=1000W, yellow for >=100W, green otherwise.
func applyPowerColor(watts float64, s string) string {
	if watts >= 1000 {
		return theme.StatusError().Render(s)
	} else if watts >= 100 {
		return theme.StatusWarn().Render(s)
	}
	return theme.StatusOK().Render(s)
}

// MeterLineOptions configures how a meter line is formatted.
type MeterLineOptions struct {
	// Required fields
	Label   string  // Line label: "PM", "EM1", "Phase A", etc.
	Power   float64 // Power in watts
	Voltage float64 // Voltage in volts
	Current float64 // Current in amps

	// Optional fields
	ID        *int     // Component ID (nil to omit)
	PF        *float64 // Power factor (nil to omit)
	Energy    *float64 // Energy total in Wh (nil to omit)
	PrevPower *float64 // Previous power for change indicator
	Indent    string   // Line prefix (default "  ")
}

// FormatMeterReading formats a complete meter reading line in the format:
// "  {Label} {ID}: {Power}  {Voltage}V  {Current}A  [PF:{pf}]  [{Energy} Wh]".
func FormatMeterReading(opts MeterLineOptions) string {
	// Format power with change indicator
	powerStr := FormatPowerValue(opts.Power, PowerValueOptions{
		ShowChange: opts.PrevPower != nil,
		PrevPower:  opts.PrevPower,
	})

	// Build label with optional ID
	label := opts.Label
	if opts.ID != nil {
		label = fmt.Sprintf("%s %d", opts.Label, *opts.ID)
	}

	// Build suffix parts
	pfStr := ""
	if opts.PF != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *opts.PF)
	}

	energyStr := ""
	if opts.Energy != nil {
		energyStr = fmt.Sprintf("  %.2f Wh", *opts.Energy)
	}

	// Use default indent
	indent := opts.Indent
	if indent == "" {
		indent = "  "
	}

	return fmt.Sprintf("%s%s: %s  %.1fV  %.2fA%s%s",
		indent, label, powerStr, opts.Voltage, opts.Current, pfStr, energyStr)
}

// PhaseLineOptions configures how a phase line is formatted (for 3-phase EM).
type PhaseLineOptions struct {
	Phase     string   // Phase name: "A", "B", "C"
	Power     float64  // Active power in watts
	Voltage   float64  // Voltage in volts
	Current   float64  // Current in amps
	PF        *float64 // Power factor (nil to omit)
	PrevPower *float64 // Previous power for change indicator
}

// FormatPhaseLine formats a single phase line for 3-phase energy meters in the format:
// "    Phase {X}: {Power}  {Voltage}V  {Current}A  [PF:{pf}]".
func FormatPhaseLine(opts PhaseLineOptions) string {
	powerStr := FormatPowerValue(opts.Power, PowerValueOptions{
		ShowChange: opts.PrevPower != nil,
		PrevPower:  opts.PrevPower,
	})

	pfStr := ""
	if opts.PF != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *opts.PF)
	}

	return fmt.Sprintf("    Phase %s: %s  %.1fV  %.2fA%s",
		opts.Phase, powerStr, opts.Voltage, opts.Current, pfStr)
}
