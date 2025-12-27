// Package create provides the extension create command.
package create

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/plugins/scaffold"
)

// NewCommand creates the extension create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		lang      string
		outputDir string
	)

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
			return run(f, args[0], lang, outputDir)
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "bash", "Extension language (bash, go, python)")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory")

	return cmd
}

func run(f *cmdutil.Factory, name, lang, outputDir string) error {
	ios := f.IOStreams()

	// Normalize name and get full extension name
	name = scaffold.NormalizeName(name)
	extName := scaffold.FullName(name)

	// Create output directory
	extDir := filepath.Join(outputDir, extName)
	if err := os.MkdirAll(extDir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate files based on language
	switch lang {
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
		return fmt.Errorf("unsupported language: %s (use bash, go, or python)", lang)
	}

	ios.Success("Created extension scaffold in %s", extDir)
	ios.Printf("\nNext steps:\n")
	ios.Printf("  1. Edit the extension code in %s\n", extDir)
	ios.Printf("  2. Test locally: ./%s/%s --help\n", extName, extName)
	ios.Printf("  3. Install: shelly extension install ./%s/%s\n", extName, extName)

	return nil
}
