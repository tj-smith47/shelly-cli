// Package create provides the script create subcommand.
package create

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	nameFlag   string
	codeFlag   string
	fileFlag   string
	enableFlag bool
)

// NewCommand creates the script create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "Script name")
	cmd.Flags().StringVar(&codeFlag, "code", "", "Script code (inline)")
	cmd.Flags().StringVarP(&fileFlag, "file", "f", "", "Script code file")
	cmd.Flags().BoolVar(&enableFlag, "enable", false, "Enable script after creation")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
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

	return cmdutil.RunWithSpinner(ctx, ios, "Creating script...", func(ctx context.Context) error {
		// Create the script
		id, err := svc.CreateScript(ctx, device, nameFlag)
		if err != nil {
			return fmt.Errorf("failed to create script: %w", err)
		}

		// Upload code if provided
		if code != "" {
			if uploadErr := svc.UpdateScriptCode(ctx, device, id, code, false); uploadErr != nil {
				return fmt.Errorf("failed to upload code: %w", uploadErr)
			}
		}

		// Enable if requested
		if enableFlag {
			enable := true
			if configErr := svc.UpdateScriptConfig(ctx, device, id, nil, &enable); configErr != nil {
				return fmt.Errorf("failed to enable script: %w", configErr)
			}
		}

		ios.Success("Created script %d", id)
		if nameFlag != "" {
			ios.Info("Name: %s", nameFlag)
		}
		if code != "" {
			ios.Info("Code uploaded (%d bytes)", len(code))
		}
		if enableFlag {
			ios.Info("Script enabled")
		}
		return nil
	})
}
