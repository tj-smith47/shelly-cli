// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// ConfigListOpts configures a config-based list command.
type ConfigListOpts[T any] struct {
	// Resource name (e.g., "alias", "group", "scene", "template").
	Resource string

	// FetchFunc retrieves items from config.
	FetchFunc func() []T

	// DisplayFunc renders table output.
	DisplayFunc func(ios *iostreams.IOStreams, items []T)

	// EmptyMsg shown when no items exist.
	// If empty, defaults to "No <resource>s configured".
	EmptyMsg string

	// HintMsg shown after empty message to suggest next action.
	HintMsg string
}

// NewConfigListCommand creates a config-based list command.
func NewConfigListCommand[T any](f *Factory, opts ConfigListOpts[T]) *cobra.Command {
	short := fmt.Sprintf("List %ss", opts.Resource)
	long := fmt.Sprintf(`List all configured %ss.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.`, opts.Resource)

	examples := fmt.Sprintf(`  # List all %ss
  shelly %s list

  # Output as JSON
  shelly %s list -o json

  # Output as YAML
  shelly %s list -o yaml`, opts.Resource, opts.Resource, opts.Resource, opts.Resource)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   short,
		Long:    long,
		Example: examples,
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigList(f, opts)
		},
	}

	return cmd
}

func runConfigList[T any](f *Factory, opts ConfigListOpts[T]) error {
	ios := f.IOStreams()
	items := opts.FetchFunc()

	if len(items) == 0 {
		msg := opts.EmptyMsg
		if msg == "" {
			msg = fmt.Sprintf("No %ss configured", opts.Resource)
		}
		ios.Info("%s", msg)
		if opts.HintMsg != "" {
			ios.Info("%s", opts.HintMsg)
		}
		return nil
	}

	// Handle structured output (JSON/YAML/template)
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, items)
	}

	// Table output
	opts.DisplayFunc(ios, items)
	return nil
}
