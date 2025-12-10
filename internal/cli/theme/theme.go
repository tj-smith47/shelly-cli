// Package theme provides theme management commands.
package theme

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the theme command and its subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "theme",
		Short: "Manage CLI themes",
		Long: `Manage CLI color themes.

Shelly CLI supports 280+ built-in themes via bubbletint.
Themes affect all CLI output including tables, status indicators, and the TUI dashboard.

Examples:
  shelly theme list              # List all available themes
  shelly theme set dracula       # Set the theme
  shelly theme preview nord      # Preview a theme
  shelly theme current           # Show current theme`,
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newPreviewCommand())
	cmd.AddCommand(newCurrentCommand())
	cmd.AddCommand(newNextCommand())
	cmd.AddCommand(newPrevCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	var outputFormat string
	var filter string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available themes",
		Long:    "List all available color themes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			themes := theme.ListThemes()

			// Filter themes if requested
			if filter != "" {
				filter = strings.ToLower(filter)
				var filtered []string
				for _, t := range themes {
					if strings.Contains(strings.ToLower(t), filter) {
						filtered = append(filtered, t)
					}
				}
				themes = filtered
			}

			// Sort alphabetically
			sort.Strings(themes)

			if len(themes) == 0 {
				fmt.Println("No themes found matching filter.")
				return nil
			}

			switch outputFormat {
			case "json":
				return output.JSON(cmd.OutOrStdout(), themes)
			case "yaml":
				return output.YAML(cmd.OutOrStdout(), themes)
			default:
				currentTheme := viper.GetString("theme")
				fmt.Printf("Available themes (%d):\n\n", len(themes))

				// Print in columns
				cols := 4
				for i, t := range themes {
					marker := "  "
					if t == currentTheme {
						marker = "* "
					}
					fmt.Printf("%s%-20s", marker, t)
					if (i+1)%cols == 0 {
						fmt.Println()
					}
				}
				if len(themes)%cols != 0 {
					fmt.Println()
				}

				fmt.Printf("\n* = current theme\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter themes by name")

	return cmd
}

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Set the current theme",
		Long: `Set the current color theme.

The theme will be saved to your configuration and applied immediately.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Try to set the theme
			if !theme.SetTheme(name) {
				return fmt.Errorf("theme '%s' not found; use 'shelly theme list' to see available themes", name)
			}

			// Save to config
			cfg := config.Get()
			cfg.Theme = name
			viper.Set("theme", name)
			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Theme set to '%s'\n", name)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return theme.ListThemes(), cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newPreviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview <name>",
		Short: "Preview a theme",
		Long: `Preview a theme without changing your settings.

This shows sample output with the specified theme applied.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Save current theme
			currentTheme := viper.GetString("theme")

			// Temporarily set the new theme
			if !theme.SetTheme(name) {
				return fmt.Errorf("theme '%s' not found; use 'shelly theme list' to see available themes", name)
			}

			// Show preview
			fmt.Printf("Preview of theme: %s\n\n", name)

			// Show color samples
			fmt.Println("Status indicators:")
			fmt.Println("  " + theme.StatusOK().Render("OK") + "  " +
				theme.StatusWarn().Render("WARNING") + "  " +
				theme.StatusError().Render("ERROR") + "  " +
				theme.StatusInfo().Render("INFO"))
			fmt.Println()

			fmt.Println("Device status:")
			fmt.Println("  " + theme.DeviceOnline() + "  " + theme.DeviceOffline() + "  " + theme.DeviceUpdating())
			fmt.Println()

			fmt.Println("Switch states:")
			fmt.Println("  " + theme.SwitchOn() + "  " + theme.SwitchOff())
			fmt.Println()

			fmt.Println("Power/Energy:")
			fmt.Println("  " + theme.FormatPower(125.5) + "  " + theme.FormatEnergy(1234.5))
			fmt.Println()

			fmt.Println("Text styles:")
			fmt.Println("  " + theme.Bold().Render("Bold") + "  " +
				theme.Dim().Render("Dim") + "  " +
				theme.Highlight().Render("Highlight"))
			fmt.Println("  " + theme.Title().Render("Title") + "  " +
				theme.Subtitle().Render("Subtitle"))
			fmt.Println("  " + theme.Link().Render("https://example.com") + "  " +
				theme.Code().Render("code"))
			fmt.Println()

			// Show sample table
			fmt.Println("Sample table:")
			t := output.NewTable("Device", "Status", "Power")
			t.AddRow("Living Room", theme.DeviceOnline(), theme.FormatPower(45.2))
			t.AddRow("Bedroom", theme.DeviceOffline(), "-")
			t.AddRow("Kitchen", theme.DeviceOnline(), theme.FormatPower(120.0))
			t.Print()

			// Restore original theme
			theme.SetTheme(currentTheme)

			fmt.Printf("\nTo use this theme: shelly theme set %s\n", name)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return theme.ListThemes(), cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func newCurrentCommand() *cobra.Command {
	var showColors bool

	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the current theme",
		Long:  "Display information about the currently active theme.",
		RunE: func(cmd *cobra.Command, args []string) error {
			currentTheme := viper.GetString("theme")
			fmt.Printf("Current theme: %s\n", currentTheme)

			if showColors {
				fmt.Println("\nTheme colors:")

				t := theme.Current()

				// Show color swatches
				colors := []struct {
					name  string
					color lipgloss.TerminalColor
				}{
					{"Foreground", t.Fg()},
					{"Background", t.Bg()},
					{"Black", t.BrightBlack()},
					{"Red", t.Red()},
					{"Green", t.Green()},
					{"Yellow", t.Yellow()},
					{"Blue", t.Blue()},
					{"Purple", t.Purple()},
					{"Cyan", t.Cyan()},
				}

				for _, c := range colors {
					style := lipgloss.NewStyle().Foreground(c.color)
					swatch := lipgloss.NewStyle().Background(c.color).Render("    ")
					fmt.Printf("  %-12s %s %s\n", c.name+":", swatch, style.Render("Sample Text"))
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showColors, "colors", "c", false, "Show theme colors")

	return cmd
}

func newNextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Cycle to the next theme",
		Long:  "Switch to the next theme in alphabetical order.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme.NextTheme()

			// Get the new theme name
			newTheme := theme.Current().ID()

			// Save to config
			cfg := config.Get()
			cfg.Theme = newTheme
			viper.Set("theme", newTheme)
			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Theme changed to '%s'\n", newTheme)
			return nil
		},
	}

	return cmd
}

func newPrevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prev",
		Aliases: []string{"previous"},
		Short:   "Cycle to the previous theme",
		Long:    "Switch to the previous theme in alphabetical order.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme.PrevTheme()

			// Get the new theme name
			newTheme := theme.Current().ID()

			// Save to config
			cfg := config.Get()
			cfg.Theme = newTheme
			viper.Set("theme", newTheme)
			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Theme changed to '%s'\n", newTheme)
			return nil
		},
	}

	return cmd
}
