// Package test provides the auth test subcommand.
package test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory     *cmdutil.Factory
	Device      string
	Credentials cmdutil.DeviceCredentials
	Timeout     time.Duration
}

// NewCommand creates the auth test command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		Timeout: 10 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "test <device>",
		Aliases: []string{"verify", "check"},
		Short:   "Test authentication credentials",
		Long: `Test authentication credentials against a device.

The device is asked for data it only returns to an authenticated caller
(Sys.GetStatus on Gen2+ devices, /settings on Gen1 devices), so a wrong
password fails the test. Without --password or --password-stdin the
credentials stored for the device are tested; with one of them the given
credentials are tested instead, and nothing is stored. The user defaults to
admin.

A device with authentication disabled accepts any credentials; the command
says so instead of reporting the password as correct.

Exit codes:
  0 - The device accepted the credentials
  1 - The device rejected the credentials, or could not be reached`,
		Example: `  # Test the credentials stored for the device
  shelly auth test living-room

  # Test other credentials
  shelly auth test living-room --user admin --password secret

  # Read the password from stdin (prompts without echo on a terminal)
  shelly auth test living-room --password-stdin < ~/.shelly-living-room-password

  # Quick test with short timeout
  shelly auth test living-room --timeout 5s`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmdutil.AddDeviceCredentialFlags(cmd, &opts.Credentials, "Password to test instead of the stored one", "Read the password to test from stdin")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 10*time.Second, "Connection timeout")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	auth, err := opts.Credentials.Auth(ios)
	if err != nil && !errors.Is(err, cmdutil.ErrNoCredentials) {
		return err
	}

	dev, err := svc.ResolveWithGeneration(ctx, opts.Device)
	if err != nil {
		return err
	}
	if auth != nil {
		dev.Auth = auth
	}

	ios.StartProgress(fmt.Sprintf("Testing authentication for %s...", opts.Device))
	info, err := svc.ProbeDevice(ctx, dev)
	if err == nil {
		dev.Generation = info.Generation
		err = svc.VerifyCredentials(ctx, dev)
	}
	ios.StopProgress()
	if err != nil {
		return fmt.Errorf("authentication test failed for %s: %w", opts.Device, err)
	}

	term.DisplayAuthTestPassed(ios, dev, info)
	return nil
}
