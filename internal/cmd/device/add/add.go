// Package add provides the device add subcommand.
package add

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Options holds the command options.
type Options struct {
	Factory    *cmdutil.Factory
	Name       string
	Address    string
	Auth       string
	Generation int
	NoVerify   bool
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
becomes "master-bathroom" as the key).`,
		Example: `  # Add a device (auto-detects generation and model)
  shelly device add kitchen 192.168.1.100

  # Add with authentication
  shelly device add secure-device 192.168.1.101 --auth admin:secret

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

	cmd.Flags().StringVar(&opts.Auth, "auth", "", "Authentication credentials (user:pass)")
	cmd.Flags().IntVarP(&opts.Generation, "generation", "g", 0, "Device generation (auto-detected if omitted)")
	cmd.Flags().BoolVar(&opts.NoVerify, "no-verify", false, "Skip connectivity check and auto-detection")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Check if name is taken
	if _, exists := config.GetDevice(opts.Name); exists {
		return fmt.Errorf("device %q already exists", opts.Name)
	}

	var generation int
	var deviceType, deviceModel string

	if !opts.NoVerify {
		ios.StartProgress("Connecting to device...")
		info, err := svc.DeviceInfoAuto(ctx, opts.Address)
		ios.StopProgress()
		if err != nil {
			return fmt.Errorf("couldn't reach device at %s: %w", opts.Address, err)
		}
		generation = info.Generation
		deviceType = info.App
		deviceModel = info.Model
	} else {
		generation = opts.Generation
	}

	// Parse auth credentials
	var auth *model.Auth
	if opts.Auth != "" {
		parts := strings.SplitN(opts.Auth, ":", 2)
		if len(parts) == 2 {
			auth = &model.Auth{Username: parts[0], Password: parts[1]}
		} else {
			return fmt.Errorf("invalid auth format, expected user:pass")
		}
	}

	if err := config.RegisterDevice(opts.Name, opts.Address, generation, deviceType, deviceModel, auth); err != nil {
		return fmt.Errorf("failed to register device: %w", err)
	}

	ios.Success("Added %q at %s", opts.Name, opts.Address)
	if deviceModel != "" {
		ios.Info("  %s (Gen%d)", deviceModel, generation)
	}

	return nil
}
