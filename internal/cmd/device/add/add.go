// Package add provides the device add subcommand.
package add

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory     *cmdutil.Factory
	Name        string
	Address     string
	Credentials cmdutil.DeviceCredentials
	Generation  int
	NoVerify    bool
	Force       bool
	Platform    string
}

// NewCommand creates the device add command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "add <name> <address>",
		Aliases: []string{"register", "new"},
		Short:   "Add a device to the registry",
		Long: `Add a Shelly device to the local registry.

The device will be verified and its generation/model auto-detected
unless --no-verify is specified.

The name is used as a friendly identifier for the device in other commands.
Names with spaces will be normalized to dashes (e.g., "Master Bathroom"
becomes "master-bathroom" as the key).

Credentials for a password-protected device are stored with the device and
used by every command run after it. Give them as --password (the user
defaults to admin, the user Shelly devices use), as --password-stdin to keep
the password out of the shell history, or as --auth user:pass. When the device
has authentication enabled the credentials are checked against it first, and
a device that rejects them is not added. --no-verify skips that check.

A name that is already registered is refused unless --force is given, which
replaces the registration (address, credentials, platform, model). Aliases
of the replaced device are kept.

--platform registers a device managed by a plugin, such as tasmota for the
shelly-tasmota plugin. The plugin's detect hook verifies the device instead
of the Shelly API.`,
		Example: `  # Add a device (auto-detects generation and model)
  shelly device add kitchen 192.168.1.100

  # Add a password-protected device
  shelly device add kitchen 192.168.1.100 --user admin --password secret

  # Read the password from stdin (prompts without echo on a terminal)
  shelly device add kitchen 192.168.1.100 --password-stdin < ~/.shelly-kitchen-password

  # Replace an existing registration with a new address or password
  shelly device add kitchen 192.168.1.110 --force --password-stdin

  # Add a Tasmota device through the shelly-tasmota plugin
  shelly device add garage-plug 192.168.1.50 --platform tasmota

  # Add without verification (offline)
  shelly device add offline-device 192.168.1.102 --no-verify --generation 2

  # Short form
  shelly dev add bedroom 192.168.1.103`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			opts.Address = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Credentials.Pair, "auth", "", "Authentication credentials (user:pass)")
	cmdutil.AddDeviceCredentialFlags(cmd, &opts.Credentials, "Device password", "Read the device password from stdin")
	cmd.Flags().IntVarP(&opts.Generation, "generation", "g", 0, "Device generation (auto-detected if omitted)")
	cmd.Flags().BoolVar(&opts.NoVerify, "no-verify", false, "Skip the connectivity check, auto-detection and the credentials check")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Replace an existing device of the same name")
	cmd.Flags().StringVar(&opts.Platform, "platform", "", "Device platform managed by a plugin (e.g. tasmota); default shelly")

	cmd.MarkFlagsMutuallyExclusive("auth", "user")
	cmd.MarkFlagsMutuallyExclusive("auth", "password")
	cmd.MarkFlagsMutuallyExclusive("auth", "password-stdin")
	cmd.MarkFlagsMutuallyExclusive("platform", "generation")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()
	mgr, err := opts.Factory.ConfigManager()
	if err != nil {
		return err
	}

	existing, exists := mgr.GetDevice(opts.Name)
	if exists && !opts.Force {
		return fmt.Errorf("device %q already exists; use --force to replace it", opts.Name)
	}

	auth, err := opts.Credentials.Auth(ios)
	if errors.Is(err, cmdutil.ErrNoCredentials) {
		auth = nil
	} else if err != nil {
		return err
	}

	dev := model.Device{
		Name:       opts.Name,
		Address:    opts.Address,
		Generation: opts.Generation,
		Auth:       auth,
	}
	// Native devices are stored without a platform, as every other registration path does.
	if opts.Platform != model.PlatformShelly {
		dev.Platform = opts.Platform
	}

	if !opts.NoVerify {
		ios.StartProgress("Connecting to device...")
		identified, authEnabled, err := svc.IdentifyNewDevice(ctx, dev)
		ios.StopProgress()
		if err != nil {
			return fmt.Errorf("couldn't verify device at %s: %w", opts.Address, err)
		}
		dev = identified
		if authEnabled && auth == nil {
			ios.Warning("%s requires authentication; add --password or --password-stdin (with --force) to store credentials", opts.Address)
		}
	}

	if err := mgr.RegisterDeviceWithPlatform(dev.Name, dev.Address, dev.Generation, dev.Type, dev.Model, dev.Platform, dev.Auth); err != nil {
		return fmt.Errorf("failed to register device: %w", err)
	}
	for _, alias := range existing.Aliases {
		if err := mgr.AddDeviceAlias(opts.Name, alias); err != nil {
			ios.Warning("couldn't keep alias %q: %v", alias, err)
		}
	}

	term.DisplayDeviceRegistered(ios, dev, exists)
	return nil
}
