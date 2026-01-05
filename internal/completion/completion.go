// Package completion provides shell completion helper functions for dynamic tab completion.
//
// These completion functions are used for shell tab completion which runs outside
// of normal command execution. They don't have access to cmd.Context() or the
// Factory pattern, so context.Background() and shelly.NewService() are used here.
package completion

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Shell type constants.
const (
	ShellBash       = "bash"
	ShellZsh        = "zsh"
	ShellFish       = "fish"
	ShellPowerShell = "powershell"
)

// DetectShell attempts to detect the user's shell.
func DetectShell() (string, error) {
	// Check SHELL environment variable first
	shell := os.Getenv("SHELL")
	if shell != "" {
		base := filepath.Base(shell)
		switch base {
		case "bash":
			return ShellBash, nil
		case "zsh":
			return ShellZsh, nil
		case "fish":
			return ShellFish, nil
		}
	}

	// On Windows, check for PowerShell
	if runtime.GOOS == "windows" {
		// Check if running in PowerShell
		if os.Getenv("PSModulePath") != "" {
			return ShellPowerShell, nil
		}
	}

	// Try to get parent process name
	ppid := os.Getppid()
	if ppid > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		//nolint:gosec // G204: ppid is from os.Getppid(), not user input
		cmd := exec.CommandContext(ctx, "ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=")
		output, err := cmd.Output()
		if err == nil {
			procName := strings.TrimSpace(string(output))
			procName = filepath.Base(procName)
			switch {
			case strings.Contains(procName, "bash"):
				return ShellBash, nil
			case strings.Contains(procName, "zsh"):
				return ShellZsh, nil
			case strings.Contains(procName, "fish"):
				return ShellFish, nil
			case strings.Contains(procName, "pwsh"), strings.Contains(procName, "powershell"):
				return ShellPowerShell, nil
			}
		}
	}

	return "", fmt.Errorf("could not detect shell")
}

// cache holds cached completion data to avoid slow network queries.
var cache = &completionCache{
	scripts:   make(map[string][]scriptEntry),
	schedules: make(map[string][]scheduleEntry),
	expiry:    make(map[string]time.Time),
}

const cacheTTL = 5 * time.Minute

type completionCache struct {
	sync.RWMutex
	scripts   map[string][]scriptEntry
	schedules map[string][]scheduleEntry
	discovery []string
	expiry    map[string]time.Time
}

type scriptEntry struct {
	ID   int
	Name string
}

type scheduleEntry struct {
	ID       int
	Timespec string
}

// DeviceNames returns a completion function for device names from the registry.
func DeviceNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// GroupNames returns a completion function for group names.
func GroupNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// AliasNames returns a completion function for alias names.
func AliasNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// ThemeNames returns a completion function for theme names.
func ThemeNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// ExtensionNames returns a completion function for extension names.
func ExtensionNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// SceneNames returns a completion function for scene names.
func SceneNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// OutputFormats returns a completion function for output format options.
func OutputFormats() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"table\tTabular format (default)",
			"json\tJSON format",
			"yaml\tYAML format",
			"template\tGo template format",
		}, cobra.ShellCompDirectiveNoFileComp
	}
}

// DevicesOrGroups returns a completion function for device or group names.
// This is useful for commands that accept either.
func DevicesOrGroups() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// NoFile returns a directive that disables file completion.
func NoFile(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// DeviceThenScriptID returns a completion function that completes
// device names for the first arg and script IDs for the second arg.
func DeviceThenScriptID() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// First arg: complete device names
		if len(args) == 0 {
			return deviceNamesFiltered(toComplete)
		}

		// Second arg: complete script IDs from the device
		if len(args) == 1 {
			return scriptIDs(args[0], toComplete)
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// DeviceThenScheduleID returns a completion function that completes
// device names for the first arg and schedule IDs for the second arg.
func DeviceThenScheduleID() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// First arg: complete device names
		if len(args) == 0 {
			return deviceNamesFiltered(toComplete)
		}

		// Second arg: complete schedule IDs from the device
		if len(args) == 1 {
			return scheduleIDs(args[0], toComplete)
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// deviceNamesFiltered returns device names that match the prefix.
func deviceNamesFiltered(toComplete string) ([]string, cobra.ShellCompDirective) {
	devices := config.ListDevices()
	var completions []string
	for name := range devices {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// scriptIDs returns script IDs from the specified device.
func scriptIDs(device, toComplete string) ([]string, cobra.ShellCompDirective) {
	scripts := getCachedScripts(device)
	if scripts == nil {
		// Try to fetch scripts (with short timeout for completion)
		scripts = fetchScripts(device)
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

// scheduleIDs returns schedule IDs from the specified device.
func scheduleIDs(device, toComplete string) ([]string, cobra.ShellCompDirective) {
	schedules := getCachedSchedules(device)
	if schedules == nil {
		// Try to fetch schedules (with short timeout for completion)
		schedules = fetchSchedules(device)
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
func getCachedScripts(device string) []scriptEntry {
	cache.RLock()
	defer cache.RUnlock()

	key := "scripts:" + device
	if exp, ok := cache.expiry[key]; ok && time.Now().Before(exp) {
		return cache.scripts[device]
	}
	return nil
}

// getCachedSchedules returns cached schedule completions for a device.
func getCachedSchedules(device string) []scheduleEntry {
	cache.RLock()
	defer cache.RUnlock()

	key := "schedules:" + device
	if exp, ok := cache.expiry[key]; ok && time.Now().Before(exp) {
		return cache.schedules[device]
	}
	return nil
}

// fetchScripts fetches scripts from a device with a short timeout.
func fetchScripts(device string) []scriptEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	shellySvc := shelly.NewService()
	autoSvc := automation.New(shellySvc, nil)
	scripts, err := autoSvc.ListScripts(ctx, device)
	if err != nil {
		return nil
	}

	result := make([]scriptEntry, len(scripts))
	for i, s := range scripts {
		result[i] = scriptEntry{ID: s.ID, Name: s.Name}
	}

	// Cache the result
	cache.Lock()
	cache.scripts[device] = result
	cache.expiry["scripts:"+device] = time.Now().Add(cacheTTL)
	cache.Unlock()

	return result
}

// fetchSchedules fetches schedules from a device with a short timeout.
func fetchSchedules(device string) []scheduleEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	shellySvc := shelly.NewService()
	autoSvc := automation.New(shellySvc, nil)
	schedules, err := autoSvc.ListSchedules(ctx, device)
	if err != nil {
		return nil
	}

	result := make([]scheduleEntry, len(schedules))
	for i, s := range schedules {
		result[i] = scheduleEntry{ID: s.ID, Timespec: s.Timespec}
	}

	// Cache the result
	cache.Lock()
	cache.schedules[device] = result
	cache.expiry["schedules:"+device] = time.Now().Add(cacheTTL)
	cache.Unlock()

	return result
}

// DiscoveredDevices returns a completion function for discovered device addresses.
// It reads from the discovery cache file if available.
func DiscoveredDevices() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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
	cache.RLock()
	if len(cache.discovery) > 0 {
		if exp, ok := cache.expiry["discovery"]; ok && time.Now().Before(exp) {
			result := cache.discovery
			cache.RUnlock()
			return result
		}
	}
	cache.RUnlock()

	// Try to read from cache file
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil
	}

	cacheFile := filepath.Join(cacheDir, "shelly", "discovery_cache.txt")
	data, err := afero.ReadFile(config.Fs(), cacheFile)
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
	cache.Lock()
	cache.discovery = result
	cache.expiry["discovery"] = time.Now().Add(cacheTTL)
	cache.Unlock()

	return result
}

// DevicesWithGroups returns a completion function that completes device names,
// group names with @ prefix, and @all for all devices.
func DevicesWithGroups() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// TemplateNames returns a completion function for template names.
func TemplateNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		templates := config.ListDeviceTemplates()
		var completions []string
		for name := range templates {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// TemplateThenDevice returns a completion function that completes
// template names for the first arg and device names for the second arg.
func TemplateThenDevice() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: template names
			templates := config.ListDeviceTemplates()
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
			return deviceNamesFiltered(toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// TemplateThenFile returns a completion function that completes
// template names for the first arg and file paths for the second arg.
func TemplateThenFile() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: template names
			templates := config.ListDeviceTemplates()
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

// DeviceThenFile returns a completion function that completes
// device names for the first arg and file paths for the second arg.
func DeviceThenFile() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: device names
			return deviceNamesFiltered(toComplete)
		}
		if len(args) == 1 {
			// Second argument: file path (use default file completion)
			return nil, cobra.ShellCompDirectiveDefault
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// DeviceThenNoComplete returns a completion function that completes
// device names for the first arg and disables completion for subsequent args.
// Useful for commands like kvs get/set/del where the second arg is a user-defined key.
func DeviceThenNoComplete() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return deviceNamesFiltered(toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// NameThenDevice returns a completion function that skips completion
// for the first arg (user-provided name) and completes device names for the second arg.
func NameThenDevice() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: name (no completion)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			// Second argument: device names
			return deviceNamesFiltered(toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// SettingKeys returns a completion function for CLI setting keys.
func SettingKeys() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.FilterSettingKeys(toComplete), cobra.ShellCompDirectiveNoFileComp
	}
}

// SettingKeysWithEquals returns a completion function for CLI setting keys with "=" suffix.
func SettingKeysWithEquals() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		keys := config.FilterSettingKeys(toComplete)
		for i, k := range keys {
			keys[i] = k + "="
		}
		return keys, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	}
}

// FileThenNoComplete returns a completion function that uses default file
// completion for the first arg and disables completion for subsequent args.
func FileThenNoComplete() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: file path (use default file completion)
			return nil, cobra.ShellCompDirectiveDefault
		}
		// Subsequent arguments: no completion
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

	fs := config.Fs()
	dir := filepath.Join(cacheDir, "shelly")
	if err := fs.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	cacheFile := filepath.Join(dir, "discovery_cache.txt")
	data := strings.Join(addresses, "\n")

	// Update memory cache
	cache.Lock()
	cache.discovery = addresses
	cache.expiry["discovery"] = time.Now().Add(cacheTTL)
	cache.Unlock()

	return afero.WriteFile(fs, cacheFile, []byte(data), 0o600)
}

// ScriptTemplateNames returns a completion function for script template names.
func ScriptTemplateNames() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return scriptTemplateNamesFiltered(toComplete), cobra.ShellCompDirectiveNoFileComp
	}
}

// DeviceThenScriptTemplate returns a completion function that completes
// device names for the first arg and script template names for the second arg.
func DeviceThenScriptTemplate() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return deviceNamesFiltered(toComplete)
		}
		if len(args) == 1 {
			return scriptTemplateNamesFiltered(toComplete), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// scriptTemplateNamesFiltered returns script template names matching the prefix.
func scriptTemplateNamesFiltered(toComplete string) []string {
	templates := automation.ListAllScriptTemplates()
	var names []string
	for name := range templates {
		if strings.HasPrefix(name, toComplete) {
			names = append(names, name)
		}
	}
	return names
}
