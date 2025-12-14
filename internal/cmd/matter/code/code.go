// Package code provides the matter code command.
package code

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the matter code command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "code <device>",
		Aliases: []string{"pairing", "qr"},
		Short:   "Show Matter pairing code",
		Long: `Show the Matter pairing code for commissioning a device.

Displays the commissioning information needed to add the device
to a Matter fabric (Apple Home, Google Home, etc.):
- Manual pairing code (11-digit number)
- QR code data (for compatible apps)
- Discriminator and setup PIN

If the pairing code is not available via the API, check the device
label or web UI at http://<device-ip>/matter for the QR code.`,
		Example: `  # Show pairing code
  shelly matter code living-room

  # Output as JSON
  shelly matter code living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// CommissioningInfo holds pairing information.
type CommissioningInfo struct {
	ManualCode    string `json:"manual_code,omitempty"`
	QRCode        string `json:"qr_code,omitempty"`
	Discriminator int    `json:"discriminator,omitempty"`
	SetupPINCode  int    `json:"setup_pin_code,omitempty"`
	Available     bool   `json:"available"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Check if device is commissionable
	commissionable, err := svc.MatterIsCommissionable(ctx, opts.Device)
	if err != nil {
		return err
	}

	if !commissionable {
		ios.Warning("Device is not commissionable.")
		ios.Info("Enable Matter first: shelly matter enable %s", opts.Device)
		if opts.JSON {
			return outputJSON(ios, CommissioningInfo{Available: false})
		}
		return nil
	}

	// Get commissioning code
	codeInfo, err := svc.MatterGetCommissioningCode(ctx, opts.Device)
	if err != nil {
		ios.Debug("failed to get commissioning code: %v", err)
		// Code not available via API, show instructions
		if opts.JSON {
			return outputJSON(ios, CommissioningInfo{Available: false})
		}
		displayNotAvailable(ios, opts.Device)
		return nil
	}

	info := CommissioningInfo{
		ManualCode:    codeInfo.ManualCode,
		QRCode:        codeInfo.QRCode,
		Discriminator: codeInfo.Discriminator,
		SetupPINCode:  codeInfo.SetupPINCode,
		Available:     codeInfo.ManualCode != "",
	}

	if opts.JSON {
		return outputJSON(ios, info)
	}

	displayCommissioningInfo(ios, info, opts.Device)
	return nil
}

func outputJSON(ios *iostreams.IOStreams, info CommissioningInfo) error {
	output, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(output))
	return nil
}

func displayCommissioningInfo(ios *iostreams.IOStreams, info CommissioningInfo, device string) {
	if info.Available && info.ManualCode != "" {
		displayAvailableCode(ios, info)
		return
	}
	displayNotAvailable(ios, device)
}

func displayAvailableCode(ios *iostreams.IOStreams, info CommissioningInfo) {
	ios.Println(theme.Bold().Render("Matter Pairing Code:"))
	ios.Println()
	ios.Printf("  Manual Code: %s\n", theme.Highlight().Render(info.ManualCode))
	if info.QRCode != "" {
		ios.Printf("  QR Data: %s\n", info.QRCode)
	}
	if info.Discriminator > 0 {
		ios.Printf("  Discriminator: %d\n", info.Discriminator)
	}
	if info.SetupPINCode > 0 {
		ios.Printf("  Setup PIN: %d\n", info.SetupPINCode)
	}
	ios.Println()
	ios.Info("Use this code in your Matter controller app.")
}

func displayNotAvailable(ios *iostreams.IOStreams, device string) {
	deviceIP := ""
	if devCfg, cfgErr := config.ResolveDevice(device); cfgErr == nil {
		deviceIP = devCfg.Address
	}

	ios.Println(theme.Bold().Render("Matter Pairing Information:"))
	ios.Println()
	ios.Info("Pairing code not available via API.")
	ios.Println()
	ios.Info("To get the pairing code:")
	ios.Info("  1. Check the device label for QR code")
	if deviceIP != "" {
		ios.Info("  2. Visit: http://%s/matter", deviceIP)
	} else {
		ios.Info("  2. Visit the device web UI at /matter")
	}
	ios.Info("  3. Use your Matter controller app to scan the QR code")
}
