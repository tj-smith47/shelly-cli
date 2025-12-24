// Package tasmota provides HTTP client and types for Tasmota device communication.
package tasmota

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is an HTTP client for Tasmota devices.
type Client struct {
	address  string
	user     string
	password string
	client   *http.Client
}

// NewClient creates a new Tasmota client.
func NewClient(address, user, password string) *Client {
	return &Client{
		address:  address,
		user:     user,
		password: password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ExecuteCommand sends a command to the Tasmota device and returns the raw response.
// Uses GET /cm?user=<user>&password=<pass>&cmnd=<command>.
func (c *Client) ExecuteCommand(ctx context.Context, command string) ([]byte, error) {
	endpoint := fmt.Sprintf("http://%s/cm", c.address)

	params := url.Values{}
	params.Set("cmnd", command)
	if c.user != "" {
		params.Set("user", c.user)
		params.Set("password", c.password)
	}

	reqURL := endpoint + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close on body

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// ExecuteCommandJSON sends a command and unmarshals the response into the target.
func (c *Client) ExecuteCommandJSON(ctx context.Context, command string, target any) error {
	body, err := c.ExecuteCommand(ctx, command)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse response: %w (body: %s)", err, string(body))
	}

	return nil
}

// GetStatus returns the full device status (Status 0).
func (c *Client) GetStatus(ctx context.Context) (*StatusAll, error) {
	var status StatusAll
	if err := c.ExecuteCommandJSON(ctx, "Status 0", &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// GetStatusInfo returns basic status info (Status 1).
func (c *Client) GetStatusInfo(ctx context.Context) (*StatusInfo, error) {
	var resp struct {
		StatusSTS StatusInfo `json:"StatusSTS"`
	}
	if err := c.ExecuteCommandJSON(ctx, "Status 1", &resp); err != nil {
		return nil, err
	}
	return &resp.StatusSTS, nil
}

// GetStatusFirmware returns firmware info (Status 2).
func (c *Client) GetStatusFirmware(ctx context.Context) (*StatusFWR, error) {
	var resp struct {
		StatusFWR StatusFWR `json:"StatusFWR"`
	}
	if err := c.ExecuteCommandJSON(ctx, "Status 2", &resp); err != nil {
		return nil, err
	}
	return &resp.StatusFWR, nil
}

// GetStatusNetwork returns network info (Status 5).
func (c *Client) GetStatusNetwork(ctx context.Context) (*StatusNET, error) {
	var resp struct {
		StatusNET StatusNET `json:"StatusNET"`
	}
	if err := c.ExecuteCommandJSON(ctx, "Status 5", &resp); err != nil {
		return nil, err
	}
	return &resp.StatusNET, nil
}

// GetStatusSensor returns sensor readings (Status 8).
func (c *Client) GetStatusSensor(ctx context.Context) (*StatusSNS, error) {
	var resp struct {
		StatusSNS StatusSNS `json:"StatusSNS"`
	}
	if err := c.ExecuteCommandJSON(ctx, "Status 8", &resp); err != nil {
		return nil, err
	}
	return &resp.StatusSNS, nil
}

// GetStatusState returns power state (Status 11).
func (c *Client) GetStatusState(ctx context.Context) (*StatusSTS, error) {
	var resp struct {
		StatusSTS StatusSTS `json:"StatusSTS"`
	}
	if err := c.ExecuteCommandJSON(ctx, "Status 11", &resp); err != nil {
		return nil, err
	}
	return &resp.StatusSTS, nil
}

// Power controls a relay. ID is 0-based (0 = Power1, 1 = Power2, etc.).
// state should be "ON", "OFF", or "TOGGLE".
func (c *Client) Power(ctx context.Context, id int, state string) (*PowerResponse, error) {
	// Tasmota uses 1-based indexing: Power1, Power2, etc.
	// Power (without number) = Power1
	cmd := "Power"
	if id > 0 {
		cmd = fmt.Sprintf("Power%d", id+1)
	}
	cmd = cmd + " " + strings.ToUpper(state)

	var resp PowerResponse
	if err := c.ExecuteCommandJSON(ctx, cmd, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetOtaURL sets the OTA update URL.
func (c *Client) SetOtaURL(ctx context.Context, otaURL string) error {
	cmd := fmt.Sprintf("OtaUrl %s", otaURL)
	_, err := c.ExecuteCommand(ctx, cmd)
	return err
}

// Upgrade triggers an OTA update.
func (c *Client) Upgrade(ctx context.Context) (*UpgradeResponse, error) {
	var resp UpgradeResponse
	if err := c.ExecuteCommandJSON(ctx, "Upgrade 1", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// IsTasmota attempts to detect if the device at the given address is running Tasmota.
// Returns true if Tasmota is detected.
func (c *Client) IsTasmota(ctx context.Context) bool {
	// Try to get status - if we can parse it as Tasmota response, it's Tasmota
	_, err := c.GetStatus(ctx)
	return err == nil
}

// CountRelays returns the number of relays based on power state responses.
func (c *Client) CountRelays(ctx context.Context) int {
	state, err := c.GetStatusState(ctx)
	if err != nil {
		return 0
	}

	count := 0
	if state.Power != "" || state.Power1 != "" {
		count = 1
	}
	if state.Power2 != "" {
		count = 2
	}
	if state.Power3 != "" {
		count = 3
	}
	if state.Power4 != "" {
		count = 4
	}

	return count
}

// GetPowerStates returns a map of relay ID to state (ON/OFF).
func (c *Client) GetPowerStates(ctx context.Context) (map[int]string, error) {
	state, err := c.GetStatusState(ctx)
	if err != nil {
		return nil, err
	}

	states := make(map[int]string)

	// Handle both single relay (Power) and multi-relay (Power1, Power2, etc.)
	if state.Power != "" {
		states[0] = state.Power
	} else if state.Power1 != "" {
		states[0] = state.Power1
	}
	if state.Power2 != "" {
		states[1] = state.Power2
	}
	if state.Power3 != "" {
		states[2] = state.Power3
	}
	if state.Power4 != "" {
		states[3] = state.Power4
	}

	return states, nil
}
