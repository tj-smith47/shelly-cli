// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Generation represents a Shelly device generation.
type Generation int

const (
	// GenerationUnknown indicates the generation could not be determined.
	GenerationUnknown Generation = 0
	// Gen1 indicates a Gen1 device (uses REST API).
	Gen1 Generation = 1
	// Gen2 indicates a Gen2+ device (uses RPC API).
	Gen2 Generation = 2
)

// DetectionResult contains device detection information.
type DetectionResult struct {
	Generation Generation
	DeviceType string
	Model      string
	MAC        string
	Firmware   string
	AuthEn     bool
}

// DetectGeneration probes a device to determine its generation.
//
// The /shelly endpoint is the universal probe: BOTH Gen1 and Gen2+ devices serve
// it and report their generation (Gen2+ via the "gen" field, Gen1 via gen<2 or
// its absence). /rpc/Shelly.GetDeviceInfo, by contrast, is Gen2-only — a Gen1
// device cannot answer it and will burn the full client timeout before failing.
// So /shelly is tried FIRST: a Gen1 device is identified on the first round-trip
// and never waits on the Gen2 probe. /rpc is kept only as a fallback for the rare
// Gen2+ device whose /shelly compatibility endpoint is unavailable. Probing Gen2
// first (the previous order) misrouted every Gen1 device reached by a bare IP —
// detection would stall on /rpc, and an inconclusive result left the generation
// unknown, which downstream routing silently treats as Gen2 (RPC).
func DetectGeneration(ctx context.Context, address string, auth *model.Auth) (*DetectionResult, error) {
	url := ensureHTTPScheme(address)

	transport := cloneDefaultTransport()
	if strings.HasPrefix(url, "https") {
		// Shelly devices ship self-signed certs, matching the convention in
		// client.go/gen1.go which skip verification for https:// endpoints.
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Shelly devices use self-signed TLS certs; skipping verification is intentional
	}

	client := &http.Client{Timeout: 5 * time.Second, Transport: transport}

	// Try the universal /shelly endpoint first — it identifies a Gen1 device on
	// the first round-trip and never stalls on the Gen2-only /rpc probe.
	gen1Result, gen1Err := tryGen1Detection(ctx, client, url, auth)
	if gen1Err == nil {
		return gen1Result, nil
	}

	// Fallback: a Gen2+ device whose /shelly compatibility endpoint did not
	// answer (or reported gen>=2, which tryGen1Detection rejects) is confirmed
	// via the Gen2 RPC endpoint.
	gen2Result, gen2Err := tryGen2Detection(ctx, client, url, auth)
	if gen2Err == nil {
		return gen2Result, nil
	}

	return nil, fmt.Errorf("failed to detect device generation: gen1 error: %w, gen2 error: %w", gen1Err, gen2Err)
}

// tryGen2Detection attempts to detect a Gen2+ device.
func tryGen2Detection(ctx context.Context, client *http.Client, baseURL string, auth *model.Auth) (*DetectionResult, error) {
	url := baseURL + "/rpc/Shelly.GetDeviceInfo"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if auth != nil && auth.Username != "" {
		req.SetBasicAuth(auth.Username, auth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			iostreams.DebugErrCat(iostreams.CategoryNetwork, "closing gen2 detection response body", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info struct {
		ID       string `json:"id"`
		MAC      string `json:"mac"`
		Model    string `json:"model"`
		Gen      int    `json:"gen"`
		FW       string `json:"fw_id"`
		App      string `json:"app"`
		AuthEn   bool   `json:"auth_en"`
		Type     string `json:"type"`
		DevType  string `json:"dev_type"`
		Firmware string `json:"ver"`
	}

	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	// Gen2+ devices have gen >= 2
	if info.Gen < 2 {
		return nil, fmt.Errorf("device reports gen %d, expected >= 2", info.Gen)
	}

	firmware := info.Firmware
	if firmware == "" {
		firmware = info.FW
	}

	return &DetectionResult{
		Generation: Gen2,
		DeviceType: firstNonEmpty(info.App, info.Type, info.DevType),
		Model:      info.Model,
		MAC:        info.MAC,
		Firmware:   firmware,
		AuthEn:     info.AuthEn,
	}, nil
}

// tryGen1Detection attempts to detect a Gen1 device.
func tryGen1Detection(ctx context.Context, client *http.Client, baseURL string, auth *model.Auth) (*DetectionResult, error) {
	url := baseURL + "/shelly"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	if auth != nil && auth.Username != "" {
		req.SetBasicAuth(auth.Username, auth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			iostreams.DebugErrCat(iostreams.CategoryNetwork, "closing gen1 detection response body", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info struct {
		Type      string `json:"type"`
		MAC       string `json:"mac"`
		Auth      bool   `json:"auth"`
		FW        string `json:"fw"`
		NumRelays int    `json:"num_outputs"`
		NumMeters int    `json:"num_meters"`
		Gen       int    `json:"gen"`
	}

	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	// Gen1 devices either have gen=1 or no gen field (defaults to 0)
	// Gen2+ devices would have gen >= 2
	if info.Gen >= 2 {
		return nil, fmt.Errorf("device reports gen %d, expected gen1", info.Gen)
	}

	return &DetectionResult{
		Generation: Gen1,
		DeviceType: info.Type,
		Model:      info.Type, // Gen1 uses type as model
		MAC:        info.MAC,
		Firmware:   info.FW,
		AuthEn:     info.Auth,
	}, nil
}

// cloneDefaultTransport returns a clone of http.DefaultTransport, falling back
// to a fresh *http.Transport if the default is ever replaced with a type that
// is not *http.Transport (so the detection client always has a transport whose
// TLSClientConfig can be customized for self-signed Shelly certs).
func cloneDefaultTransport() *http.Transport {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		return t.Clone()
	}
	return &http.Transport{}
}

// httpStatusError returns a user-friendly error for common HTTP status codes.
func httpStatusError(statusCode int) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication required (HTTP 401)")
	case http.StatusForbidden:
		return fmt.Errorf("access denied (HTTP 403)")
	case http.StatusNotFound:
		return fmt.Errorf("device not found or endpoint not available (HTTP 404)")
	case http.StatusServiceUnavailable:
		return fmt.Errorf("device busy or unavailable (HTTP 503)")
	case http.StatusGatewayTimeout:
		return fmt.Errorf("device timeout (HTTP 504)")
	default:
		return fmt.Errorf("unexpected HTTP status: %d", statusCode)
	}
}

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}

// IsGen1 returns true if the device is Gen1.
func (r *DetectionResult) IsGen1() bool {
	return r.Generation == Gen1
}

// IsGen2 returns true if the device is Gen2+.
func (r *DetectionResult) IsGen2() bool {
	return r.Generation == Gen2
}
