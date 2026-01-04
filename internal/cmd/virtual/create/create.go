// Package create provides the virtual create command.
package create

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	Type    string
	Name    string
	ID      int
	Factory *cmdutil.Factory
}

// NewCommand creates the virtual create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "create <device> <type>",
		Aliases: []string{"add", "new"},
		Short:   "Create a virtual component",
		Long: `Create a new virtual component on a Shelly Gen2+ device.

Available types:
  boolean  - True/false value
  number   - Numeric value with optional min/max
  text     - Text string value
  enum     - Selection from predefined options
  button   - Triggerable button
  group    - Component group

Virtual components are automatically assigned IDs in the range 200-299.`,
		Example: `  # Create a boolean component
  shelly virtual create kitchen boolean --name "Away Mode"

  # Create a number component
  shelly virtual create kitchen number --name "Temperature Offset"

  # Create with specific ID
  shelly virtual create kitchen boolean --name "Override" --id 205`,
		Args: cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completion.DeviceNames()(cmd, args, toComplete)
			}
			if len(args) == 1 {
				return shelly.ValidVirtualTypes, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Type = strings.ToLower(args[1])
			if !shelly.IsValidVirtualType(opts.Type) {
				return fmt.Errorf("invalid type %q, must be one of: %s", opts.Type, strings.Join(shelly.ValidVirtualTypes, ", "))
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Component display name")
	cmd.Flags().IntVar(&opts.ID, "id", 0, "Specific component ID (200-299, auto-assigned if not specified)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	params := shelly.AddVirtualComponentParams{
		Type: shelly.VirtualComponentType(opts.Type),
		Name: opts.Name,
		ID:   opts.ID,
	}

	var id int
	err := cmdutil.RunWithSpinner(ctx, ios, "Creating virtual component...", func(ctx context.Context) error {
		var createErr error
		id, createErr = svc.AddVirtualComponent(ctx, opts.Device, params)
		return createErr
	})
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%d", opts.Type, id)
	if opts.Name != "" {
		ios.Success("Created virtual component %s (%s)", key, opts.Name)
	} else {
		ios.Success("Created virtual component %s", key)
	}

	// Invalidate cached virtual component list
	cmdutil.InvalidateCache(opts.Factory, opts.Device, cache.TypeVirtuals)
	return nil
}
