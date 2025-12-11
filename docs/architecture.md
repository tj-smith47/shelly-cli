# Architecture Reference

This document provides detailed patterns from industry-standard CLI tools that guide the shelly-cli implementation. These patterns are derived from comprehensive audits of gh, kubectl, docker, jira-cli, gh-dash, and k9s.

## Table of Contents

1. [Factory Pattern (gh/kubectl)](#factory-pattern)
2. [IOStreams Pattern (gh)](#iostreams-pattern)
3. [Command Utilities (gh/kubectl/jira-cli)](#command-utilities)
4. [Directory Structure](#directory-structure)
5. [TUI Architecture (gh-dash/BubbleTea)](#tui-architecture)
6. [Concurrency Patterns](#concurrency-patterns)
7. [Testing Patterns](#testing-patterns)

---

## Factory Pattern

**Source:** gh CLI (`pkg/cmdutil/factory.go`), kubectl

The Factory pattern provides centralized dependency injection for commands. Instead of creating dependencies directly in each command, the Factory provides them lazily on demand.

### Why Use Factory?

1. **Testability**: Replace real dependencies with mocks
2. **Lazy Loading**: Dependencies created only when needed
3. **Consistency**: Single source for all dependencies
4. **Plugin Support**: Plugins can receive the same dependencies

### Implementation

```go
// internal/cmdutil/factory.go

package cmdutil

import (
    "github.com/tj-smith47/shelly-cli/internal/config"
    "github.com/tj-smith47/shelly-cli/internal/iostreams"
    "github.com/tj-smith47/shelly-cli/internal/client"
    "github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Factory provides dependencies to commands
type Factory struct {
    // Lazy initializers - called on first access
    IOStreams    func() *iostreams.IOStreams
    Config       func() (*config.Config, error)
    ShellyClient func(device string) (*client.Client, error)
    Resolver     func() shelly.DeviceResolver

    // Cached instances (set after first call)
    ioStreams    *iostreams.IOStreams
    config       *config.Config
}

// New creates a Factory with production dependencies
func New() *Factory {
    var f Factory

    f.IOStreams = func() *iostreams.IOStreams {
        if f.ioStreams == nil {
            f.ioStreams = iostreams.System()
        }
        return f.ioStreams
    }

    f.Config = func() (*config.Config, error) {
        if f.config == nil {
            cfg, err := config.Load()
            if err != nil {
                return nil, err
            }
            f.config = cfg
        }
        return f.config, nil
    }

    f.ShellyClient = func(device string) (*client.Client, error) {
        cfg, err := f.Config()
        if err != nil {
            return nil, err
        }
        return client.New(device, cfg)
    }

    f.Resolver = func() shelly.DeviceResolver {
        return shelly.NewConfigResolver(f.Config)
    }

    return &f
}
```

### Usage in Commands

```go
// internal/cmd/switch/on/on.go

func NewCommand(f *cmdutil.Factory) *cobra.Command {
    var switchID int

    cmd := &cobra.Command{
        Use:   "on <device>",
        Short: "Turn switch on",
        RunE: func(cmd *cobra.Command, args []string) error {
            ios := f.IOStreams()
            client, err := f.ShellyClient(args[0])
            if err != nil {
                return err
            }

            return runOn(cmd.Context(), ios, client, switchID)
        },
    }

    cmd.Flags().IntVarP(&switchID, "id", "i", 0, "Switch ID")
    return cmd
}
```

---

## IOStreams Pattern

**Source:** gh CLI (`pkg/iostreams/iostreams.go`)

IOStreams provides a unified abstraction for terminal I/O, enabling consistent handling of color, TTY detection, progress indicators, and paging.

### Why Use IOStreams?

1. **Testability**: Capture output in tests
2. **TTY Detection**: Adjust output based on terminal capabilities
3. **Color Management**: Respect NO_COLOR, FORCE_COLOR, etc.
4. **Progress Indicators**: Unified spinner/progress handling
5. **Paging**: Automatic paging for long output

### Implementation

```go
// internal/iostreams/iostreams.go

package iostreams

import (
    "io"
    "os"

    "github.com/briandowns/spinner"
    "github.com/mattn/go-isatty"
)

// IOStreams holds I/O streams and terminal state
type IOStreams struct {
    In     io.Reader
    Out    io.Writer
    ErrOut io.Writer

    // Terminal state (detected once)
    isStdinTTY  bool
    isStdoutTTY bool
    isStderrTTY bool

    // Color settings
    colorEnabled     bool
    colorForced      bool

    // Progress indicator
    progressIndicator *spinner.Spinner
}

// System creates IOStreams connected to stdin/stdout/stderr
func System() *IOStreams {
    ios := &IOStreams{
        In:     os.Stdin,
        Out:    os.Stdout,
        ErrOut: os.Stderr,
    }

    // Detect TTY
    if f, ok := os.Stdin.(*os.File); ok {
        ios.isStdinTTY = isatty.IsTerminal(f.Fd())
    }
    if f, ok := os.Stdout.(*os.File); ok {
        ios.isStdoutTTY = isatty.IsTerminal(f.Fd())
    }
    if f, ok := os.Stderr.(*os.File); ok {
        ios.isStderrTTY = isatty.IsTerminal(f.Fd())
    }

    // Determine color settings
    ios.colorEnabled = ios.isStdoutTTY && !isColorDisabled()

    return ios
}

func isColorDisabled() bool {
    // Check NO_COLOR (https://no-color.org/)
    if _, ok := os.LookupEnv("NO_COLOR"); ok {
        return true
    }
    // Check SHELLY_NO_COLOR
    if _, ok := os.LookupEnv("SHELLY_NO_COLOR"); ok {
        return true
    }
    return false
}

// IsStdoutTTY returns true if stdout is a terminal
func (s *IOStreams) IsStdoutTTY() bool {
    return s.isStdoutTTY
}

// ColorEnabled returns true if color output is enabled
func (s *IOStreams) ColorEnabled() bool {
    return s.colorEnabled
}

// StartProgress starts a spinner with the given message
func (s *IOStreams) StartProgress(msg string) {
    if !s.isStdoutTTY {
        // No spinner for non-TTY, just print message
        fmt.Fprintln(s.ErrOut, msg)
        return
    }

    s.progressIndicator = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
    s.progressIndicator.Suffix = " " + msg
    s.progressIndicator.Writer = s.ErrOut
    s.progressIndicator.Start()
}

// StopProgress stops the current spinner
func (s *IOStreams) StopProgress() {
    if s.progressIndicator != nil {
        s.progressIndicator.Stop()
        s.progressIndicator = nil
    }
}
```

### Test Helper

```go
// internal/testutil/iostreams.go

func NewTestIOStreams() (*IOStreams, *bytes.Buffer, *bytes.Buffer) {
    stdin := &bytes.Buffer{}
    stdout := &bytes.Buffer{}
    stderr := &bytes.Buffer{}

    ios := &IOStreams{
        In:           stdin,
        Out:          stdout,
        ErrOut:       stderr,
        colorEnabled: false, // Disable color in tests
    }

    return ios, stdout, stderr
}
```

---

## Command Utilities

**Source:** gh (`pkg/cmdutil/`), kubectl, jira-cli (`internal/cmdutil/`, `internal/cmdcommon/`)

Shared utilities reduce duplication across commands.

### Generic Command Runner

```go
// internal/cmdutil/runner.go

package cmdutil

import (
    "context"
    "fmt"

    "golang.org/x/sync/errgroup"

    "github.com/tj-smith47/shelly-cli/internal/iostreams"
    "github.com/tj-smith47/shelly-cli/internal/shelly"
)

// ComponentAction is a function that operates on a device component
type ComponentAction func(ctx context.Context, svc *shelly.Service, device string, id int) error

// RunWithSpinner executes an action with a progress spinner
func RunWithSpinner(ctx context.Context, ios *iostreams.IOStreams, msg string, action func(context.Context) error) error {
    ios.StartProgress(msg)
    err := action(ctx)
    ios.StopProgress()
    return err
}

// RunBatch executes an action on multiple devices concurrently
func RunBatch(ctx context.Context, ios *iostreams.IOStreams, targets []string, concurrent int, action ComponentAction) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(concurrent)

    svc := shelly.NewService()

    for _, target := range targets {
        t := target
        g.Go(func() error {
            if err := action(ctx, svc, t, 0); err != nil {
                // Log error but continue with other devices
                fmt.Fprintf(ios.ErrOut, "Error on %s: %v\n", t, err)
                return nil // Don't fail the whole batch
            }
            return nil
        })
    }

    return g.Wait()
}
```

### Output Format Routing

```go
// internal/cmdutil/output.go

package cmdutil

import (
    "io"

    "github.com/tj-smith47/shelly-cli/internal/output"
)

// PrintResult outputs data in the specified format
func PrintResult(w io.Writer, format string, data any, tableFn func(io.Writer, any)) error {
    switch format {
    case "json":
        return output.JSON(w, data)
    case "yaml":
        return output.YAML(w, data)
    case "template":
        // Template handled separately with template string
        return nil
    default:
        tableFn(w, data)
        return nil
    }
}
```

### Shared Flag Helpers

```go
// internal/cmdutil/flags.go

package cmdutil

import (
    "time"

    "github.com/spf13/cobra"
)

// AddComponentIDFlag adds the standard component ID flag
func AddComponentIDFlag(cmd *cobra.Command, target *int, componentName string) {
    cmd.Flags().IntVarP(target, "id", "i", 0, fmt.Sprintf("%s ID (default 0)", componentName))
}

// AddOutputFlag adds the standard output format flag
func AddOutputFlag(cmd *cobra.Command) {
    cmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml, template)")
}

// AddTimeoutFlag adds a timeout flag
func AddTimeoutFlag(cmd *cobra.Command, target *time.Duration, defaultValue time.Duration) {
    cmd.Flags().DurationVar(target, "timeout", defaultValue, "Operation timeout")
}

// AddConcurrencyFlag adds a concurrency flag for batch operations
func AddConcurrencyFlag(cmd *cobra.Command, target *int) {
    cmd.Flags().IntVarP(target, "parallel", "p", 10, "Number of parallel operations")
}
```

---

## Directory Structure

**Source:** gh (`pkg/cmd/`), docker (`cli/command/`), jira-cli (`internal/cmd/`)

### Key Principle: Commands Only Under `internal/cmd/`

The `internal/cmd/` directory contains **ONLY** command definitions. All shared utilities, helpers, and infrastructure live elsewhere.

```
internal/
├── cmd/                    # ONLY command definitions
│   ├── root.go
│   ├── switch/
│   │   ├── switch.go       # Parent command
│   │   ├── on/
│   │   │   └── on.go       # `shelly switch on`
│   │   ├── off/
│   │   │   └── off.go      # `shelly switch off`
│   │   └── status/
│   │       └── status.go   # `shelly switch status`
│   └── ...
│
├── cmdutil/                # Command utilities (NOT under cmd/)
│   ├── factory.go
│   ├── runner.go
│   ├── flags.go
│   └── output.go
│
├── iostreams/              # I/O abstraction (NOT under cmd/)
│   ├── iostreams.go
│   ├── color.go
│   └── progress.go
│
├── client/                 # Device client
├── config/                 # Configuration
├── model/                  # Domain models
├── output/                 # Output formatters
├── shelly/                 # Business logic
└── theme/                  # Theming
```

### Command Structure Pattern

Each command directory contains:

```go
// internal/cmd/switch/on/on.go

package on

import (
    "context"
    "fmt"

    "github.com/spf13/cobra"

    "github.com/tj-smith47/shelly-cli/internal/cmdutil"
    "github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch on command
func NewCommand(f *cmdutil.Factory) *cobra.Command {
    var switchID int

    cmd := &cobra.Command{
        Use:   "on <device>",
        Short: "Turn switch on",
        Long:  `Turn on a switch component on the specified device.`,
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return run(cmd.Context(), f, args[0], switchID)
        },
    }

    cmdutil.AddComponentIDFlag(cmd, &switchID, "Switch")

    return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, switchID int) error {
    ios := f.IOStreams()
    svc := shelly.NewService()

    return cmdutil.RunWithSpinner(ctx, ios, "Turning switch on...", func(ctx context.Context) error {
        if err := svc.SwitchOn(ctx, device, switchID); err != nil {
            return fmt.Errorf("failed to turn switch on: %w", err)
        }

        fmt.Fprintf(ios.Out, "Switch %d turned on\n", switchID)
        return nil
    })
}
```

---

## TUI Architecture

**Source:** gh-dash (dlvhdr/gh-dash), BubbleTea (charmbracelet/bubbletea)

The TUI uses the Elm Architecture via BubbleTea:

- **Model**: Application state
- **Init**: Initial command (data fetching)
- **Update**: Handle messages, return new model + commands
- **View**: Render model to string

### Component Structure

Each TUI component follows the same pattern:

```
internal/tui/components/devicelist/
├── model.go      # Model struct and constructor
├── view.go       # View() string method
├── update.go     # Update(msg) method
├── keys.go       # Component-specific key bindings
└── styles.go     # Component styles
```

### Component Implementation

```go
// internal/tui/components/devicelist/model.go

package devicelist

import (
    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/tj-smith47/shelly-cli/internal/model"
)

// Model holds the device list state
type Model struct {
    table    table.Model
    devices  []model.Device
    loading  bool
    err      error
    width    int
    height   int
}

// New creates a new device list model
func New() Model {
    columns := []table.Column{
        {Title: "Name", Width: 20},
        {Title: "IP", Width: 15},
        {Title: "Type", Width: 15},
        {Title: "Status", Width: 10},
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithFocused(true),
    )

    return Model{
        table:   t,
        loading: true,
    }
}

// Init returns the initial command
func (m Model) Init() tea.Cmd {
    return fetchDevices()
}
```

```go
// internal/tui/components/devicelist/update.go

package devicelist

import (
    tea "github.com/charmbracelet/bubbletea"
)

// DevicesLoadedMsg signals that devices were loaded
type DevicesLoadedMsg struct {
    Devices []model.Device
    Err     error
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
            // Handle selection
            return m, nil
        }

    case DevicesLoadedMsg:
        m.loading = false
        if msg.Err != nil {
            m.err = msg.Err
            return m, nil
        }
        m.devices = msg.Devices
        m.table.SetRows(devicesToRows(m.devices))
        return m, nil

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.table.SetWidth(msg.Width)
        m.table.SetHeight(msg.Height - 4) // Leave room for status
        return m, nil
    }

    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}
```

```go
// internal/tui/components/devicelist/view.go

package devicelist

import (
    "github.com/charmbracelet/lipgloss"
)

// View renders the component
func (m Model) View() string {
    if m.loading {
        return "Loading devices..."
    }

    if m.err != nil {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color("9")).
            Render("Error: " + m.err.Error())
    }

    return m.table.View()
}
```

### Async Data Fetching

```go
// internal/tui/data/devices.go

package data

import (
    tea "github.com/charmbracelet/bubbletea"

    "github.com/tj-smith47/shelly-cli/internal/shelly"
)

// FetchDevices returns a command that fetches devices
func FetchDevices() tea.Cmd {
    return func() tea.Msg {
        svc := shelly.NewService()
        devices, err := svc.ListDevices()
        return DevicesLoadedMsg{
            Devices: devices,
            Err:     err,
        }
    }
}
```

---

## Multi-Writer Output Pattern

**Source:** Docker CLI (`docker build`, `docker compose up`)

Docker's build output shows multiple concurrent operations with per-line progress updates. Each layer/service gets its own line that updates in place. This pattern is ideal for:

- Batch device operations
- Subnet scanning
- Firmware updates across multiple devices
- Scene activation

### Why Multi-Writer?

1. **Visual Clarity**: See all operations at once
2. **Real-time Feedback**: Each target shows its own status
3. **Professional UX**: Modern CLI expectation for concurrent ops

### Implementation with lipgloss

```go
// internal/iostreams/multiwriter.go

package iostreams

import (
    "fmt"
    "io"
    "sync"

    "github.com/charmbracelet/lipgloss"
)

// MultiWriter manages multiple concurrent output lines
type MultiWriter struct {
    mu      sync.Mutex
    out     io.Writer
    lines   map[string]*Line
    order   []string  // Preserve insertion order
    isTTY   bool
}

// Line represents a single output line that can be updated
type Line struct {
    ID      string
    Status  Status
    Message string
}

type Status int

const (
    StatusPending Status = iota
    StatusRunning
    StatusSuccess
    StatusError
)

// NewMultiWriter creates a multi-line writer
func NewMultiWriter(out io.Writer, isTTY bool) *MultiWriter {
    return &MultiWriter{
        out:   out,
        lines: make(map[string]*Line),
        isTTY: isTTY,
    }
}

// AddLine adds a new tracked line
func (m *MultiWriter) AddLine(id, message string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.lines[id] = &Line{
        ID:      id,
        Status:  StatusPending,
        Message: message,
    }
    m.order = append(m.order, id)
}

// UpdateLine updates an existing line
func (m *MultiWriter) UpdateLine(id string, status Status, message string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    if line, ok := m.lines[id]; ok {
        line.Status = status
        line.Message = message
    }

    m.render()
}

// render redraws all lines (TTY only)
func (m *MultiWriter) render() {
    if !m.isTTY {
        return
    }

    // Move cursor up to start of our output
    if len(m.order) > 1 {
        fmt.Fprintf(m.out, "\033[%dA", len(m.order)-1)
    }

    for _, id := range m.order {
        line := m.lines[id]
        fmt.Fprintf(m.out, "\033[2K") // Clear line

        icon := m.statusIcon(line.Status)
        style := m.statusStyle(line.Status)

        fmt.Fprintf(m.out, "%s %s: %s\n",
            icon,
            style.Render(line.ID),
            line.Message,
        )
    }
}

func (m *MultiWriter) statusIcon(s Status) string {
    switch s {
    case StatusPending:
        return "○"
    case StatusRunning:
        return "◐" // Or use spinner
    case StatusSuccess:
        return "✓"
    case StatusError:
        return "✗"
    default:
        return "?"
    }
}

func (m *MultiWriter) statusStyle(s Status) lipgloss.Style {
    switch s {
    case StatusSuccess:
        return lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
    case StatusError:
        return lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
    case StatusRunning:
        return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
    default:
        return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
    }
}

// Finalize prints final state (for non-TTY or completion)
func (m *MultiWriter) Finalize() {
    m.mu.Lock()
    defer m.mu.Unlock()

    if !m.isTTY {
        // Non-TTY: print each line once at the end
        for _, id := range m.order {
            line := m.lines[id]
            icon := m.statusIcon(line.Status)
            fmt.Fprintf(m.out, "%s %s: %s\n", icon, line.ID, line.Message)
        }
    }
}
```

### Usage in Batch Operations

```go
// internal/cmd/batch/on/on.go

func run(ctx context.Context, ios *iostreams.IOStreams, targets []string, switchID int) error {
    mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())

    // Add all lines upfront
    for _, target := range targets {
        mw.AddLine(target, "pending")
    }

    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10)

    for _, target := range targets {
        t := target
        g.Go(func() error {
            mw.UpdateLine(t, iostreams.StatusRunning, "turning on...")

            err := svc.SwitchOn(ctx, t, switchID)
            if err != nil {
                mw.UpdateLine(t, iostreams.StatusError, err.Error())
                return nil // Don't fail whole batch
            }

            mw.UpdateLine(t, iostreams.StatusSuccess, "on")
            return nil
        })
    }

    g.Wait()
    mw.Finalize()
    return nil
}
```

### Example Output

```
✓ living-room-light: on
◐ bedroom-switch: turning on...
✓ kitchen-dimmer: on
○ garage-relay: pending
✗ basement-plug: connection timeout
```

### Opportunities in shelly-cli

| Command | Current | With Multi-Writer |
|---------|---------|-------------------|
| `batch on/off/toggle` | Sequential success/error messages | Per-device progress lines |
| `discover scan` | Single spinner | Per-IP status with progress |
| `firmware update --all` | Single spinner | Per-device update progress |
| `scene activate` | Sequential messages | Per-device activation status |
| `provision bulk` | Unknown | Per-device provisioning progress |

---

## Concurrency Patterns

**Source:** gh, kubectl, best practices

### Use errgroup Instead of WaitGroup

```go
// BEFORE (verbose, error-prone)
var wg sync.WaitGroup
sem := make(chan struct{}, concurrent)
results := make(chan Result, len(targets))
for _, target := range targets {
    wg.Add(1)
    go func(device string) {
        defer wg.Done()
        sem <- struct{}{}
        defer func() { <-sem }()
        // work...
    }(target)
}

// AFTER (cleaner)
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(concurrent)
for _, target := range targets {
    t := target
    g.Go(func() error {
        // work...
        return nil
    })
}
return g.Wait()
```

### Context Propagation

Always pass context through the call chain:

```go
// Get context from Cobra command
func (cmd *cobra.Command) RunE: func(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()  // Use this, NOT context.Background()
    return run(ctx, args[0])
}

// Pass context to all operations
func run(ctx context.Context, device string) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    return client.Call(ctx, device, "method", nil)
}
```

---

## Testing Patterns

### TUI Testing with teatest

For TUI components, use the experimental teatest package from Charm:

```go
// Example TUI test using teatest
// See: https://github.com/charmbracelet/x/tree/main/exp/teatest

import (
    "testing"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/x/exp/teatest"
)

func TestDeviceListView(t *testing.T) {
    m := devicelist.New()
    tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

    // Wait for initial render
    teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
        return bytes.Contains(bts, []byte("Loading"))
    })

    // Send devices loaded message
    tm.Send(devicelist.DevicesLoadedMsg{
        Devices: []model.Device{{Name: "test", IP: "192.168.1.1"}},
    })

    // Verify table renders
    teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
        return bytes.Contains(bts, []byte("test"))
    })

    // Test keyboard navigation
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})

    // Verify quit
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
    tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
```

### Table-Driven Tests

```go
func TestSwitchOn(t *testing.T) {
    tests := []struct {
        name      string
        device    string
        switchID  int
        mockResp  any
        mockErr   error
        wantErr   bool
        wantOut   string
    }{
        {
            name:     "success",
            device:   "test-device",
            switchID: 0,
            mockResp: map[string]any{"was_on": false},
            wantOut:  "Switch 0 turned on\n",
        },
        {
            name:     "device not found",
            device:   "unknown",
            switchID: 0,
            mockErr:  client.ErrDeviceNotFound,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ios, stdout, _ := testutil.NewTestIOStreams()
            f := testutil.NewTestFactory(t)
            f.MockClient.SetResponse("Switch.Set", tt.mockResp, tt.mockErr)

            cmd := on.NewCommand(f)
            cmd.SetArgs([]string{tt.device, "--id", strconv.Itoa(tt.switchID)})

            err := cmd.Execute()

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.wantOut, stdout.String())
        })
    }
}
```

### Mock Factory

```go
// internal/testutil/factory.go

func NewTestFactory(t *testing.T) *cmdutil.Factory {
    t.Helper()

    ios, _, _ := NewTestIOStreams()
    mockClient := NewMockClient()

    return &cmdutil.Factory{
        IOStreams: func() *iostreams.IOStreams {
            return ios
        },
        Config: func() (*config.Config, error) {
            return &config.Config{}, nil
        },
        ShellyClient: func(device string) (*client.Client, error) {
            return mockClient, nil
        },
    }
}
```

---

## References

### CLI Architecture
- **gh CLI**: https://github.com/cli/cli
- **kubectl**: https://github.com/kubernetes/kubectl
- **docker CLI**: https://github.com/docker/cli
- **jira-cli**: https://github.com/ankitpokhrel/jira-cli

### TUI Frameworks
- **BubbleTea**: https://github.com/charmbracelet/bubbletea
- **Bubbles** (components): https://github.com/charmbracelet/bubbles
- **Lipgloss** (styling): https://github.com/charmbracelet/lipgloss
- **Glamour** (markdown): https://github.com/charmbracelet/glamour
- **bubbletint** (themes): https://github.com/lrstanley/bubbletint

### TUI Examples
- **gh-dash**: https://github.com/dlvhdr/gh-dash
- **k9s**: https://github.com/derailed/k9s

### Testing
- **teatest** (TUI testing): https://github.com/charmbracelet/x/tree/main/exp/teatest
