// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// ConfigExportOpts configures a config-based export command (scene, template, etc.).
type ConfigExportOpts[T any] struct {
	// Component is the type being exported (e.g., "scene", "template").
	Component string

	// Aliases are alternate command names.
	Aliases []string

	// Short is the short description.
	Short string

	// Long is the detailed description.
	Long string

	// Example shows usage examples.
	Example string

	// ValidArgsFunc provides completion for the item name argument.
	ValidArgsFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

	// Fetcher retrieves the item by name from config.
	// Returns the item and whether it exists.
	Fetcher func(name string) (T, bool)

	// DefaultFormat is the default output format (output.FormatJSON or output.FormatYAML).
	DefaultFormat output.Format
}

// configExportOptions holds runtime options for a config export command.
type configExportOptions[T any] struct {
	flags.OutputFlags
	Name    string
	File    string
	Factory *cmdutil.Factory
	Config  ConfigExportOpts[T]
}

// NewConfigExportCommand creates a config-based export command.
func NewConfigExportCommand[T any](f *cmdutil.Factory, config ConfigExportOpts[T]) *cobra.Command {
	opts := &configExportOptions[T]{
		Factory: f,
		Config:  config,
	}

	// Set default format
	defaultFormat := string(config.DefaultFormat)
	if defaultFormat == "" {
		defaultFormat = "yaml"
	}

	cmd := &cobra.Command{
		Use:               "export <name> [file]",
		Aliases:           config.Aliases,
		Short:             config.Short,
		Long:              config.Long,
		Example:           config.Example,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: config.ValidArgsFunc,
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Name = args[0]
			if len(args) > 1 {
				opts.File = args[1]
			}
			return runConfigExport(opts)
		},
	}

	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, defaultFormat, "json", "yaml")

	return cmd
}

func runConfigExport[T any](opts *configExportOptions[T]) error {
	ios := opts.Factory.IOStreams()

	// Fetch item
	item, exists := opts.Config.Fetcher(opts.Name)
	if !exists {
		return fmt.Errorf("%s %q not found", opts.Config.Component, opts.Name)
	}

	// Determine format from file extension if file specified
	format, err := output.ParseFormat(opts.Format)
	if err != nil {
		return fmt.Errorf("unsupported format: %s (use json or yaml)", opts.Format)
	}

	if opts.File != "" {
		ext := strings.ToLower(filepath.Ext(opts.File))
		switch ext {
		case ".json":
			format = output.FormatJSON
		case ".yaml", ".yml":
			format = output.FormatYAML
		}
	}

	// Use output.ExportToFile for consistent behavior
	if err := output.ExportToFile(ios, item, opts.File, format, strings.ToUpper(string(format))); err != nil {
		return err
	}

	// ExportToFile handles the success message for file output
	// For stdout, we're done
	return nil
}
