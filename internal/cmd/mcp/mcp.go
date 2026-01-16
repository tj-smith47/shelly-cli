// Package mcp provides the MCP (Model Context Protocol) server command.
package mcp

import (
	"github.com/njayp/ophis"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/mcp/configure"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// excludedCommands lists commands that should not be exposed via MCP.
// These are interactive, TUI, or otherwise incompatible with AI assistant usage.
var excludedCommands = []string{
	// TUI commands
	"dash", "dashboard", "ui",
	// Interactive commands
	"repl", "interactive", "i", "shell", "sh", "console",
	// Monitoring (requires terminal)
	"monitor", "mon",
	// Setup wizards
	"init", "setup",
	// Provisioning (requires hardware access)
	"ble", "wifi",
	// Streaming commands
	"follow", "tail",
	// External browser commands
	"feedback",
}

// NewCommand creates the mcp command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mcp",
		Aliases: []string{"rpc"},
		Short:   "MCP server for AI assistant integration",
		Long: `MCP (Model Context Protocol) server for AI assistant integration.

This command allows AI assistants like Claude, Gemini, and others to interact
with your Shelly devices through the CLI. The MCP server exposes CLI commands
as tools that AI assistants can invoke.

Use 'mcp start' to run the MCP server, or 'mcp configure' to set up AI
assistant configuration files automatically.`,
		Example: `  # Start the MCP server
  shelly mcp start

  # Enable in Claude Desktop
  shelly mcp claude enable

  # Enable in VS Code
  shelly mcp vscode enable

  # Configure Gemini CLI (or other assistants)
  shelly mcp configure --gemini

  # List available MCP tools
  shelly mcp tools`,
	}

	// Configure ophis with excluded commands
	cfg := &ophis.Config{
		Selectors: []ophis.Selector{
			{
				CmdSelector: ophis.ExcludeCmdsContaining(excludedCommands...),
			},
		},
	}

	// Get the ophis command and extract its subcommands
	ophisCmd := ophis.Command(cfg)
	for _, subCmd := range ophisCmd.Commands() {
		cmd.AddCommand(subCmd)
	}

	// Add configure subcommand for additional AI assistants
	cmd.AddCommand(configure.NewCommand(f))

	return cmd
}
