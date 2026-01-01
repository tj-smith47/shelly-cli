// Package create provides the extension create command.
package create

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/plugins/scaffold"
)

// Options holds the options for the create command.
type Options struct {
	Factory   *cmdutil.Factory
	Lang      string
	Name      string
	OutputDir string
}

// NewCommand creates the extension create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"new", "init", "scaffold"},
		Short:   "Create a new extension scaffold",
		Long: `Create a new extension scaffold with boilerplate code.

Supported languages:
  - bash (default): Shell script extension
  - go: Go language extension
  - python: Python extension

The extension will be created in the current directory or the directory
specified with --output.`,
		Example: `  # Create a bash extension
  shelly extension create myext

  # Create a Go extension
  shelly extension create myext --lang go

  # Create in specific directory
  shelly extension create myext --output ~/projects`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Name = args[0]
			return run(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Lang, "lang", "l", "bash", "Extension language (bash, go, python)")
	cmd.Flags().StringVarP(&opts.OutputDir, "output", "o", ".", "Output directory")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Normalize name and get full extension name
	name := scaffold.NormalizeName(opts.Name)
	extName := scaffold.FullName(name)

	// Create output directory
	extDir := filepath.Join(opts.OutputDir, extName)
	if err := config.Fs().MkdirAll(extDir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate files based on language
	switch opts.Lang {
	case "bash", "sh":
		if err := scaffold.Bash(extDir, extName, name); err != nil {
			return err
		}
	case "go", "golang":
		if err := scaffold.Go(extDir, extName, name); err != nil {
			return err
		}
	case "python", "py":
		if err := scaffold.Python(extDir, extName, name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported language: %s (use bash, go, or python)", opts.Lang)
	}

	ios.Success("Created extension scaffold in %s", extDir)
	ios.Printf("\nNext steps:\n")
	ios.Printf("  1. Edit the extension code in %s\n", extDir)
	ios.Printf("  2. Test locally: ./%s/%s --help\n", extName, extName)
	ios.Printf("  3. Install: shelly extension install ./%s/%s\n", extName, extName)

	return nil
}
