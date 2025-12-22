// Package init provides the init command for first-run setup.
package init

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the init command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &cmdutil.InitOptions{}

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

  # Headless: register devices directly
  shelly init --device kitchen=192.168.1.100 --device bedroom=192.168.1.101

  # Headless: device with authentication
  shelly init --device secure=192.168.1.102:admin:secret

  # Headless: import from JSON file
  shelly init --devices-json devices.json

  # Headless: inline JSON
  shelly init --devices-json '{"name":"kitchen","address":"192.168.1.100"}'

  # Headless: with discovery and completions
  shelly init --discover --discover-modes http,mdns --completions bash,zsh

  # Headless: full CI/CD setup
  shelly init \
    --device kitchen=192.168.1.100 \
    --theme dracula \
    --api-mode local \
    --no-color

  # Headless: with cloud credentials
  shelly init --cloud-email user@example.com --cloud-password secret`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, f, opts)
		},
	}

	// Device flags
	cmd.Flags().StringArrayVar(&opts.Devices, "device", nil, "Device spec: name=ip[:user:pass] (repeatable)")
	cmd.Flags().StringArrayVar(&opts.DevicesJSON, "devices-json", nil, "JSON device(s): file path, array, or single object (repeatable)")

	// Discovery flags
	cmd.Flags().BoolVar(&opts.Discover, "discover", false, "Enable device discovery (opt-in in non-interactive mode)")
	cmd.Flags().DurationVar(&opts.DiscoverTimeout, "discover-timeout", 15*time.Second, "Discovery timeout")
	cmd.Flags().StringVar(&opts.DiscoverModes, "discover-modes", "http", "Discovery modes: http,mdns,coiot,ble,all (comma-separated)")
	cmd.Flags().StringVar(&opts.Network, "network", "", "Subnet for HTTP probe discovery (e.g., 192.168.1.0/24)")

	// Completion flags
	cmd.Flags().StringVar(&opts.Completions, "completions", "", "Install completions for shells: bash,zsh,fish,powershell (comma-separated)")
	cmd.Flags().BoolVar(&opts.Aliases, "aliases", false, "Install default command aliases (opt-in)")

	// Config flags
	cmd.Flags().StringVar(&opts.Theme, "theme", "", "Set theme (default: dracula)")
	cmd.Flags().StringVar(&opts.OutputFormat, "output-format", "", "Set output format: table,json,yaml (default: table)")
	cmd.Flags().BoolVar(&opts.NoColor, "no-color", false, "Disable colors in output")
	cmd.Flags().StringVar(&opts.APIMode, "api-mode", "", "API mode: local,cloud,auto (default: local)")

	// Cloud flags
	cmd.Flags().StringVar(&opts.CloudEmail, "cloud-email", "", "Shelly Cloud email (enables cloud setup)")
	cmd.Flags().StringVar(&opts.CloudPassword, "cloud-password", "", "Shelly Cloud password (enables cloud setup)")

	// Control flags
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite existing configuration")
	cmd.Flags().BoolP("check", "", false, "Verify current setup without making changes")

	return cmd
}

func run(cmd *cobra.Command, f *cmdutil.Factory, opts *cmdutil.InitOptions) error {
	check, err := cmd.Flags().GetBool("check")
	if err != nil {
		return fmt.Errorf("failed to read --check flag: %w", err)
	}
	if check {
		return cmdutil.RunInitCheck(f)
	}
	return cmdutil.RunInitWizard(cmd.Context(), f, cmd.Root(), opts)
}
