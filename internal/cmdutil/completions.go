package cmdutil

// Shell completion helper functions for dynamic tab completion.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// completionCache holds cached completion data to avoid slow network queries.
var completionCache = &struct {
	sync.RWMutex
	scripts   map[string][]scriptCompletion
	schedules map[string][]scheduleCompletion
	discovery []string
	expiry    map[string]time.Time
}{
	scripts:   make(map[string][]scriptCompletion),
	schedules: make(map[string][]scheduleCompletion),
	expiry:    make(map[string]time.Time),
}

const completionCacheTTL = 5 * time.Minute

type scriptCompletion struct {
	ID   int
	Name string
}

type scheduleCompletion struct {
	ID       int
	Timespec string
}

// CompleteDeviceNames returns a completion function for device names from the registry.
func CompleteDeviceNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		devices := config.ListDevices()

		var completions []string
		for name := range devices {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteGroupNames returns a completion function for group names.
func CompleteGroupNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		groups := config.ListGroups()

		var completions []string
		for name := range groups {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteAliasNames returns a completion function for alias names.
func CompleteAliasNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		aliases := config.ListAliases()

		var completions []string
		for name, alias := range aliases {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name+"\t"+alias.Command)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteThemeNames returns a completion function for theme names.
func CompleteThemeNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		themes := theme.ListThemes()

		var completions []string
		for _, name := range themes {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteExtensionNames returns a completion function for extension names.
func CompleteExtensionNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		loader := plugins.NewLoader()
		exts, err := loader.Discover()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var completions []string
		for _, ext := range exts {
			if strings.HasPrefix(ext.Name, toComplete) {
				completions = append(completions, ext.Name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteSceneNames returns a completion function for scene names.
func CompleteSceneNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		scenes := config.ListScenes()

		var completions []string
		for name := range scenes {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteOutputFormats returns a completion function for output format options.
func CompleteOutputFormats() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"table\tTabular format (default)",
			"json\tJSON format",
			"yaml\tYAML format",
			"template\tGo template format",
		}, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteDevicesOrGroups returns a completion function for device or group names.
// This is useful for commands that accept either.
func CompleteDevicesOrGroups() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string

		// Add devices
		devices := config.ListDevices()
		for name := range devices {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name+"\tdevice")
			}
		}

		// Add groups
		groups := config.ListGroups()
		for name := range groups {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name+"\tgroup")
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// NoFileCompletion returns a directive that disables file completion.
func NoFileCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// CompleteDeviceThenScriptID returns a completion function that completes
// device names for the first arg and script IDs for the second arg.
func CompleteDeviceThenScriptID() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// First arg: complete device names
		if len(args) == 0 {
			return completeDeviceNamesFiltered(toComplete)
		}

		// Second arg: complete script IDs from the device
		if len(args) == 1 {
			return completeScriptIDs(args[0], toComplete)
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteDeviceThenScheduleID returns a completion function that completes
// device names for the first arg and schedule IDs for the second arg.
func CompleteDeviceThenScheduleID() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// First arg: complete device names
		if len(args) == 0 {
			return completeDeviceNamesFiltered(toComplete)
		}

		// Second arg: complete schedule IDs from the device
		if len(args) == 1 {
			return completeScheduleIDs(args[0], toComplete)
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// completeDeviceNamesFiltered returns device names that match the prefix.
func completeDeviceNamesFiltered(toComplete string) ([]string, cobra.ShellCompDirective) {
	devices := config.ListDevices()
	var completions []string
	for name := range devices {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeScriptIDs returns script IDs from the specified device.
func completeScriptIDs(device, toComplete string) ([]string, cobra.ShellCompDirective) {
	scripts := getCachedScripts(device)
	if scripts == nil {
		// Try to fetch scripts (with short timeout for completion)
		scripts = fetchScriptsForCompletion(device)
	}

	var completions []string
	for _, s := range scripts {
		idStr := fmt.Sprintf("%d", s.ID)
		if strings.HasPrefix(idStr, toComplete) {
			desc := idStr
			if s.Name != "" {
				desc = fmt.Sprintf("%d\t%s", s.ID, s.Name)
			}
			completions = append(completions, desc)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeScheduleIDs returns schedule IDs from the specified device.
func completeScheduleIDs(device, toComplete string) ([]string, cobra.ShellCompDirective) {
	schedules := getCachedSchedules(device)
	if schedules == nil {
		// Try to fetch schedules (with short timeout for completion)
		schedules = fetchSchedulesForCompletion(device)
	}

	var completions []string
	for _, s := range schedules {
		idStr := fmt.Sprintf("%d", s.ID)
		if strings.HasPrefix(idStr, toComplete) {
			desc := fmt.Sprintf("%d\t%s", s.ID, s.Timespec)
			completions = append(completions, desc)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// getCachedScripts returns cached script completions for a device.
func getCachedScripts(device string) []scriptCompletion {
	completionCache.RLock()
	defer completionCache.RUnlock()

	key := "scripts:" + device
	if exp, ok := completionCache.expiry[key]; ok && time.Now().Before(exp) {
		return completionCache.scripts[device]
	}
	return nil
}

// getCachedSchedules returns cached schedule completions for a device.
func getCachedSchedules(device string) []scheduleCompletion {
	completionCache.RLock()
	defer completionCache.RUnlock()

	key := "schedules:" + device
	if exp, ok := completionCache.expiry[key]; ok && time.Now().Before(exp) {
		return completionCache.schedules[device]
	}
	return nil
}

// fetchScriptsForCompletion fetches scripts from a device with a short timeout.
func fetchScriptsForCompletion(device string) []scriptCompletion {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	svc := shelly.NewService()
	scripts, err := svc.ListScripts(ctx, device)
	if err != nil {
		return nil
	}

	result := make([]scriptCompletion, len(scripts))
	for i, s := range scripts {
		result[i] = scriptCompletion{ID: s.ID, Name: s.Name}
	}

	// Cache the result
	completionCache.Lock()
	completionCache.scripts[device] = result
	completionCache.expiry["scripts:"+device] = time.Now().Add(completionCacheTTL)
	completionCache.Unlock()

	return result
}

// fetchSchedulesForCompletion fetches schedules from a device with a short timeout.
func fetchSchedulesForCompletion(device string) []scheduleCompletion {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	svc := shelly.NewService()
	schedules, err := svc.ListSchedules(ctx, device)
	if err != nil {
		return nil
	}

	result := make([]scheduleCompletion, len(schedules))
	for i, s := range schedules {
		result[i] = scheduleCompletion{ID: s.ID, Timespec: s.Timespec}
	}

	// Cache the result
	completionCache.Lock()
	completionCache.schedules[device] = result
	completionCache.expiry["schedules:"+device] = time.Now().Add(completionCacheTTL)
	completionCache.Unlock()

	return result
}

// CompleteDiscoveredDevices returns a completion function for discovered device addresses.
// It reads from the discovery cache file if available.
func CompleteDiscoveredDevices() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		addresses := getDiscoveryCache()

		var completions []string
		for _, addr := range addresses {
			if strings.HasPrefix(addr, toComplete) {
				completions = append(completions, addr)
			}
		}

		// Also include registered device names
		devices := config.ListDevices()
		for name := range devices {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// getDiscoveryCache reads cached discovery results from the cache directory.
func getDiscoveryCache() []string {
	completionCache.RLock()
	if len(completionCache.discovery) > 0 {
		if exp, ok := completionCache.expiry["discovery"]; ok && time.Now().Before(exp) {
			result := completionCache.discovery
			completionCache.RUnlock()
			return result
		}
	}
	completionCache.RUnlock()

	// Try to read from cache file
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil
	}

	cacheFile := filepath.Join(cacheDir, "shelly", "discovery_cache.txt")
	//nolint:gosec // G304: cacheFile is constructed from UserCacheDir() + constant paths, not user input
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	// Cache in memory
	completionCache.Lock()
	completionCache.discovery = result
	completionCache.expiry["discovery"] = time.Now().Add(completionCacheTTL)
	completionCache.Unlock()

	return result
}

// CompleteDevicesWithGroups returns a completion function that completes device names,
// group names with @ prefix, and @all for all devices.
func CompleteDevicesWithGroups() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		devices := config.ListDevices()
		groups := config.ListGroups()
		completions := make([]string, 0, len(devices)+len(groups)+1)
		completions = append(completions, "@all\tall registered devices")
		for name := range groups {
			completions = append(completions, "@"+name+"\tgroup")
		}
		for name := range devices {
			completions = append(completions, name)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// ExpandDeviceArgs expands @all to all registered devices and @groupname to group members.
func ExpandDeviceArgs(devices []string) []string {
	var result []string
	for _, d := range devices {
		switch {
		case d == "@all":
			for name := range config.ListDevices() {
				result = append(result, name)
			}
		case strings.HasPrefix(d, "@"):
			groupName := strings.TrimPrefix(d, "@")
			if g, exists := config.GetGroup(groupName); exists {
				result = append(result, g.Devices...)
			}
		default:
			result = append(result, d)
		}
	}
	return result
}

// CompleteTemplateNames returns a completion function for template names.
func CompleteTemplateNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		templates := config.ListTemplates()
		var completions []string
		for name := range templates {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteTemplateThenDevice returns a completion function that completes
// template names for the first arg and device names for the second arg.
func CompleteTemplateThenDevice() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: template names
			templates := config.ListTemplates()
			var completions []string
			for name := range templates {
				if strings.HasPrefix(name, toComplete) {
					completions = append(completions, name)
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			// Second argument: device names
			return completeDeviceNamesFiltered(toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteTemplateThenFile returns a completion function that completes
// template names for the first arg and file paths for the second arg.
func CompleteTemplateThenFile() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: template names
			templates := config.ListTemplates()
			var completions []string
			for name := range templates {
				if strings.HasPrefix(name, toComplete) {
					completions = append(completions, name)
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			// Second argument: file path (use default file completion)
			return nil, cobra.ShellCompDirectiveDefault
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// SaveDiscoveryCache saves discovered addresses to the cache file.
// This should be called by the discover command after a successful scan.
func SaveDiscoveryCache(addresses []string) error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(cacheDir, "shelly")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	cacheFile := filepath.Join(dir, "discovery_cache.txt")
	data := strings.Join(addresses, "\n")

	// Update memory cache
	completionCache.Lock()
	completionCache.discovery = addresses
	completionCache.expiry["discovery"] = time.Now().Add(completionCacheTTL)
	completionCache.Unlock()

	return os.WriteFile(cacheFile, []byte(data), 0o600)
}
