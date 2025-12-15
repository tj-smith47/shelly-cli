// Package importcmd provides the auth import subcommand.
package importcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	Input  string
	DryRun bool
}

// NewCommand creates the auth import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "import <file>",
		Aliases: []string{"imp", "restore"},
		Short:   "Import device credentials",
		Long: `Import device authentication credentials from a file.

Imports credentials that were previously exported with auth export.`,
		Example: `  # Import credentials
  shelly auth import credentials.json

  # Preview without applying
  shelly auth import credentials.json --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Input = args[0]
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview import without applying changes")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	data, err := os.ReadFile(opts.Input)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var export struct {
		ExportedAt  string                                         `json:"exported_at"`
		Version     string                                         `json:"version"`
		Credentials map[string]struct{ Username, Password string } `json:"credentials"`
	}

	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	if len(export.Credentials) == 0 {
		ios.Warning("No credentials found in file")
		return nil
	}

	ios.Info("Found %d credential(s) exported at %s", len(export.Credentials), export.ExportedAt)

	if opts.DryRun {
		ios.Println("")
		ios.Info("Credentials to import (dry-run):")
		for device, cred := range export.Credentials {
			ios.Printf("  %s: %s/***\n", device, cred.Username)
		}
		return nil
	}

	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	importCount := 0
	for device, cred := range export.Credentials {
		if err := cfg.SetDeviceAuth(device, cred.Username, cred.Password); err != nil {
			ios.Warning("Failed to set credentials for %s: %v", device, err)
			continue
		}
		importCount++
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	ios.Success("Imported %d credential(s)", importCount)
	return nil
}
