// Package diff provides the template diff subcommand.
package diff

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Template string
	Device   string
	Factory  *cmdutil.Factory
}

// NewCommand creates the template diff command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "diff <template> <device>",
		Aliases: []string{"compare", "cmp"},
		Short:   "Compare a template with a device",
		Long: `Compare a saved configuration template with a device's current configuration.

Shows differences between the template and the device, highlighting
what would change if the template were applied.`,
		Example: `  # Compare template with device
  shelly template diff my-config bedroom

  # Output as JSON
  shelly template diff my-config bedroom -o json`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.TemplateThenDevice(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Template = args[0]
			opts.Device = args[1]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Get template
	tpl, exists := config.GetTemplate(opts.Template)
	if !exists {
		return fmt.Errorf("template %q not found", opts.Template)
	}

	// Compare template with device
	var diffs []model.ConfigDiff
	err := cmdutil.RunWithSpinner(ctx, ios, "Comparing configurations...", func(ctx context.Context) error {
		var compareErr error
		diffs, compareErr = svc.CompareTemplate(ctx, opts.Device, tpl.Config)
		return compareErr
	})
	if err != nil {
		return err
	}

	// Handle output formats
	if output.WantsStructured() {
		return cmdutil.PrintListResult(ios, diffs, nil)
	}

	term.DisplayTemplateDiffs(ios, opts.Template, opts.Device, diffs)
	return nil
}
