// Package mcp provides MCP (Model Context Protocol) configuration utilities.
package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ServerConfig represents the MCP server configuration for AI assistants.
type ServerConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// ConfigOptions holds options for writing MCP configuration.
type ConfigOptions struct {
	DryRun bool
}

// ConfigureClaudeDesktop writes MCP configuration for Claude Desktop.
func ConfigureClaudeDesktop(ios *iostreams.IOStreams, opts *ConfigOptions, serverCfg ServerConfig) error {
	// Claude Desktop config path varies by OS
	var configPath string
	switch runtime.GOOS {
	case "darwin":
		configPath = theme.ExpandPath("~/Library/Application Support/Claude/claude_desktop_config.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), "Claude", "claude_desktop_config.json")
	default: // linux and others
		configPath = theme.ExpandPath("~/.config/claude-desktop/claude_desktop_config.json")
	}

	return WriteConfig(ios, opts, configPath, "shelly", serverCfg)
}

// ConfigureClaudeCode writes MCP configuration for Claude Code.
func ConfigureClaudeCode(ios *iostreams.IOStreams, opts *ConfigOptions, serverCfg ServerConfig) error {
	configPath := theme.ExpandPath("~/.claude/settings.json")
	return WriteConfig(ios, opts, configPath, "shelly", serverCfg)
}

// ConfigureGemini writes MCP configuration for Gemini CLI.
func ConfigureGemini(ios *iostreams.IOStreams, opts *ConfigOptions, serverCfg ServerConfig) error {
	configPath := theme.ExpandPath("~/.gemini/settings.json")
	return WriteConfig(ios, opts, configPath, "shelly", serverCfg)
}

// WriteConfig reads existing config, adds/updates the MCP server entry, and writes back.
func WriteConfig(ios *iostreams.IOStreams, opts *ConfigOptions, configPath, serverName string, serverCfg ServerConfig) error {
	fs := config.Fs()

	// Ensure parent directory exists
	dir := filepath.Dir(configPath)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read existing config or create empty one
	var cfg map[string]any
	data, err := afero.ReadFile(fs, configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read config: %w", err)
		}
		cfg = make(map[string]any)
	} else {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Get or create mcpServers map
	mcpServers, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		mcpServers = make(map[string]any)
	}

	// Add our server config
	mcpServers[serverName] = map[string]any{
		"command": serverCfg.Command,
		"args":    serverCfg.Args,
	}
	cfg["mcpServers"] = mcpServers

	// Marshal with indentation
	output, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Show styled path
	styledPath := theme.SemanticPrimary().Render(configPath)
	ios.Printf("  %s\n", styledPath)

	if opts.DryRun {
		ios.Printf("  Would write:\n")
		ios.Printf("  %s\n", string(output))
		return nil
	}

	// Write config file
	if err := afero.WriteFile(fs, configPath, output, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
