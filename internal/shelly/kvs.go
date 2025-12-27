// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
)

// KVSItem is an alias for kvs.Item for backward compatibility.
type KVSItem = kvs.Item

// KVSListResult is an alias for kvs.ListResult for backward compatibility.
type KVSListResult = kvs.ListResult

// KVSGetResult is an alias for kvs.GetResult for backward compatibility.
type KVSGetResult = kvs.GetResult

// KVSExport is an alias for kvs.Export for backward compatibility.
type KVSExport = kvs.Export

// ListKVS lists all KVS keys on a device.
func (s *Service) ListKVS(ctx context.Context, identifier string) (*KVSListResult, error) {
	var result *KVSListResult
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = kvs.List(ctx, conn)
		return err
	})
	return result, err
}

// GetKVS retrieves a value from device KVS.
func (s *Service) GetKVS(ctx context.Context, identifier, key string) (*KVSGetResult, error) {
	var result *KVSGetResult
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = kvs.Get(ctx, conn, key)
		return err
	})
	return result, err
}

// GetManyKVS retrieves multiple values matching a pattern.
func (s *Service) GetManyKVS(ctx context.Context, identifier, match string) ([]KVSItem, error) {
	var result []KVSItem
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = kvs.GetMany(ctx, conn, match)
		return err
	})
	return result, err
}

// GetAllKVS retrieves all key-value pairs from device KVS.
func (s *Service) GetAllKVS(ctx context.Context, identifier string) ([]KVSItem, error) {
	return s.GetManyKVS(ctx, identifier, "*")
}

// SetKVS stores a value in device KVS.
func (s *Service) SetKVS(ctx context.Context, identifier, key string, value any) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return kvs.Set(ctx, conn, key, value)
	})
}

// DeleteKVS removes a key from device KVS.
func (s *Service) DeleteKVS(ctx context.Context, identifier, key string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return kvs.Delete(ctx, conn, key)
	})
}

// ExportKVS exports all KVS data from a device.
func (s *Service) ExportKVS(ctx context.Context, identifier string) (*KVSExport, error) {
	var result *KVSExport
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = kvs.ExportAll(ctx, conn)
		return err
	})
	return result, err
}

// ImportKVS imports KVS data to a device.
// If overwrite is true, existing keys will be overwritten.
// If overwrite is false, existing keys will be skipped.
func (s *Service) ImportKVS(ctx context.Context, identifier string, data *KVSExport, overwrite bool) (imported, skipped int, err error) {
	err = s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		imported, skipped, err = kvs.Import(ctx, conn, data, overwrite)
		return err
	})
	return imported, skipped, err
}

// ParseKVSValue parses a string value into the appropriate type.
//
// Deprecated: Use kvs.ParseValue directly.
func ParseKVSValue(valueStr string) any {
	return kvs.ParseValue(valueStr)
}

// ParseKVSImportFile reads and parses a KVS import file (JSON or YAML).
//
// Deprecated: Use kvs.ParseImportFile directly.
func ParseKVSImportFile(file string) (*KVSExport, error) {
	return kvs.ParseImportFile(file)
}
