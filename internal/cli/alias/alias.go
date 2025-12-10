// Package alias provides alias management commands.
package alias

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the alias command and its subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage command aliases",
		Long: `Create, list, and manage command aliases.

Aliases allow you to create shortcuts for commonly used commands.
You can also create shell aliases (prefixed with !) that execute
shell commands directly.

Examples:
  # Create a simple alias
  shelly alias set ss "switch status"

  # Create an alias with arguments
  shelly alias set toggle-office "switch toggle office-light $1"

  # Create a shell alias
  shelly alias set notify "!notify-send 'Shelly' '$@'"

  # List all aliases
  shelly alias list

  # Delete an alias
  shelly alias delete ss`,
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newDeleteCommand())
	cmd.AddCommand(newImportCommand())
	cmd.AddCommand(newExportCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all aliases",
		Long:    "Display all configured command aliases.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			aliases := cfg.ListAliases()

			if len(aliases) == 0 {
				fmt.Println("No aliases configured.")
				fmt.Println("\nUse 'shelly alias set <name> <command>' to create an alias.")
				return nil
			}

			switch outputFormat {
			case "json":
				return output.JSON(cmd.OutOrStdout(), aliases)
			case "yaml":
				return output.YAML(cmd.OutOrStdout(), aliases)
			default:
				t := output.NewTable("Name", "Command", "Shell")
				for _, a := range aliases {
					shell := "no"
					if a.Shell {
						shell = "yes"
					}
					t.AddRow(a.Name, a.Command, shell)
				}
				t.Print()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")

	return cmd
}

func newSetCommand() *cobra.Command {
	var shell bool

	cmd := &cobra.Command{
		Use:   "set <name> <command>",
		Short: "Create or update an alias",
		Long: `Create or update a command alias.

The command can include argument placeholders:
  $1, $2, ... - Individual arguments
  $@          - All arguments

Shell aliases (prefixed with !) are executed directly in the shell.

Examples:
  shelly alias set ss "switch status"
  shelly alias set toggle "switch toggle $1"
  shelly alias set notify "!notify-send 'Shelly' '$@'"`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			// Join all args after the name as the command
			command := strings.Join(args[1:], " ")

			cfg := config.Get()

			// Check if updating existing
			existing := cfg.GetAlias(name)
			action := "Created"
			if existing != nil {
				action = "Updated"
			}

			if err := cfg.AddAlias(name, command, shell); err != nil {
				return fmt.Errorf("failed to create alias: %w", err)
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("%s alias '%s' -> '%s'\n", action, name, command)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&shell, "shell", "s", false, "Execute as shell command")

	return cmd
}

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete an alias",
		Long:    "Remove an existing command alias.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg := config.Get()

			if cfg.GetAlias(name) == nil {
				return fmt.Errorf("alias '%s' not found", name)
			}

			cfg.RemoveAlias(name)

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Deleted alias '%s'\n", name)
			return nil
		},
	}

	return cmd
}

func newImportCommand() *cobra.Command {
	var merge bool

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import aliases from a file",
		Long: `Import aliases from a YAML file.

By default, existing aliases with the same name will be overwritten.
Use --merge to skip existing aliases.

File format:
  aliases:
    ss: "switch status"
    toggle: "switch toggle $1"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]
			cfg := config.Get()

			imported, skipped, err := cfg.ImportAliases(filename, merge)
			if err != nil {
				return fmt.Errorf("failed to import aliases: %w", err)
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Imported %d alias(es)", imported)
			if skipped > 0 {
				fmt.Printf(" (skipped %d existing)", skipped)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&merge, "merge", "m", false, "Skip existing aliases instead of overwriting")

	return cmd
}

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export aliases to a file",
		Long: `Export all aliases to a YAML file.

If no file is specified, aliases are written to stdout.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()

			var filename string
			if len(args) > 0 {
				filename = args[0]
			}

			if err := cfg.ExportAliases(filename); err != nil {
				return fmt.Errorf("failed to export aliases: %w", err)
			}

			if filename != "" {
				fmt.Printf("Exported aliases to %s\n", filename)
			}

			return nil
		},
	}

	return cmd
}
