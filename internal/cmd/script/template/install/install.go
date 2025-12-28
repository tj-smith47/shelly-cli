// Package install provides the script template install subcommand.
package install

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
)

// Options holds command options.
type Options struct {
	Device    string
	Template  string
	Configure bool
	Enable    bool
	Name      string
	Factory   *cmdutil.Factory
}

// NewCommand creates the script template install command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "install <device> <template>",
		Aliases: []string{"add", "deploy"},
		Short:   "Install a script template on a device",
		Long: `Install a script template on a Shelly device.

Creates a new script on the device with the template code. Template
variables are substituted with their default values, or you can use
--configure for interactive configuration.`,
		Example: `  # Install with default values
  shelly script template install living-room motion-light

  # Install with interactive configuration
  shelly script template install living-room motion-light --configure

  # Install and enable immediately
  shelly script template install living-room motion-light --enable

  # Install with custom script name
  shelly script template install living-room motion-light --name "Motion Sensor"`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceThenScriptTemplate(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Template = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Configure, "configure", false, "Interactive variable configuration")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable script after installation")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Custom script name (defaults to template name)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	// Get template
	tpl, ok := automation.GetScriptTemplate(opts.Template)
	if !ok {
		return fmt.Errorf("script template %q not found", opts.Template)
	}

	// Prompt for variable values if requested
	values := ios.PromptScriptVariables(tpl.Variables, opts.Configure)

	// Substitute variables in code
	code := automation.SubstituteVariables(tpl.Code, values)

	// Determine script name
	scriptName := opts.Name
	if scriptName == "" {
		scriptName = tpl.Name
	}

	// Install script on device
	var result *automation.InstallScriptResult
	err := cmdutil.RunWithSpinner(ctx, ios, "Installing script template...", func(ctx context.Context) error {
		var installErr error
		result, installErr = svc.InstallScript(ctx, opts.Device, scriptName, code, opts.Enable)
		return installErr
	})
	if err != nil {
		return err
	}

	ios.Success("Installed template %q as script %d on %s", opts.Template, result.ID, opts.Device)
	if result.Enabled {
		ios.Info("Script enabled")
	}
	return nil
}
