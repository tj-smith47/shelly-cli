// Package export provides the auth export subcommand.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
)

// Options holds the command options.
type Options struct {
	flags.DeviceTargetFlags
	Output   string
	Password string
}

// NewCommand creates the auth export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

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
			return run(cmd.Context(), f, args, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "credentials.json", "Output file path")
	flags.AddAllOnlyFlag(cmd, &opts.DeviceTargetFlags)

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, devices []string, opts *Options) error {
	ios := f.IOStreams()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Collect credentials
	creds := cfg.GetAllDeviceCredentials()
	if len(creds) == 0 {
		ios.Warning("No credentials found to export")
		return nil
	}

	// Filter if not --all
	if !opts.All && len(devices) > 0 {
		filtered := make(map[string]struct{ Username, Password string })
		for _, d := range devices {
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
	if dir := filepath.Dir(opts.Output); dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	if err := os.WriteFile(opts.Output, data, 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	ios.Success("Exported %d credential(s) to %s", len(creds), opts.Output)
	return nil
}
