// Package interactive provides the interactive REPL command.
package interactive

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
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Action constants for device control.
const (
	actionOn     = "on"
	actionOff    = "off"
	actionToggle = "toggle"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	NoPrompt bool
}

// NewCommand creates the interactive command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:     "interactive",
		Aliases: []string{"repl", "i"},
		Short:   "Launch interactive REPL",
		Long: `Launch an interactive REPL (Read-Eval-Print Loop) for Shelly CLI.

This provides a command-line shell where you can enter Shelly commands
without prefixing them with 'shelly'. It supports command history within
the session and provides a more interactive experience.

Available commands in REPL:
  help             Show available commands
  devices          List registered devices
  connect <device> Set active device for subsequent commands
  disconnect       Clear active device
  status           Show status of active device
  on               Turn on active device
  off              Turn off active device
  toggle           Toggle active device
  rpc <method>     Execute raw RPC call on active device
  exit, quit, q    Exit the REPL

You can also run any shelly subcommand by typing it directly.`,
		Example: `  # Start interactive mode
  shelly interactive

  # Start with a default device
  shelly interactive --device living-room

  # Example session:
  > devices
  > connect living-room
  > status
  > on
  > exit`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Device, "device", "d", "", "Default device to connect to")
	cmd.Flags().BoolVar(&opts.NoPrompt, "no-prompt", false, "Disable interactive prompt (for scripting)")

	return cmd
}

// replState holds the state of the REPL session.
type replState struct {
	activeDevice string
	svc          *shelly.Service
	ios          *iostreams.IOStreams
}

// run executes the interactive REPL.
func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	if !ios.IsStdinTTY() && !opts.NoPrompt {
		ios.Info("Running in non-interactive mode (stdin is not a TTY)")
	}

	ios.Println(theme.Bold().Render("Shelly Interactive REPL"))
	ios.Info("Type 'help' for available commands, 'exit' to quit")
	ios.Println()

	// Create service
	svc := shelly.NewService()

	state := &replState{
		activeDevice: opts.Device,
		svc:          svc,
		ios:          ios,
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			ios.Println("\nSession terminated")
			return nil
		default:
		}

		// Print prompt
		prompt := buildPrompt(state.activeDevice)
		if _, err := fmt.Fprint(ios.Out, prompt); err != nil {
			ios.DebugErr("failed to print prompt", err)
		}

		// Read input
		if !scanner.Scan() {
			// EOF or error
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

		// Parse and execute command
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		shouldExit := executeCommand(ctx, state, cmd, args)
		if shouldExit {
			ios.Println("Goodbye!")
			return nil
		}
	}
}

// buildPrompt creates the REPL prompt string.
func buildPrompt(activeDevice string) string {
	if activeDevice != "" {
		return fmt.Sprintf("shelly [%s]> ", theme.Highlight().Render(activeDevice))
	}
	return "shelly> "
}

// executeCommand handles a single REPL command.
// Returns true if the REPL should exit.
func executeCommand(ctx context.Context, state *replState, cmd string, args []string) bool {
	// Handle exit commands
	if cmd == "exit" || cmd == "quit" || cmd == "q" {
		return true
	}

	// Handle other commands
	switch cmd {
	case "help", "h", "?":
		showHelp(state.ios)
	case "devices", "ls":
		listDevices(state.ios)
	case "connect", "use", "cd":
		handleConnect(state, args)
	case "disconnect", "clear":
		handleDisconnect(state)
	case "status", "st":
		handleStatus(ctx, state, args)
	case actionOn, actionOff, actionToggle:
		handleControl(ctx, state, args, cmd)
	case "rpc", "call":
		handleRPC(ctx, state, args)
	case "methods":
		handleMethods(ctx, state, args)
	case "info":
		handleInfo(ctx, state, args)
	default:
		state.ios.Warning("Unknown command: %s", cmd)
		state.ios.Info("Type 'help' for available commands")
	}

	return false
}

// handleConnect processes the connect command.
func handleConnect(state *replState, args []string) {
	if len(args) == 0 {
		state.ios.Error("Usage: connect <device>")
		return
	}
	state.activeDevice = args[0]
	state.ios.Success("Connected to %s", state.activeDevice)
}

// handleDisconnect processes the disconnect command.
func handleDisconnect(state *replState) {
	state.activeDevice = ""
	state.ios.Info("Disconnected from device")
}

// handleStatus processes the status command.
func handleStatus(ctx context.Context, state *replState, args []string) {
	device := resolveDevice(state.activeDevice, args)
	if device == "" {
		state.ios.Error("No device specified. Use 'connect <device>' or provide device argument")
		return
	}
	showDeviceStatus(ctx, state, device)
}

// handleControl processes on/off/toggle commands.
func handleControl(ctx context.Context, state *replState, args []string, action string) {
	device := resolveDevice(state.activeDevice, args)
	if device == "" {
		state.ios.Error("No device specified")
		return
	}
	controlDevice(ctx, state, device, action)
}

// handleRPC processes the rpc command.
func handleRPC(ctx context.Context, state *replState, args []string) {
	if len(args) == 0 {
		state.ios.Error("Usage: rpc <method> [params_json]")
		return
	}
	device := state.activeDevice
	method := args[0]
	params := args[1:]
	if device == "" {
		// First arg might be device
		if len(args) < 2 {
			state.ios.Error("No device connected. Use 'connect <device>' first or provide device argument")
			return
		}
		device = args[0]
		method = args[1]
		params = args[2:]
	}
	executeRPC(ctx, state, device, method, params)
}

// handleMethods processes the methods command.
func handleMethods(ctx context.Context, state *replState, args []string) {
	device := resolveDevice(state.activeDevice, args)
	if device == "" {
		state.ios.Error("No device specified")
		return
	}
	listMethods(ctx, state, device)
}

// handleInfo processes the info command.
func handleInfo(ctx context.Context, state *replState, args []string) {
	device := resolveDevice(state.activeDevice, args)
	if device == "" {
		state.ios.Error("No device specified")
		return
	}
	showDeviceInfo(ctx, state, device)
}

// resolveDevice returns the device to use - either from args or the active device.
func resolveDevice(activeDevice string, args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return activeDevice
}

// showHelp displays available commands.
func showHelp(ios *iostreams.IOStreams) {
	ios.Println(theme.Bold().Render("Available Commands:"))
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Navigation:"))
	ios.Println("    help              Show this help message")
	ios.Println("    devices           List registered devices")
	ios.Println("    connect <device>  Set active device")
	ios.Println("    disconnect        Clear active device")
	ios.Println("    exit              Exit the REPL")
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Device Control:"))
	ios.Println("    status [device]   Show device status")
	ios.Println("    info [device]     Show device info")
	ios.Println("    on [device]       Turn on device components")
	ios.Println("    off [device]      Turn off device components")
	ios.Println("    toggle [device]   Toggle device components")
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Advanced:"))
	ios.Println("    rpc <method>      Execute raw RPC call")
	ios.Println("    methods           List available RPC methods")
	ios.Println()
	ios.Info("Tip: Connect to a device first, then omit [device] arguments")
}

// listDevices shows registered devices.
func listDevices(ios *iostreams.IOStreams) {
	cfg := config.Get()
	if cfg == nil || len(cfg.Devices) == 0 {
		ios.Info("No devices registered. Use 'shelly device add' to add devices.")
		return
	}

	ios.Println(theme.Bold().Render("Registered Devices:"))
	for name, dev := range cfg.Devices {
		status := theme.Dim().Render(dev.Address)
		ios.Println("  " + theme.Highlight().Render(name) + " " + status)
	}
}

// showDeviceStatus displays the device status.
func showDeviceStatus(ctx context.Context, state *replState, device string) {
	err := state.svc.WithConnection(ctx, device, func(c *client.Client) error {
		status, err := c.GetStatus(ctx)
		if err != nil {
			return err
		}

		jsonBytes, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return err
		}

		state.ios.Println(theme.Bold().Render("Device Status:"))
		state.ios.Println(string(jsonBytes))
		return nil
	})
	if err != nil {
		state.ios.Error("Failed to get status: %v", err)
	}
}

// showDeviceInfo displays device information.
func showDeviceInfo(ctx context.Context, state *replState, device string) {
	err := state.svc.WithConnection(ctx, device, func(c *client.Client) error {
		info := c.Info()
		state.ios.Println(theme.Bold().Render("Device Info:"))
		state.ios.Println("  ID:       " + info.ID)
		state.ios.Println("  Model:    " + info.Model)
		state.ios.Println("  MAC:      " + info.MAC)
		state.ios.Println("  App:      " + info.App)
		state.ios.Println("  Firmware: " + info.Firmware)
		state.ios.Printf("  Gen:      %d\n", info.Generation)
		state.ios.Printf("  Auth:     %v\n", info.AuthEn)
		return nil
	})
	if err != nil {
		state.ios.Error("Failed to get device info: %v", err)
	}
}

// controlDevice performs on/off/toggle operations on device components.
func controlDevice(ctx context.Context, state *replState, device, action string) {
	// Capitalize first letter of action
	actionTitle := strings.ToUpper(action[:1]) + action[1:]
	state.ios.Info("%s %s...", actionTitle, device)

	err := state.svc.WithConnection(ctx, device, func(c *client.Client) error {
		comps, err := c.ListComponents(ctx)
		if err != nil {
			return err
		}

		for _, comp := range comps {
			var opErr error
			switch comp.Type {
			case model.ComponentSwitch:
				opErr = controlSwitch(ctx, c, comp.ID, action)
			case model.ComponentLight:
				opErr = controlLight(ctx, c, comp.ID, action)
			case model.ComponentRGB:
				opErr = controlRGB(ctx, c, comp.ID, action)
			case model.ComponentCover:
				opErr = controlCover(ctx, c, comp.ID, action)
			default:
				state.ios.Debug("skipping unsupported component type %s:%d", comp.Type, comp.ID)
				continue
			}

			if opErr != nil {
				state.ios.DebugErr(fmt.Sprintf("%s %s:%d failed", action, comp.Type, comp.ID), opErr)
			} else {
				state.ios.Success("%s:%d %s", comp.Type, comp.ID, action)
			}
		}
		return nil
	})
	if err != nil {
		state.ios.Error("Failed: %v", err)
	}
}

// controlSwitch performs on/off/toggle on a switch component.
func controlSwitch(ctx context.Context, c *client.Client, id int, action string) error {
	switch action {
	case actionOn:
		return c.Switch(id).On(ctx)
	case actionOff:
		return c.Switch(id).Off(ctx)
	case actionToggle:
		_, err := c.Switch(id).Toggle(ctx)
		return err
	default:
		return nil
	}
}

// controlLight performs on/off/toggle on a light component.
func controlLight(ctx context.Context, c *client.Client, id int, action string) error {
	switch action {
	case actionOn:
		return c.Light(id).On(ctx)
	case actionOff:
		return c.Light(id).Off(ctx)
	case actionToggle:
		_, err := c.Light(id).Toggle(ctx)
		return err
	default:
		return nil
	}
}

// controlRGB performs on/off/toggle on an RGB component.
func controlRGB(ctx context.Context, c *client.Client, id int, action string) error {
	switch action {
	case actionOn:
		return c.RGB(id).On(ctx)
	case actionOff:
		return c.RGB(id).Off(ctx)
	case actionToggle:
		_, err := c.RGB(id).Toggle(ctx)
		return err
	default:
		return nil
	}
}

// controlCover performs open/close on a cover component.
func controlCover(ctx context.Context, c *client.Client, id int, action string) error {
	switch action {
	case actionOn:
		return c.Cover(id).Open(ctx, nil)
	case actionOff:
		return c.Cover(id).Close(ctx, nil)
	case actionToggle:
		return c.Cover(id).Stop(ctx)
	default:
		return nil
	}
}

// executeRPC executes a raw RPC call.
func executeRPC(ctx context.Context, state *replState, device, method string, args []string) {
	var params map[string]any
	if len(args) > 0 {
		paramsJSON := strings.Join(args, " ")
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			state.ios.Error("Invalid JSON params: %v", err)
			return
		}
	}

	result, err := state.svc.RawRPC(ctx, device, method, params)
	if err != nil {
		state.ios.Error("RPC call failed: %v", err)
		return
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		state.ios.Error("Failed to format response: %v", err)
		return
	}

	state.ios.Println(string(jsonBytes))
}

// listMethods shows available RPC methods.
func listMethods(ctx context.Context, state *replState, device string) {
	result, err := state.svc.RawRPC(ctx, device, "Shelly.ListMethods", nil)
	if err != nil {
		state.ios.Error("Failed to list methods: %v", err)
		return
	}

	// Parse the result
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
