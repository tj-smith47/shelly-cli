package term

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// RenderThemePreview builds a sample-output preview of the currently active theme,
// labeled with themeID. It reads the global theme accessors, so the caller must
// activate the theme to preview (theme.SetTheme) before calling.
func RenderThemePreview(themeID string) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(theme.Purple())
	b.WriteString(title.Render(fmt.Sprintf("Theme: %s", themeID)))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Color Palette:"))
	b.WriteString("\n")

	colors := []struct {
		name  string
		color func() color.Color
	}{
		{"Foreground", theme.Fg},
		{"Background", theme.Bg},
		{"Red", theme.Red},
		{"Green", theme.Green},
		{"Yellow", theme.Yellow},
		{"Blue", theme.Blue},
		{"Purple", theme.Purple},
		{"Cyan", theme.Cyan},
	}
	for _, c := range colors {
		style := lipgloss.NewStyle().Foreground(c.color())
		block := lipgloss.NewStyle().Background(c.color()).Render("    ")
		fmt.Fprintf(&b, "  %s %s\n", block, style.Render(c.name))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Status Indicators:"))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s  Success message\n", theme.StatusOK().Render("✓"))
	fmt.Fprintf(&b, "  %s  Warning message\n", theme.StatusWarn().Render("!"))
	fmt.Fprintf(&b, "  %s  Error message\n", theme.StatusError().Render("✗"))
	fmt.Fprintf(&b, "  %s  Info message\n", theme.StatusInfo().Render("i"))

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Device Status:"))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s\n", theme.DeviceOnline())
	fmt.Fprintf(&b, "  %s\n", theme.DeviceOffline())
	fmt.Fprintf(&b, "  %s\n", theme.DeviceUpdating())

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Switch State:"))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s  %s\n", theme.SwitchOn(), theme.SwitchOff())

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Power/Energy:"))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s  %s\n", theme.StyledPower(1234.5), theme.StyledEnergy(5678.9))

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Table Header:"))
	b.WriteString("\n")
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Cyan())
	dimStyle := lipgloss.NewStyle().Foreground(theme.BrightBlack())
	fmt.Fprintf(&b, "  %s  %s  %s\n",
		headerStyle.Render("DEVICE"),
		headerStyle.Render("STATUS"),
		headerStyle.Render("POWER"))
	fmt.Fprintf(&b, "  %s  %s  %s\n",
		lipgloss.NewStyle().Foreground(theme.Fg()).Render("kitchen"),
		theme.StatusOnline().Render("online"),
		theme.StyledPower(45.2))
	fmt.Fprintf(&b, "  %s  %s  %s\n",
		dimStyle.Render("bedroom"),
		theme.StatusOffline().Render("offline"),
		dimStyle.Render("--"))

	return b.String()
}

// ApplyImportedTheme applies an imported theme and displays success message.
func ApplyImportedTheme(ios *iostreams.IOStreams, themeName string, colors map[string]string) error {
	if err := theme.ApplyConfig(themeName, colors, nil); err != nil {
		return fmt.Errorf("failed to apply theme: %w", err)
	}

	if len(colors) > 0 {
		ios.Success("Theme '%s' with %d color overrides imported and applied", themeName, len(colors))
	} else {
		ios.Success("Theme '%s' imported and applied", themeName)
	}
	return nil
}

// DisplayValidationResult displays theme validation results.
func DisplayValidationResult(ios *iostreams.IOStreams, themeName string, colors map[string]string) {
	ios.Success("Theme file validated successfully")
	if themeName != "" {
		ios.Info("Base theme: %s", themeName)
	}
	if len(colors) > 0 {
		ios.Info("Color overrides: %d", len(colors))
	}
	ios.Info("Use --apply to apply the theme")
}
