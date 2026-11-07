// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// ConfigImportOpts configures a config-based import command (scene, template, etc.).
type ConfigImportOpts struct {
	// Component is the type being imported (e.g., "scene", "template").
	Component string

	// Aliases are alternate command names.
	Aliases []string

	// Short is the short description.
	Short string

	// Long is the detailed description.
	Long string

	// Example shows usage examples.
	Example string

	// SupportsNameArg indicates if name can be provided as second positional arg.
	// If false, only file is accepted (ExactArgs(1)).
	// If true, file and optional name are accepted (RangeArgs(1,2)).
	SupportsNameArg bool

	// NameFlagEnabled adds a --name flag for name override (used when SupportsNameArg is false).
	NameFlagEnabled bool

	// ForceFlagName is the name of the replace-existing flag. Empty means
	// "force" with shorthand "f".
	ForceFlagName string

	// ForceFlagAlias, when set, registers a hidden flag of that name bound to
	// the same variable as the force flag, so scripts written against an older
	// spelling keep working.
	ForceFlagAlias string

	// ValidArgsFunc provides completion for arguments.
	ValidArgsFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

	// Importer performs the import: parse file, validate, check exists, save.
	// Returns a success message on success, or an error.
	Importer func(file, nameOverride string, overwrite bool) (successMsg string, err error)
}

// configImportOptions holds runtime options for a config import command.
type configImportOptions struct {
	File      string
	Name      string
	Overwrite bool
	Factory   *cmdutil.Factory
	Config    ConfigImportOpts
}

// NewConfigImportCommand creates a config-based import command.
func NewConfigImportCommand(f *cmdutil.Factory, config ConfigImportOpts) *cobra.Command {
	opts := &configImportOptions{
		Factory: f,
		Config:  config,
	}

	// Determine args based on config
	var args cobra.PositionalArgs
	usePattern := "import <file>"
	if config.SupportsNameArg {
		args = cobra.RangeArgs(1, 2)
		usePattern = "import <file> [name]"
	} else {
		args = cobra.ExactArgs(1)
	}

	cmd := &cobra.Command{
		Use:               usePattern,
		Aliases:           config.Aliases,
		Short:             config.Short,
		Long:              config.Long,
		Example:           config.Example,
		Args:              args,
		ValidArgsFunction: config.ValidArgsFunc,
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			opts.File = cmdArgs[0]
			if config.SupportsNameArg && len(cmdArgs) > 1 {
				opts.Name = cmdArgs[1]
			}
			return runConfigImport(opts)
		},
	}

	// Add name flag if enabled
	if config.NameFlagEnabled {
		cmd.Flags().StringVarP(&opts.Name, "name", "n", "", fmt.Sprintf("Override %s name from file", config.Component))
	}

	flagName := config.ForceFlagName
	if flagName == "" {
		flagName = "force"
	}
	shorthand := ""
	if flagName == "force" {
		shorthand = "f"
	}
	usage := fmt.Sprintf("Replace an existing %s with the same name", config.Component)
	cmd.Flags().BoolVarP(&opts.Overwrite, flagName, shorthand, false, usage)
	if alias := config.ForceFlagAlias; alias != "" {
		cmd.Flags().BoolVar(&opts.Overwrite, alias, false, usage)
		if err := cmd.Flags().MarkHidden(alias); err != nil {
			// MarkHidden only fails on an unknown flag name; the alias is defined above.
			panic(err)
		}
	}

	return cmd
}

func runConfigImport(opts *configImportOptions) error {
	ios := opts.Factory.IOStreams()

	successMsg, err := opts.Config.Importer(opts.File, opts.Name, opts.Overwrite)
	if err != nil {
		return err
	}

	ios.Success("%s", successMsg)
	return nil
}
