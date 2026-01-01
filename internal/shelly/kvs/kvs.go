// Package kvs provides Key-Value Store operations for Shelly devices.
package kvs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Item represents a key-value pair with optional etag.
type Item struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
	Etag  string `json:"etag,omitempty"`
}

// ListResult represents the result of listing KVS keys.
type ListResult struct {
	Keys []string `json:"keys"`
	Rev  int      `json:"rev"`
}

// GetResult represents the result of getting a KVS value.
type GetResult struct {
	Value any    `json:"value"`
	Etag  string `json:"etag"`
}

// Export holds exported KVS data.
type Export struct {
	Items   []Item `json:"items"`
	Version int    `json:"version"`
	Rev     int    `json:"rev"`
}

// List lists all KVS keys on the device.
func List(ctx context.Context, conn *client.Client) (*ListResult, error) {
	kvs := conn.KVS()
	res, err := kvs.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list KVS keys: %w", err)
	}
	return &ListResult{
		Keys: res.Keys,
		Rev:  res.Rev,
	}, nil
}

// Get retrieves a value from device KVS.
func Get(ctx context.Context, conn *client.Client, key string) (*GetResult, error) {
	kvs := conn.KVS()
	res, err := kvs.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get KVS key %q: %w", key, err)
	}
	return &GetResult{
		Value: res.Value,
		Etag:  res.Etag,
	}, nil
}

// GetMany retrieves multiple values matching a pattern.
func GetMany(ctx context.Context, conn *client.Client, match string) ([]Item, error) {
	kvs := conn.KVS()
	items, err := kvs.GetMany(ctx, match)
	if err != nil {
		return nil, fmt.Errorf("failed to get KVS keys matching %q: %w", match, err)
	}
	result := make([]Item, len(items))
	for i, item := range items {
		result[i] = Item{
			Key:   item.Key,
			Value: item.Value,
			Etag:  item.Etag,
		}
	}
	return result, nil
}

// GetAll retrieves all key-value pairs from device KVS.
func GetAll(ctx context.Context, conn *client.Client) ([]Item, error) {
	return GetMany(ctx, conn, "*")
}

// Set stores a value in device KVS.
func Set(ctx context.Context, conn *client.Client, key string, value any) error {
	kvs := conn.KVS()
	_, err := kvs.Set(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to set KVS key %q: %w", key, err)
	}
	return nil
}

// Delete removes a key from device KVS.
func Delete(ctx context.Context, conn *client.Client, key string) error {
	kvs := conn.KVS()
	_, err := kvs.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete KVS key %q: %w", key, err)
	}
	return nil
}

// ExportAll exports all KVS data from a device.
func ExportAll(ctx context.Context, conn *client.Client) (*Export, error) {
	kvs := conn.KVS()

	// Get list for revision info
	listRes, err := kvs.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list KVS keys: %w", err)
	}

	// Get all items with values
	items, err := kvs.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get KVS data: %w", err)
	}

	exported := make([]Item, len(items))
	for i, item := range items {
		exported[i] = Item{
			Key:   item.Key,
			Value: item.Value,
			Etag:  item.Etag,
		}
	}

	return &Export{
		Items:   exported,
		Version: 1,
		Rev:     listRes.Rev,
	}, nil
}

// Import imports KVS data to a device.
// If overwrite is true, existing keys will be overwritten.
// If overwrite is false, existing keys will be skipped.
// Returns the count of imported and skipped items.
func Import(ctx context.Context, conn *client.Client, data *Export, overwrite bool) (imported, skipped int, err error) {
	kvs := conn.KVS()

	// Get existing keys if not overwriting
	var existingKeys map[string]bool
	if !overwrite {
		listRes, listErr := kvs.List(ctx)
		if listErr != nil {
			return 0, 0, fmt.Errorf("failed to list existing keys: %w", listErr)
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
			return imported, skipped, fmt.Errorf("failed to set key %q: %w", item.Key, setErr)
		}
		imported++
	}
	return imported, skipped, nil
}

// ParseValue parses a string value into the appropriate type.
// Tries to parse as JSON first, then falls back to string.
func ParseValue(valueStr string) any {
	// Try to parse as JSON
	var jsonValue any
	if err := json.Unmarshal([]byte(valueStr), &jsonValue); err == nil {
		return jsonValue
	}
	// Fall back to string
	return valueStr
}

// ParseImportFile reads and parses a KVS import file (JSON or YAML).
func ParseImportFile(file string) (*Export, error) {
	content, err := afero.ReadFile(config.Fs(), file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data Export
	if err := json.Unmarshal(content, &data); err != nil {
		// Try YAML
		if yamlErr := yaml.Unmarshal(content, &data); yamlErr != nil {
			return nil, fmt.Errorf("failed to parse file (tried JSON and YAML): %w", err)
		}
	}

	return &data, nil
}
