// Package export provides the auth export subcommand.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	flags.DeviceTargetFlags
	Devices  []string
	Output   string
	Password string
}

// NewCommand creates the auth export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "export [device...]",
		Aliases: []string{"exp", "backup"},
		Short:   "Export device credentials",
		Long: `Export device authentication credentials to a file.

Exports credentials from the CLI config for backup or transfer.
Use with auth import to restore credentials on another system.`,
		Example: `  # Export all credentials
  shelly auth export --all -o credentials.json

  # Export specific devices
  shelly auth export kitchen bedroom -o creds.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Devices = args
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "credentials.json", "Output file path")
	flags.AddAllOnlyFlag(cmd, &opts.DeviceTargetFlags)

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	mgr, err := opts.Factory.ConfigManager()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Collect credentials
	creds := mgr.GetAllDeviceCredentials()
	if len(creds) == 0 {
		ios.Warning("No credentials found to export")
		return nil
	}

	// Filter if not --all
	if !opts.All && len(opts.Devices) > 0 {
		filtered := make(map[string]struct{ Username, Password string })
		for _, d := range opts.Devices {
			if c, ok := creds[d]; ok {
				filtered[d] = c
			}
		}
		creds = filtered
	} else if !opts.All {
		return fmt.Errorf("specify devices or use --all")
	}

	if len(creds) == 0 {
		ios.Warning("No matching credentials found")
		return nil
	}

	// Build export
	export := map[string]any{
		"exported_at": time.Now().Format(time.RFC3339),
		"version":     "1.0",
		"credentials": creds,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	// Ensure output directory exists
	fs := config.Fs()
	if dir := filepath.Dir(opts.Output); dir != "." {
		if err := fs.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	if err := afero.WriteFile(fs, opts.Output, data, 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	ios.Success("Exported %d credential(s) to %s", len(creds), opts.Output)
	return nil
}
