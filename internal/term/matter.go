// Package term provides composed terminal presentation for the CLI.
package term

import (
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayMatterStatus displays Matter status.
func DisplayMatterStatus(ios *iostreams.IOStreams, status model.MatterStatus, device string) {
	ios.Println(theme.Bold().Render("Matter Status:"))
	ios.Println()

	ios.Printf("  Enabled: %s\n", output.RenderEnabledState(status.Enabled))

	if status.Enabled {
		ios.Printf("  Status: %s\n", output.RenderBoolState(status.Commissionable, "Commissionable", "Not Commissionable"))
		ios.Printf("  Paired Fabrics: %d\n", status.FabricsCount)

		if status.Commissionable {
			ios.Println()
			ios.Info("Device is ready to be added to a Matter fabric.")
			ios.Info("Use 'shelly matter code %s' to get the pairing code.", device)
		}
	} else {
		ios.Println()
		ios.Info("Enable Matter with: shelly matter enable %s", device)
	}
}

// OutputMatterStatusJSON outputs Matter status as JSON.
func OutputMatterStatusJSON(ios *iostreams.IOStreams, status model.MatterStatus) error {
	jsonBytes, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(jsonBytes))
	return nil
}

// DisplayCommissioningInfo displays Matter commissioning/pairing information.
func DisplayCommissioningInfo(ios *iostreams.IOStreams, info model.CommissioningInfo, deviceIP string) {
	if info.Available && info.ManualCode != "" {
		DisplayAvailableCode(ios, info)
		return
	}
	DisplayNotAvailable(ios, deviceIP)
}

// DisplayAvailableCode displays available pairing code.
func DisplayAvailableCode(ios *iostreams.IOStreams, info model.CommissioningInfo) {
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

// DisplayNotAvailable displays instructions when pairing code is not available via API.
func DisplayNotAvailable(ios *iostreams.IOStreams, deviceIP string) {
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
