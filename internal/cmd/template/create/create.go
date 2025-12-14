// Package create provides the template create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Name        string
	Device      string
	Description string
	IncludeWiFi bool
	Force       bool
	Factory     *cmdutil.Factory
}

// NewCommand creates the template create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "create <name> <device>",
		Aliases: []string{"new", "save"},
		Short:   "Create a template from a device",
		Long: `Create a configuration template by capturing settings from a device.

The template captures:
  - Device configuration and settings
  - Component configurations (switches, lights, etc.)
  - Schedules and webhooks
  - Script configurations (not code)

WiFi credentials are excluded by default for security.
Use --include-wifi to include them.`,
		Example: `  # Create a template from a device
  shelly template create my-config living-room

  # Create with description
  shelly template create my-config living-room --description "Standard switch config"

  # Include WiFi settings
  shelly template create my-config living-room --include-wifi

  # Overwrite existing template
  shelly template create my-config living-room --force`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completeNameThenDevice(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			opts.Device = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Template description")
	cmd.Flags().BoolVar(&opts.IncludeWiFi, "include-wifi", false, "Include WiFi credentials (security risk)")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing template")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate template name
	if err := config.ValidateTemplateName(opts.Name); err != nil {
		return err
	}

	// Check if template exists
	if _, exists := config.GetTemplate(opts.Name); exists && !opts.Force {
		return fmt.Errorf("template %q already exists (use --force to overwrite)", opts.Name)
	}

	// Capture device configuration using service method
	var deviceTpl *shelly.DeviceTemplate
	err := cmdutil.RunWithSpinner(ctx, ios, "Capturing device configuration...", func(ctx context.Context) error {
		var captureErr error
		deviceTpl, captureErr = svc.CaptureTemplate(ctx, opts.Device, opts.IncludeWiFi)
		return captureErr
	})
	if err != nil {
		return err
	}

	// Save template - if force flag and exists, delete first
	if opts.Force {
		if _, exists := config.GetTemplate(opts.Name); exists {
			if delErr := config.DeleteTemplate(opts.Name); delErr != nil {
				return fmt.Errorf("failed to delete existing template: %w", delErr)
			}
		}
	}

	err = config.CreateTemplate(
		opts.Name,
		opts.Description,
		deviceTpl.Model,
		deviceTpl.App,
		deviceTpl.Generation,
		deviceTpl.Config,
		opts.Device,
	)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	ios.Success("Template %q created from %s (%s)", opts.Name, opts.Device, deviceTpl.Model)
	return nil
}

// completeNameThenDevice provides completion for name and device arguments.
func completeNameThenDevice() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: template name (no completion)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			// Second argument: device names
			devices := config.ListDevices()
			completions := make([]string, 0, len(devices))
			for name := range devices {
				completions = append(completions, name)
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
