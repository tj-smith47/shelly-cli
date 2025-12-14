// Package shellcmd provides the device shell command.
package shellcmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the shell command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:     "shell <device>",
		Aliases: []string{"sh", "console"},
		Short:   "Interactive shell for a specific device",
		Long: `Open an interactive shell for a specific Shelly device.

This provides direct access to execute RPC commands on the device.
It maintains a persistent connection and allows you to explore the
device's capabilities interactively.

Available commands:
  help           Show available commands
  info           Show device information
  status         Show device status
  config         Show device configuration
  methods        List available RPC methods
  <method>       Execute RPC method (e.g., Switch.GetStatus, Shelly.GetConfig)
  exit           Close shell

RPC methods can be called directly by typing the method name.
For methods requiring parameters, provide JSON after the method name.`,
		Example: `  # Open shell for a device
  shelly shell living-room

  # Example session:
  shell> info
  shell> methods
  shell> Switch.GetStatus {"id":0}
  shell> Shelly.GetConfig
  shell> exit`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

// shellState holds the state of the device shell session.
type shellState struct {
	device string
	conn   *client.Client
	ios    *iostreams.IOStreams
}

// run executes the device shell.
func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Connect to device
	ios.Info("Connecting to %s...", opts.Device)
	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", opts.Device, err)
	}
	defer iostreams.CloseWithDebug("closing device shell connection", conn)

	info := conn.Info()
	ios.Println()
	ios.Println(theme.Bold().Render(fmt.Sprintf("Connected to %s", info.Model)))
	ios.Printf("  ID: %s | MAC: %s | FW: %s\n", info.ID, info.MAC, info.Firmware)
	ios.Info("Type 'help' for commands, 'methods' to list RPC methods, 'exit' to quit")
	ios.Println()

	state := &shellState{
		device: opts.Device,
		conn:   conn,
		ios:    ios,
	}

	scanner := bufio.NewScanner(os.Stdin)
	prompt := fmt.Sprintf("%s> ", theme.Highlight().Render(opts.Device))

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			ios.Println("\nSession terminated")
			return nil
		default:
		}

		// Print prompt
		if _, err := fmt.Fprint(ios.Out, prompt); err != nil {
			ios.DebugErr("failed to print prompt", err)
		}

		// Read input
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			ios.Println("\nGoodbye!")
			return nil
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		shouldExit := executeShellCommand(ctx, state, line)
		if shouldExit {
			ios.Println("Goodbye!")
			return nil
		}
	}
}

// executeShellCommand handles a shell command.
// Returns true if the shell should exit.
func executeShellCommand(ctx context.Context, state *shellState, line string) bool {
	ios := state.ios

	// Split line into command and params
	parts := strings.SplitN(line, " ", 2)
	cmd := parts[0]
	paramsStr := ""
	if len(parts) > 1 {
		paramsStr = strings.TrimSpace(parts[1])
	}

	// Handle built-in commands (case-insensitive)
	switch strings.ToLower(cmd) {
	case "exit", "quit", "q":
		return true

	case "help", "h", "?":
		showShellHelp(ios)
		return false

	case "info":
		showInfo(state)
		return false

	case "status":
		showStatus(ctx, state)
		return false

	case "config":
		showConfig(ctx, state)
		return false

	case "methods":
		showMethods(ctx, state)
		return false

	case "components":
		showComponents(ctx, state)
		return false
	}

	// Otherwise, treat as RPC method call
	executeRPCMethod(ctx, state, cmd, paramsStr)
	return false
}

// showShellHelp displays available commands.
func showShellHelp(ios *iostreams.IOStreams) {
	ios.Println(theme.Bold().Render("Shell Commands:"))
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Built-in:"))
	ios.Println("    help         Show this help")
	ios.Println("    info         Device information")
	ios.Println("    status       Device status (JSON)")
	ios.Println("    config       Device configuration (JSON)")
	ios.Println("    methods      List available RPC methods")
	ios.Println("    components   List device components")
	ios.Println("    exit         Close shell")
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("RPC Methods:"))
	ios.Println("    <Method.Name>               Call method without params")
	ios.Println("    <Method.Name> {\"key\":val}   Call method with JSON params")
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Examples:"))
	ios.Println("    Switch.GetStatus {\"id\":0}")
	ios.Println("    Switch.Set {\"id\":0,\"on\":true}")
	ios.Println("    Shelly.GetDeviceInfo")
	ios.Println("    Script.List")
}

// showInfo displays device information.
func showInfo(state *shellState) {
	info := state.conn.Info()
	state.ios.Println(theme.Bold().Render("Device Information:"))
	state.ios.Println("  ID:         " + info.ID)
	state.ios.Println("  Model:      " + info.Model)
	state.ios.Println("  MAC:        " + info.MAC)
	state.ios.Println("  App:        " + info.App)
	state.ios.Println("  Firmware:   " + info.Firmware)
	state.ios.Printf("  Generation: %d\n", info.Generation)
	state.ios.Printf("  Auth:       %v\n", info.AuthEn)
}

// showStatus displays device status.
func showStatus(ctx context.Context, state *shellState) {
	status, err := state.conn.GetStatus(ctx)
	if err != nil {
		state.ios.Error("Failed to get status: %v", err)
		return
	}

	jsonBytes, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		state.ios.Error("Failed to format status: %v", err)
		return
	}

	state.ios.Println(string(jsonBytes))
}

// showConfig displays device configuration.
func showConfig(ctx context.Context, state *shellState) {
	cfg, err := state.conn.GetConfig(ctx)
	if err != nil {
		state.ios.Error("Failed to get config: %v", err)
		return
	}

	jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		state.ios.Error("Failed to format config: %v", err)
		return
	}

	state.ios.Println(string(jsonBytes))
}

// showMethods lists available RPC methods.
func showMethods(ctx context.Context, state *shellState) {
	result, err := state.conn.Call(ctx, "Shelly.ListMethods", nil)
	if err != nil {
		state.ios.Error("Failed to list methods: %v", err)
		return
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		state.ios.Error("Failed to parse response: %v", err)
		return
	}

	var resp struct {
		Methods []string `json:"methods"`
	}
	if err := json.Unmarshal(jsonBytes, &resp); err != nil {
		state.ios.Error("Failed to parse methods: %v", err)
		return
	}

	state.ios.Println(theme.Bold().Render("Available RPC Methods:"))
	for _, method := range resp.Methods {
		state.ios.Println("  " + method)
	}
}

// showComponents lists device components.
func showComponents(ctx context.Context, state *shellState) {
	comps, err := state.conn.ListComponents(ctx)
	if err != nil {
		state.ios.Error("Failed to list components: %v", err)
		return
	}

	state.ios.Println(theme.Bold().Render("Device Components:"))
	for _, comp := range comps {
		state.ios.Printf("  %s:%d (%s)\n", comp.Type, comp.ID, comp.Key)
	}
}

// executeRPCMethod executes an RPC method call.
func executeRPCMethod(ctx context.Context, state *shellState, method, paramsStr string) {
	var params map[string]any
	if paramsStr != "" {
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			state.ios.Error("Invalid JSON params: %v", err)
			state.ios.Info("Usage: %s {\"key\": \"value\"}", method)
			return
		}
	}

	result, err := state.conn.Call(ctx, method, params)
	if err != nil {
		state.ios.Error("RPC error: %v", err)
		return
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		state.ios.Error("Failed to format response: %v", err)
		return
	}

	state.ios.Println(string(jsonBytes))
}
