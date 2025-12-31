// Package set provides the alias set command.
package set

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the set command.
type Options struct {
	Factory *cmdutil.Factory
	Command string
	Name    string
	Shell   bool
}

// NewCommand creates the alias set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Name = args[0]
			opts.Command = strings.Join(args[1:], " ")
			return run(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Shell, "shell", "s", false, "Create a shell alias (executes in shell)")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	command := opts.Command
	shell := opts.Shell

	// Detect shell alias from ! prefix
	if strings.HasPrefix(command, "!") {
		shell = true
		command = strings.TrimPrefix(command, "!")
	}

	// Check if updating existing alias
	_, isUpdate := config.GetAlias(opts.Name)

	if err := config.AddAlias(opts.Name, command, shell); err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	if isUpdate {
		ios.Success("Updated alias '%s'", opts.Name)
	} else {
		ios.Success("Created alias '%s'", opts.Name)
	}

	// Show the alias
	aliasType := "command"
	if shell {
		aliasType = "shell"
	}
	ios.Printf("  %s -> %s (%s)\n", opts.Name, command, aliasType)

	return nil
}
