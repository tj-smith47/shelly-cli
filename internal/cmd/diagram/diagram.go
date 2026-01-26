// Package diagram provides the diagram command for displaying device wiring diagrams.
package diagram

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/diagram"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// Options holds command options.
type Options struct {
	Factory    *cmdutil.Factory
	Model      string
	Generation string
	Style      string
}

// NewCommand creates the diagram command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "diagram",
		Aliases: []string{"wiring", "diag"},
		Short:   "Display ASCII wiring diagrams for Shelly devices",
		Long: `Display ASCII wiring diagrams showing terminal connections and wiring
layouts for Shelly device models. Useful for installation reference.

Use --style to choose between schematic (circuit-style), compact (minimal box),
or detailed (installer-friendly with annotations) diagram styles.`,
		Example: `  # Show wiring diagram for Shelly Plus 1
  shelly diagram -m plus-1

  # Compact layout for Pro 4PM
  shelly diagram -m pro-4pm -s compact

  # Detailed installer view for Dimmer 2
  shelly diagram -m dimmer-2 -s detailed

  # Disambiguate model with generation
  shelly diagram -m 1 -g 1

  # Show Gen3 variant
  shelly diagram -m 1pm-gen3`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Model, "model", "m", "", "Device model slug (e.g., plus-1, pro-4pm, dimmer-2)")
	cmd.Flags().StringVarP(&opts.Generation, "generation", "g", "", "Device generation (1, 2, 3, 4, gen1, gen2, gen3, gen4)")
	cmd.Flags().StringVarP(&opts.Style, "style", "s", "schematic", "Diagram style (schematic, compact, detailed)")

	utils.Must(cmd.MarkFlagRequired("model"))

	// Flag completions
	utils.Must(cmd.RegisterFlagCompletionFunc("model", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return diagram.ModelSlugs(0), cobra.ShellCompDirectiveNoFileComp
	}))
	utils.Must(cmd.RegisterFlagCompletionFunc("generation", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return diagram.ValidGenerations(), cobra.ShellCompDirectiveNoFileComp
	}))
	utils.Must(cmd.RegisterFlagCompletionFunc("style", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return diagram.ValidStyles(), cobra.ShellCompDirectiveNoFileComp
	}))

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Parse generation filter (if provided)
	var gen int
	if opts.Generation != "" {
		var err error
		gen, err = diagram.ParseGeneration(opts.Generation)
		if err != nil {
			return err
		}
	}

	// Parse style
	style, err := diagram.ParseStyle(opts.Style)
	if err != nil {
		return err
	}

	// Look up model
	model, err := diagram.LookupModel(opts.Model, gen)
	if err != nil {
		return err
	}

	// Render and display
	renderer := diagram.NewRenderer(style)
	rendered := renderer.Render(model)
	term.DisplayDiagram(ios, rendered)

	// Show generation note when the slug matches multiple generations
	if gen == 0 {
		gens := diagram.ModelGenerations(opts.Model)
		if len(gens) > 1 {
			term.DisplayDiagramGenerationNote(ios, model.Generation, gens)
		}
	}

	return nil
}
