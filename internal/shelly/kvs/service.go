// Package kvs provides Key-Value Store operations for Shelly devices.
package kvs

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// ConnectionFunc is a function that executes operations with a device connection.
type ConnectionFunc func(ctx context.Context, identifier string, fn func(*client.Client) error) error

// Service provides KVS operations for Shelly devices.
type Service struct {
	withConnection ConnectionFunc
	cache          *cache.FileCache
	ios            *iostreams.IOStreams
}

// ServiceOption configures a KVS Service.
type ServiceOption func(*Service)

// WithFileCache configures the service with a file cache for cache invalidation.
func WithFileCache(fc *cache.FileCache) ServiceOption {
	return func(s *Service) {
		s.cache = fc
	}
}

// WithIOStreams configures the service with IOStreams for debug logging.
func WithIOStreams(ios *iostreams.IOStreams) ServiceOption {
	return func(s *Service) {
		s.ios = ios
	}
}

// NewService creates a new KVS service.
func NewService(withConnection ConnectionFunc, opts ...ServiceOption) *Service {
	s := &Service{
		withConnection: withConnection,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// invalidateCache invalidates cached KVS data for a device after mutations.
func (s *Service) invalidateCache(device string) {
	if s.cache == nil {
		return
	}
	if err := s.cache.Invalidate(device, cache.TypeKVS); err != nil && s.ios != nil {
		s.ios.DebugErr("cache invalidate "+device+"/"+cache.TypeKVS, err)
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
	err := s.withConnection(ctx, identifier, func(conn *client.Client) error {
		return Set(ctx, conn, key, value)
	})
	if err == nil {
		s.invalidateCache(identifier)
	}
	return err
}

// Delete removes a key from device KVS.
func (s *Service) Delete(ctx context.Context, identifier, key string) error {
	err := s.withConnection(ctx, identifier, func(conn *client.Client) error {
		return Delete(ctx, conn, key)
	})
	if err == nil {
		s.invalidateCache(identifier)
	}
	return err
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
	if err == nil && imported > 0 {
		s.invalidateCache(identifier)
	}
	return imported, skipped, err
}
