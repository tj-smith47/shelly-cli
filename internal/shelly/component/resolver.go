// Package component provides component-related operations for Shelly devices.
package component

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// Type represents a type of device component.
type Type string

// Component type constants.
const (
	TypeSwitch Type = "switch"
	TypeLight  Type = "light"
	TypeCover  Type = "cover"
	TypeInput  Type = "input"
)

// Config holds name and ID extracted from device config.
type Config struct {
	ID   int
	Name string
}

// Resolver resolves component names to IDs.
type Resolver struct {
	configFetcher func(ctx context.Context, device string) (map[string]json.RawMessage, error)
}

// NewResolver creates a new component resolver.
// The configFetcher function should return the full device config as a map.
func NewResolver(configFetcher func(ctx context.Context, device string) (map[string]json.RawMessage, error)) *Resolver {
	return &Resolver{configFetcher: configFetcher}
}

// ResolveByName resolves a component name to its ID for the given device and component type.
// It first checks cached component names in config, then fetches from the device if needed.
// Returns an error if the name is not found.
func (r *Resolver) ResolveByName(ctx context.Context, device string, componentType Type, name string) (int, error) {
	// First check cached component names in device config
	if cachedID, found := r.findInCache(device, componentType, name); found {
		return cachedID, nil
	}

	// Fetch from device
	if r.configFetcher == nil {
		return 0, fmt.Errorf("no config fetcher available")
	}

	configMap, err := r.configFetcher(ctx, device)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch device config: %w", err)
	}

	// Parse components from config
	components := ParseConfigs(configMap, componentType)

	// Find by name (case-insensitive)
	for _, comp := range components {
		if strings.EqualFold(comp.Name, name) {
			return comp.ID, nil
		}
	}

	return 0, fmt.Errorf("%s component with name %q not found", componentType, name)
}

// findInCache checks cached component names in the device config.
func (r *Resolver) findInCache(device string, componentType Type, name string) (int, bool) {
	dev, ok := config.GetDevice(device)
	if !ok {
		return 0, false
	}

	if dev.Components == nil {
		return 0, false
	}

	typeKey := string(componentType)
	typeMap, ok := dev.Components[typeKey]
	if !ok {
		return 0, false
	}

	for id, compName := range typeMap {
		if strings.EqualFold(compName, name) {
			return id, true
		}
	}

	return 0, false
}

// ParseConfigs extracts component configs from a full device config map.
func ParseConfigs(configMap map[string]json.RawMessage, componentType Type) []Config {
	prefix := string(componentType) + ":"

	// Count matching keys for preallocation
	count := 0
	for key := range configMap {
		if strings.HasPrefix(key, prefix) {
			count++
		}
	}

	components := make([]Config, 0, count)
	for key, raw := range configMap {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		var cfg struct {
			ID   int     `json:"id"`
			Name *string `json:"name,omitempty"`
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			continue
		}

		comp := Config{ID: cfg.ID}
		if cfg.Name != nil {
			comp.Name = *cfg.Name
		}
		components = append(components, comp)
	}

	return components
}

// UpdateDeviceComponentNames updates the cached component names for a device.
// This is called after fetching device config to persist names for offline use.
func UpdateDeviceComponentNames(deviceName string, names map[string]map[int]string) error {
	return config.UpdateDeviceComponents(deviceName, names)
}

// NamesFromConfig extracts all component names from a device config map.
// Returns a map of component type -> map of ID -> name.
func NamesFromConfig(configMap map[string]json.RawMessage) map[string]map[int]string {
	result := make(map[string]map[int]string)

	for _, compType := range []Type{TypeSwitch, TypeLight, TypeCover, TypeInput} {
		configs := ParseConfigs(configMap, compType)
		if len(configs) == 0 {
			continue
		}

		typeKey := string(compType)
		result[typeKey] = make(map[int]string)

		for _, cfg := range configs {
			if cfg.Name != "" {
				result[typeKey][cfg.ID] = cfg.Name
			}
		}

		// Only keep if we have at least one named component
		if len(result[typeKey]) == 0 {
			delete(result, typeKey)
		}
	}

	return result
}

// ResolveID resolves a component identifier to an ID.
// If name is provided, it resolves by name; otherwise returns the provided id.
func ResolveID(ctx context.Context, fetcher func(ctx context.Context, device string) (map[string]json.RawMessage, error), device string, componentType Type, id int, name string) (int, error) {
	if name == "" {
		return id, nil
	}

	resolver := NewResolver(fetcher)
	return resolver.ResolveByName(ctx, device, componentType, name)
}

// Info holds basic component information.
type Info struct {
	Type Type
	ID   int
	Name string
}

// FindComponentByName searches for a component by name across all types.
// Returns the component info if found.
func FindComponentByName(ctx context.Context, fetcher func(ctx context.Context, device string) (map[string]json.RawMessage, error), device, name string) (*Info, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("no config fetcher available")
	}

	configMap, err := fetcher(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch device config: %w", err)
	}

	for _, compType := range []Type{TypeSwitch, TypeLight, TypeCover, TypeInput} {
		configs := ParseConfigs(configMap, compType)
		for _, cfg := range configs {
			if strings.EqualFold(cfg.Name, name) {
				return &Info{
					Type: compType,
					ID:   cfg.ID,
					Name: cfg.Name,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("component with name %q not found", name)
}

// GetType returns the Type from a string.
func GetType(typeStr string) (Type, bool) {
	switch strings.ToLower(typeStr) {
	case "switch":
		return TypeSwitch, true
	case "light":
		return TypeLight, true
	case "cover":
		return TypeCover, true
	case "input":
		return TypeInput, true
	default:
		return "", false
	}
}

// UpdateFromFullConfig extracts component names from a full config response
// and updates the device's cached component names.
func UpdateFromFullConfig(deviceName string, configMap map[string]json.RawMessage) error {
	names := NamesFromConfig(configMap)
	if len(names) == 0 {
		return nil // No names to cache
	}
	return UpdateDeviceComponentNames(deviceName, names)
}

// DeviceComponent represents a device component with its cached name.
type DeviceComponent struct {
	Type Type
	ID   int
	Name string
}

// GetCachedComponents returns cached component names for a device.
func GetCachedComponents(deviceName string) []DeviceComponent {
	dev, ok := config.GetDevice(deviceName)
	if !ok || dev.Components == nil {
		return nil
	}

	var result []DeviceComponent
	for typeStr, idMap := range dev.Components {
		compType, ok := GetType(typeStr)
		if !ok {
			continue
		}
		for id, name := range idMap {
			result = append(result, DeviceComponent{
				Type: compType,
				ID:   id,
				Name: name,
			})
		}
	}
	return result
}

// NameLookup provides quick lookup from name to component info.
type NameLookup struct {
	byName map[string]DeviceComponent
}

// NewNameLookup creates a lookup table from cached component names.
func NewNameLookup(deviceName string) *NameLookup {
	lookup := &NameLookup{
		byName: make(map[string]DeviceComponent),
	}

	components := GetCachedComponents(deviceName)
	for _, comp := range components {
		key := strings.ToLower(comp.Name)
		lookup.byName[key] = comp
	}

	return lookup
}

// Find looks up a component by name (case-insensitive).
func (l *NameLookup) Find(name string) (DeviceComponent, bool) {
	comp, ok := l.byName[strings.ToLower(name)]
	return comp, ok
}

// FindByType looks up a component by name and type (case-insensitive).
func (l *NameLookup) FindByType(name string, componentType Type) (DeviceComponent, bool) {
	comp, ok := l.byName[strings.ToLower(name)]
	if !ok {
		return DeviceComponent{}, false
	}
	if comp.Type != componentType {
		return DeviceComponent{}, false
	}
	return comp, ok
}

// GetDeviceComponentNames returns cached component names for a device.
// Returns nil if not cached.
func GetDeviceComponentNames(deviceName string) map[string]map[int]string {
	dev, ok := config.GetDevice(deviceName)
	if !ok || dev.Components == nil {
		return nil
	}

	// Deep copy to prevent mutation
	result := make(map[string]map[int]string)
	for typeStr, idMap := range dev.Components {
		result[typeStr] = make(map[int]string)
		for id, name := range idMap {
			result[typeStr][id] = name
		}
	}
	return result
}

// FormatComponentName returns a display name for a component.
// If the component has a user-configured name, returns it; otherwise returns a generated name.
func FormatComponentName(name string, componentType Type, id int) string {
	if name != "" {
		return name
	}
	return fmt.Sprintf("%s:%d", componentType, id)
}

// MustResolveID resolves a component identifier to an ID.
// If name is provided but resolution fails, returns the provided id with no error.
// This is useful for cases where we want to try name resolution but fall back to ID.
func MustResolveID(ctx context.Context, fetcher func(ctx context.Context, device string) (map[string]json.RawMessage, error), device string, componentType Type, id int, name string) int {
	if name == "" {
		return id
	}

	resolved, err := ResolveID(ctx, fetcher, device, componentType, id, name)
	if err != nil {
		return id // Fall back to provided ID
	}
	return resolved
}

// ResolveGen1ComponentName resolves a component name for Gen1 devices.
// Gen1 uses different config structure: relays[], lights[], rollers[].
func ResolveGen1ComponentName(configMap map[string]json.RawMessage, componentType Type, name string) (int, error) {
	var arrayKey string
	switch componentType {
	case TypeSwitch:
		arrayKey = "relays"
	case TypeLight:
		arrayKey = "lights"
	case TypeCover:
		arrayKey = "rollers"
	default:
		return 0, fmt.Errorf("unsupported Gen1 component type: %s", componentType)
	}

	raw, ok := configMap[arrayKey]
	if !ok {
		return 0, fmt.Errorf("no %s found in config", arrayKey)
	}

	var items []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0, fmt.Errorf("failed to parse %s config: %w", arrayKey, err)
	}

	for i, item := range items {
		if strings.EqualFold(item.Name, name) {
			return i, nil
		}
	}

	return 0, fmt.Errorf("%s with name %q not found", componentType, name)
}

// ResolveIDWithGen attempts to resolve a component ID by name for any generation.
// It checks the generation from device config and uses the appropriate method.
// After fetching config from device, it caches component names for offline use.
func ResolveIDWithGen(ctx context.Context, fetcher func(ctx context.Context, device string) (map[string]json.RawMessage, error), device string, componentType Type, id int, name string) (int, error) {
	if name == "" {
		return id, nil
	}

	// First check cached names (works for both Gen1 and Gen2)
	dev, ok := config.GetDevice(device)
	if ok && dev.Components != nil {
		typeKey := string(componentType)
		if typeMap, ok := dev.Components[typeKey]; ok {
			for cachedID, cachedName := range typeMap {
				if strings.EqualFold(cachedName, name) {
					return cachedID, nil
				}
			}
		}
	}

	// Fetch config and resolve
	if fetcher == nil {
		return 0, fmt.Errorf("no config fetcher available")
	}

	configMap, err := fetcher(ctx, device)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch device config: %w", err)
	}

	// Cache component names for offline use (best-effort, don't fail resolution if caching fails)
	if cacheErr := UpdateFromConfig(device, configMap); cacheErr != nil {
		debug.TraceEvent("component: failed to cache names for %s: %v", device, cacheErr)
	}

	// Check if Gen1 by looking for Gen1-specific keys
	isGen1 := hasGen1Keys(configMap)
	if isGen1 {
		resolvedID, err := ResolveGen1ComponentName(configMap, componentType, name)
		if err != nil {
			return 0, err
		}
		return resolvedID, nil
	}

	// Gen2+ resolution
	components := ParseConfigs(configMap, componentType)
	for _, comp := range components {
		if strings.EqualFold(comp.Name, name) {
			return comp.ID, nil
		}
	}

	return 0, fmt.Errorf("%s component with name %q not found", componentType, name)
}

// hasGen1Keys checks if the config map contains Gen1-specific keys.
func hasGen1Keys(configMap map[string]json.RawMessage) bool {
	gen1Keys := []string{"relays", "lights", "rollers", "meters", "emeters"}
	for _, key := range gen1Keys {
		if _, ok := configMap[key]; ok {
			return true
		}
	}
	return false
}

// gen1NamedItem holds a name field for Gen1 config parsing.
type gen1NamedItem struct {
	Name string `json:"name"`
}

// parseGen1Names extracts names from a Gen1 config array.
func parseGen1Names(raw json.RawMessage) map[int]string {
	var items []gen1NamedItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}

	names := make(map[int]string)
	for i, item := range items {
		if item.Name != "" {
			names[i] = item.Name
		}
	}

	if len(names) == 0 {
		return nil
	}
	return names
}

// Gen1NamesFromConfig extracts component names from Gen1 config.
func Gen1NamesFromConfig(configMap map[string]json.RawMessage) map[string]map[int]string {
	result := make(map[string]map[int]string)

	// Map Gen1 keys to component types
	gen1Mappings := []struct {
		key     string
		typeKey string
	}{
		{"relays", "switch"},
		{"lights", "light"},
		{"rollers", "cover"},
	}

	for _, mapping := range gen1Mappings {
		raw, ok := configMap[mapping.key]
		if !ok {
			continue
		}
		names := parseGen1Names(raw)
		if names != nil {
			result[mapping.typeKey] = names
		}
	}

	return result
}

// UpdateFromConfig extracts and caches component names from a config map.
// Automatically detects Gen1 vs Gen2 format.
func UpdateFromConfig(deviceName string, configMap map[string]json.RawMessage) error {
	var names map[string]map[int]string

	if hasGen1Keys(configMap) {
		names = Gen1NamesFromConfig(configMap)
	} else {
		names = NamesFromConfig(configMap)
	}

	if len(names) == 0 {
		return nil
	}

	return UpdateDeviceComponentNames(deviceName, names)
}
