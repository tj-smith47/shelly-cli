// Package plugins provides plugin discovery and execution functionality.
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// Executor runs plugins with proper environment setup.
type Executor struct{}

// NewExecutor creates a new plugin executor.
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute runs a plugin with the given arguments.
func (e *Executor) Execute(plugin *Plugin, args []string) error {
	return e.ExecuteContext(context.Background(), plugin, args)
}

// ExecuteContext runs a plugin with context for cancellation support.
func (e *Executor) ExecuteContext(ctx context.Context, plugin *Plugin, args []string) error {
	//nolint:gosec // G204: Plugin path is validated by loader, not arbitrary user input
	cmd := exec.CommandContext(ctx, plugin.Path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set up environment
	cmd.Env = e.buildEnvironment(plugin)

	// Run the plugin
	return cmd.Run()
}

// ExecuteCapture runs a plugin and captures its output.
func (e *Executor) ExecuteCapture(plugin *Plugin, args []string) ([]byte, error) {
	return e.ExecuteCaptureContext(context.Background(), plugin, args)
}

// ExecuteCaptureContext runs a plugin with context and captures its output.
func (e *Executor) ExecuteCaptureContext(ctx context.Context, plugin *Plugin, args []string) ([]byte, error) {
	//nolint:gosec // G204: Plugin path is validated by loader, not arbitrary user input
	cmd := exec.CommandContext(ctx, plugin.Path, args...)
	cmd.Env = e.buildEnvironment(plugin)

	return cmd.Output()
}

// buildEnvironment creates the environment for plugin execution.
func (e *Executor) buildEnvironment(plugin *Plugin) []string {
	// Start with current environment
	env := os.Environ()

	// Add Shelly-specific variables
	cfg := config.Get()

	// SHELLY_CONFIG_PATH: Config file location
	if configFile := viper.ConfigFileUsed(); configFile != "" {
		env = append(env, "SHELLY_CONFIG_PATH="+configFile)
	}

	// SHELLY_OUTPUT_FORMAT: Current output format
	env = append(env, "SHELLY_OUTPUT_FORMAT="+cfg.Output)

	// SHELLY_NO_COLOR: Color disabled flag
	if !cfg.Color {
		env = append(env, "SHELLY_NO_COLOR=1")
	}

	// SHELLY_VERBOSE: Verbose mode flag
	if cfg.Verbose {
		env = append(env, "SHELLY_VERBOSE=1")
	}

	// SHELLY_QUIET: Quiet mode flag
	if cfg.Quiet {
		env = append(env, "SHELLY_QUIET=1")
	}

	// SHELLY_API_MODE: API mode (local, cloud, auto)
	// SHELLY_THEME: Current theme name
	env = append(env, "SHELLY_API_MODE="+cfg.APIMode, "SHELLY_THEME="+cfg.GetThemeConfig().Name)

	// SHELLY_DEVICES_JSON: JSON of registered devices
	devicesJSON, err := json.Marshal(cfg.Devices)
	if err != nil {
		// Log warning but continue - plugins can still work without device list
		fmt.Fprintf(os.Stderr, "Warning: failed to marshal devices for plugin: %v\n", err)
	} else {
		env = append(env, "SHELLY_DEVICES_JSON="+string(devicesJSON))
	}

	// SHELLY_PLUGIN_DIR: Directory where plugin is installed
	if plugin != nil && plugin.Dir != "" {
		env = append(env, "SHELLY_PLUGIN_DIR="+plugin.Dir)
	}

	// SHELLY_CLI_VERSION: CLI version for compatibility checks
	env = append(env, "SHELLY_CLI_VERSION="+version.Version)

	return env
}

// RunPlugin is a convenience function to find and execute a plugin with context.
func RunPlugin(ctx context.Context, name string, args []string) error {
	loader := NewLoader()
	plugin, err := loader.Find(name)
	if err != nil {
		return fmt.Errorf("error finding plugin: %w", err)
	}
	if plugin == nil {
		return fmt.Errorf("plugin %q not found", name)
	}

	executor := NewExecutor()
	return executor.ExecuteContext(ctx, plugin, args)
}
