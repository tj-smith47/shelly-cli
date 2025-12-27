// Package kvs provides Key-Value Store operations for Shelly devices.
package kvs

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// ConnectionFunc is a function that executes operations with a device connection.
type ConnectionFunc func(ctx context.Context, identifier string, fn func(*client.Client) error) error

// Service provides KVS operations for Shelly devices.
type Service struct {
	withConnection ConnectionFunc
}

// NewService creates a new KVS service.
func NewService(withConnection ConnectionFunc) *Service {
	return &Service{
		withConnection: withConnection,
	}
}

// List lists all KVS keys on a device.
func (s *Service) List(ctx context.Context, identifier string) (*ListResult, error) {
	var result *ListResult
	err := s.withConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = List(ctx, conn)
		return err
	})
	return result, err
}

// Get retrieves a value from device KVS.
func (s *Service) Get(ctx context.Context, identifier, key string) (*GetResult, error) {
	var result *GetResult
	err := s.withConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = Get(ctx, conn, key)
		return err
	})
	return result, err
}

// GetMany retrieves multiple values matching a pattern.
func (s *Service) GetMany(ctx context.Context, identifier, match string) ([]Item, error) {
	var result []Item
	err := s.withConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = GetMany(ctx, conn, match)
		return err
	})
	return result, err
}

// GetAll retrieves all key-value pairs from device KVS.
func (s *Service) GetAll(ctx context.Context, identifier string) ([]Item, error) {
	return s.GetMany(ctx, identifier, "*")
}

// Set stores a value in device KVS.
func (s *Service) Set(ctx context.Context, identifier, key string, value any) error {
	return s.withConnection(ctx, identifier, func(conn *client.Client) error {
		return Set(ctx, conn, key, value)
	})
}

// Delete removes a key from device KVS.
func (s *Service) Delete(ctx context.Context, identifier, key string) error {
	return s.withConnection(ctx, identifier, func(conn *client.Client) error {
		return Delete(ctx, conn, key)
	})
}

// Export exports all KVS data from a device.
func (s *Service) Export(ctx context.Context, identifier string) (*Export, error) {
	var result *Export
	err := s.withConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = ExportAll(ctx, conn)
		return err
	})
	return result, err
}

// Import imports KVS data to a device.
// If overwrite is true, existing keys will be overwritten.
// If overwrite is false, existing keys will be skipped.
func (s *Service) Import(ctx context.Context, identifier string, data *Export, overwrite bool) (imported, skipped int, err error) {
	err = s.withConnection(ctx, identifier, func(conn *client.Client) error {
		imported, skipped, err = Import(ctx, conn, data, overwrite)
		return err
	})
	return imported, skipped, err
}
