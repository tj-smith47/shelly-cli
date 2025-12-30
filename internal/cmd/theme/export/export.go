// Package exportcmd provides the theme export command.
package export

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

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
	export := theme.Export{
		Name: current.ID,
		RenderedColors: theme.RenderedColors{
			Foreground:  theme.ColorToHex(theme.Fg()),
			Background:  theme.ColorToHex(theme.Bg()),
			Green:       theme.ColorToHex(theme.Green()),
			Red:         theme.ColorToHex(theme.Red()),
			Yellow:      theme.ColorToHex(theme.Yellow()),
			Blue:        theme.ColorToHex(theme.Blue()),
			Cyan:        theme.ColorToHex(theme.Cyan()),
			Purple:      theme.ColorToHex(theme.Purple()),
			BrightBlack: theme.ColorToHex(theme.BrightBlack()),
		},
	}

	// Include custom color overrides if any are set
	export.ColorOverrides = theme.BuildColorOverrides(theme.GetCustomColors())

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
