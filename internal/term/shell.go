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
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ShellSession manages an interactive device shell session.
type ShellSession struct {
	Device string
	Conn   *client.Client
	IOS    *iostreams.IOStreams
}

// NewShellSession creates a new shell session.
func NewShellSession(ios *iostreams.IOStreams, conn *client.Client, device string) *ShellSession {
	return &ShellSession{
		Device: device,
		Conn:   conn,
		IOS:    ios,
	}
}

// FormatShellPrompt creates the shell prompt string.
func FormatShellPrompt(device string) string {
	return fmt.Sprintf("%s> ", theme.Highlight().Render(device))
}

// ExecuteCommand handles a shell command.
// Returns true if the shell should exit.
func (s *ShellSession) ExecuteCommand(ctx context.Context, line string) bool {
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
		DisplayShellHelp(s.IOS)
		return false

	case "info":
		s.showInfo()
		return false

	case "status":
		s.showStatus(ctx)
		return false

	case "config":
		s.showConfig(ctx)
		return false

	case "methods":
		s.showMethods(ctx)
		return false

	case "components":
		s.showComponents(ctx)
		return false
	}

	// Otherwise, treat as RPC method call
	s.executeRPC(ctx, cmd, paramsStr)
	return false
}

// DisplayShellHelp displays available shell commands.
func DisplayShellHelp(ios *iostreams.IOStreams) {
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
func (s *ShellSession) showInfo() {
	info := s.Conn.Info()
	s.IOS.Println(theme.Bold().Render("Device Information:"))
	s.IOS.Println("  ID:         " + info.ID)
	s.IOS.Println("  Model:      " + info.Model)
	s.IOS.Println("  MAC:        " + info.MAC)
	s.IOS.Println("  App:        " + info.App)
	s.IOS.Println("  Firmware:   " + info.Firmware)
	s.IOS.Printf("  Generation: %d\n", info.Generation)
	s.IOS.Printf("  Auth:       %v\n", info.AuthEn)
}

// showStatus displays device status.
func (s *ShellSession) showStatus(ctx context.Context) {
	status, err := s.Conn.GetStatus(ctx)
	if err != nil {
		s.IOS.Error("Failed to get status: %v", err)
		return
	}

	jsonBytes, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		s.IOS.Error("Failed to format status: %v", err)
		return
	}

	s.IOS.Println(string(jsonBytes))
}

// showConfig displays device configuration.
func (s *ShellSession) showConfig(ctx context.Context) {
	cfg, err := s.Conn.GetConfig(ctx)
	if err != nil {
		s.IOS.Error("Failed to get config: %v", err)
		return
	}

	jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		s.IOS.Error("Failed to format config: %v", err)
		return
	}

	s.IOS.Println(string(jsonBytes))
}

// showMethods lists available RPC methods.
func (s *ShellSession) showMethods(ctx context.Context) {
	result, err := s.Conn.Call(ctx, "Shelly.ListMethods", nil)
	if err != nil {
		s.IOS.Error("Failed to list methods: %v", err)
		return
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		s.IOS.Error("Failed to parse response: %v", err)
		return
	}

	var resp struct {
		Methods []string `json:"methods"`
	}
	if err := json.Unmarshal(jsonBytes, &resp); err != nil {
		s.IOS.Error("Failed to parse methods: %v", err)
		return
	}

	s.IOS.Println(theme.Bold().Render("Available RPC Methods:"))
	for _, method := range resp.Methods {
		s.IOS.Println("  " + method)
	}
}

// showComponents lists device components.
func (s *ShellSession) showComponents(ctx context.Context) {
	comps, err := s.Conn.ListComponents(ctx)
	if err != nil {
		s.IOS.Error("Failed to list components: %v", err)
		return
	}

	s.IOS.Println(theme.Bold().Render("Device Components:"))
	for _, comp := range comps {
		s.IOS.Printf("  %s:%d (%s)\n", comp.Type, comp.ID, comp.Key)
	}
}

// executeRPC executes an RPC method call.
func (s *ShellSession) executeRPC(ctx context.Context, method, paramsStr string) {
	var params map[string]any
	if paramsStr != "" {
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			s.IOS.Error("Invalid JSON params: %v", err)
			s.IOS.Info("Usage: %s {\"key\": \"value\"}", method)
			return
		}
	}

	result, err := s.Conn.Call(ctx, method, params)
	if err != nil {
		s.IOS.Error("RPC error: %v", err)
		return
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		s.IOS.Error("Failed to format response: %v", err)
		return
	}

	s.IOS.Println(string(jsonBytes))
}

// RunShellLoop runs the shell loop with readline support.
// Returns nil on clean exit, error on failure.
func (s *ShellSession) RunShellLoop(ctx context.Context) error {
	// Set up readline with history
	historyFile := ""
	configDir, err := config.Dir()
	if err == nil {
		historyFile = filepath.Join(configDir, "shell_history")
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          FormatShellPrompt(s.Device),
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

		shouldExit := s.ExecuteCommand(ctx, line)
		if shouldExit {
			s.IOS.Println("Goodbye!")
			return nil
		}
	}
}
