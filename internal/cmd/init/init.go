// Package init provides the init command for first-run setup.
package init

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/wizard"
)

// Options holds the command options.
type Options struct {
	Factory     *cmdutil.Factory
	WizardOpts  *wizard.Options
	Check       bool
	RootCommand *cobra.Command // Needed for wizard.Run
}

// NewCommand creates the init command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	wizardOpts := &wizard.Options{}
	opts := &Options{Factory: f, WizardOpts: wizardOpts}

	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"setup", "configure"},
		Short:   "Initialize shelly CLI for first-time use",
		Long: `Initialize the shelly CLI with a guided setup wizard.

This command helps new users get started by:
  - Creating a configuration file with sensible defaults
  - Discovering Shelly devices on your network
  - Registering discovered devices with friendly names
  - Installing shell completions for tab completion
  - Optionally setting up Shelly Cloud access

INTERACTIVE MODE (default):
  Run without flags to use the interactive wizard.

NON-INTERACTIVE MODE:
  Automatically enabled when any device or config flags are provided.
  Discovery, completions, aliases, and cloud are opt-in via flags.

Use --check to verify your current setup without making changes.`,
		Example: `  # Interactive setup wizard
  shelly init

  # Check current setup without changes
  shelly init --check

  # Non-interactive: register devices directly
  shelly init --device kitchen=192.168.1.100 --device bedroom=192.168.1.101

  # Non-interactive: device with authentication
  shelly init --device secure=192.168.1.102:admin:secret

  # Non-interactive: import from JSON file
  shelly init --devices-json devices.json

  # Non-interactive: inline JSON
  shelly init --devices-json '{"name":"kitchen","address":"192.168.1.100"}'

  # Non-interactive: with discovery and completions
  shelly init --discover --discover-modes http,mdns --completions bash,zsh

  # Non-interactive: full CI/CD setup
  shelly init \
    --device kitchen=192.168.1.100 \
    --theme dracula \
    --api-mode local \
    --no-color

  # Non-interactive: with cloud credentials
  shelly init --cloud-email user@example.com --cloud-password secret

  # Non-interactive: enable anonymous telemetry
  shelly init --telemetry`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.RootCommand = cmd.Root()
			return run(cmd.Context(), opts)
		},
	}

	// Device flags
	cmd.Flags().StringArrayVar(&wizardOpts.Devices, "device", nil, "Device spec: name=ip[:user:pass] (repeatable)")
	cmd.Flags().StringArrayVar(&wizardOpts.DevicesJSON, "devices-json", nil, "JSON device(s): file path, array, or single object (repeatable)")

	// Discovery flags
	cmd.Flags().BoolVar(&wizardOpts.Discover, "discover", false, "Enable device discovery (opt-in in non-interactive mode)")
	cmd.Flags().DurationVar(&wizardOpts.DiscoverTimeout, "discover-timeout", 15*time.Second, "Discovery timeout")
	cmd.Flags().StringVar(&wizardOpts.DiscoverModes, "discover-modes", "http", "Discovery modes: http,mdns,coiot,ble,all (comma-separated)")
	cmd.Flags().StringVar(&wizardOpts.Network, "network", "", "Subnet for HTTP probe discovery (e.g., 192.168.1.0/24)")

	// Completion flags
	cmd.Flags().StringVar(&wizardOpts.Completions, "completions", "", "Install completions for shells: bash,zsh,fish,powershell (comma-separated)")
	cmd.Flags().BoolVar(&wizardOpts.Aliases, "aliases", false, "Install default command aliases (opt-in)")

	// Config flags
	cmd.Flags().StringVar(&wizardOpts.Theme, "theme", "", "Set theme (default: dracula)")
	cmd.Flags().StringVar(&wizardOpts.OutputFormat, "output-format", "", "Set output format: table,json,yaml (default: table)")
	cmd.Flags().BoolVar(&wizardOpts.NoColor, "no-color", false, "Disable colors in output")
	cmd.Flags().StringVar(&wizardOpts.APIMode, "api-mode", "", "API mode: local,cloud,auto (default: local)")

	// Cloud flags
	cmd.Flags().StringVar(&wizardOpts.CloudEmail, "cloud-email", "", "Shelly Cloud email (enables cloud setup)")
	cmd.Flags().StringVar(&wizardOpts.CloudPassword, "cloud-password", "", "Shelly Cloud password (enables cloud setup)")

	// Telemetry flags
	cmd.Flags().BoolVar(&wizardOpts.Telemetry, "telemetry", false, "Enable anonymous usage telemetry (opt-in)")

	// Control flags
	cmd.Flags().BoolVar(&wizardOpts.Force, "force", false, "Overwrite existing configuration")
	cmd.Flags().BoolVar(&opts.Check, "check", false, "Verify current setup without making changes")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	if opts.Check {
		return wizard.RunCheck(opts.Factory)
	}
	return wizard.Run(ctx, opts.Factory, opts.RootCommand, opts.WizardOpts)
}
