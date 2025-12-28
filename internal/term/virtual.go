// Package term provides terminal display functions for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayVirtualComponents displays a list of virtual components.
func DisplayVirtualComponents(ios *iostreams.IOStreams, components []shelly.VirtualComponent) {
	if len(components) == 0 {
		ios.NoResults("virtual components")
		return
	}

	ios.Title("Virtual Components")
	ios.Println()

	for _, c := range components {
		displayVirtualComponentRow(ios, &c)
	}

	ios.Println()
	ios.Count("virtual component", len(components))
}

// DisplayVirtualComponent displays a single virtual component in detail.
func DisplayVirtualComponent(ios *iostreams.IOStreams, c *shelly.VirtualComponent) {
	ios.Title("Virtual Component")
	ios.Println()

	ios.Printf("  %s: %s\n", theme.Dim().Render("Key"), theme.Highlight().Render(c.Key))
	ios.Printf("  %s: %s\n", theme.Dim().Render("Type"), string(c.Type))
	ios.Printf("  %s: %d\n", theme.Dim().Render("ID"), c.ID)

	if c.Name != "" {
		ios.Printf("  %s: %s\n", theme.Dim().Render("Name"), c.Name)
	}

	// Display value based on type
	ios.Printf("  %s: %s\n", theme.Dim().Render("Value"), formatVirtualValue(c))

	// Display constraints if present
	if c.Min != nil || c.Max != nil {
		var rangeStr string
		switch {
		case c.Min != nil && c.Max != nil:
			rangeStr = fmt.Sprintf("%.2f - %.2f", *c.Min, *c.Max)
		case c.Min != nil:
			rangeStr = fmt.Sprintf(">= %.2f", *c.Min)
		case c.Max != nil:
			rangeStr = fmt.Sprintf("<= %.2f", *c.Max)
		}
		ios.Printf("  %s: %s\n", theme.Dim().Render("Range"), rangeStr)
	}

	if c.Unit != nil && *c.Unit != "" {
		ios.Printf("  %s: %s\n", theme.Dim().Render("Unit"), *c.Unit)
	}

	if len(c.Options) > 0 {
		ios.Printf("  %s: %v\n", theme.Dim().Render("Options"), c.Options)
	}

	ios.Println()
}

func displayVirtualComponentRow(ios *iostreams.IOStreams, c *shelly.VirtualComponent) {
	typeStr := theme.Dim().Render(fmt.Sprintf("[%s]", c.Type))
	keyStr := theme.Highlight().Render(c.Key)
	valueStr := formatVirtualValue(c)

	if c.Name != "" {
		ios.Printf("  %s %s (%s): %s\n", typeStr, keyStr, c.Name, valueStr)
	} else {
		ios.Printf("  %s %s: %s\n", typeStr, keyStr, valueStr)
	}
}

func formatVirtualValue(c *shelly.VirtualComponent) string {
	switch c.Type {
	case shelly.VirtualBoolean:
		if c.BoolValue != nil {
			if *c.BoolValue {
				return theme.StatusOK().Render("true")
			}
			return theme.StatusError().Render("false")
		}
	case shelly.VirtualNumber:
		if c.NumValue != nil {
			val := fmt.Sprintf("%.2f", *c.NumValue)
			if c.Unit != nil && *c.Unit != "" {
				val += " " + *c.Unit
			}
			return val
		}
	case shelly.VirtualText, shelly.VirtualEnum:
		if c.StrValue != nil {
			return fmt.Sprintf("%q", *c.StrValue)
		}
	case shelly.VirtualButton:
		return theme.Dim().Render("(button)")
	case shelly.VirtualGroup:
		return theme.Dim().Render("(group)")
	}

	if c.Value != nil {
		return fmt.Sprintf("%v", c.Value)
	}

	return theme.Dim().Render("(no value)")
}
