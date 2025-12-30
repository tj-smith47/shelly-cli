// Package cmd implements the plugin commands for the Tasmota plugin.
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/examples/plugins/shelly-tasmota/tasmota"
)

// UpdateResult is returned by the apply_update hook.
// Matches the structure expected by shelly-cli.
type UpdateResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	Rebooting bool   `json:"rebooting,omitempty"`
}

var applyUpdateFlags struct {
	address  string
	authUser string
	authPass string
	stage    string
	url      string
	timeout  time.Duration
}

// NewApplyUpdateCmd creates the apply-update command.
func NewApplyUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply-update",
		Short: "Apply a firmware update to a Tasmota device",
		Long: `Trigger an OTA firmware update on a Tasmota device.

You can either:
  - Provide a custom --url for the firmware binary
  - Use --stage=stable or --stage=beta to use official Tasmota OTA URLs

The command sets the OTA URL on the device and triggers the upgrade process.
The device will reboot automatically after downloading and applying the update.`,
		Example: `  # Update to latest stable
  shelly-tasmota apply-update --address=192.168.1.50 --stage=stable

  # Update to beta
  shelly-tasmota apply-update --address=192.168.1.50 --stage=beta

  # Update from custom URL
  shelly-tasmota apply-update --address=192.168.1.50 --url=http://example.com/tasmota.bin.gz`,
		RunE: runApplyUpdate,
	}

	cmd.Flags().StringVar(&applyUpdateFlags.address, "address", "", "Device IP address (required)")
	cmd.Flags().StringVar(&applyUpdateFlags.authUser, "auth-user", "", "HTTP auth username")
	cmd.Flags().StringVar(&applyUpdateFlags.authPass, "auth-pass", "", "HTTP auth password")
	cmd.Flags().StringVar(&applyUpdateFlags.stage, "stage", "", "Release stage: stable or beta")
	cmd.Flags().StringVar(&applyUpdateFlags.url, "url", "", "Custom firmware URL (overrides --stage)")
	cmd.Flags().DurationVar(&applyUpdateFlags.timeout, "timeout", 30*time.Second, "Request timeout")

	if err := cmd.MarkFlagRequired("address"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to mark flag required: %v\n", err)
	}

	return cmd
}

func runApplyUpdate(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), applyUpdateFlags.timeout)
	defer cancel()

	client := tasmota.NewClient(applyUpdateFlags.address, applyUpdateFlags.authUser, applyUpdateFlags.authPass)

	// Determine the OTA URL
	otaURL := applyUpdateFlags.url
	if otaURL == "" {
		// Build URL based on stage and device info
		var err error
		otaURL, err = buildUpdateURL(ctx, client, applyUpdateFlags.stage)
		if err != nil {
			result := UpdateResult{
				Success: false,
				Error:   err.Error(),
			}
			return outputJSON(result)
		}
	}

	// Set the OTA URL on the device
	if err := client.SetOtaURL(ctx, otaURL); err != nil {
		result := UpdateResult{
			Success: false,
			Error:   fmt.Sprintf("failed to set OTA URL: %v", err),
		}
		return outputJSON(result)
	}

	// Trigger the upgrade
	upgradeResp, err := client.Upgrade(ctx)
	if err != nil {
		result := UpdateResult{
			Success: false,
			Error:   fmt.Sprintf("failed to trigger upgrade: %v", err),
		}
		return outputJSON(result)
	}

	// Check the response
	result := UpdateResult{
		Success:   true,
		Message:   fmt.Sprintf("Upgrade initiated from %s", otaURL),
		Rebooting: true,
	}

	// Tasmota returns "Upgrade started" or similar message
	if upgradeResp.Upgrade != "" {
		result.Message = upgradeResp.Upgrade
	}

	return outputJSON(result)
}

// buildUpdateURL constructs the appropriate OTA URL based on device info and stage.
func buildUpdateURL(ctx context.Context, client *tasmota.Client, stage string) (string, error) {
	if stage == "" {
		stage = "stable"
	}

	// Get firmware info to determine chip type and variant
	fwInfo, err := client.GetStatusFirmware(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get firmware info: %w", err)
	}

	chipType := determineChipType(fwInfo.Hardware, fwInfo.Version)
	variant := determineVariant(fwInfo.Version)

	stable := stage == "stable"
	return buildOTAURL(chipType, variant, stable), nil
}
