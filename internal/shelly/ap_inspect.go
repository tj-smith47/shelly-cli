package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// APInspection is the live, persisted state read from a device sitting at its
// factory WiFi AP: its identity plus the WiFi station config it would use to join
// a network. It answers "did the provisioning actually take?" without the device
// having to successfully join the LAN first — invaluable when a device configures
// but never appears on the network.
type APInspection struct {
	Model        string
	MAC          string
	Firmware     string
	StaSSID      string
	Ipv4Method   string
	StaIP        string
	StaGateway   string
	Generation   int
	StaKeySet    bool
	StaConnected bool
}

// InspectAtAP hops the host onto a device's factory WiFi AP, reads the device's
// identity and persisted WiFi station configuration at discovery.DefaultAPIP, and
// returns the host to its home network. Use it to verify what a --to-ap restore or
// onboard actually wrote to a device that isn't reaching the LAN.
func (s *Service) InspectAtAP(ctx context.Context, apSSID, apHostIP string) (*APInspection, error) {
	var (
		insp    *APInspection
		readErr error
	)
	hopErr := s.withAPHop(ctx, apSSID, apHostIP, &OnboardWiFiConfig{}, func(ctx context.Context) error {
		insp, readErr = s.readDeviceAtAP(ctx, discovery.DefaultAPIP)
		return readErr
	})
	if readErr != nil {
		return nil, fmt.Errorf("read device at AP %q: %w", apSSID, readErr)
	}
	if hopErr != nil {
		return nil, fmt.Errorf("AP hop for %q failed: %w", apSSID, hopErr)
	}
	return insp, nil
}

// readDeviceAtAP identifies the device via the universal /shelly probe and, for a
// Gen1 device, reads its persisted WiFi station config and live connection state.
// Identity is returned even when the Gen1 WiFi read fails, so the caller still
// learns what the device is.
func (s *Service) readDeviceAtAP(ctx context.Context, addr string) (*APInspection, error) {
	det, err := client.DetectGeneration(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("identify device: %w", err)
	}
	insp := &APInspection{
		Generation: int(det.Generation),
		Model:      det.Model,
		MAC:        det.MAC,
		Firmware:   det.Firmware,
	}
	if det.Generation == client.Gen1 {
		if wifiErr := s.readGen1WiFiAtAP(ctx, addr, insp); wifiErr != nil {
			return insp, fmt.Errorf("read Gen1 WiFi config: %w", wifiErr)
		}
	}
	return insp, nil
}

// readGen1WiFiAtAP fills insp with the device's persisted station settings (from
// /settings) and its live connection state (from /status). The key is masked by
// the device, so only its presence is reported.
func (s *Service) readGen1WiFiAtAP(ctx context.Context, addr string, insp *APInspection) error {
	return s.WithGen1Connection(ctx, addr, func(conn *client.Gen1Client) error {
		settings, err := conn.GetSettings(ctx)
		if err != nil {
			return err
		}
		if settings.WiFiSta != nil {
			insp.StaSSID = settings.WiFiSta.SSID
			insp.StaKeySet = settings.WiFiSta.Key != ""
			insp.Ipv4Method = settings.WiFiSta.Ipv4Method
			insp.StaIP = settings.WiFiSta.IP
			insp.StaGateway = settings.WiFiSta.Gw
		}
		// Live status is best-effort: a device that just had its STA reconfigured
		// may not have associated yet, which is exactly what we want to observe.
		if status, statusErr := conn.GetStatus(ctx); statusErr == nil && status.WiFiSta != nil {
			insp.StaConnected = status.WiFiSta.Connected
			if status.WiFiSta.IP != "" {
				insp.StaIP = status.WiFiSta.IP
			}
		}
		return nil
	})
}
