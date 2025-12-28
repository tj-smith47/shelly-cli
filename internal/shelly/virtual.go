package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// VirtualComponentType represents the type of a virtual component.
type VirtualComponentType string

// Virtual component types.
const (
	VirtualBoolean VirtualComponentType = "boolean"
	VirtualNumber  VirtualComponentType = "number"
	VirtualText    VirtualComponentType = "text"
	VirtualEnum    VirtualComponentType = "enum"
	VirtualButton  VirtualComponentType = "button"
	VirtualGroup   VirtualComponentType = "group"
)

// VirtualComponent represents a virtual component on a device.
type VirtualComponent struct {
	Key       string               `json:"key"`
	Type      VirtualComponentType `json:"type"`
	ID        int                  `json:"id"`
	Name      string               `json:"name,omitempty"`
	Value     any                  `json:"value,omitempty"`
	BoolValue *bool                `json:"bool_value,omitempty"`
	NumValue  *float64             `json:"num_value,omitempty"`
	StrValue  *string              `json:"str_value,omitempty"`
	Options   []string             `json:"options,omitempty"`
	Min       *float64             `json:"min,omitempty"`
	Max       *float64             `json:"max,omitempty"`
	Unit      *string              `json:"unit,omitempty"`
}

// componentInfo contains information about a component from GetComponents.
type componentInfo struct {
	Key    string          `json:"key"`
	Status json.RawMessage `json:"status,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
}

// componentListResponse is the response from Shelly.GetComponents.
type componentListResponse struct {
	Components []componentInfo `json:"components"`
}

// ListVirtualComponents returns all virtual components on a device.
func (s *Service) ListVirtualComponents(ctx context.Context, identifier string) ([]VirtualComponent, error) {
	var results []VirtualComponent

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"dynamic_only":   true,
			"include_status": true,
			"include_config": true,
		}
		rawResult, err := conn.Call(ctx, "Shelly.GetComponents", params)
		if err != nil {
			return err
		}

		jsonBytes, err := json.Marshal(rawResult)
		if err != nil {
			return fmt.Errorf("marshal components: %w", err)
		}

		var compList componentListResponse
		if err := json.Unmarshal(jsonBytes, &compList); err != nil {
			return fmt.Errorf("unmarshal components: %w", err)
		}

		for _, comp := range compList.Components {
			vc, ok := parseVirtualComponent(comp)
			if ok {
				results = append(results, vc)
			}
		}
		return nil
	})

	return results, err
}

// parseVirtualComponent attempts to parse a componentInfo as a virtual component.
func parseVirtualComponent(comp componentInfo) (VirtualComponent, bool) {
	parts := strings.Split(comp.Key, ":")
	if len(parts) != 2 {
		return VirtualComponent{}, false
	}

	compType := parts[0]
	var id int
	if _, err := fmt.Sscanf(parts[1], "%d", &id); err != nil {
		return VirtualComponent{}, false
	}

	// Virtual components have IDs in range 200-299.
	if id < 200 || id > 299 {
		return VirtualComponent{}, false
	}

	vt := VirtualComponentType(compType)
	if !isVirtualType(vt) {
		return VirtualComponent{}, false
	}

	vc := VirtualComponent{
		Key:  comp.Key,
		Type: vt,
		ID:   id,
	}

	// Parse config for name.
	if len(comp.Config) > 0 {
		var cfg struct {
			Name    *string  `json:"name,omitempty"`
			Options []string `json:"options,omitempty"`
			Min     *float64 `json:"min,omitempty"`
			Max     *float64 `json:"max,omitempty"`
			Unit    *string  `json:"unit,omitempty"`
		}
		if err := json.Unmarshal(comp.Config, &cfg); err == nil {
			if cfg.Name != nil {
				vc.Name = *cfg.Name
			}
			vc.Options = cfg.Options
			vc.Min = cfg.Min
			vc.Max = cfg.Max
			vc.Unit = cfg.Unit
		}
	}

	// Parse status for value.
	if len(comp.Status) > 0 {
		var status struct {
			Value any `json:"value,omitempty"`
		}
		if err := json.Unmarshal(comp.Status, &status); err == nil {
			vc.Value = status.Value
			switch v := status.Value.(type) {
			case bool:
				vc.BoolValue = &v
			case float64:
				vc.NumValue = &v
			case string:
				vc.StrValue = &v
			}
		}
	}

	return vc, true
}

func isVirtualType(t VirtualComponentType) bool {
	switch t {
	case VirtualBoolean, VirtualNumber, VirtualText, VirtualEnum, VirtualButton, VirtualGroup:
		return true
	default:
		return false
	}
}

// GetVirtualBoolean gets the status of a virtual boolean component.
func (s *Service) GetVirtualBoolean(ctx context.Context, identifier string, id int) (*bool, error) {
	var value *bool
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		vb := components.NewVirtualBoolean(conn.RPCClient(), id)
		status, err := vb.GetStatus(ctx)
		if err != nil {
			return err
		}
		value = status.Value
		return nil
	})
	return value, err
}

// SetVirtualBoolean sets the value of a virtual boolean component.
func (s *Service) SetVirtualBoolean(ctx context.Context, identifier string, id int, val bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		vb := components.NewVirtualBoolean(conn.RPCClient(), id)
		return vb.Set(ctx, val)
	})
}

// ToggleVirtualBoolean toggles a virtual boolean component.
func (s *Service) ToggleVirtualBoolean(ctx context.Context, identifier string, id int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		vb := components.NewVirtualBoolean(conn.RPCClient(), id)
		return vb.Toggle(ctx)
	})
}

// GetVirtualNumber gets the status of a virtual number component.
func (s *Service) GetVirtualNumber(ctx context.Context, identifier string, id int) (*float64, error) {
	var value *float64
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		vn := components.NewVirtualNumber(conn.RPCClient(), id)
		status, err := vn.GetStatus(ctx)
		if err != nil {
			return err
		}
		value = status.Value
		return nil
	})
	return value, err
}

// SetVirtualNumber sets the value of a virtual number component.
func (s *Service) SetVirtualNumber(ctx context.Context, identifier string, id int, val float64) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		vn := components.NewVirtualNumber(conn.RPCClient(), id)
		return vn.Set(ctx, val)
	})
}

// GetVirtualText gets the status of a virtual text component.
func (s *Service) GetVirtualText(ctx context.Context, identifier string, id int) (*string, error) {
	var value *string
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		vt := components.NewVirtualText(conn.RPCClient(), id)
		status, err := vt.GetStatus(ctx)
		if err != nil {
			return err
		}
		value = status.Value
		return nil
	})
	return value, err
}

// SetVirtualText sets the value of a virtual text component.
func (s *Service) SetVirtualText(ctx context.Context, identifier string, id int, val string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		vt := components.NewVirtualText(conn.RPCClient(), id)
		return vt.Set(ctx, val)
	})
}

// GetVirtualEnum gets the status of a virtual enum component.
func (s *Service) GetVirtualEnum(ctx context.Context, identifier string, id int) (*string, error) {
	var value *string
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		ve := components.NewVirtualEnum(conn.RPCClient(), id)
		status, err := ve.GetStatus(ctx)
		if err != nil {
			return err
		}
		value = status.Value
		return nil
	})
	return value, err
}

// SetVirtualEnum sets the value of a virtual enum component.
func (s *Service) SetVirtualEnum(ctx context.Context, identifier string, id int, val string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		ve := components.NewVirtualEnum(conn.RPCClient(), id)
		return ve.Set(ctx, val)
	})
}

// TriggerVirtualButton triggers a virtual button component.
func (s *Service) TriggerVirtualButton(ctx context.Context, identifier string, id int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		vb := components.NewVirtualButton(conn.RPCClient(), id)
		return vb.Trigger(ctx)
	})
}

// AddVirtualComponentParams holds parameters for adding a virtual component.
type AddVirtualComponentParams struct {
	Type   VirtualComponentType
	ID     int
	Name   string
	Config map[string]any
}

// AddVirtualComponent adds a new virtual component.
func (s *Service) AddVirtualComponent(ctx context.Context, identifier string, params AddVirtualComponentParams) (int, error) {
	var id int
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		virtual := components.NewVirtual(conn.RPCClient())

		config := params.Config
		if config == nil {
			config = make(map[string]any)
		}
		if params.Name != "" {
			config["name"] = params.Name
		}

		result, err := virtual.Add(ctx, string(params.Type), config, params.ID)
		if err != nil {
			return err
		}
		id = result.ID
		return nil
	})
	return id, err
}

// DeleteVirtualComponent deletes a virtual component by key.
func (s *Service) DeleteVirtualComponent(ctx context.Context, identifier, key string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		virtual := components.NewVirtual(conn.RPCClient())
		return virtual.Delete(ctx, key)
	})
}

// ValidVirtualTypes is the list of valid virtual component types.
var ValidVirtualTypes = []string{"boolean", "number", "text", "enum", "button", "group"}

// IsValidVirtualType checks if a type string is a valid virtual component type.
func IsValidVirtualType(t string) bool {
	for _, valid := range ValidVirtualTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// ParseVirtualKey parses a virtual component key (e.g., "boolean:200") into type and ID.
func ParseVirtualKey(key string) (compType string, id int, err error) {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid key format %q, expected type:id (e.g., boolean:200)", key)
	}

	id, err = strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid component ID %q: %w", parts[1], err)
	}

	if id < 200 || id > 299 {
		return "", 0, fmt.Errorf("virtual component ID must be in range 200-299, got %d", id)
	}

	return parts[0], id, nil
}
