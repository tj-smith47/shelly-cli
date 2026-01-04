// Package create provides the script create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	Name    string
	Code    string
	File    string
	Enable  bool
}

// NewCommand creates the script create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "create <device>",
		Aliases: []string{"new"},
		Short:   "Create a new script",
		Long: `Create a new script on a Gen2+ Shelly device.

You can provide the script code inline with --code or from a file with --file.
Use --enable to automatically enable the script after creation.`,
		Example: `  # Create an empty script
  shelly script create living-room --name "My Script"

  # Create with inline code
  shelly script create living-room --name "Hello" --code "print('Hello!');" --enable

  # Create from file
  shelly script create living-room --name "Auto Light" --file auto-light.js --enable`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Script name")
	cmd.Flags().StringVar(&opts.Code, "code", "", "Script code (inline)")
	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Script code file")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable script after creation")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	// Get code from file if specified
	code := opts.Code
	if opts.File != "" {
		data, err := afero.ReadFile(config.Fs(), opts.File)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		code = string(data)
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Creating script...", func(ctx context.Context) error {
		// Create the script
		id, createErr := svc.CreateScript(ctx, opts.Device, opts.Name)
		if createErr != nil {
			return fmt.Errorf("failed to create script: %w", createErr)
		}

		// Upload code if provided
		if code != "" {
			if uploadErr := svc.UpdateScriptCode(ctx, opts.Device, id, code, false); uploadErr != nil {
				return fmt.Errorf("failed to upload code: %w", uploadErr)
			}
		}

		// Enable if requested
		if opts.Enable {
			enable := true
			if configErr := svc.UpdateScriptConfig(ctx, opts.Device, id, nil, &enable); configErr != nil {
				return fmt.Errorf("failed to enable script: %w", configErr)
			}
		}

		ios.Success("Created script %d", id)
		if opts.Name != "" {
			ios.Info("Name: %s", opts.Name)
		}
		if code != "" {
			ios.Info("Code uploaded (%d bytes)", len(code))
		}
		if opts.Enable {
			ios.Info("Script enabled")
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cached script list
	cmdutil.InvalidateCache(opts.Factory, opts.Device, cache.TypeScripts)
	return nil
}
