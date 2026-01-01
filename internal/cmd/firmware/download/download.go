// Package download provides the firmware download subcommand.
package download

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/firmware"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Beta    bool
	Device  string
	Latest  bool
	Output  string
}

// NewCommand creates the firmware download command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		Latest:  true,
	}

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
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path")
	cmd.Flags().BoolVar(&opts.Latest, "latest", true, "Download latest version")
	cmd.Flags().BoolVar(&opts.Beta, "beta", false, "Download beta version")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Determine stage
	stage := "stable"
	if opts.Beta {
		stage = "beta"
	}

	// Get firmware URL from device
	var fwURL string
	err := cmdutil.RunWithSpinner(ctx, ios, "Getting firmware information...", func(ctx context.Context) error {
		var urlErr error
		fwURL, urlErr = svc.GetFirmwareURL(ctx, opts.Device, stage)
		return urlErr
	})
	if err != nil {
		return fmt.Errorf("failed to get firmware URL: %w", err)
	}

	// Determine output filename
	outputPath := opts.Output
	if outputPath == "" {
		// Default to firmware_<device>_<stage>.zip
		outputPath = fmt.Sprintf("firmware_%s_%s.zip", opts.Device, stage)
	}

	// Ensure parent directory exists
	fs := config.Fs()
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := fs.MkdirAll(dir, 0o750); err != nil {
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

		if writeErr := afero.WriteFile(fs, outputPath, result.Data, 0o644); writeErr != nil {
			return fmt.Errorf("failed to write file: %w", writeErr)
		}

		ios.Success("Downloaded firmware to %s (%d bytes)", outputPath, result.Size)
		return nil
	})
}
