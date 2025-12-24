// Package plugins provides plugin discovery, loading, and hook execution.
package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// HookExecutor executes plugin hooks.
type HookExecutor struct {
	plugin *Plugin
}

// NewHookExecutor creates a hook executor for a plugin.
func NewHookExecutor(plugin *Plugin) *HookExecutor {
	return &HookExecutor{plugin: plugin}
}

// ExecuteDetect runs the detect hook to probe if an address belongs to this platform.
func (e *HookExecutor) ExecuteDetect(ctx context.Context, address string, auth *model.Auth) (*DeviceDetectionResult, error) {
	if e.plugin.Manifest == nil || e.plugin.Manifest.Hooks == nil || e.plugin.Manifest.Hooks.Detect == "" {
		return nil, fmt.Errorf("plugin %s does not have detect hook", e.plugin.Name)
	}

	args := []string{"--address", address}
	if auth != nil && auth.Username != "" {
		args = append(args, "--auth-user", auth.Username, "--auth-pass", auth.Password)
	}

	output, err := e.executeHook(ctx, e.plugin.Manifest.Hooks.Detect, args)
	if err != nil {
		return nil, err
	}

	var result DeviceDetectionResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse detect response: %w", err)
	}

	return &result, nil
}

// ExecuteStatus runs the status hook to get device status.
func (e *HookExecutor) ExecuteStatus(ctx context.Context, address string, auth *model.Auth) (*DeviceStatusResult, error) {
	if e.plugin.Manifest == nil || e.plugin.Manifest.Hooks == nil || e.plugin.Manifest.Hooks.Status == "" {
		return nil, fmt.Errorf("plugin %s does not have status hook", e.plugin.Name)
	}

	args := []string{"--address", address}
	if auth != nil && auth.Username != "" {
		args = append(args, "--auth-user", auth.Username, "--auth-pass", auth.Password)
	}

	output, err := e.executeHook(ctx, e.plugin.Manifest.Hooks.Status, args)
	if err != nil {
		return nil, err
	}

	var result DeviceStatusResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	return &result, nil
}

// ExecuteControl runs the control hook to execute device commands.
func (e *HookExecutor) ExecuteControl(ctx context.Context, address string, auth *model.Auth, action, component string, id int) (*ControlResult, error) {
	if e.plugin.Manifest == nil || e.plugin.Manifest.Hooks == nil || e.plugin.Manifest.Hooks.Control == "" {
		return nil, fmt.Errorf("plugin %s does not have control hook", e.plugin.Name)
	}

	args := []string{
		"--address", address,
		"--action", action,
		"--component", component,
		"--id", fmt.Sprintf("%d", id),
	}
	if auth != nil && auth.Username != "" {
		args = append(args, "--auth-user", auth.Username, "--auth-pass", auth.Password)
	}

	output, err := e.executeHook(ctx, e.plugin.Manifest.Hooks.Control, args)
	if err != nil {
		return nil, err
	}

	var result ControlResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse control response: %w", err)
	}

	return &result, nil
}

// ExecuteCheckUpdates runs the check_updates hook to check for firmware updates.
func (e *HookExecutor) ExecuteCheckUpdates(ctx context.Context, address string, auth *model.Auth) (*FirmwareUpdateInfo, error) {
	if e.plugin.Manifest == nil || e.plugin.Manifest.Hooks == nil || e.plugin.Manifest.Hooks.CheckUpdates == "" {
		return nil, fmt.Errorf("plugin %s does not have check_updates hook", e.plugin.Name)
	}

	args := []string{"--address", address}
	if auth != nil && auth.Username != "" {
		args = append(args, "--auth-user", auth.Username, "--auth-pass", auth.Password)
	}

	output, err := e.executeHook(ctx, e.plugin.Manifest.Hooks.CheckUpdates, args)
	if err != nil {
		return nil, err
	}

	var result FirmwareUpdateInfo
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse check_updates response: %w", err)
	}

	return &result, nil
}

// ExecuteApplyUpdate runs the apply_update hook to apply firmware updates.
func (e *HookExecutor) ExecuteApplyUpdate(ctx context.Context, address string, auth *model.Auth, stage, url string) (*UpdateResult, error) {
	if e.plugin.Manifest == nil || e.plugin.Manifest.Hooks == nil || e.plugin.Manifest.Hooks.ApplyUpdate == "" {
		return nil, fmt.Errorf("plugin %s does not have apply_update hook", e.plugin.Name)
	}

	args := []string{"--address", address}
	if stage != "" {
		args = append(args, "--stage", stage)
	}
	if url != "" {
		args = append(args, "--url", url)
	}
	if auth != nil && auth.Username != "" {
		args = append(args, "--auth-user", auth.Username, "--auth-pass", auth.Password)
	}

	output, err := e.executeHook(ctx, e.plugin.Manifest.Hooks.ApplyUpdate, args)
	if err != nil {
		return nil, err
	}

	var result UpdateResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse apply_update response: %w", err)
	}

	return &result, nil
}

// executeHook is the internal hook runner.
// It handles relative paths (prepending plugin directory) and captures stdout/stderr.
func (e *HookExecutor) executeHook(ctx context.Context, hookCmd string, args []string) ([]byte, error) {
	// Parse the hook command - it may include the binary and subcommand
	// e.g., "./shelly-tasmota detect" or just "detect"
	parts := strings.Fields(hookCmd)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty hook command")
	}

	cmdPath := parts[0]
	hookArgs := parts[1:] // Any arguments from the hook definition

	// If the command is relative (doesn't start with /), prepend plugin directory
	if !filepath.IsAbs(cmdPath) && e.plugin.Dir != "" {
		// Handle ./ prefix
		cmdPath = strings.TrimPrefix(cmdPath, "./")
		cmdPath = filepath.Join(e.plugin.Dir, cmdPath)
	}

	// Combine hook definition args with caller args
	allArgs := make([]string, 0, len(hookArgs)+len(args))
	allArgs = append(allArgs, hookArgs...)
	allArgs = append(allArgs, args...)

	//nolint:gosec // G204: Plugin path is from manifest in known plugins directory
	cmd := exec.CommandContext(ctx, cmdPath, allArgs...)

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set up environment
	cmd.Env = e.buildHookEnvironment()

	// Run the hook
	if err := cmd.Run(); err != nil {
		// Include stderr in error message if available
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("hook execution failed: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, fmt.Errorf("hook execution failed: %w", err)
	}

	return stdout.Bytes(), nil
}

// buildHookEnvironment creates the environment for hook execution.
func (e *HookExecutor) buildHookEnvironment() []string {
	// Start with current environment
	env := os.Environ()

	// Add plugin-specific variables
	if e.plugin.Dir != "" {
		env = append(env, "SHELLY_PLUGIN_DIR="+e.plugin.Dir)
	}

	// Add CLI version for compatibility checks
	env = append(env, "SHELLY_CLI_VERSION="+version.Version)

	return env
}
