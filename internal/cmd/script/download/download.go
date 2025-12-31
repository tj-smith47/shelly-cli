// Package download provides the script download subcommand.
package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
}

// NewCommand creates the script download command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "download <device> <id> <file>",
		Aliases: []string{"save"},
		Short:   "Download script to file",
		Long:    `Download script code from a device to a local file.`,
		Example: `  # Download script to file
  shelly script download living-room 1 script.js

  # Download to a specific directory
  shelly script download living-room 1 scripts/auto-light.js`,
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

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	return cmdutil.RunWithSpinner(ctx, ios, "Downloading script...", func(ctx context.Context) error {
		code, err := svc.GetScriptCode(ctx, opts.Device, opts.ID)
		if err != nil {
			return fmt.Errorf("failed to get script code: %w", err)
		}

		if code == "" {
			ios.Warning("Script %d has no code", opts.ID)
			return nil
		}

		// Ensure directory exists
		dir := filepath.Dir(opts.File)
		if dir != "." && dir != "" {
			if mkErr := os.MkdirAll(dir, 0o750); mkErr != nil {
				return fmt.Errorf("failed to create directory: %w", mkErr)
			}
		}

		// Write file
		//nolint:gosec // G306: 0o644 is appropriate for script files
		if writeErr := os.WriteFile(opts.File, []byte(code), 0o644); writeErr != nil {
			return fmt.Errorf("failed to write file: %w", writeErr)
		}

		ios.Success("Downloaded script %d to %s (%d bytes)", opts.ID, opts.File, len(code))
		return nil
	})
}
