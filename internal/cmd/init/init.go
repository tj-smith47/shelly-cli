// Package init provides the init command for first-run setup.
package init

import (
	"context"

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

This command walks you through:
  1. Configuration (theme, output format, aliases)
  2. Device discovery (find Shelly devices on your network)
  3. Device registration (give discovered devices friendly names)
  4. Shell completions (tab completion for commands)
  5. Cloud access (optional remote control via Shelly Cloud)
  6. Telemetry (optional anonymous usage statistics)

INTERACTIVE MODE (default):
  Run without flags for the full guided wizard.

DEFAULTS MODE (--defaults):
  Runs the full wizard with sensible defaults, no prompts.
  Discovers devices, registers with default names, installs completions.
  Combine with flags to override individual defaults (e.g., --defaults --theme nord).

INDIVIDUAL FLAGS:
  Pass specific flags to set values; other steps remain interactive.

Use --check to verify your current setup without making changes.`,
		Example: `  # Interactive setup wizard (recommended for first use)
  shelly init

  # Quick setup with sensible defaults (no prompts)
  shelly init --defaults

  # Defaults with a specific theme
  shelly init --defaults --theme nord

  # Defaults with force overwrite of existing config
  shelly init --defaults --force

  # Check current setup without changes
  shelly init --check

  # Register devices directly (other steps remain interactive)
  shelly init --device kitchen=192.168.1.100 --device bedroom=192.168.1.101

  # Run discovery without prompting (other steps remain interactive)
  shelly init --discover

  # Scan specific subnets
  shelly init --discover --network 192.168.1.0/24 --network 10.0.0.0/24

  # Full CI/CD setup (no prompts)
  shelly init --defaults \
    --device kitchen=192.168.1.100 \
    --theme dracula \
    --no-color

  # With cloud credentials
  shelly init --defaults --cloud-email user@example.com --cloud-password secret`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.RootCommand = cmd.Root()
			return run(cmd.Context(), opts)
		},
	}

	// Device flags
	cmd.Flags().StringArrayVar(&wizardOpts.Devices, "device", nil, "Device spec: name=ip[:user:pass] (repeatable)")
	cmd.Flags().StringArrayVar(&wizardOpts.DevicesJSON, "devices-json", nil, "JSON device(s): file path, array, or single object (repeatable)")

	// Discovery flags
	cmd.Flags().BoolVar(&wizardOpts.Discover, "discover", false, "Run device discovery without prompting")
	cmd.Flags().DurationVar(&wizardOpts.DiscoverTimeout, "discover-timeout", cmdutil.DefaultScanTimeout, "Discovery timeout")
	cmd.Flags().StringVar(&wizardOpts.DiscoverModes, "discover-modes", "http", "Discovery modes: http,mdns,coiot,ble,all (comma-separated)")
	cmd.Flags().StringArrayVar(&wizardOpts.Networks, "network", nil, "Subnet(s) for HTTP probe discovery (repeatable)")
	cmd.Flags().BoolVar(&wizardOpts.AllNetworks, "all-networks", false, "Scan all detected subnets without prompting")

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
	cmd.Flags().BoolVar(&wizardOpts.Defaults, "defaults", false, "Use sensible defaults for all prompts (no interactive questions)")
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
