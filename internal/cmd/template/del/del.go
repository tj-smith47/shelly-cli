// Package del provides the template delete subcommand.
package del

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds command options.
type Options struct {
	Template string
	Yes      bool
	Factory  *cmdutil.Factory
}

// NewCommand creates the template delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "delete <template>",
		Aliases: []string{"del", "rm", "remove"},
		Short:   "Delete a template",
		Long:    `Delete a saved configuration template.`,
		Example: `  # Delete a template
  shelly template delete my-config

  # Delete without confirmation
  shelly template delete my-config --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TemplateNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Template = args[0]
			return run(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Check if template exists
	tpl, exists := config.GetTemplate(opts.Template)
	if !exists {
		return fmt.Errorf("template %q not found", opts.Template)
	}

	// Confirm unless --yes
	if !opts.Yes {
		msg := fmt.Sprintf("Delete template %q (model: %s)?", tpl.Name, tpl.Model)
		confirmed, err := ios.Confirm(msg, false)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Info("Cancelled")
			return nil
		}
	}

	// Delete template
	if err := config.DeleteTemplate(opts.Template); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	ios.Success("Template %q deleted", opts.Template)
	return nil
}
