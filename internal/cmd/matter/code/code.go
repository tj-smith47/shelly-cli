// Package code provides the matter code command.
package code

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
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
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
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

	var info CommissioningInfo

	// Get device address for the web UI hint
	var deviceIP string
	if devCfg, cfgErr := config.ResolveDevice(opts.Device); cfgErr == nil {
		deviceIP = devCfg.Address
	}

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		// First check if Matter is enabled and commissionable
		statusResult, err := conn.Call(ctx, "Matter.GetStatus", nil)
		if err != nil {
			ios.Debug("Matter.GetStatus failed: %v", err)
			return fmt.Errorf("matter not available on this device: %w", err)
		}

		var st struct {
			Commissionable bool `json:"commissionable"`
		}
		statusBytes, statusMarshalErr := json.Marshal(statusResult)
		if statusMarshalErr != nil {
			return fmt.Errorf("failed to marshal status: %w", statusMarshalErr)
		}
		if err := json.Unmarshal(statusBytes, &st); err != nil {
			return fmt.Errorf("failed to parse status: %w", err)
		}

		if !st.Commissionable {
			ios.Warning("Device is not commissionable.")
			ios.Info("Enable Matter first: shelly matter enable %s", opts.Device)
			return nil
		}

		// Try to get commissioning code via RPC
		// Note: Not all devices expose this via API
		codeResult, err := conn.Call(ctx, "Matter.GetCommissioningCode", nil)
		if err != nil {
			ios.Debug("Matter.GetCommissioningCode not available: %v", err)
			// Commissioning code may not be available via API
			return nil
		}

		codeBytes, codeMarshalErr := json.Marshal(codeResult)
		if codeMarshalErr != nil {
			ios.Debug("failed to marshal code: %v", codeMarshalErr)
			return nil
		}
		if err := json.Unmarshal(codeBytes, &info); err != nil {
			ios.Debug("failed to parse code: %v", err)
			return nil
		}
		info.Available = true

		return nil
	})
	if err != nil {
		return err
	}

	if opts.JSON {
		output, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	if info.Available && info.ManualCode != "" {
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
	} else {
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

	return nil
}
