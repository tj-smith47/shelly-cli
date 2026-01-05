// Package update provides the script update subcommand.
package update

import (
	"context"
	"fmt"
	"strconv"

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
	ID      int
	Name    string
	Code    string
	File    string
	Append  bool
	Enable  bool
}

// NewCommand creates the script update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "update <device> <id>",
		Aliases: []string{"up"},
		Short:   "Update a script",
		Long: `Update an existing script on a Gen2+ Shelly device.

You can update the script name, code, or enabled status.
Use --append to add code to the existing script instead of replacing it.`,
		Example: `  # Update script name
  shelly script update living-room 1 --name "New Name"

  # Update script code
  shelly script update living-room 1 --file script.js

  # Append code to existing script
  shelly script update living-room 1 --code "// More code" --append

  # Enable/disable script
  shelly script update living-room 1 --enable`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceThenScriptID(),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid script ID: %s", args[1])
			}
			opts.Device = args[0]
			opts.ID = id
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Script name")
	cmd.Flags().StringVar(&opts.Code, "code", "", "Script code (inline)")
	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Script code file")
	cmd.Flags().BoolVar(&opts.Append, "append", false, "Append code instead of replacing")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable the script")

	return cmd
}

//nolint:gocyclo,nestif // Complexity from handling multiple optional update fields in one function
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

	// Check if anything was specified
	hasCode := code != ""
	hasConfig := opts.Name != "" || opts.Enable
	if !hasCode && !hasConfig {
		ios.Warning("No changes specified")
		return nil
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Updating script...", func(ctx context.Context) error {
		// Update code if provided
		if hasCode {
			if uploadErr := svc.UpdateScriptCode(ctx, opts.Device, opts.ID, code, opts.Append); uploadErr != nil {
				return fmt.Errorf("failed to update code: %w", uploadErr)
			}
			if opts.Append {
				ios.Info("Code appended (%d bytes)", len(code))
			} else {
				ios.Info("Code updated (%d bytes)", len(code))
			}
		}

		// Update config if name or enable specified
		if hasConfig {
			var namePtr *string
			var enablePtr *bool
			if opts.Name != "" {
				namePtr = &opts.Name
			}
			if opts.Enable {
				enable := true
				enablePtr = &enable
			}

			if configErr := svc.UpdateScriptConfig(ctx, opts.Device, opts.ID, namePtr, enablePtr); configErr != nil {
				return fmt.Errorf("failed to update config: %w", configErr)
			}
			if namePtr != nil {
				ios.Info("Name updated: %s", *namePtr)
			}
			if enablePtr != nil {
				ios.Info("Script enabled")
			}
		}

		ios.Success("Script %d updated", opts.ID)
		return nil
	})
	if err != nil {
		return err
	}

	cmdutil.InvalidateCache(opts.Factory, opts.Device, cache.TypeScripts)
	return nil
}
