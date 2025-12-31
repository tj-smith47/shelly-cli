// Package importcmd provides the template import subcommand.
package importcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the template import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigImportCommand(f, factories.ConfigImportOpts{
		Component: "template",
		Aliases:   []string{"load"},
		Short:     "Import a template from a file",
		Long: `Import a configuration template from a JSON or YAML file.

If no name is specified, the template name from the file is used.
Use --force to overwrite an existing template with the same name.`,
		Example: `  # Import a template
  shelly template import template.yaml

  # Import with a different name
  shelly template import template.yaml my-new-config

  # Overwrite existing template
  shelly template import template.yaml --force`,
		SupportsNameArg: true,
		NameFlagEnabled: false,
		ForceFlagName:   "force",
		ValidArgsFunc:   completion.FileThenNoComplete(),
		Importer:        importTemplate,
	})
}

func importTemplate(file, nameOverride string, overwrite bool) (string, error) {
	// #nosec G304 -- file path comes from user CLI argument
	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Parse template
	tpl, err := config.ParseDeviceTemplateFile(file, data)
	if err != nil {
		return "", err
	}

	// Override name if specified
	if nameOverride != "" {
		tpl.Name = nameOverride
	}

	// Validate name
	if err := config.ValidateTemplateName(tpl.Name); err != nil {
		return "", err
	}

	// Check if exists
	if _, exists := config.GetDeviceTemplate(tpl.Name); exists && !overwrite {
		return "", fmt.Errorf("template %q already exists (use --force to overwrite)", tpl.Name)
	}

	// Save template
	if err := config.SaveDeviceTemplate(tpl); err != nil {
		return "", fmt.Errorf("failed to save template: %w", err)
	}

	return fmt.Sprintf("Template %q imported from %s", tpl.Name, file), nil
}
