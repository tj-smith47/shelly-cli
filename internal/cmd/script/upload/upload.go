// Package upload provides the script upload subcommand.
package upload

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var appendFlag bool

// NewCommand creates the script upload command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upload <device> <id> <file>",
		Aliases: []string{"put"},
		Short:   "Upload script from file",
		Long: `Upload script code to a device from a file.

By default, replaces the existing code. Use --append to add to existing code.`,
		Example: `  # Upload script from file
  shelly script upload living-room 1 script.js

  # Append code from file
  shelly script upload living-room 1 additions.js --append`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid script ID: %s", args[1])
			}
			return run(cmd.Context(), args[0], id, args[2])
		},
	}

	cmd.Flags().BoolVar(&appendFlag, "append", false, "Append to existing code")

	return cmd
}

func run(ctx context.Context, device string, id int, file string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

	// Read file
	//nolint:gosec // G304: User-provided file path is intentional for this command
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	code := string(data)

	return cmdutil.RunWithSpinner(ctx, ios, "Uploading script...", func(ctx context.Context) error {
		if uploadErr := svc.UpdateScriptCode(ctx, device, id, code, appendFlag); uploadErr != nil {
			return fmt.Errorf("failed to upload script: %w", uploadErr)
		}

		if appendFlag {
			ios.Success("Appended %d bytes to script %d", len(code), id)
		} else {
			ios.Success("Uploaded %d bytes to script %d", len(code), id)
		}
		return nil
	})
}
