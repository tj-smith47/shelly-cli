package cmdutil

import "github.com/spf13/cobra"

// cmdNameBLE and cmdNameWiFi name the hardware-provisioning subtrees excluded below.
// They are extracted to consts because cache.go's component parser also references
// "ble"/"wifi", and a third raw copy here trips goconst (which counts occurrences
// across the whole package, test files included).
const (
	cmdNameBLE  = "ble"
	cmdNameWiFi = "wifi"
)

// excludedMCPCommands is the set of command names and aliases whose subtrees must not
// be exposed as MCP tools — interactive prompts, TUIs, terminal monitors, setup
// wizards, hardware provisioning, and streaming commands have no request/response
// shape an AI assistant can drive non-interactively.
var excludedMCPCommands = map[string]bool{
	// TUI commands
	"dash": true, "dashboard": true, "ui": true,
	// Interactive commands
	"repl": true, "interactive": true, "i": true, "shell": true, "sh": true, "console": true,
	// Monitoring (requires terminal)
	"monitor": true, "mon": true,
	// Setup wizards
	"init": true, "setup": true,
	// Provisioning (requires hardware access)
	cmdNameBLE: true, cmdNameWiFi: true,
	// Streaming commands
	"follow": true, "tail": true,
	// External browser commands
	"feedback": true,
}

// IncludeCommandAsMCPTool reports whether cmd should be exposed as an MCP tool. It
// returns false when cmd's own name or alias — or that of any ancestor — is in the
// excluded set, so excluding a parent drops its whole subtree.
//
// Matching is EXACT against each command's name and aliases, never a substring of the
// command path. The distinction is load-bearing: ophis's ExcludeCmdsContaining matches
// substrings of cmd.CommandPath(), which is unusable here because the binary is
// "shelly" (contains "sh") and most subcommands contain the letter "i" (config, switch,
// …) — substring tokens like "sh"/"i" would reject every command and expose zero tools.
func IncludeCommandAsMCPTool(cmd *cobra.Command) bool {
	for cur := cmd; cur != nil; cur = cur.Parent() {
		if excludedMCPCommands[cur.Name()] {
			return false
		}
		for _, alias := range cur.Aliases {
			if excludedMCPCommands[alias] {
				return false
			}
		}
	}
	return true
}
