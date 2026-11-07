// Package importcmd provides the alias import command.
package importcmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

const aliasLoad = "load"

// Options holds the options for the import command.
type Options struct {
	Factory  *cmdutil.Factory
	Filename string
	Merge    bool
	Yes      bool
}

// NewCommand creates the alias import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "import <file>",
		Aliases: []string{aliasLoad},
		Short:   "Import aliases from a YAML file",
		Long: `Import command aliases from a YAML file.

By default, existing aliases with the same name are overwritten after a
confirmation prompt; pass --yes to skip the prompt in scripts. Use --merge to
skip existing aliases instead of overwriting them.

The file format is:
  aliases:
    name1: "command1"
    name2: "command2"
    shellalias: "!shell command"`,
		Example: `  # Import aliases (overwrite existing)
  shelly alias import aliases.yaml

  # Import without overwriting existing aliases
  shelly alias import aliases.yaml --merge

  # Overwrite existing aliases without the prompt (scripts)
  shelly alias import aliases.yaml --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Filename = args[0]
			return run(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Merge, "merge", "m", false, "Merge with existing aliases (skip conflicts)")
	flags.AddYesFlag(cmd, &opts.Yes)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Only an import that replaces aliases already in the config is destructive;
	// --merge keeps them, so it never prompts.
	overwritten, err := opts.overwrittenAliases()
	if err != nil {
		return err
	}
	if len(overwritten) > 0 {
		confirmed, err := opts.Factory.ConfirmAction(
			fmt.Sprintf("Overwrite %d existing alias(es) (%s)?", len(overwritten), strings.Join(overwritten, ", ")), opts.Yes)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Info("Cancelled")
			return nil
		}
	}

	imported, skipped, err := config.ImportAliases(opts.Filename, opts.Merge)
	if err != nil {
		return fmt.Errorf("failed to import aliases: %w", err)
	}

	ios.Success("Imported %d alias(es) from %s", imported, opts.Filename)
	if skipped > 0 {
		ios.Info("Skipped %d existing alias(es)", skipped)
	}

	return nil
}

// overwrittenAliases lists the aliases in the import file that already exist
// in the config, in name order. It is empty with --merge, which keeps them.
func (o *Options) overwrittenAliases() ([]string, error) {
	if o.Merge {
		return nil, nil
	}
	data, err := afero.ReadFile(config.Fs(), o.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to import aliases: failed to read file: %w", err)
	}
	var af struct {
		Aliases map[string]string `yaml:"aliases"`
	}
	if err := yaml.Unmarshal(data, &af); err != nil {
		return nil, fmt.Errorf("failed to import aliases: failed to parse file: %w", err)
	}
	var overwritten []string
	for name := range af.Aliases {
		if _, exists := config.GetAlias(name); exists {
			overwritten = append(overwritten, name)
		}
	}
	sort.Strings(overwritten)
	return overwritten, nil
}
