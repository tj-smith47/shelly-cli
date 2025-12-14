package ota

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// CheckOptions holds check command options.
type CheckOptions struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

func newCheckCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &CheckOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "check <device>",
		Short: "Check for firmware updates",
		Long: `Check if a firmware update is available for a Gen1 device.

This queries the device's OTA endpoint to see if a newer
firmware version is available.`,
		Example: `  # Check for updates
  shelly gen1 ota check living-room

  # Output as JSON
  shelly gen1 ota check living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runCheck(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func runCheck(ctx context.Context, opts *CheckOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	ios.StartProgress("Checking for updates...")
	info, err := gen1Client.CheckForUpdate(ctx)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if opts.JSON {
		output, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	ios.Println(theme.Bold().Render("Firmware Update Status:"))
	ios.Println()

	if info.HasUpdate {
		ios.Printf("  Current:   %s\n", gen1Client.Info().Firmware)
		ios.Printf("  Available: %s\n", theme.StatusOK().Render(info.NewVersion))
		ios.Println()
		ios.Info("Update available! Run: shelly gen1 ota update %s", opts.Device)
	} else {
		ios.Printf("  Current: %s\n", gen1Client.Info().Firmware)
		ios.Success("Firmware is up to date")
	}

	return nil
}

// connectGen1 resolves device config and connects to a Gen1 device.
func connectGen1(ctx context.Context, ios *iostreams.IOStreams, deviceName string) (*client.Gen1Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	devCfg, err := config.ResolveDevice(deviceName)
	if err != nil {
		return nil, err
	}

	device := model.Device{
		Name:    devCfg.Name,
		Address: devCfg.Address,
	}
	if devCfg.Auth != nil {
		device.Auth = &model.Auth{
			Username: devCfg.Auth.Username,
			Password: devCfg.Auth.Password,
		}
	}

	ios.StartProgress("Connecting to device...")
	gen1Client, err := client.ConnectGen1(ctx, device)
	ios.StopProgress()

	if err != nil {
		return nil, err
	}

	return gen1Client, nil
}
