// Package device provides the cloud device subcommand.
package device

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var statusFlag bool

// NewCommand creates the cloud device command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device <id>",
		Aliases: []string{"get"},
		Short:   "Show cloud device details",
		Long: `Show details for a specific device from Shelly Cloud.

Displays device information including status, settings, and online state.`,
		Example: `  # Get device details
  shelly cloud device abc123

  # Get device with full status
  shelly cloud device abc123 --status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().BoolVar(&statusFlag, "status", false, "Show full device status")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, deviceID string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2)
	defer cancel()

	ios := f.IOStreams()

	// Check if logged in
	cfg := config.Get()
	if cfg.Cloud.AccessToken == "" {
		ios.Error("Not logged in to Shelly Cloud")
		ios.Info("Use 'shelly cloud login' to authenticate")
		return fmt.Errorf("not logged in")
	}

	// Create cloud client
	client := shelly.NewCloudClient(cfg.Cloud.AccessToken)

	return cmdutil.RunWithSpinner(ctx, ios, "Fetching device from cloud...", func(ctx context.Context) error {
		device, err := client.GetDevice(ctx, deviceID)
		if err != nil {
			return fmt.Errorf("failed to get device: %w", err)
		}

		displayDevice(ios, device)
		return nil
	})
}

func displayDevice(ios *iostreams.IOStreams, device *shelly.CloudDevice) {
	ios.Title("Cloud Device")
	ios.Println()

	ios.Printf("  ID:     %s\n", device.ID)

	if device.Model != "" {
		ios.Printf("  Model:  %s\n", device.Model)
	}

	if device.Generation > 0 {
		ios.Printf("  Gen:    %d\n", device.Generation)
	}

	if device.MAC != "" {
		ios.Printf("  MAC:    %s\n", device.MAC)
	}

	if device.FirmwareVersion != "" {
		ios.Printf("  FW:     %s\n", device.FirmwareVersion)
	}

	ios.Printf("  Status: %s\n", output.RenderOnline(device.Online, output.CaseTitle))

	// Show status JSON if requested and available
	if statusFlag && len(device.Status) > 0 {
		ios.Println()
		ios.Title("Device Status")
		ios.Println()
		printJSON(ios, device.Status)
	}

	// Show settings if available
	if statusFlag && len(device.Settings) > 0 {
		ios.Println()
		ios.Title("Device Settings")
		ios.Println()
		printJSON(ios, device.Settings)
	}
}

func printJSON(ios *iostreams.IOStreams, data json.RawMessage) {
	var prettyJSON map[string]any
	if err := json.Unmarshal(data, &prettyJSON); err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	formatted, err := json.MarshalIndent(prettyJSON, "  ", "  ")
	if err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	ios.Printf("  %s\n", string(formatted))
}
