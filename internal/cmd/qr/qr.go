// Package qr provides the qr command for generating device QR codes.
package qr

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds the command options.
type Options struct {
	WiFi       bool
	JSONOutput bool
}

// DeviceQRInfo holds device information for QR code generation.
type DeviceQRInfo struct {
	Device    string `json:"device"`
	IP        string `json:"ip"`
	MAC       string `json:"mac,omitempty"`
	Model     string `json:"model,omitempty"`
	Firmware  string `json:"firmware,omitempty"`
	WebURL    string `json:"web_url"`
	WiFiSSID  string `json:"wifi_ssid,omitempty"`
	QRContent string `json:"qr_content"`
}

// extractWiFiSSID extracts the station SSID from WiFi config.
func extractWiFiSSID(rawResult any) string {
	wifiBytes, err := json.Marshal(rawResult)
	if err != nil {
		return ""
	}
	var wifiConfig struct {
		Sta struct {
			SSID string `json:"ssid"`
		} `json:"sta"`
	}
	if err := json.Unmarshal(wifiBytes, &wifiConfig); err != nil {
		return ""
	}
	return wifiConfig.Sta.SSID
}

// NewCommand creates the qr command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "qr <device>",
		Aliases: []string{"qrcode"},
		Short:   "Generate device QR code information",
		Long: `Generate QR code information for a Shelly device.

The QR code can contain:
  - Device web interface URL (default)
  - WiFi network configuration (with --wifi flag)

The command outputs the QR content that can be used with any QR
code generator to create a scannable code.`,
		Example: `  # Show QR code content for device web UI
  shelly qr kitchen-light

  # Show WiFi configuration QR content
  shelly qr kitchen-light --wifi

  # JSON output for processing
  shelly qr kitchen-light -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.WiFi, "wifi", false, "Generate WiFi config QR content")
	cmd.Flags().BoolVarP(&opts.JSONOutput, "json", "o", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.StartProgress("Getting device info...")

	conn, err := svc.Connect(ctx, device)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing qr connection", conn)

	// Get device info
	rawResult, err := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to get device info: %w", err)
	}

	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	var deviceInfo struct {
		ID   string `json:"id"`
		MAC  string `json:"mac"`
		App  string `json:"app"`
		Ver  string `json:"ver"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jsonBytes, &deviceInfo); err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to parse device info: %w", err)
	}

	// Get WiFi config if requested
	var wifiSSID string
	if opts.WiFi {
		if wifiResult, err := conn.Call(ctx, "WiFi.GetConfig", nil); err == nil {
			wifiSSID = extractWiFiSSID(wifiResult)
		}
	}

	ios.StopProgress()

	// Build QR info - use the device identifier (which may be IP or name)
	// For the web URL, we use the device identifier since that's how we connected
	webURL := fmt.Sprintf("http://%s", device)

	qrContent := webURL
	if opts.WiFi && wifiSSID != "" {
		// WiFi QR format: WIFI:S:<SSID>;T:<TYPE>;P:<PASSWORD>;;
		// Since we don't have the password, just show the SSID
		qrContent = fmt.Sprintf("WIFI:S:%s;T:WPA;;", escapeWiFiQR(wifiSSID))
	}

	info := DeviceQRInfo{
		Device:    device,
		IP:        device, // Use device identifier (may be IP, hostname, or alias)
		MAC:       deviceInfo.MAC,
		Model:     deviceInfo.App,
		Firmware:  deviceInfo.Ver,
		WebURL:    webURL,
		WiFiSSID:  wifiSSID,
		QRContent: qrContent,
	}

	if opts.JSONOutput {
		return outputJSON(ios, info)
	}

	// Display QR info
	ios.Success("QR Code Information")
	ios.Println("")

	ios.Printf("Device: %s\n", info.Device)
	ios.Printf("IP: %s\n", info.IP)
	if info.MAC != "" {
		ios.Printf("MAC: %s\n", info.MAC)
	}
	if info.Model != "" {
		ios.Printf("Model: %s\n", info.Model)
	}
	ios.Println("")

	ios.Info("QR Content:")
	ios.Println(info.QRContent)
	ios.Println("")

	// Show ASCII QR representation placeholder
	ios.Info("To generate a QR code, use:")
	ios.Printf("  echo '%s' | qrencode -t UTF8\n", info.QRContent)
	ios.Println("  (requires qrencode package)")
	ios.Println("")

	ios.Info("Or use any online QR generator with the above content.")

	return nil
}

// escapeWiFiQR escapes special characters in WiFi QR content.
func escapeWiFiQR(s string) string {
	// Escape special characters for WiFi QR format
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, ":", "\\:")
	return s
}

func outputJSON(ios *iostreams.IOStreams, info DeviceQRInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	ios.Println(string(data))
	return nil
}
