// Package auth provides authentication configuration for Shelly devices.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

const (
	// DefaultPasswordLength is the default length for generated passwords.
	DefaultPasswordLength = 16
	// PasswordCharset contains characters used for password generation.
	PasswordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	// DefaultUser is the Shelly device's built-in authentication user.
	DefaultUser = "admin"
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
	WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error
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

// Set turns on authentication on a device with the given user and password,
// or turns it off when password is empty. gen1 selects the protocol.
//
// A Gen1 device takes any user name (empty means "admin") through
// /settings/login. A Gen2+ device has a single user, admin, and a single
// realm, its device ID, which is read from the device; any other user is an
// error (errors.Is ErrInvalidAuthParams), since the device would store a hash
// no request can match.
func (s *Service) Set(ctx context.Context, identifier string, gen1 bool, user, password string) error {
	if user == "" {
		user = DefaultUser
	}
	if gen1 {
		return s.provider.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
			if password == "" {
				return conn.Device().SetAuth(ctx, false, "", "")
			}
			return conn.Device().SetAuth(ctx, true, user, password)
		})
	}
	if user != DefaultUser {
		return fmt.Errorf("%w: user %q", ErrInvalidAuthParams, user)
	}
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		id := conn.Info().ID
		params := map[string]any{
			"user":  DefaultUser,
			"realm": id,
			"ha1":   nil,
		}
		if password != "" {
			params["ha1"] = CalculateHA1(DefaultUser, id, password)
		}
		_, err := conn.Call(ctx, "Shelly.SetAuth", params)
		return err
	})
}

// Disable turns authentication off on a device; gen1 selects the protocol.
func (s *Service) Disable(ctx context.Context, identifier string, gen1 bool) error {
	return s.Set(ctx, identifier, gen1, "", "")
}

// ErrInvalidAuthParams reports a user a Gen2+ device cannot have.
var ErrInvalidAuthParams = errors.New("invalid user: Gen2+ devices have a single user, admin")

// CalculateHA1 returns the HA1 a Gen2+ device stores through Shelly.SetAuth:
// the hex SHA-256 of "user:realm:password".
func CalculateHA1(user, realm, password string) string {
	hash := sha256.Sum256([]byte(user + ":" + realm + ":" + password))
	return hex.EncodeToString(hash[:])
}

// GeneratePassword creates a cryptographically secure random password of the specified length.
// If length is less than 8, it defaults to 8 for security.
func GeneratePassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(PasswordCharset)))

	for i := range length {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = PasswordCharset[n.Int64()]
	}

	return string(result), nil
}
