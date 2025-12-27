// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// KVSItem represents a key-value pair with optional etag.
type KVSItem struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
	Etag  string `json:"etag,omitempty"`
}

// KVSListResult represents the result of listing KVS keys.
type KVSListResult struct {
	Keys []string `json:"keys"`
	Rev  int      `json:"rev"`
}

// KVSGetResult represents the result of getting a KVS value.
type KVSGetResult struct {
	Value any    `json:"value"`
	Etag  string `json:"etag"`
}

// ListKVS lists all KVS keys on a device.
func (s *Service) ListKVS(ctx context.Context, identifier string) (*KVSListResult, error) {
	var result *KVSListResult
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		kvs := conn.KVS()
		res, err := kvs.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list KVS keys: %w", err)
		}
		result = &KVSListResult{
			Keys: res.Keys,
			Rev:  res.Rev,
		}
		return nil
	})
	return result, err
}

// GetKVS retrieves a value from device KVS.
func (s *Service) GetKVS(ctx context.Context, identifier, key string) (*KVSGetResult, error) {
	var result *KVSGetResult
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		kvs := conn.KVS()
		res, err := kvs.Get(ctx, key)
		if err != nil {
			return fmt.Errorf("failed to get KVS key %q: %w", key, err)
		}
		result = &KVSGetResult{
			Value: res.Value,
			Etag:  res.Etag,
		}
		return nil
	})
	return result, err
}

// GetManyKVS retrieves multiple values matching a pattern.
func (s *Service) GetManyKVS(ctx context.Context, identifier, match string) ([]KVSItem, error) {
	var result []KVSItem
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		kvs := conn.KVS()
		items, err := kvs.GetMany(ctx, match)
		if err != nil {
			return fmt.Errorf("failed to get KVS keys matching %q: %w", match, err)
		}
		result = make([]KVSItem, len(items))
		for i, item := range items {
			result[i] = KVSItem{
				Key:   item.Key,
				Value: item.Value,
				Etag:  item.Etag,
			}
		}
		return nil
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
		kvs := conn.KVS()
		_, err := kvs.Set(ctx, key, value)
		if err != nil {
			return fmt.Errorf("failed to set KVS key %q: %w", key, err)
		}
		return nil
	})
}

// DeleteKVS removes a key from device KVS.
func (s *Service) DeleteKVS(ctx context.Context, identifier, key string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		kvs := conn.KVS()
		_, err := kvs.Delete(ctx, key)
		if err != nil {
			return fmt.Errorf("failed to delete KVS key %q: %w", key, err)
		}
		return nil
	})
}

// KVSExport holds exported KVS data.
type KVSExport struct {
	Items   []KVSItem `json:"items"`
	Version int       `json:"version"`
	Rev     int       `json:"rev"`
}

// ExportKVS exports all KVS data from a device.
func (s *Service) ExportKVS(ctx context.Context, identifier string) (*KVSExport, error) {
	var result *KVSExport
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		kvs := conn.KVS()

		// Get list for revision info
		listRes, err := kvs.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list KVS keys: %w", err)
		}

		// Get all items with values
		items, err := kvs.GetAll(ctx)
		if err != nil {
			return fmt.Errorf("failed to get KVS data: %w", err)
		}

		exported := make([]KVSItem, len(items))
		for i, item := range items {
			exported[i] = KVSItem{
				Key:   item.Key,
				Value: item.Value,
				Etag:  item.Etag,
			}
		}

		result = &KVSExport{
			Items:   exported,
			Version: 1,
			Rev:     listRes.Rev,
		}
		return nil
	})
	return result, err
}

// ImportKVS imports KVS data to a device.
// If overwrite is true, existing keys will be overwritten.
// If overwrite is false, existing keys will be skipped.
func (s *Service) ImportKVS(ctx context.Context, identifier string, data *KVSExport, overwrite bool) (imported, skipped int, err error) {
	err = s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		kvs := conn.KVS()

		// Get existing keys if not overwriting
		var existingKeys map[string]bool
		if !overwrite {
			listRes, listErr := kvs.List(ctx)
			if listErr != nil {
				return fmt.Errorf("failed to list existing keys: %w", listErr)
			}
			existingKeys = make(map[string]bool, len(listRes.Keys))
			for _, k := range listRes.Keys {
				existingKeys[k] = true
			}
		}

		for _, item := range data.Items {
			// Skip existing keys if not overwriting
			if !overwrite && existingKeys[item.Key] {
				skipped++
				continue
			}

			_, setErr := kvs.Set(ctx, item.Key, item.Value)
			if setErr != nil {
				return fmt.Errorf("failed to set key %q: %w", item.Key, setErr)
			}
			imported++
		}
		return nil
	})
	return imported, skipped, err
}

// ParseKVSValue parses a string value into the appropriate type.
// Tries to parse as JSON first, then falls back to string.
func ParseKVSValue(valueStr string) any {
	// Try to parse as JSON
	var jsonValue any
	if err := json.Unmarshal([]byte(valueStr), &jsonValue); err == nil {
		return jsonValue
	}
	// Fall back to string
	return valueStr
}

// ParseKVSImportFile reads and parses a KVS import file (JSON or YAML).
func ParseKVSImportFile(file string) (*KVSExport, error) {
	//nolint:gosec // G304: file path is from user command line argument
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data KVSExport
	if err := json.Unmarshal(content, &data); err != nil {
		// Try YAML
		if yamlErr := yaml.Unmarshal(content, &data); yamlErr != nil {
			return nil, fmt.Errorf("failed to parse file (tried JSON and YAML): %w", err)
		}
	}

	return &data, nil
}
