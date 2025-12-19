// Package qr provides the qr command for generating device QR codes.
package qr

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	WiFi    bool
	NoQR    bool
	Size    int
	Factory *cmdutil.Factory
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

// NewCommand creates the qr command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f, Size: 256}

	cmd := &cobra.Command{
		Use:     "qr <device>",
		Aliases: []string{"qrcode"},
		Short:   "Generate device QR code",
		Long: `Generate a QR code for a Shelly device.

The QR code can contain:
  - Device web interface URL (default)
  - WiFi network configuration (with --wifi flag)

By default, displays the QR code as ASCII art in the terminal.`,
		Example: `  # Generate QR code for device web UI
  shelly qr kitchen-light

  # Generate WiFi configuration QR code
  shelly qr kitchen-light --wifi

  # Show only the content without QR display
  shelly qr kitchen-light --no-qr

  # JSON output with QR content
  shelly qr kitchen-light -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.WiFi, "wifi", false, "Generate WiFi config QR content")
	cmd.Flags().BoolVar(&opts.NoQR, "no-qr", false, "Don't display QR code, just show content")
	cmd.Flags().IntVar(&opts.Size, "size", 256, "QR code size in pixels (for --save)")

	return cmd
}

func run(ctx context.Context, device string, opts *Options) error {
	f := opts.Factory
	ios := f.IOStreams()
	svc := f.ShellyService()

	var deviceInfo struct {
		ID   string `json:"id"`
		MAC  string `json:"mac"`
		App  string `json:"app"`
		Ver  string `json:"ver"`
		Name string `json:"name"`
	}
	var wifiSSID string

	err := cmdutil.RunWithSpinner(ctx, ios, "Getting device info...", func(ctx context.Context) error {
		conn, connErr := svc.Connect(ctx, device)
		if connErr != nil {
			return fmt.Errorf("failed to connect to device: %w", connErr)
		}
		defer iostreams.CloseWithDebug("closing qr connection", conn)

		// Get device info
		rawResult, callErr := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
		if callErr != nil {
			return fmt.Errorf("failed to get device info: %w", callErr)
		}

		jsonBytes, marshalErr := json.Marshal(rawResult)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal device info: %w", marshalErr)
		}

		if unmarshalErr := json.Unmarshal(jsonBytes, &deviceInfo); unmarshalErr != nil {
			return fmt.Errorf("failed to parse device info: %w", unmarshalErr)
		}

		// Get WiFi config if requested
		if opts.WiFi {
			if wifiResult, wifiErr := conn.Call(ctx, "WiFi.GetConfig", nil); wifiErr == nil {
				wifiSSID = shelly.ExtractWiFiSSID(wifiResult)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Build QR info
	webURL := fmt.Sprintf("http://%s", device)

	qrContent := webURL
	if opts.WiFi && wifiSSID != "" {
		// WiFi QR format: WIFI:S:<SSID>;T:<TYPE>;P:<PASSWORD>;;
		// Since we don't have the password, just show the SSID
		qrContent = fmt.Sprintf("WIFI:S:%s;T:WPA;;", output.EscapeWiFiQR(wifiSSID))
	}

	info := DeviceQRInfo{
		Device:    device,
		IP:        device,
		MAC:       deviceInfo.MAC,
		Model:     deviceInfo.App,
		Firmware:  deviceInfo.Ver,
		WebURL:    webURL,
		WiFiSSID:  wifiSSID,
		QRContent: qrContent,
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, info)
	}

	// Display device info
	ios.Success("QR Code for %s", info.Device)
	ios.Println()

	if info.MAC != "" {
		ios.Printf("MAC: %s\n", info.MAC)
	}
	if info.Model != "" {
		ios.Printf("Model: %s\n", info.Model)
	}
	ios.Println()

	// Generate and display QR code
	if !opts.NoQR {
		qr, qrErr := qrcode.New(qrContent, qrcode.Medium)
		if qrErr != nil {
			ios.Warning("Failed to generate QR code: %v", qrErr)
		} else {
			// Display ASCII QR code
			ios.Println(qr.ToSmallString(false))
		}
	}

	// Show the content
	ios.Info("Content: %s", qrContent)

	return nil
}
