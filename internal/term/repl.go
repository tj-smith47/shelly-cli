package term

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayREPLHelp displays available REPL commands.
func DisplayREPLHelp(ios *iostreams.IOStreams) {
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

// DisplayRegisteredDevices shows registered devices for the REPL.
func DisplayRegisteredDevices(ios *iostreams.IOStreams) {
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

// DisplayRPCMethods displays available RPC methods.
func DisplayRPCMethods(ios *iostreams.IOStreams, methods []string) {
	ios.Println(theme.Bold().Render("Available RPC Methods:"))
	for _, method := range methods {
		ios.Println("  " + method)
	}
}

// DisplayControlResults displays the results of controlling device components.
func DisplayControlResults(ios *iostreams.IOStreams, results []shelly.ComponentControlResult, action string) {
	var failures []string
	for _, r := range results {
		if r.Success {
			ios.Success("%s:%d %s", r.Type, r.ID, action)
		} else {
			failures = append(failures, fmt.Sprintf("%s:%d", r.Type, r.ID))
			ios.DebugErr(fmt.Sprintf("%s %s:%d failed", action, r.Type, r.ID), r.Err)
		}
	}

	if len(failures) > 0 {
		ios.Warning("Failed on: %s", strings.Join(failures, ", "))
	}
}

// DisplayREPLDeviceStatus displays device status JSON in the REPL.
func DisplayREPLDeviceStatus(ios *iostreams.IOStreams, svc *shelly.Service, ctx context.Context, device string) {
	err := svc.WithConnection(ctx, device, func(c *client.Client) error {
		status, err := c.GetStatus(ctx)
		if err != nil {
			return err
		}

		jsonBytes, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return err
		}

		ios.Println(theme.Bold().Render("Device Status:"))
		ios.Println(string(jsonBytes))
		return nil
	})
	if err != nil {
		ios.Error("Failed to get status: %v", err)
	}
}

// DisplayREPLDeviceInfo displays device info in the REPL.
func DisplayREPLDeviceInfo(ios *iostreams.IOStreams, svc *shelly.Service, ctx context.Context, device string) {
	err := svc.WithConnection(ctx, device, func(c *client.Client) error {
		info := c.Info()
		ios.Println(theme.Bold().Render("Device Info:"))
		ios.Println("  ID:       " + info.ID)
		ios.Println("  Model:    " + info.Model)
		ios.Println("  MAC:      " + model.NormalizeMAC(info.MAC))
		ios.Println("  App:      " + info.App)
		ios.Println("  Firmware: " + info.Firmware)
		ios.Printf("  Gen:      %d\n", info.Generation)
		ios.Printf("  Auth:     %v\n", info.AuthEn)
		return nil
	})
	if err != nil {
		ios.Error("Failed to get device info: %v", err)
	}
}

// DisplayRPCResult executes and displays an RPC result in the REPL.
func DisplayRPCResult(ios *iostreams.IOStreams, svc *shelly.Service, ctx context.Context, device, method string, params map[string]any) {
	result, err := svc.RawRPC(ctx, device, method, params)
	if err != nil {
		ios.Error("RPC call failed: %v", err)
		return
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ios.Error("Failed to format response: %v", err)
		return
	}

	ios.Println(string(jsonBytes))
}

// FetchRPCMethods fetches and returns available RPC methods.
func FetchRPCMethods(ios *iostreams.IOStreams, svc *shelly.Service, ctx context.Context, device string) []string {
	result, err := svc.RawRPC(ctx, device, "Shelly.ListMethods", nil)
	if err != nil {
		ios.Error("Failed to list methods: %v", err)
		return nil
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		ios.Error("Failed to parse response: %v", err)
		return nil
	}

	var resp struct {
		Methods []string `json:"methods"`
	}
	if err := json.Unmarshal(jsonBytes, &resp); err != nil {
		ios.Error("Failed to parse methods: %v", err)
		return nil
	}

	return resp.Methods
}

// REPLSession manages an interactive REPL session.
type REPLSession struct {
	ActiveDevice string
	Svc          *shelly.Service
	IOS          *iostreams.IOStreams
}

// NewREPLSession creates a new REPL session.
func NewREPLSession(ios *iostreams.IOStreams, svc *shelly.Service, initialDevice string) *REPLSession {
	return &REPLSession{
		ActiveDevice: initialDevice,
		Svc:          svc,
		IOS:          ios,
	}
}

// ExecuteCommand handles a single REPL command.
// Returns true if the REPL should exit.
func (s *REPLSession) ExecuteCommand(ctx context.Context, cmd string, args []string) bool {
	// Handle exit commands
	if cmd == "exit" || cmd == "quit" || cmd == "q" {
		return true
	}

	// Handle other commands
	switch cmd {
	case "help", "h", "?":
		DisplayREPLHelp(s.IOS)
	case "devices", "ls":
		DisplayRegisteredDevices(s.IOS)
	case "connect", "use", "cd":
		s.handleConnect(args)
	case "disconnect", "clear":
		s.handleDisconnect()
	case "status", "st":
		s.handleStatus(ctx, args)
	case shelly.ActionOn, shelly.ActionOff, shelly.ActionToggle:
		s.handleControl(ctx, args, cmd)
	case "rpc", "call":
		s.handleRPC(ctx, args)
	case "methods":
		s.handleMethods(ctx, args)
	case "info":
		s.handleInfo(ctx, args)
	default:
		s.IOS.Warning("Unknown command: %s", cmd)
		s.IOS.Info("Type 'help' for available commands")
	}

	return false
}

// handleConnect processes the connect command.
func (s *REPLSession) handleConnect(args []string) {
	if len(args) == 0 {
		s.IOS.Error("Usage: connect <device>")
		return
	}
	s.ActiveDevice = args[0]
	s.IOS.Success("Connected to %s", s.ActiveDevice)
}

// handleDisconnect processes the disconnect command.
func (s *REPLSession) handleDisconnect() {
	s.ActiveDevice = ""
	s.IOS.Info("Disconnected from device")
}

// handleStatus processes the status command.
func (s *REPLSession) handleStatus(ctx context.Context, args []string) {
	device := s.resolveDevice(args)
	if device == "" {
		s.IOS.Error("No device specified. Use 'connect <device>' or provide device argument")
		return
	}
	DisplayREPLDeviceStatus(s.IOS, s.Svc, ctx, device)
}

// handleControl processes on/off/toggle commands.
func (s *REPLSession) handleControl(ctx context.Context, args []string, action string) {
	device := s.resolveDevice(args)
	if device == "" {
		s.IOS.Error("No device specified")
		return
	}
	// Capitalize first letter of action
	actionTitle := strings.ToUpper(action[:1]) + action[1:]
	s.IOS.Info("%s %s...", actionTitle, device)

	results, err := s.Svc.ControlAllComponents(ctx, device, action)
	if err != nil {
		s.IOS.Error("Failed: %v", err)
		return
	}

	DisplayControlResults(s.IOS, results, action)
}

// handleRPC processes the rpc command.
func (s *REPLSession) handleRPC(ctx context.Context, args []string) {
	if len(args) == 0 {
		s.IOS.Error("Usage: rpc <method> [params_json]")
		return
	}
	device := s.ActiveDevice
	method := args[0]
	params := args[1:]
	if device == "" {
		// First arg might be device
		if len(args) < 2 {
			s.IOS.Error("No device connected. Use 'connect <device>' first or provide device argument")
			return
		}
		device = args[0]
		method = args[1]
		params = args[2:]
	}

	var paramsMap map[string]any
	if len(params) > 0 {
		paramsJSON := strings.Join(params, " ")
		if err := json.Unmarshal([]byte(paramsJSON), &paramsMap); err != nil {
			s.IOS.Error("Invalid JSON params: %v", err)
			return
		}
	}

	DisplayRPCResult(s.IOS, s.Svc, ctx, device, method, paramsMap)
}

// handleMethods processes the methods command.
func (s *REPLSession) handleMethods(ctx context.Context, args []string) {
	device := s.resolveDevice(args)
	if device == "" {
		s.IOS.Error("No device specified")
		return
	}
	methods := FetchRPCMethods(s.IOS, s.Svc, ctx, device)
	if methods != nil {
		DisplayRPCMethods(s.IOS, methods)
	}
}

// handleInfo processes the info command.
func (s *REPLSession) handleInfo(ctx context.Context, args []string) {
	device := s.resolveDevice(args)
	if device == "" {
		s.IOS.Error("No device specified")
		return
	}
	DisplayREPLDeviceInfo(s.IOS, s.Svc, ctx, device)
}

// resolveDevice returns the device to use - either from args or the active device.
func (s *REPLSession) resolveDevice(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return s.ActiveDevice
}

// RunREPLLoop runs the REPL loop with readline support.
// Returns nil on clean exit, error on failure.
func (s *REPLSession) RunREPLLoop(ctx context.Context) error {
	// Set up readline with history
	historyFile := ""
	configDir, err := config.Dir()
	if err == nil {
		historyFile = filepath.Join(configDir, "repl_history")
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          output.FormatREPLPrompt(s.ActiveDevice),
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer func() {
		if closeErr := rl.Close(); closeErr != nil {
			s.IOS.DebugErr("failed to close readline", closeErr)
		}
	}()

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			s.IOS.Println("\nSession terminated")
			return nil
		default:
		}

		// Update prompt with current device
		rl.SetPrompt(output.FormatREPLPrompt(s.ActiveDevice))

		// Read input with readline
		line, err := rl.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				continue // Ctrl+C, just show new prompt
			}
			if errors.Is(err, io.EOF) {
				s.IOS.Println("\nGoodbye!")
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
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

		shouldExit := s.ExecuteCommand(ctx, cmd, args)
		if shouldExit {
			s.IOS.Println("Goodbye!")
			return nil
		}
	}
}
