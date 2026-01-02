// Package list provides the extension list command.
package list

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// Options holds the options for the list command.
type Options struct {
	Factory *cmdutil.Factory
	All     bool
}

// NewCommand creates the extension list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List installed extensions",
		Long: `List installed extensions.

Extensions are external plugins that add new commands to the CLI. They
are standalone executables named 'shelly-*' that can be installed from
git repositories or created locally.

By default, only shows extensions installed in the user plugins directory.
Use --all to show all discovered extensions including those in PATH.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: Name, Version, Source (github/url/local/unknown), Path`,
		Example: `  # List installed extensions
  shelly extension list

  # List all discovered extensions
  shelly extension list --all

  # Output as JSON
  shelly extension list -o json

  # Get extension names only
  shelly extension list -o json | jq -r '.[].name'

  # Find extensions without versions
  shelly extension list -o json | jq '.[] | select(.version == "")'

  # Short form
  shelly ext ls`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "List all discovered extensions, not just installed ones")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	var extensionList []plugins.Plugin
	var err error

	if opts.All {
		// List all discovered extensions
		loader := plugins.NewLoader()
		extensionList, err = loader.Discover()
	} else {
		// List only installed extensions
		registry, rerr := plugins.NewRegistry()
		if rerr != nil {
			return rerr
		}
		extensionList, err = registry.List()
	}

	if err != nil {
		return err
	}

	if len(extensionList) == 0 {
		if opts.All {
			ios.Info("No extensions found")
		} else {
			ios.Info("No extensions installed. Use 'shelly extension install' to install one.")
		}
		return nil
	}

	// Handle JSON/YAML output
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, extensionList)
	}

	// Table output
	builder := table.NewBuilder("Name", "Version", "Source", "Path")
	for _, ext := range extensionList {
		version := ext.Version
		if version == "" {
			version = "-"
		}
		source := "-"
		if ext.Manifest != nil {
			source = ext.Manifest.Source.Type
		}
		builder.AddRow(ext.Name, version, source, ext.Path)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print extension list table", err)
	}
	ios.Println()
	ios.Count("extension", len(extensionList))

	return nil
}
