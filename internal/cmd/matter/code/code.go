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

	info, err := getCommissioningInfo(ctx, svc, opts.Device, ios)
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputJSON(ios, info)
	}

	displayCommissioningInfo(ios, info, opts.Device)
	return nil
}

func getCommissioningInfo(ctx context.Context, svc *shelly.Service, device string, ios *iostreams.IOStreams) (CommissioningInfo, error) {
	var info CommissioningInfo

	err := svc.WithConnection(ctx, device, func(conn *client.Client) error {
		commissionable, err := checkCommissionable(ctx, conn, ios)
		if err != nil {
			return err
		}
		if !commissionable {
			ios.Warning("Device is not commissionable.")
			ios.Info("Enable Matter first: shelly matter enable %s", device)
			return nil
		}

		fetchedInfo, fetchErr := fetchCommissioningCode(ctx, conn, ios)
		if fetchErr == nil {
			info = fetchedInfo
		}
		return nil
	})

	return info, err
}

func checkCommissionable(ctx context.Context, conn *client.Client, ios *iostreams.IOStreams) (bool, error) {
	statusResult, err := conn.Call(ctx, "Matter.GetStatus", nil)
	if err != nil {
		ios.Debug("Matter.GetStatus failed: %v", err)
		return false, fmt.Errorf("matter not available on this device: %w", err)
	}

	var st struct {
		Commissionable bool `json:"commissionable"`
	}
	statusBytes, err := json.Marshal(statusResult)
	if err != nil {
		return false, fmt.Errorf("failed to marshal status: %w", err)
	}
	if err := json.Unmarshal(statusBytes, &st); err != nil {
		return false, fmt.Errorf("failed to parse status: %w", err)
	}

	return st.Commissionable, nil
}

func fetchCommissioningCode(ctx context.Context, conn *client.Client, ios *iostreams.IOStreams) (CommissioningInfo, error) {
	var info CommissioningInfo

	codeResult, err := conn.Call(ctx, "Matter.GetCommissioningCode", nil)
	if err != nil {
		ios.Debug("Matter.GetCommissioningCode not available: %v", err)
		return info, err
	}

	codeBytes, err := json.Marshal(codeResult)
	if err != nil {
		ios.Debug("failed to marshal code: %v", err)
		return info, err
	}
	if err := json.Unmarshal(codeBytes, &info); err != nil {
		ios.Debug("failed to parse code: %v", err)
		return info, err
	}
	info.Available = true

	return info, nil
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
