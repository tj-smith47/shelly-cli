// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"
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
// Gen2+ devices respond to /rpc/Shelly.GetDeviceInfo.
// Gen1 devices respond to /shelly.
func DetectGeneration(ctx context.Context, address string, auth *model.Auth) (*DetectionResult, error) {
	url := address
	if url != "" && !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// Try Gen2 RPC endpoint first (more common for newer devices)
	gen2Result, gen2Err := tryGen2Detection(ctx, client, url, auth)
	if gen2Err == nil {
		return gen2Result, nil
	}

	// Try Gen1 endpoint
	gen1Result, gen1Err := tryGen1Detection(ctx, client, url, auth)
	if gen1Err == nil {
		return gen1Result, nil
	}

	return nil, fmt.Errorf("failed to detect device generation: gen2 error: %w, gen1 error: %w", gen2Err, gen1Err)
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
