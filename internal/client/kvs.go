// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"
)

// KVSComponent provides access to a device's Key-Value Storage.
type KVSComponent struct {
	kvs *components.KVS
	rpc *rpc.Client
}

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

// KVSSetResult represents the result of setting a KVS value.
type KVSSetResult struct {
	Etag string `json:"etag"`
	Rev  int    `json:"rev"`
}

// KVSDeleteResult represents the result of deleting a KVS key.
type KVSDeleteResult struct {
	Rev int `json:"rev"`
}

// KVS returns a KVS component accessor.
func (c *Client) KVS() *KVSComponent {
	return &KVSComponent{
		kvs: components.NewKVS(c.rpcClient),
		rpc: c.rpcClient,
	}
}

// List returns all KVS key names.
func (k *KVSComponent) List(ctx context.Context) (*KVSListResult, error) {
	result, err := k.kvs.List(ctx)
	if err != nil {
		return nil, err
	}
	// Extract key names from the map
	keys := make([]string, 0, len(result.Keys))
	for key := range result.Keys {
		keys = append(keys, key)
	}
	return &KVSListResult{
		Keys: keys,
		Rev:  result.Rev,
	}, nil
}

// Get retrieves a value by key.
func (k *KVSComponent) Get(ctx context.Context, key string) (*KVSGetResult, error) {
	result, err := k.kvs.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return &KVSGetResult{
		Value: result.Value,
		Etag:  result.Etag,
	}, nil
}

// GetMany retrieves multiple values matching a pattern.
func (k *KVSComponent) GetMany(ctx context.Context, match string) ([]KVSItem, error) {
	result, err := k.kvs.GetMany(ctx, match)
	if err != nil {
		return nil, err
	}

	items := make([]KVSItem, len(result.Items))
	for i, item := range result.Items {
		etag := ""
		if item.Etag != nil {
			etag = *item.Etag
		}
		items[i] = KVSItem{
			Key:   item.Key,
			Value: item.Value,
			Etag:  etag,
		}
	}
	return items, nil
}

// Set stores a value for a key.
func (k *KVSComponent) Set(ctx context.Context, key string, value any) (*KVSSetResult, error) {
	result, err := k.kvs.Set(ctx, key, value)
	if err != nil {
		return nil, err
	}
	return &KVSSetResult{
		Etag: result.Etag,
		Rev:  result.Rev,
	}, nil
}

// Delete removes a key-value pair.
func (k *KVSComponent) Delete(ctx context.Context, key string) (*KVSDeleteResult, error) {
	result, err := k.kvs.Delete(ctx, key)
	if err != nil {
		return nil, err
	}
	return &KVSDeleteResult{
		Rev: result.Rev,
	}, nil
}

// GetAll retrieves all key-value pairs.
func (k *KVSComponent) GetAll(ctx context.Context) ([]KVSItem, error) {
	return k.GetMany(ctx, "*")
}
