// Package download provides the firmware download subcommand.
package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/firmware"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	outputFlag string
	latestFlag bool
	betaFlag   bool
)

// NewCommand creates the firmware download command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download <device>",
		Aliases: []string{"dl"},
		Short:   "Download firmware file",
		Long: `Download firmware file for a device.

Downloads the latest available firmware for the device.
The firmware URL is determined by querying the device.`,
		Example: `  # Download latest firmware for a device
  shelly firmware download living-room

  # Download to specific file
  shelly firmware download living-room --output firmware.zip

  # Download beta firmware
  shelly firmware download living-room --beta`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path")
	cmd.Flags().BoolVar(&latestFlag, "latest", true, "Download latest version")
	cmd.Flags().BoolVar(&betaFlag, "beta", false, "Download beta version")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Determine stage
	stage := "stable"
	if betaFlag {
		stage = "beta"
	}

	// Get firmware URL from device
	ios.StartProgress("Getting firmware information...")
	fwURL, err := svc.GetFirmwareURL(ctx, device, stage)
	ios.StopProgress()
	if err != nil {
		return fmt.Errorf("failed to get firmware URL: %w", err)
	}

	// Determine output filename
	outputPath := outputFlag
	if outputPath == "" {
		// Default to firmware_<device>_<stage>.zip
		outputPath = fmt.Sprintf("firmware_%s_%s.zip", device, stage)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	ios.Printf("Downloading from: %s\n", fwURL)
	ios.Printf("Saving to: %s\n", outputPath)

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*10) // Extended timeout for downloads
	defer cancel()

	return cmdutil.RunWithSpinner(ctx, ios, "Downloading firmware...", func(ctx context.Context) error {
		downloader := firmware.NewDownloader()
		result, downloadErr := downloader.Download(ctx, fwURL)
		if downloadErr != nil {
			return fmt.Errorf("download failed: %w", downloadErr)
		}

		//nolint:gosec // G306: 0o644 is appropriate for downloaded firmware files
		if writeErr := os.WriteFile(outputPath, result.Data, 0o644); writeErr != nil {
			return fmt.Errorf("failed to write file: %w", writeErr)
		}

		ios.Success("Downloaded firmware to %s (%d bytes)", outputPath, result.Size)
		return nil
	})
}
