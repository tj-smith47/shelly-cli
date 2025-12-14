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
	ID          string     `yaml:"id" json:"id"`
	DisplayName string     `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Colors      ThemeColor `yaml:"colors" json:"colors"`
}

// ThemeColor represents the color values of a theme.
type ThemeColor struct {
	Foreground  string `yaml:"foreground" json:"foreground"`
	Background  string `yaml:"background" json:"background"`
	Black       string `yaml:"black" json:"black"`
	Red         string `yaml:"red" json:"red"`
	Green       string `yaml:"green" json:"green"`
	Yellow      string `yaml:"yellow" json:"yellow"`
	Blue        string `yaml:"blue" json:"blue"`
	Purple      string `yaml:"purple" json:"purple"`
	Cyan        string `yaml:"cyan" json:"cyan"`
	White       string `yaml:"white" json:"white"`
	BrightBlack string `yaml:"bright_black" json:"bright_black"`
}

// NewCommand creates the theme export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [file]",
		Aliases: []string{"exp", "save"},
		Short:   "Export current theme",
		Long: `Export the current theme configuration to a file.

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
		ID:          current.ID,
		DisplayName: current.DisplayName,
		Colors: ThemeColor{
			Foreground:  colorToHex(current.Fg),
			Background:  colorToHex(current.Bg),
			Black:       colorToHex(current.Black),
			Red:         colorToHex(current.Red),
			Green:       colorToHex(current.Green),
			Yellow:      colorToHex(current.Yellow),
			Blue:        colorToHex(current.Blue),
			Purple:      colorToHex(current.Purple),
			Cyan:        colorToHex(current.Cyan),
			White:       colorToHex(current.White),
			BrightBlack: colorToHex(current.BrightBlack),
		},
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

// colorToHex converts a tint.Color to a hex string.
func colorToHex(c interface{ RGBA() (r, g, b, a uint32) }) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}
