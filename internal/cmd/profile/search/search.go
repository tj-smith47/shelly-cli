// Package search provides the profile search command.
package search

import (
	"sort"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/profiles"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Query      string
	Capability string
	Protocol   string
	Factory    *cmdutil.Factory
}

// NewCommand creates the profile search command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "search <query>",
		Aliases: []string{"find", "s"},
		Short:   "Search device profiles",
		Long: `Search for device profiles by name, model, or feature.

Search by text query, or filter by capability or protocol.`,
		Example: `  # Search by name
  shelly profile search "plug"

  # Find devices with dimming
  shelly profile search --capability dimming

  # Find Z-Wave devices
  shelly profile search --protocol zwave

  # Combine filters
  shelly profile search --capability power_metering --protocol mqtt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Query = args[0]
			}
			return run(opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)
	cmd.Flags().StringVar(&opts.Capability, "capability", "", "Filter by capability (e.g., dimming, scripting, power_metering)")
	cmd.Flags().StringVar(&opts.Protocol, "protocol", "", "Filter by protocol (e.g., mqtt, ble, zwave, matter)")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	var result []*profiles.Profile

	switch {
	case opts.Query != "":
		result = profiles.Search(opts.Query)
	case opts.Capability != "":
		result = profiles.ListByCapability(opts.Capability)
	case opts.Protocol != "":
		result = profiles.ListByProtocol(opts.Protocol)
	default:
		result = profiles.List()
	}

	// Apply additional filters
	if opts.Capability != "" && opts.Query != "" {
		result = term.FilterProfilesByCapability(result, opts.Capability)
	}
	if opts.Protocol != "" && (opts.Query != "" || opts.Capability != "") {
		result = term.FilterProfilesByProtocol(result, opts.Protocol)
	}

	// Sort by model
	sort.Slice(result, func(i, j int) bool {
		return result[i].Model < result[j].Model
	})

	if len(result) == 0 {
		ios.Info("No profiles found matching your criteria")
		return nil
	}

	return cmdutil.PrintListResult(ios, result, func(ios *iostreams.IOStreams, items []*profiles.Profile) {
		table := output.NewTable("Model", "Name", "Generation", "Series", "Form Factor")
		for _, p := range items {
			table.AddRow(
				p.Model,
				p.Name,
				p.Generation.String(),
				string(p.Series),
				string(p.FormFactor),
			)
		}
		if err := table.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print table", err)
		}
		ios.Printf("\nFound %d profile(s)\n", len(items))
	})
}
