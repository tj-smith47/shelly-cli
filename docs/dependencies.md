# Dependencies

## Core Dependencies

```go
require (
    github.com/tj-smith47/shelly-go v0.1.0  // Shelly device library
    github.com/spf13/cobra v1.8.x           // CLI framework
    github.com/spf13/viper v1.18.x          // Configuration
    github.com/charmbracelet/bubbletea      // TUI framework
    github.com/charmbracelet/bubbles        // TUI components
    github.com/charmbracelet/lipgloss       // Styling (CLI + TUI)
    github.com/charmbracelet/glamour        // Markdown rendering (help screens)
    github.com/charmbracelet/x/exp/teatest  // TUI testing
    github.com/lrstanley/bubbletint         // Themes for ALL output (CLI + TUI)
    github.com/AlecAivazis/survey/v2        // Interactive prompts
    github.com/briandowns/spinner           // Progress spinners
    github.com/olekukonko/tablewriter       // Table output (non-TUI)
    golang.org/x/sync/errgroup              // Concurrent operations
    gopkg.in/yaml.v3                        // YAML support
)
```

## Development Dependencies

```go
require (
    github.com/stretchr/testify             // Testing assertions
    github.com/golang/mock                  // Mocking (or use Go 1.25 testing features)
)
```

---

## Usage Guidelines

### Theming (bubbletint)

`github.com/lrstanley/bubbletint` provides theming for ALL CLI output, not just TUI:
- Table output colors
- Status indicators (success/warning/error)
- Device state colors (online/offline/updating)
- Spinner colors
- All lipgloss-styled output
- TUI dashboard components

**Example:**
```go
import "github.com/lrstanley/bubbletint"

// Get current theme
t := theme.Current()

// Use theme colors
style := lipgloss.NewStyle().
    Foreground(t.Green()).  // Success color
    Bold(true)
```

### Spinners (briandowns/spinner)

Use for long-running operations (>1 second):
- Device discovery (mDNS, BLE, subnet scan)
- Firmware updates and downloads
- Backup/restore operations
- Bulk provisioning
- Cloud authentication

**Example:**
```go
import "github.com/briandowns/spinner"

s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
s.Suffix = " Discovering devices..."
s.Start()
defer s.Stop()

// Do work...
```

### Interactive Prompts (survey)

Use `github.com/AlecAivazis/survey/v2` for interactive input (outside TUI):
- Confirmations (factory reset, delete operations)
- Selections (choose device, select WiFi network)
- Text input (device names, credentials)
- Multi-select (choose devices for batch)
- Password input (auth credentials, encryption)

**Example:**
```go
import "github.com/AlecAivazis/survey/v2"

var confirm bool
prompt := &survey.Confirm{
    Message: "Factory reset device? This cannot be undone.",
}
survey.AskOne(prompt, &confirm)
```

### Concurrent Operations (errgroup)

Use `golang.org/x/sync/errgroup` for batch operations:

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx)
g.SetLimit(10)  // Max concurrent

for _, device := range devices {
    d := device
    g.Go(func() error {
        return svc.TurnOn(ctx, d)
    })
}

return g.Wait()
```

### TUI Framework (BubbleTea)

Use Charm ecosystem for TUI:
- `bubbletea` - Core framework (Elm Architecture)
- `bubbles` - Pre-built components (table, list, textinput, viewport)
- `lipgloss` - Styling and layout
- `glamour` - Markdown rendering

**Component pattern:**
```go
type Model struct {
    // State
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* handle messages */ }
func (m Model) View() string { /* render */ }
```

---

## Go 1.25.5 Features

Use these Go 1.25 features where applicable:

| Feature | Use Case |
|---------|----------|
| `sync.WaitGroup.Go()` | Cleaner goroutine spawning in batch ops |
| `testing/synctest` | Virtualized time for time-dependent tests |
| Range over functions | Iterator patterns |
| Swiss map | Automatic (faster maps) |
| GreenTea GC | Automatic (better GC) |
| Container-aware GOMAXPROCS | Automatic (K8s/Docker) |

**Example - WaitGroup.Go():**
```go
var wg sync.WaitGroup
for _, item := range items {
    i := item
    wg.Go(func() {  // Instead of Add(1) + go func() { defer Done() }
        process(i)
    })
}
wg.Wait()
```
