// Package apply provides the template apply subcommand.
package apply

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Template string
	Device   string
	DryRun   bool
	Factory  *cmdutil.Factory
}

// NewCommand creates the template apply command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "apply <template> <device>",
		Aliases: []string{"set", "push"},
		Short:   "Apply a template to a device",
		Long: `Apply a saved configuration template to a device.

The template configuration will be merged with the device's current
settings. Use --dry-run to preview changes without applying them.

Note: Only devices of the same model/generation are fully compatible.`,
		Example: `  # Apply a template to a device
  shelly template apply my-config bedroom

  # Preview changes without applying
  shelly template apply my-config bedroom --dry-run

  # Apply without confirmation
  shelly template apply my-config bedroom --yes`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.TemplateThenDevice(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Template = args[0]
			opts.Device = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview changes without applying")
	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Get template
	tpl, exists := config.GetDeviceTemplate(opts.Template)
	if !exists {
		return fmt.Errorf("template %q not found", opts.Template)
	}

	// Warn about model/generation compatibility
	info, err := svc.GetDeviceInfo(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to get device info: %w", err)
	}

	if !config.IsCompatibleModel(tpl, info.Model) {
		ios.Warning("Template was created for %s but device is %s", tpl.Model, info.Model)
	}
	if !config.IsCompatibleGeneration(tpl, info.Generation) {
		ios.Warning("Template was created for Gen%d but device is Gen%d", tpl.Generation, info.Generation)
	}

	// Handle dry run mode
	if opts.DryRun {
		var changes []string
		err := cmdutil.RunWithSpinner(ctx, ios, "Comparing configurations...", func(ctx context.Context) error {
			var dryRunErr error
			changes, dryRunErr = svc.ApplyTemplate(ctx, opts.Device, tpl.Config, true)
			return dryRunErr
		})
		if err != nil {
			return err
		}

		if len(changes) == 0 {
			ios.Info("No changes would be made")
			return nil
		}

		ios.Title("Changes that would be applied")
		ios.Println()
		for _, change := range changes {
			ios.Printf("  %s\n", change)
		}
		ios.Printf("\n%d change(s) would be applied\n", len(changes))
		return nil
	}

	// Confirm unless --yes
	confirmed, err := opts.Factory.ConfirmAction(fmt.Sprintf("Apply template %q to device %q?", opts.Template, opts.Device), opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Cancelled")
		return nil
	}

	// Apply template
	var changes []string
	err = cmdutil.RunWithSpinner(ctx, ios, "Applying template...", func(ctx context.Context) error {
		var applyErr error
		changes, applyErr = svc.ApplyTemplate(ctx, opts.Device, tpl.Config, false)
		return applyErr
	})
	if err != nil {
		return err
	}

	ios.Success("Template %q applied to %s", opts.Template, opts.Device)
	for _, change := range changes {
		ios.Printf("  %s\n", change)
	}

	return nil
}
