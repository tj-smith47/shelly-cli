// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-go/gen2"
	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Client represents a connection to a Shelly device.
type Client struct {
	device    *gen2.Device
	rpcClient *rpc.Client
	transport transport.Transport
	info      *DeviceInfo
}

// DeviceInfo holds information about the connected device.
type DeviceInfo struct {
	ID         string
	MAC        string
	Model      string
	Generation int
	Firmware   string
	App        string
	AuthEn     bool
}

// Connect establishes a connection to a Shelly device.
func Connect(ctx context.Context, device model.Device) (*Client, error) {
	url := device.Address
	if url != "" && url[0] != 'h' {
		url = "http://" + url
	}

	var opts []transport.Option
	if device.HasAuth() {
		opts = append(opts, transport.WithAuth(device.Auth.Username, device.Auth.Password))
	}

	httpTransport := transport.NewHTTP(url, opts...)
	rpcClient := rpc.NewClient(httpTransport)
	gen2Device := gen2.NewDevice(rpcClient)

	info, err := gen2Device.GetDeviceInfo(ctx)
	if err != nil {
		iostreams.CloseWithDebug("closing device after connection failure", gen2Device)
		return nil, fmt.Errorf("%w: %w", model.ErrConnectionFailed, err)
	}

	return &Client{
		device:    gen2Device,
		rpcClient: rpcClient,
		transport: httpTransport,
		info: &DeviceInfo{
			ID:         info.ID,
			MAC:        info.MAC,
			Model:      info.Model,
			Generation: info.Gen,
			Firmware:   info.FirmwareVersion,
			App:        info.App,
			AuthEn:     info.AuthEnabled,
		},
	}, nil
}

// Close closes the device connection.
func (c *Client) Close() error {
	if c.device != nil {
		return c.device.Close()
	}
	if c.rpcClient != nil {
		return c.rpcClient.Close()
	}
	return nil
}

// Info returns the device information.
func (c *Client) Info() *DeviceInfo {
	return c.info
}

// Call makes a raw RPC call to the device.
func (c *Client) Call(ctx context.Context, method string, params map[string]any) (any, error) {
	return c.rpcClient.Call(ctx, method, params)
}

// Switch returns a switch component accessor.
func (c *Client) Switch(id int) *SwitchComponent {
	return &SwitchComponent{
		sw:  components.NewSwitch(c.rpcClient, id),
		rpc: c.rpcClient,
		id:  id,
	}
}

// Cover returns a cover component accessor.
func (c *Client) Cover(id int) *CoverComponent {
	return &CoverComponent{
		cv:  components.NewCover(c.rpcClient, id),
		rpc: c.rpcClient,
		id:  id,
	}
}

// Light returns a light component accessor.
func (c *Client) Light(id int) *LightComponent {
	return &LightComponent{
		lt:  components.NewLight(c.rpcClient, id),
		rpc: c.rpcClient,
		id:  id,
	}
}

// RGB returns an RGB component accessor.
func (c *Client) RGB(id int) *RGBComponent {
	return &RGBComponent{
		rgb: components.NewRGB(c.rpcClient, id),
		rpc: c.rpcClient,
		id:  id,
	}
}

// Input returns an input component accessor.
func (c *Client) Input(id int) *InputComponent {
	return &InputComponent{
		input: components.NewInput(c.rpcClient, id),
		rpc:   c.rpcClient,
		id:    id,
	}
}

// Thermostat returns a thermostat component accessor.
func (c *Client) Thermostat(id int) *ThermostatComponent {
	return &ThermostatComponent{
		th:  components.NewThermostat(c.rpcClient, id),
		rpc: c.rpcClient,
		id:  id,
	}
}

// ListComponents returns all components on the device.
func (c *Client) ListComponents(ctx context.Context) ([]model.Component, error) {
	rpcResult, err := c.rpcClient.Call(ctx, "Shelly.GetComponents", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to get components: %w", err)
	}

	var resp struct {
		Components []struct {
			Key string `json:"key"`
		} `json:"components"`
	}

	if err := unmarshalResponse(rpcResult, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse components: %w", err)
	}

	comps := make([]model.Component, 0, len(resp.Components))
	for _, rc := range resp.Components {
		parsed, ok := parseComponentKey(rc.Key)
		if ok {
			comps = append(comps, parsed)
		}
	}

	return comps, nil
}

// FilterComponents returns components matching the given type.
func (c *Client) FilterComponents(ctx context.Context, compType model.ComponentType) ([]model.Component, error) {
	all, err := c.ListComponents(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []model.Component
	for _, comp := range all {
		if comp.Type == compType {
			filtered = append(filtered, comp)
		}
	}

	return filtered, nil
}

// componentPrefixes maps component key prefixes to their types.
var componentPrefixes = map[string]model.ComponentType{
	"switch:": model.ComponentSwitch,
	"cover:":  model.ComponentCover,
	"light:":  model.ComponentLight,
	"rgb:":    model.ComponentRGB,
	"input:":  model.ComponentInput,
}

// parseComponentKey parses a component key like "switch:0" into a Component.
func parseComponentKey(key string) (model.Component, bool) {
	for prefix, compType := range componentPrefixes {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			var id int
			if _, err := fmt.Sscanf(key[len(prefix):], "%d", &id); err != nil {
				return model.Component{}, false
			}
			return model.Component{
				Type: compType,
				ID:   id,
				Key:  key,
			}, true
		}
	}
	return model.Component{}, false
}

// Reboot reboots the device.
// delayMS specifies the delay in milliseconds before reboot (0 for immediate).
func (c *Client) Reboot(ctx context.Context, delayMS int) error {
	params := map[string]any{}
	if delayMS > 0 {
		params["delay_ms"] = delayMS
	}
	_, err := c.rpcClient.Call(ctx, "Shelly.Reboot", params)
	return err
}

// FactoryReset performs a factory reset on the device.
// WARNING: This will erase all settings and return the device to factory defaults.
func (c *Client) FactoryReset(ctx context.Context) error {
	_, err := c.rpcClient.Call(ctx, "Shelly.FactoryReset", map[string]any{})
	return err
}

// GetStatus returns the full device status as a map.
func (c *Client) GetStatus(ctx context.Context) (map[string]any, error) {
	result, err := c.rpcClient.Call(ctx, "Shelly.GetStatus", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to get device status: %w", err)
	}

	// Convert the response to map[string]any
	var status map[string]any
	if err := unmarshalResponse(result, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return status, nil
}

// GetConfig returns the full device config as a map.
func (c *Client) GetConfig(ctx context.Context) (map[string]any, error) {
	result, err := c.rpcClient.Call(ctx, "Shelly.GetConfig", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to get device config: %w", err)
	}

	// Convert the response to map[string]any
	var config map[string]any
	if err := unmarshalResponse(result, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// SetConfig updates device configuration.
func (c *Client) SetConfig(ctx context.Context, config map[string]any) error {
	_, err := c.rpcClient.Call(ctx, "Shelly.SetConfig", map[string]any{
		"config": config,
	})
	if err != nil {
		return fmt.Errorf("failed to set device config: %w", err)
	}
	return nil
}

// RPCClient returns the underlying RPC client.
func (c *Client) RPCClient() *rpc.Client {
	return c.rpcClient
}

// unmarshalResponse converts an RPC response to a typed struct.
func unmarshalResponse(data, v any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, v); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}
