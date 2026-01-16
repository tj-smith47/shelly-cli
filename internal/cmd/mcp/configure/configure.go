// Package configure provides the mcp configure subcommand.
package configure

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mcp"
)

// Options holds command options.
type Options struct {
	Factory    *cmdutil.Factory
	ClaudeCode bool
	ClaudeDesk bool
	Gemini     bool
	ShellyPath string
	DryRun     bool
}

// NewCommand creates the mcp configure command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "configure",
		Aliases: []string{"cfg", "setup"},
		Short:   "Configure AI assistant MCP integration",
		Long: `Configure AI assistant MCP integration.

This command writes the necessary configuration to enable AI assistants to
use the Shelly CLI via MCP (Model Context Protocol).

Supported AI assistants:
  --claude-desktop  Configure Claude Desktop app
  --claude-code     Configure Claude Code (VS Code extension / CLI)
  --gemini          Configure Gemini CLI

You can specify multiple flags to configure all at once.`,
		Example: `  # Configure Claude Desktop
  shelly mcp configure --claude-desktop

  # Configure Claude Code
  shelly mcp configure --claude-code

  # Configure all supported assistants
  shelly mcp configure --claude-desktop --claude-code --gemini

  # Preview changes without writing (dry run)
  shelly mcp configure --claude-desktop --dry-run

  # Use custom shelly binary path
  shelly mcp configure --claude-code --shelly-path /usr/local/bin/shelly`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Default to all if none specified
			if !opts.ClaudeCode && !opts.ClaudeDesk && !opts.Gemini {
				opts.ClaudeCode = true
				opts.ClaudeDesk = true
				opts.Gemini = true
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.ClaudeDesk, "claude-desktop", false, "Configure Claude Desktop")
	cmd.Flags().BoolVar(&opts.ClaudeCode, "claude-code", false, "Configure Claude Code")
	cmd.Flags().BoolVar(&opts.Gemini, "gemini", false, "Configure Gemini CLI")
	cmd.Flags().StringVar(&opts.ShellyPath, "shelly-path", "", "Path to shelly binary (auto-detected if not specified)")
	cmd.Flags().BoolVarP(&opts.DryRun, "dry-run", "n", false, "Preview changes without writing files")

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Detect shelly binary path
	shellyPath := opts.ShellyPath
	if shellyPath == "" {
		var err error
		shellyPath, err = os.Executable()
		if err != nil {
			return fmt.Errorf("failed to detect shelly path: %w", err)
		}
	}

	serverCfg := mcp.ServerConfig{
		Command: shellyPath,
		Args:    []string{"mcp", "start"},
	}

	cfgOpts := &mcp.ConfigOptions{
		DryRun: opts.DryRun,
	}

	var configured []string

	if opts.ClaudeDesk {
		if err := mcp.ConfigureClaudeDesktop(ios, cfgOpts, serverCfg); err != nil {
			ios.Error("Claude Desktop: %v", err)
		} else {
			configured = append(configured, "Claude Desktop")
		}
	}

	if opts.ClaudeCode {
		if err := mcp.ConfigureClaudeCode(ios, cfgOpts, serverCfg); err != nil {
			ios.Error("Claude Code: %v", err)
		} else {
			configured = append(configured, "Claude Code")
		}
	}

	if opts.Gemini {
		if err := mcp.ConfigureGemini(ios, cfgOpts, serverCfg); err != nil {
			ios.Error("Gemini: %v", err)
		} else {
			configured = append(configured, "Gemini")
		}
	}

	if len(configured) == 0 {
		return fmt.Errorf("failed to configure any AI assistant")
	}

	if opts.DryRun {
		ios.Info("Dry run complete - no files were modified")
	} else {
		for _, name := range configured {
			ios.Success("%s configured successfully", name)
		}
		ios.Printf("\nRestart your AI assistant to apply changes.\n")
	}

	return nil
}
