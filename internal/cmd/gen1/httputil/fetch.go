// Package httputil provides HTTP utilities for Gen1 device commands.
package httputil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DefaultTimeout is the default HTTP request timeout.
const DefaultTimeout = 10 * time.Second

// FetchEndpoint fetches JSON from a Gen1 device HTTP endpoint.
// It resolves the device configuration, handles authentication, and
// returns the parsed JSON response as a map.
func FetchEndpoint(ctx context.Context, ios *iostreams.IOStreams, device, endpoint string) (map[string]any, error) {
	devCfg, err := config.ResolveDevice(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve device: %w", err)
	}

	address := devCfg.Address
	if address == "" {
		return nil, fmt.Errorf("device %s has no address configured", device)
	}

	// Ensure http:// prefix
	if len(address) < 7 || address[:7] != "http://" {
		address = "http://" + address
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address+endpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if devCfg.Auth != nil && devCfg.Auth.Username != "" {
		req.SetBasicAuth(devCfg.Auth.Username, devCfg.Auth.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			ios.Debug("failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}
