// Package set provides the alias set command.
package set

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the alias set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var shell bool

	cmd := &cobra.Command{
		Use:     "set <name> <command>",
		Aliases: []string{"add", "create"},
		Short:   "Create or update a command alias",
		Long: `Create or update a command alias.

The command can include argument placeholders:
  - $1, $2, etc. are replaced with positional arguments
  - $@ is replaced with all arguments

Shell aliases (prefixed with ! or using --shell) are executed in your shell.`,
		Example: `  # Simple alias
  shelly alias set lights "batch on living-room kitchen bedroom"

  # Alias with argument interpolation
  shelly alias set sw "switch $1 $2"
  # Usage: shelly sw on kitchen -> switch on kitchen

  # Alias with all arguments
  shelly alias set dev "device $@"
  # Usage: shelly dev info kitchen -> device info kitchen

  # Shell alias
  shelly alias set backup --shell 'tar -czf shelly-backup.tar.gz ~/.config/shelly'`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			command := strings.Join(args[1:], " ")
			return run(f, name, command, shell)
		},
	}

	cmd.Flags().BoolVarP(&shell, "shell", "s", false, "Create a shell alias (executes in shell)")

	return cmd
}

func run(f *cmdutil.Factory, name, command string, shell bool) error {
	ios := f.IOStreams()

	// Detect shell alias from ! prefix
	if strings.HasPrefix(command, "!") {
		shell = true
		command = strings.TrimPrefix(command, "!")
	}

	// Check if updating existing alias
	existing := config.Get().GetAlias(name)

	if err := config.AddAlias(name, command, shell); err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	if existing != nil {
		ios.Success("Updated alias '%s'", name)
	} else {
		ios.Success("Created alias '%s'", name)
	}

	// Show the alias
	aliasType := "command"
	if shell {
		aliasType = "shell"
	}
	ios.Printf("  %s -> %s (%s)\n", name, command, aliasType)

	return nil
}
