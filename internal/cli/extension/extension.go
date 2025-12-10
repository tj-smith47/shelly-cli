// Package extension provides plugin/extension management commands.
package extension

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// NewCommand creates the extension command and its subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"ext", "plugin"},
		Short:   "Manage shelly extensions",
		Long: `Manage shelly extensions (plugins).

Extensions are executables named 'shelly-*' that extend the CLI functionality.
They can be installed in ~/.config/shelly/plugins/ or anywhere in your PATH.

When you run 'shelly foo', if 'foo' is not a built-in command, shelly will
look for an extension named 'shelly-foo' and execute it.

Extensions receive environment variables with configuration:
  SHELLY_CONFIG_PATH    - Path to the config file
  SHELLY_OUTPUT_FORMAT  - Current output format
  SHELLY_NO_COLOR       - "1" if color is disabled
  SHELLY_VERBOSE        - "1" if verbose mode is enabled
  SHELLY_DEVICES_JSON   - JSON of registered devices`,
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newInstallCommand())
	cmd.AddCommand(newRemoveCommand())
	cmd.AddCommand(newExecCommand())
	cmd.AddCommand(newCreateCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	var all bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed extensions",
		Long: `List installed extensions.

By default, shows only extensions installed in ~/.config/shelly/plugins/.
Use --all to show all extensions found in PATH.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var pluginList []plugins.Plugin
			var err error

			if all {
				loader := plugins.NewLoader()
				pluginList, err = loader.Discover()
			} else {
				registry, regErr := plugins.NewRegistry()
				if regErr != nil {
					return fmt.Errorf("failed to access plugin registry: %w", regErr)
				}
				pluginList, err = registry.List()
			}

			if err != nil {
				return fmt.Errorf("failed to list extensions: %w", err)
			}

			if len(pluginList) == 0 {
				if all {
					fmt.Println("No extensions found.")
				} else {
					fmt.Println("No extensions installed.")
					fmt.Println("\nUse 'shelly extension list --all' to see all available extensions.")
				}
				return nil
			}

			switch outputFormat {
			case "json":
				return output.JSON(cmd.OutOrStdout(), pluginList)
			case "yaml":
				return output.YAML(cmd.OutOrStdout(), pluginList)
			default:
				t := output.NewTable("Name", "Version", "Path")
				for _, p := range pluginList {
					version := p.Version
					if version == "" {
						version = "-"
					}
					t.AddRow(p.Name, version, p.Path)
				}
				t.Print()
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "List all extensions in PATH")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")

	return cmd
}

func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <path>",
		Short: "Install an extension from a local file",
		Long: `Install an extension from a local file.

The file must be an executable named 'shelly-<name>'.

Examples:
  shelly extension install ./shelly-myplugin
  shelly extension install ~/downloads/shelly-notify`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourcePath := args[0]

			// Expand path
			if strings.HasPrefix(sourcePath, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}
				sourcePath = filepath.Join(home, sourcePath[1:])
			}

			// Resolve to absolute path
			absPath, err := filepath.Abs(sourcePath)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}

			// Check file exists
			if _, err := os.Stat(absPath); err != nil {
				return fmt.Errorf("file not found: %s", absPath)
			}

			registry, err := plugins.NewRegistry()
			if err != nil {
				return fmt.Errorf("failed to access plugin registry: %w", err)
			}

			if err := registry.Install(absPath); err != nil {
				return fmt.Errorf("failed to install extension: %w", err)
			}

			name := strings.TrimPrefix(filepath.Base(absPath), plugins.PluginPrefix)
			fmt.Printf("Installed extension '%s'\n", name)
			return nil
		},
	}

	return cmd
}

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "uninstall"},
		Short:   "Remove an installed extension",
		Long:    "Remove an extension installed in ~/.config/shelly/plugins/.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			registry, err := plugins.NewRegistry()
			if err != nil {
				return fmt.Errorf("failed to access plugin registry: %w", err)
			}

			if err := registry.Remove(name); err != nil {
				return err
			}

			fmt.Printf("Removed extension '%s'\n", name)
			return nil
		},
	}

	return cmd
}

func newExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <name> [args...]",
		Short: "Execute an extension explicitly",
		Long: `Execute an extension explicitly.

This is useful for running extensions that might conflict with
built-in command names.

Example:
  shelly extension exec myplugin --flag value`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			pluginArgs := args[1:]

			loader := plugins.NewLoader()
			plugin, err := loader.Find(name)
			if err != nil {
				return fmt.Errorf("error finding extension: %w", err)
			}
			if plugin == nil {
				return fmt.Errorf("extension '%s' not found", name)
			}

			executor := plugins.NewExecutor()
			return executor.Execute(plugin, pluginArgs)
		},
	}

	return cmd
}

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new extension scaffold",
		Long: `Create a new extension scaffold.

This creates a basic shell script extension template in the current directory.

Example:
  shelly extension create myplugin`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			filename := plugins.PluginPrefix + name

			// Check if file already exists
			if _, err := os.Stat(filename); err == nil {
				return fmt.Errorf("file '%s' already exists", filename)
			}

			content := fmt.Sprintf(`#!/bin/bash
# shelly-%s - A shelly extension
#
# This extension receives environment variables from shelly:
#   SHELLY_CONFIG_PATH    - Path to the config file
#   SHELLY_OUTPUT_FORMAT  - Current output format (json, yaml, table, text)
#   SHELLY_NO_COLOR       - "1" if color is disabled
#   SHELLY_VERBOSE        - "1" if verbose mode is enabled
#   SHELLY_DEVICES_JSON   - JSON of registered devices

# Handle --version flag
if [[ "$1" == "--version" ]]; then
    echo "%s 0.1.0"
    exit 0
fi

# Handle --help flag
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Usage: shelly %s [options]"
    echo ""
    echo "A custom shelly extension."
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --version      Show version"
    exit 0
fi

# Your extension logic here
echo "Hello from %s extension!"
echo "Received arguments: $@"
`, name, name, name, name)

			if err := os.WriteFile(filename, []byte(content), 0o755); err != nil {
				return fmt.Errorf("failed to create extension: %w", err)
			}

			fmt.Printf("Created extension scaffold: %s\n", filename)
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Edit the script to add your functionality")
			fmt.Printf("  2. Test it: ./%s\n", filename)
			fmt.Printf("  3. Install it: shelly extension install %s\n", filename)

			return nil
		},
	}

	return cmd
}
