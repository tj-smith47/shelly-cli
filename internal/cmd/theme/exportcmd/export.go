// Package exportcmd provides the theme export command.
package exportcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ThemeExport represents an exported theme configuration.
type ThemeExport struct {
	Name           string            `yaml:"name" json:"name"`
	ColorOverrides map[string]string `yaml:"color_overrides,omitempty" json:"color_overrides,omitempty"`
	RenderedColors RenderedColors    `yaml:"rendered_colors" json:"rendered_colors"`
}

// RenderedColors represents the actual color values being used (base + overrides).
type RenderedColors struct {
	Foreground  string `yaml:"foreground" json:"foreground"`
	Background  string `yaml:"background" json:"background"`
	Green       string `yaml:"green" json:"green"`
	Red         string `yaml:"red" json:"red"`
	Yellow      string `yaml:"yellow" json:"yellow"`
	Blue        string `yaml:"blue" json:"blue"`
	Cyan        string `yaml:"cyan" json:"cyan"`
	Purple      string `yaml:"purple" json:"purple"`
	BrightBlack string `yaml:"bright_black" json:"bright_black"`
}

// NewCommand creates the theme export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [file]",
		Aliases: []string{"exp", "save"},
		Short:   "Export current theme",
		Long: `Export the current theme configuration to a file.

Exports the base theme name, any custom color overrides, and the effective
colors (what you actually see). The exported file can be imported back with
'shelly theme import'.

If no file is specified, outputs to stdout.`,
		Example: `  # Export to file
  shelly theme export mytheme.yaml

  # Export to stdout
  shelly theme export`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := ""
			if len(args) > 0 {
				file = args[0]
			}
			return run(f, file)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, file string) error {
	ios := f.IOStreams()

	current := theme.Current()
	if current == nil {
		return fmt.Errorf("no theme is currently set")
	}

	// Build export data
	export := ThemeExport{
		Name: current.ID,
		RenderedColors: RenderedColors{
			Foreground:  colorToHex(theme.Fg()),
			Background:  colorToHex(theme.Bg()),
			Green:       colorToHex(theme.Green()),
			Red:         colorToHex(theme.Red()),
			Yellow:      colorToHex(theme.Yellow()),
			Blue:        colorToHex(theme.Blue()),
			Cyan:        colorToHex(theme.Cyan()),
			Purple:      colorToHex(theme.Purple()),
			BrightBlack: colorToHex(theme.BrightBlack()),
		},
	}

	// Include custom color overrides if any are set
	if custom := theme.GetCustomColors(); custom != nil {
		colors := make(map[string]string)
		if custom.Foreground != "" {
			colors["foreground"] = custom.Foreground
		}
		if custom.Background != "" {
			colors["background"] = custom.Background
		}
		if custom.Green != "" {
			colors["green"] = custom.Green
		}
		if custom.Red != "" {
			colors["red"] = custom.Red
		}
		if custom.Yellow != "" {
			colors["yellow"] = custom.Yellow
		}
		if custom.Blue != "" {
			colors["blue"] = custom.Blue
		}
		if custom.Cyan != "" {
			colors["cyan"] = custom.Cyan
		}
		if custom.Purple != "" {
			colors["purple"] = custom.Purple
		}
		if custom.BrightBlack != "" {
			colors["bright_black"] = custom.BrightBlack
		}
		if len(colors) > 0 {
			export.ColorOverrides = colors
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(export)
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	// Write to file or stdout
	if file == "" {
		ios.Printf("%s", string(data))
	} else {
		if err := os.WriteFile(file, data, 0o600); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		ios.Success("Theme exported to %s", file)
	}

	return nil
}

// colorToHex converts a color.Color to a hex string.
func colorToHex(c interface{ RGBA() (r, g, b, a uint32) }) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}
