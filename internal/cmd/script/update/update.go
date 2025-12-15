// Package update provides the script update subcommand.
package update

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

var (
	nameFlag   string
	codeFlag   string
	fileFlag   string
	appendFlag bool
	enableFlag bool
)

// NewCommand creates the script update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0], id)
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "Script name")
	cmd.Flags().StringVar(&codeFlag, "code", "", "Script code (inline)")
	cmd.Flags().StringVarP(&fileFlag, "file", "f", "", "Script code file")
	cmd.Flags().BoolVar(&appendFlag, "append", false, "Append code instead of replacing")
	cmd.Flags().BoolVar(&enableFlag, "enable", false, "Enable the script")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Get code from file if specified
	code := codeFlag
	if fileFlag != "" {
		//nolint:gosec // G304: User-provided file path is intentional for this command
		data, err := os.ReadFile(fileFlag)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		code = string(data)
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Updating script...", func(ctx context.Context) error {
		updated := false

		// Update code if provided
		if code != "" {
			if uploadErr := svc.UpdateScriptCode(ctx, device, id, code, appendFlag); uploadErr != nil {
				return fmt.Errorf("failed to update code: %w", uploadErr)
			}
			if appendFlag {
				ios.Info("Code appended (%d bytes)", len(code))
			} else {
				ios.Info("Code updated (%d bytes)", len(code))
			}
			updated = true
		}

		// Update config if name or enable specified
		var namePtr *string
		var enablePtr *bool
		if nameFlag != "" {
			namePtr = &nameFlag
		}
		if enableFlag {
			enable := true
			enablePtr = &enable
		}

		if namePtr != nil || enablePtr != nil {
			if configErr := svc.UpdateScriptConfig(ctx, device, id, namePtr, enablePtr); configErr != nil {
				return fmt.Errorf("failed to update config: %w", configErr)
			}
			if namePtr != nil {
				ios.Info("Name updated: %s", *namePtr)
			}
			if enablePtr != nil {
				ios.Info("Script enabled")
			}
			updated = true
		}

		if !updated {
			ios.Warning("No changes specified")
			return nil
		}

		ios.Success("Script %d updated", id)
		return nil
	})
}
