// Package list provides the profile list command.
package list

import (
	"sort"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/profiles"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Generation string
	Series     string
	Factory    *cmdutil.Factory
}

// NewCommand creates the profile list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List device profiles",
		Long: `List all known Shelly device profiles.

Optionally filter by generation or series.`,
		Example: `  # List all profiles
  shelly profile list

  # List Gen2 devices
  shelly profile list --gen gen2

  # List Pro series devices
  shelly profile list --series pro

  # JSON output
  shelly profile list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)
	cmd.Flags().StringVar(&opts.Generation, "gen", "", "Filter by generation (gen1, gen2, gen3, gen4)")
	cmd.Flags().StringVar(&opts.Series, "series", "", "Filter by series (classic, plus, pro, mini, blu, wave)")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	var result []*profiles.Profile

	switch {
	case opts.Generation != "":
		gen := term.ParseProfileGeneration(opts.Generation)
		if gen == types.GenerationUnknown {
			ios.Warning("Unknown generation: %s", opts.Generation)
			return nil
		}
		result = profiles.ListByGeneration(gen)
	case opts.Series != "":
		series := term.ParseProfileSeries(opts.Series)
		if series == "" {
			ios.Warning("Unknown series: %s", opts.Series)
			return nil
		}
		result = profiles.ListBySeries(series)
	default:
		result = profiles.List()
	}

	// Sort by model
	sort.Slice(result, func(i, j int) bool {
		return result[i].Model < result[j].Model
	})

	return cmdutil.PrintListResult(ios, result, func(ios *iostreams.IOStreams, items []*profiles.Profile) {
		builder := table.NewBuilder("Model", "Name", "Generation", "Series", "Form Factor")
		for _, p := range items {
			builder.AddRow(
				p.Model,
				p.Name,
				p.Generation.String(),
				string(p.Series),
				string(p.FormFactor),
			)
		}
		table := builder.WithModeStyle(ios).Build()
		if err := table.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print table", err)
		}
	})
}
