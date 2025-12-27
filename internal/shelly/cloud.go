// Package shelly provides the cloud service layer for Shelly Cloud API.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tj-smith47/shelly-go/cloud"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// CloudDevice represents a device from the Shelly Cloud API.
type CloudDevice struct {
	ID              string
	Name            string
	Model           string
	MAC             string
	FirmwareVersion string
	Generation      int
	Online          bool
	CloudEnabled    bool
	Status          json.RawMessage
	Settings        json.RawMessage
}

// CloudToken represents a parsed Shelly Cloud token.
type CloudToken struct {
	AccessToken string
	UserAPIURL  string
	Email       string
	UserID      int
	Expiry      time.Time
}

// CloudAuthStatus represents the cloud authentication status.
type CloudAuthStatus struct {
	Authenticated bool
	Email         string
	UserAPIURL    string
	TokenExpiry   time.Time
	TokenValid    bool
}

// CloudLoginResult represents the result of a cloud login.
type CloudLoginResult struct {
	Token      string
	UserAPIURL string
	Expiry     time.Time
}

// CloudClient wraps the shelly-go cloud client.
type CloudClient struct {
	client *cloud.Client
}

// NewCloudClient creates a new cloud client with the given access token.
func NewCloudClient(accessToken string) *CloudClient {
	return &CloudClient{
		client: cloud.NewClient(cloud.WithAccessToken(accessToken)),
	}
}

// NewCloudClientWithCredentials creates a new cloud client and authenticates.
func NewCloudClientWithCredentials(ctx context.Context, email, password string) (*CloudClient, *CloudLoginResult, error) {
	passwordHash := cloud.HashPassword(password)

	client := cloud.NewClient()
	token, err := client.Login(ctx, email, passwordHash)
	if err != nil {
		return nil, nil, err
	}

	client.SetToken(token)

	return &CloudClient{client: client}, &CloudLoginResult{
		Token:      token.AccessToken,
		UserAPIURL: token.UserAPIURL,
		Expiry:     token.Expiry,
	}, nil
}

// Login authenticates with the Shelly Cloud API.
func (c *CloudClient) Login(ctx context.Context, email, password string) (*CloudLoginResult, error) {
	passwordHash := cloud.HashPassword(password)

	token, err := c.client.Login(ctx, email, passwordHash)
	if err != nil {
		return nil, err
	}

	c.client.SetToken(token)

	return &CloudLoginResult{
		Token:      token.AccessToken,
		UserAPIURL: token.UserAPIURL,
		Expiry:     token.Expiry,
	}, nil
}

// GetAllDevices returns all devices from the Shelly Cloud.
func (c *CloudClient) GetAllDevices(ctx context.Context) ([]CloudDevice, error) {
	devicesMap, err := c.client.GetAllDevices(ctx)
	if err != nil {
		return nil, err
	}

	devices := make([]CloudDevice, 0, len(devicesMap))
	for id, status := range devicesMap {
		dev := CloudDevice{
			ID:       id,
			Online:   status.Online,
			Status:   status.Status,
			Settings: status.Settings,
		}

		// Extract device info if available
		if status.DevInfo != nil {
			dev.Model = status.DevInfo.Code
			dev.Generation = status.DevInfo.Generation
		}

		devices = append(devices, dev)
	}

	return devices, nil
}

// GetDevice returns a specific device from the Shelly Cloud.
func (c *CloudClient) GetDevice(ctx context.Context, deviceID string) (*CloudDevice, error) {
	status, err := c.client.GetDeviceStatus(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	dev := &CloudDevice{
		ID:       deviceID,
		Online:   status.Online,
		Status:   status.Status,
		Settings: status.Settings,
	}

	if status.DevInfo != nil {
		dev.Model = status.DevInfo.Code
		dev.Generation = status.DevInfo.Generation
	}

	return dev, nil
}

// SetSwitch controls a switch via the cloud.
func (c *CloudClient) SetSwitch(ctx context.Context, deviceID string, channel int, on bool) error {
	return c.client.SetSwitch(ctx, deviceID, channel, on)
}

// ToggleSwitch toggles a switch via the cloud.
func (c *CloudClient) ToggleSwitch(ctx context.Context, deviceID string, channel int) error {
	return c.client.ToggleSwitch(ctx, deviceID, channel)
}

// OpenCover opens a cover via the cloud.
func (c *CloudClient) OpenCover(ctx context.Context, deviceID string, channel int) error {
	return c.client.OpenCover(ctx, deviceID, channel)
}

// CloseCover closes a cover via the cloud.
func (c *CloudClient) CloseCover(ctx context.Context, deviceID string, channel int) error {
	return c.client.CloseCover(ctx, deviceID, channel)
}

// StopCover stops a cover via the cloud.
func (c *CloudClient) StopCover(ctx context.Context, deviceID string, channel int) error {
	return c.client.StopCover(ctx, deviceID, channel)
}

// SetCoverPosition sets the cover position via the cloud.
func (c *CloudClient) SetCoverPosition(ctx context.Context, deviceID string, channel, position int) error {
	return c.client.SetCoverPosition(ctx, deviceID, channel, position)
}

// SetLight controls a light via the cloud.
func (c *CloudClient) SetLight(ctx context.Context, deviceID string, channel int, on bool) error {
	return c.client.SetLight(ctx, deviceID, channel, on)
}

// ToggleLight toggles a light via the cloud.
func (c *CloudClient) ToggleLight(ctx context.Context, deviceID string, channel int) error {
	return c.client.ToggleLight(ctx, deviceID, channel)
}

// SetLightBrightness sets the light brightness via the cloud.
func (c *CloudClient) SetLightBrightness(ctx context.Context, deviceID string, channel, brightness int) error {
	return c.client.SetLightBrightness(ctx, deviceID, channel, brightness)
}

// SetLightRGB sets the light RGB color via the cloud.
func (c *CloudClient) SetLightRGB(ctx context.Context, deviceID string, channel, red, green, blue int) error {
	return c.client.SetLightRGB(ctx, deviceID, channel, red, green, blue)
}

// ParseToken parses a JWT token string and extracts the claims.
func ParseToken(tokenString string) (*CloudToken, error) {
	token, err := cloud.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &CloudToken{
		AccessToken: token.AccessToken,
		UserAPIURL:  token.UserAPIURL,
		Expiry:      token.Expiry,
	}, nil
}

// ValidateToken checks if a token is valid and not expired.
func ValidateToken(tokenString string) error {
	return cloud.ValidateToken(tokenString)
}

// IsTokenExpired checks if a token is expired.
func IsTokenExpired(tokenString string) bool {
	return cloud.IsTokenExpired(tokenString)
}

// TimeUntilExpiry returns the time until the token expires.
func TimeUntilExpiry(tokenString string) time.Duration {
	return cloud.TimeUntilExpiry(tokenString)
}

// HashPassword returns the SHA1 hash of a password for cloud auth.
func HashPassword(password string) string {
	return cloud.HashPassword(password)
}

// BuildCloudWebSocketURL builds the WebSocket URL for cloud events.
func BuildCloudWebSocketURL(serverURL, token string) (string, error) {
	if serverURL == "" {
		// Try to extract from token
		parsedToken, err := ParseToken(token)
		if err != nil {
			return "", fmt.Errorf("no server URL and failed to parse token: %w", err)
		}
		serverURL = parsedToken.UserAPIURL
	}

	if serverURL == "" {
		return "", fmt.Errorf("no server URL available")
	}

	// Parse to get hostname
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}

	hostname := u.Hostname()
	if hostname == "" {
		hostname = serverURL
	}

	// Build WebSocket URL: wss://{host}:6113/shelly/wss/hk_sock?t={token}
	return fmt.Sprintf("wss://%s:6113/shelly/wss/hk_sock?t=%s", hostname, url.QueryEscape(token)), nil
}

// CloudEventStreamOptions configures cloud event streaming.
type CloudEventStreamOptions struct {
	DeviceFilter string // Filter events by device ID
	EventFilter  string // Filter events by event type substring
	Raw          bool   // Output raw JSON without parsing
}

// CloudEventHandler is called for each cloud event received.
// Return an error to stop streaming.
type CloudEventHandler func(event *model.CloudEvent, raw []byte) error

// StreamCloudEvents reads cloud events from a websocket and calls the handler for each.
// Blocks until the context is cancelled or an error occurs.
func StreamCloudEvents(ctx context.Context, conn *websocket.Conn, opts CloudEventStreamOptions, handler CloudEventHandler) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Set read deadline
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return handleCloudEventReadError(ctx, err)
		}

		// Parse the event
		var event model.CloudEvent
		if err := json.Unmarshal(message, &event); err != nil {
			continue // Skip unparseable messages
		}

		// Apply filters
		if opts.DeviceFilter != "" && event.GetDeviceID() != opts.DeviceFilter {
			continue
		}
		if opts.EventFilter != "" && !strings.Contains(event.Event, opts.EventFilter) {
			continue
		}

		// Call handler
		if err := handler(&event, message); err != nil {
			return err
		}
	}
}

// handleCloudEventReadError processes WebSocket read errors.
// Returns nil for expected closures (normal close, context cancelled).
func handleCloudEventReadError(ctx context.Context, err error) error {
	if isExpectedCloudClosure(ctx, err) {
		return nil
	}
	return fmt.Errorf("read error: %w", err)
}

// isExpectedCloudClosure checks if the error represents a normal termination.
func isExpectedCloudClosure(ctx context.Context, err error) bool {
	// Normal WebSocket closure
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		return true
	}
	// Context was cancelled (user pressed Ctrl+C)
	return ctx.Err() != nil
}
