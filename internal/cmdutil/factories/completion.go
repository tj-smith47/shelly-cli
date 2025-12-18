// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// CompletionGenerator generates shell completion script.
type CompletionGenerator func(cmd *cobra.Command, w io.Writer) error

// CompletionOpts configures a shell completion command.
type CompletionOpts struct {
	// Shell is the shell name (e.g., "bash", "zsh", "fish", "powershell").
	Shell string

	// Aliases are the command aliases (e.g., []string{"b"}).
	Aliases []string

	// Long is the detailed description with shell-specific setup instructions.
	Long string

	// Example shows usage examples.
	Example string

	// Generator produces the completion script.
	Generator CompletionGenerator
}

// NewCompletionCommand creates a shell completion command.
// This factory consolidates the common pattern across bash, zsh, fish, powershell completion commands.
func NewCompletionCommand(f *cmdutil.Factory, opts CompletionOpts) *cobra.Command {
	short := fmt.Sprintf("Generate %s completion script", opts.Shell)

	cmd := &cobra.Command{
		Use:                   opts.Shell,
		Aliases:               opts.Aliases,
		Short:                 short,
		Long:                  opts.Long,
		Example:               opts.Example,
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return opts.Generator(cmd.Root(), f.IOStreams().Out)
		},
	}

	return cmd
}
