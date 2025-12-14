// Package list provides the extension list command.
package list

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// NewCommand creates the extension list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List installed extensions",
		Long: `List installed extensions.

By default, only shows extensions installed in the user plugins directory.
Use --all to show all discovered extensions including those in PATH.`,
		Example: `  # List installed extensions
  shelly extension list

  # List all discovered extensions
  shelly extension list --all`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f, all)
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "List all discovered extensions, not just installed ones")

	return cmd
}

func run(f *cmdutil.Factory, all bool) error {
	ios := f.IOStreams()

	var extensionList []plugins.Plugin
	var err error

	if all {
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
		if all {
			ios.Info("No extensions found")
		} else {
			ios.Info("No extensions installed. Use 'shelly extension install' to install one.")
		}
		return nil
	}

	// Handle JSON/YAML output
	if cmdutil.WantsStructured() {
		return cmdutil.FormatOutput(ios, extensionList)
	}

	// Table output
	table := output.NewTable("Name", "Version", "Path")
	for _, ext := range extensionList {
		version := ext.Version
		if version == "" {
			version = "-"
		}
		table.AddRow(ext.Name, version, ext.Path)
	}

	table.Print()
	ios.Count("extension", len(extensionList))

	return nil
}
