// Package upload provides the script upload subcommand.
package upload

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	File    string
	Append  bool
}

// NewCommand creates the script upload command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.DeviceThenScriptID(),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid script ID: %s", args[1])
			}
			opts.Device = args[0]
			opts.ID = id
			opts.File = args[2]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Append, "append", false, "Append to existing code")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	// Read file
	//nolint:gosec // G304: User-provided file path is intentional for this command
	data, err := os.ReadFile(opts.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	code := string(data)

	return cmdutil.RunWithSpinner(ctx, ios, "Uploading script...", func(ctx context.Context) error {
		if uploadErr := svc.UpdateScriptCode(ctx, opts.Device, opts.ID, code, opts.Append); uploadErr != nil {
			return fmt.Errorf("failed to upload script: %w", uploadErr)
		}

		if opts.Append {
			ios.Success("Appended %d bytes to script %d", len(code), opts.ID)
		} else {
			ios.Success("Uploaded %d bytes to script %d", len(code), opts.ID)
		}
		return nil
	})
}
