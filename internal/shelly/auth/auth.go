// Package auth provides authentication configuration for Shelly devices.
package auth

import (
	"context"
	"crypto/md5" //nolint:gosec // Required for Shelly digest auth (HA1 hash)
	"encoding/hex"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// Status holds authentication status information.
type Status struct {
	Enabled bool   `json:"enabled"`
	User    string `json:"user,omitempty"`
	Realm   string `json:"realm,omitempty"`
}

// ConnectionProvider allows executing operations with a device connection.
type ConnectionProvider interface {
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

// DeviceInfoProvider provides device info access.
type DeviceInfoProvider interface {
	GetAuthEnabled(ctx context.Context, identifier string) (bool, error)
}

// Service provides authentication-related operations for Shelly devices.
type Service struct {
	provider     ConnectionProvider
	infoProvider DeviceInfoProvider
}

// New creates a new auth service.
func New(provider ConnectionProvider, infoProvider DeviceInfoProvider) *Service {
	return &Service{
		provider:     provider,
		infoProvider: infoProvider,
	}
}

// GetStatus returns the authentication status for a device.
func (s *Service) GetStatus(ctx context.Context, identifier string) (*Status, error) {
	enabled, err := s.infoProvider.GetAuthEnabled(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return &Status{
		Enabled: enabled,
	}, nil
}

// Set configures device authentication.
// If password is empty, authentication is disabled.
func (s *Service) Set(ctx context.Context, identifier, user, realm, password string) error {
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"user":  user,
			"realm": realm,
		}
		if password != "" {
			// Calculate HA1 = MD5(user:realm:password)
			ha1 := CalculateHA1(user, realm, password)
			params["ha1"] = ha1
		}
		_, err := conn.Call(ctx, "Shelly.SetAuth", params)
		return err
	})
}

// Disable disables device authentication.
func (s *Service) Disable(ctx context.Context, identifier string) error {
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		// Setting ha1 to null disables authentication
		params := map[string]any{
			"user":  "admin",
			"realm": "",
			"ha1":   nil,
		}
		_, err := conn.Call(ctx, "Shelly.SetAuth", params)
		return err
	})
}

// CalculateHA1 calculates the HA1 hash for digest authentication.
// MD5 is required by the Shelly device protocol - not a security concern since
// this is a password hash transmitted over a local network to the device.
func CalculateHA1(user, realm, password string) string {
	data := user + ":" + realm + ":" + password
	hash := md5.Sum([]byte(data)) //nolint:gosec // Required by Shelly digest auth protocol
	return hex.EncodeToString(hash[:])
}
