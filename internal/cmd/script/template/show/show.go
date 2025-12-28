// Package show provides the script template show subcommand.
package show

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Name    string
	Code    bool
	Factory *cmdutil.Factory
}

// NewCommand creates the script template show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "show <name>",
		Aliases: []string{"get", "view"},
		Short:   "Show script template details",
		Long: `Show details of a script template including its code.

Displays the template metadata, configurable variables, and the
JavaScript source code.`,
		Example: `  # Show template details
  shelly script template show motion-light

  # Show only the code (for piping)
  shelly script template show motion-light --code

  # Output as JSON
  shelly script template show motion-light -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ScriptTemplateNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Name = args[0]
			return run(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Code, "code", false, "Show only the code (for piping)")
	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Get template
	tpl, ok := automation.GetScriptTemplate(opts.Name)
	if !ok {
		return fmt.Errorf("script template %q not found", opts.Name)
	}

	// Code-only mode for piping
	if opts.Code {
		ios.Printf("%s", tpl.Code)
		return nil
	}

	// Handle output formats
	if output.WantsStructured() {
		return cmdutil.PrintResult(ios, tpl, nil)
	}

	term.DisplayScriptTemplate(ios, tpl)
	return nil
}
