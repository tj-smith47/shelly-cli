# Shelly CLI - Comprehensive Implementation Plan

> ⛔ **STOP. READ [RULES.md](RULES.md) FIRST.** All mandatory workflow rules are there.

---

## Session Quick Start

**For LLM sessions:** This section contains everything needed to start working immediately.

### What This Project Is
A production-ready Cobra CLI for Shelly smart home devices, targeting adoption by ALLTERCO Robotics as the official Shelly CLI. Built on the `shelly-go` library (at `/db/appdata/shelly-go`).

### Before You Code
1. **Read [RULES.md](RULES.md)** - Mandatory workflow rules (factory pattern, IOStreams, error handling)
2. **Check Current Phase** (see status table below) - Pick up where we left off
3. **Verify shelly-go API** - Check `gen2/components/` before assuming features exist
4. **Run existing tests** - `go test ./...` to ensure you don't break anything

### Key Patterns (Enforced)
```go
// ✅ CORRECT: Use factory pattern
func NewCmdSwitch(f *cmdutil.Factory) *cobra.Command

// ✅ CORRECT: Get IOStreams from factory
ios := f.IOStreams()

// ✅ CORRECT: Context from command
ctx := cmd.Context()

// ✅ CORRECT: Themed output
fmt.Fprintf(ios.Out, "%s: %s\n", ios.ColorScheme().Cyan("Power"), value)

// ❌ WRONG: Direct instantiation, fmt.Println, context.Background()
```

### Commands to Run
```bash
go build ./...                    # Must pass
golangci-lint run ./...           # Must pass
go test ./...                     # Must pass
./shelly <command> --help         # Verify help text
./shelly <command> -o json        # Verify JSON output works
```

### Where to Find Things
- **Commands**: `internal/cmd/<command>/` (subdirectory per subcommand)
- **Service layer**: `internal/shelly/` (business logic, device communication)
- **Output/theming**: `internal/iostreams/` (IOStreams, colors, spinners)
- **Utilities**: `internal/cmdutil/` (RunWithSpinner, PrintResult, flags)
- **shelly-go components**: `/db/appdata/shelly-go/gen2/components/`

---

> ⚠️ **CRITICAL:** NEVER mark any task as "optional", "nice to have", or "future improvement". All tasks are REQUIRED. Do not defer work without explicit user approval.

> **Session context:** See [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md) for historical implementation details.

## Project Overview

**Repository:** `github.com/tj-smith47/shelly-cli`
**Command:** `shelly`
**Go Version:** 1.25.5 (Released December 4, 2024)
**Library Dependency:** `github.com/tj-smith47/shelly-go` (comprehensive Shelly device library)

### Goals

Create a production-ready, open-source Cobra CLI that:
1. Provides intuitive access to the full Shelly API via the shelly-go library
2. Offers TUI dashboards inspired by k9s and gh-dash
3. Includes convenience features (aliases, auto-completions, plugins)
4. Has full theme support via bubbletint
5. Achieves 90%+ test coverage
6. Includes comprehensive documentation and examples

---

## Current Status

**Last Updated:** 2025-12-14 | **Current Phase:** 20.3 - Debug Commands | **shelly-go:** v0.1.6

### Phase Progress

| Phase | Name | Status |
|-------|------|--------|
| 0.1-0.6 | Architecture Refactoring | ✅ Complete |
| 0.7 | Test Coverage Foundation | ⏳ Deferred to Phase 29 |
| 1-2 | Project Foundation & Infrastructure | ✅ Complete |
| 3-4 | Device Management & Control | 🟡 Commands done, completions TBD |
| 5 | Configuration Commands | ✅ Complete |
| 6 | Firmware Commands | ✅ Complete |
| 7 | Script Commands | ✅ Complete |
| 8 | Schedule Commands | ✅ Complete |
| 9 | Cloud Commands | ✅ Complete |
| 10 | Backup & Restore | ✅ Complete |
| 11.1-11.3 | Monitoring (status, energy, power) | ✅ Complete |
| 11.4 | Energy Dashboard | ✅ Complete |
| 11.5 | Metrics Export | ✅ Complete |
| 12 | Alias System | ✅ Complete |
| 13 | Plugin System | ✅ Complete |
| 14 | TUI Dashboard | ✅ Complete |
| 15 | Theme System | ✅ Complete |
| 16 | Shell Completions | ✅ Complete |
| 17 | Update Command | ✅ Complete |
| 18-25 | Advanced Features | ⏳ Pending |
| 26-27 | Documentation & Examples | ⏳ Pending |
| 28 | Testing (90%+ coverage) | ⏳ Pending |
| 29 | Innovative Commands (82 new) | ⏳ Pending |
| 30 | Polish & Release (FINAL) | ⏳ Pending |

**Test Coverage:** ~19% (target: 90%+ in Phase 28)

> **Note:** Session-specific implementation notes moved to [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md)

---

## Lessons Learned

> See [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md) for full details.

**Key lessons:**
1. **Verify shelly-go API** - Check component files before planning features
2. **Use concrete types** - Never `interface{}` for API data
3. **Never suppress errors** - Use `ios.DebugErr()` instead of `//nolint:errcheck`
4. **Check existing patterns** - Search cmdutil before reimplementing
5. **Verify before marking complete** - Build, lint, test, manual verify
6. **No bulk formatting** - Only format files YOU changed

---

## Global Flags & Output Control

> ⚠️ **IMPORTANT:** All commands MUST support these global flags. Theming must be consistent across all output.

### Global Flags (Available on ALL Commands)

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | `text` | Output format: `text`, `json`, `yaml`, `table`, `csv` |
| `--no-color` | | bool | auto | Disable colored output (also respects `NO_COLOR` env var) |
| `--plain` | | bool | false | Plain output: no colors, no spinners, no progress bars |
| `--quiet` | `-q` | bool | false | Suppress non-essential output (only errors and direct results) |
| `--verbose` | `-v` | bool | false | Verbose output with extra details |
| `--debug` | | bool | false | Debug mode with request/response logging |
| `--timeout` | | duration | 30s | Request timeout for device operations |
| `--config` | | string | auto | Config file path (default: `~/.config/shelly/config.yaml`) |
| `--device` | `-d` | string | | Default device for commands that require one |

### Output Format Examples

```bash
# JSON output (pipeable)
shelly device list -o json | jq '.[] | select(.online)'

# CSV for spreadsheets
shelly energy history kitchen -o csv > energy.csv

# YAML for config management
shelly config export kitchen -o yaml > kitchen.yaml

# Plain for scripting (no ANSI codes)
shelly switch status kitchen --plain

# Quiet for cron jobs
shelly batch on all-lights -q
```

### Theming Requirements

All output MUST use themed colors via IOStreams:
- **Keys/labels**: Use `ios.ColorScheme().Cyan()` for property names
- **Values**: Use theme-appropriate colors (green=success, red=error, yellow=warning)
- **Tables**: Headers themed, alternating row colors optional
- **Status indicators**: Use consistent symbols (✓, ✗, ●, ○, ◐)

**Anti-pattern:** Never use `fmt.Println()` or hardcoded ANSI codes in commands.

### Pipeable Commands

These commands MUST produce clean, parseable output when piped:

| Command | Pipeable Output | Example Use |
|---------|-----------------|-------------|
| `device list` | Device names/IPs | `shelly device list -o json \| jq -r '.[].name'` |
| `device info` | Device JSON | `shelly device info kitchen -o json \| jq '.firmware'` |
| `switch status` | State (on/off/json) | `shelly switch status kitchen --plain` returns `on` or `off` |
| `energy status` | Power/energy values | `shelly energy status -o json \| jq '.power'` |
| `energy history` | CSV/JSON data | `shelly energy history kitchen -o csv >> log.csv` |
| `discover scan` | Discovered IPs | `shelly discover scan -o json \| jq -r '.[].ip'` |
| `config export` | Config JSON/YAML | `shelly config export kitchen -o yaml > backup.yaml` |
| `backup create` | Backup file path | `shelly backup create kitchen -o plain` returns filename |
| `script get` | Script source code | `shelly script get kitchen 1 > script.js` |

**Requirements for pipeable commands:**
1. When stdout is not a TTY, automatically disable colors and spinners
2. Support `--plain` to force machine-readable output even in TTY
3. JSON output must be valid, parseable JSON (no extra text)
4. Exit codes must reflect success (0) or failure (non-zero)

---

## Phase 0: Architecture Refactoring (PREREQUISITE)

> **IMPORTANT:** This phase must be completed before Phase 5. It addresses critical code quality issues identified during the architecture audit against gh, kubectl, docker, jira-cli, gh-dash, and k9s.

### 0.1 IOStreams Package (gh pattern) ✅
Create unified I/O handling following gh CLI's iostreams pattern:

- [x] Create `internal/iostreams/iostreams.go`:
  ```go
  type IOStreams struct {
      In       io.Reader
      Out      io.Writer
      ErrOut   io.Writer

      // Terminal detection
      IsStdinTTY   bool
      IsStdoutTTY  bool
      IsStderrTTY  bool

      // Color management
      colorEnabled bool

      // Progress indicators
      progressIndicator *spinner.Spinner
  }

  func (s *IOStreams) StartProgress(msg string)
  func (s *IOStreams) StopProgress()
  func (s *IOStreams) ColorEnabled() bool
  ```
- [x] Create `internal/iostreams/color.go` - Theme-aware color utilities
- [x] Create `internal/iostreams/progress.go` - Spinner/progress indicators
- [x] Create `internal/iostreams/multiwriter.go` - Docker-style multi-line output (see `docs/architecture.md`)
- [x] Add comprehensive tests (92.3% coverage)

### 0.2 Command Utilities Package (gh/kubectl pattern) ✅
Create shared utilities for commands (NOT under `internal/cmd/` - only commands live there):

- [x] Create `internal/cmdutil/factory.go`:
  ```go
  // Factory provides dependencies to commands (gh pattern)
  type Factory struct {
      IOStreams    func() *IOStreams
      Config       func() (*config.Config, error)
      ShellyClient func(device string) (*client.Client, error)
      Resolver     func() shelly.DeviceResolver
  }
  ```
- [x] Create `internal/cmdutil/runner.go`:
  ```go
  // Generic command runner (kubectl pattern)
  type ComponentAction func(ctx context.Context, svc *shelly.Service, device string, id int) error

  func RunWithSpinner(ctx context.Context, ios *IOStreams, msg string, action func(context.Context) error) error
  func RunBatch(ctx context.Context, ios *IOStreams, targets []string, action ComponentAction) error
  ```
- [x] Create `internal/cmdutil/flags.go`:
  ```go
  func AddComponentIDFlag(cmd *cobra.Command, target *int, name string)
  func AddOutputFlag(cmd *cobra.Command)
  func AddTimeoutFlag(cmd *cobra.Command, target *time.Duration)
  ```
- [x] Create `internal/cmdutil/output.go`:
  ```go
  func PrintResult(ios *IOStreams, format string, data any, tableFn func(any)) error
  ```
- [x] Add comprehensive tests (92.3% coverage)

### 0.3 Package Consolidation ✅
Consolidated duplicate packages into `internal/iostreams/`:
- [x] All message functions consolidated to `iostreams/color.go`
- [x] All spinner functions consolidated to `iostreams/progress.go`
- [x] All prompt functions consolidated to `iostreams/prompt.go`
- [x] Debug functions added to `iostreams/debug.go`
- [x] Deleted `internal/ui/` directory
- [x] Deleted `internal/output/messages.go`, `spinner.go`, `prompt.go`
- [x] Kept `internal/output/format.go` and `table.go` (formatters)
- [x] Fixed batch/command and scene/activate to use `errgroup` instead of `WaitGroup`

### 0.4 Context Propagation ✅
Fix context handling throughout codebase:

- [x] Update `internal/cmd/root.go` to setup cancellation-aware context
- [x] Update all 41+ command run() functions to accept context parameter
- [x] Use `cmd.Context()` instead of `context.Background()`
- [x] Ensure Ctrl+C cancels in-flight HTTP requests (inherent via context propagation)

### 0.5 Concurrency Patterns ✅
Replace manual patterns with errgroup and add multi-writer output:

- [x] Refactor `internal/cmd/batch/on/on.go` to use errgroup + MultiWriter (already used cmdutil.RunBatch)
- [x] Refactor `internal/cmd/batch/off/off.go` to use errgroup + MultiWriter (already used cmdutil.RunBatch)
- [x] Refactor `internal/cmd/batch/toggle/toggle.go` to use errgroup + MultiWriter (already used cmdutil.RunBatch)
- [x] Refactor `internal/cmd/batch/command/command.go` to use errgroup + MultiWriter
- [x] Refactor `internal/cmd/scene/activate/activate.go` to use errgroup + MultiWriter
- [x] Refactor `internal/cmd/discover/scan/scan.go` to use MultiWriter with progress callback
- [x] Refactor `internal/plugins/loader.go` to parallelize version detection

**Current pattern (to replace):**
```go
// internal/cmd/batch/on/on.go - Current verbose pattern
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
```

**Target pattern:**
```go
// Using errgroup + MultiWriter
mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())
for _, t := range targets {
    mw.AddLine(t, "pending")
}
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(concurrent)
for _, t := range targets {
    target := t
    g.Go(func() error {
        mw.UpdateLine(target, StatusRunning, "turning on...")
        if err := svc.SwitchOn(ctx, target, switchID); err != nil {
            mw.UpdateLine(target, StatusError, err.Error())
            return nil
        }
        mw.UpdateLine(target, StatusSuccess, "on")
        return nil
    })
}
return g.Wait()
```

### 0.6 Command Refactoring (DRY)
Refactor all duplicate command patterns to use cmdutil:

**Problem:** 38+ nearly identical `run()` functions across command files. Example from `internal/cmd/light/on/on.go`:
```go
// This pattern repeats 38+ times with only method/message changes:
func run(device string, componentID int) error {
    ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
    defer cancel()
    svc := shelly.NewService()
    spin := ui.NewSpinner("Turning light on...")  // Only these 3 things change
    spin.Start()
    err := svc.LightOn(ctx, device, componentID)  // Method name
    spin.Stop()
    if err != nil {
        return fmt.Errorf("failed to turn light on: %w", err)  // Error message
    }
    ui.Success("Light %d turned on", componentID)  // Success message
    return nil
}
```

**Solution:** Use `cmdutil.RunWithSpinner`:
```go
func run(ctx context.Context, f *cmdutil.Factory, device string, lightID int) error {
    ios := f.IOStreams()
    svc := shelly.NewService()
    return cmdutil.RunWithSpinner(ctx, ios, "Turning light on...", func(ctx context.Context) error {
        if err := svc.LightOn(ctx, device, lightID); err != nil {
            return fmt.Errorf("failed to turn light on: %w", err)
        }
        fmt.Fprintf(ios.Out, "Light %d turned on\n", lightID)
        return nil
    })
}
```

**Files to refactor:**
- [x] Refactor all `on.go` commands (switch, light, rgb) to use `cmdutil.RunSimple`
- [x] Refactor all `off.go` commands to use `cmdutil.RunSimple`
- [x] Refactor all `toggle.go` commands to use `cmdutil.RunWithSpinner`
- [x] Refactor all `status.go` commands to use `cmdutil.RunStatus`/`cmdutil.PrintResult`
- [x] Refactor component `list.go` commands to use `cmdutil.RunList`/`cmdutil.PrintListResult`
- [x] Apply consistent flag patterns via `cmdutil.AddComponentIDFlag`

### 0.7 Test Coverage Foundation
Establish testing patterns for new packages:

- [ ] Add comprehensive tests for `internal/iostreams/` (target: 90%+) - **currently 79.1%** - DEFERRED to Phase 25
  - Note: Gap is primarily interactive prompt functions that require terminal input
- [x] Add comprehensive tests for `internal/cmdutil/` (target: 90%+) - **93.6% achieved**
- [x] Add table-driven tests for runner patterns (RunStatus, RunList, PrintResult)
- [x] Update existing command tests to use new patterns (all tests pass with refactored code)

**Note:** Remaining test coverage work deferred to Phase 25 (Testing). Proceeding with functionality.

---

## Phase 1: Project Foundation

### 1.1 Repository Setup
- [x] Initialize Go module: `go mod init github.com/tj-smith47/shelly-cli`
- [x] Configure `go.mod` with Go 1.25.5 minimum version
- [x] Add shelly-go library dependency: `github.com/tj-smith47/shelly-go` v0.1.3
- [x] Create `.gitignore` with Go-specific patterns and build artifacts
- [x] Create `LICENSE` (MIT)
- [x] Create `CONTRIBUTING.md` with contribution guidelines
- [x] Create `SECURITY.md` with security policy
- [x] Create `CODE_OF_CONDUCT.md`

### 1.2 Directory Structure
```
shelly-cli/
├── cmd/
│   └── shelly/
│       └── main.go                 # Entry point
├── internal/
│   ├── cmd/                        # CLI commands ONLY (subdirectory-per-command)
│   │   ├── root.go                 # Root command
│   │   ├── switchcmd/              # Switch control (on/, off/, toggle/, status/, list/)
│   │   ├── cover/                  # Cover control (open/, closecmd/, stop/, position/, status/, list/)
│   │   ├── light/                  # Light control (on/, off/, toggle/, set/, status/, list/)
│   │   ├── rgb/                    # RGB control (on/, off/, toggle/, set/, status/, list/)
│   │   ├── input/                  # Input status (list/, status/, trigger/)
│   │   ├── discover/               # Discovery (mdns/, ble/, coiot/, scan/)
│   │   ├── device/                 # Device management (list/, add/, remove/, info/, etc.)
│   │   ├── group/                  # Group management (list/, create/, delete/, add/, etc.)
│   │   ├── batch/                  # Batch operations (on/, off/, toggle/, command/)
│   │   ├── scene/                  # Scene management (list/, create/, activate/, etc.)
│   │   ├── config/                 # Config commands (get/, set/, export/, import/)
│   │   ├── firmware/               # Firmware commands (check/, update/, rollback/)
│   │   ├── script/                 # Script commands (list/, create/, start/, stop/)
│   │   ├── schedule/               # Schedule commands (list/, create/, enable/)
│   │   ├── cloud/                  # Cloud commands (login/, devices/, events/)
│   │   ├── wifi/                   # WiFi configuration (status/, scan/, set/)
│   │   ├── mqtt/                   # MQTT configuration (status/, set/, disable/)
│   │   ├── webhook/                # Webhook management (list/, create/, delete/)
│   │   ├── kvs/                    # KVS storage (list/, get/, set/, delete/)
│   │   ├── energy/                 # Energy monitoring (status/, history/, export/)
│   │   ├── bthome/                 # BTHome devices (list/, add/, remove/, status/)
│   │   ├── zigbee/                 # Zigbee gateway (status/, discover/, pair/)
│   │   └── matter/                 # Matter protocol (status/, enable/, code/)
│   │   # NOTE: NO shared utilities here - only command definitions!
│   │
│   ├── iostreams/                  # Unified I/O handling (gh pattern) [NEW]
│   │   ├── iostreams.go            # IOStreams struct and methods
│   │   ├── color.go                # Theme-aware color utilities
│   │   └── progress.go             # Spinner/progress indicators
│   │
│   ├── cmdutil/                    # Command utilities (gh/kubectl pattern) [NEW]
│   │   ├── factory.go              # Factory for dependency injection
│   │   ├── runner.go               # Generic command runners
│   │   ├── flags.go                # Shared flag definitions
│   │   └── output.go               # Output format routing
│   │
│   ├── model/                      # Core domain types
│   │   ├── device.go               # Device types
│   │   ├── component.go            # Component types (Switch, Cover, Light, RGB status/config)
│   │   └── errors.go               # Domain errors
│   ├── client/                     # SDK wrapper for device communication
│   │   ├── client.go               # Connection management
│   │   ├── switch.go               # Switch operations
│   │   ├── cover.go                # Cover operations
│   │   ├── light.go                # Light operations
│   │   └── rgb.go                  # RGB operations
│   ├── shelly/                     # Business logic service layer
│   │   ├── shelly.go               # Service with connection management
│   │   ├── resolver.go             # Device resolution (name/IP/alias)
│   │   ├── switch.go               # Switch business logic
│   │   ├── cover.go                # Cover business logic
│   │   ├── light.go                # Light business logic
│   │   └── rgb.go                  # RGB business logic
│   ├── helpers/                    # Utility functions
│   │   ├── device.go               # Device helpers
│   │   └── display.go              # Display helpers
│   ├── tui/                        # TUI dashboard (gh-dash/BubbleTea pattern)
│   │   ├── app.go                  # Main app model (Elm Architecture)
│   │   ├── keys.go                 # Keyboard bindings (vim-style)
│   │   └── components/             # Reusable TUI components
│   │       ├── devicelist/         # Device list table
│   │       ├── devicedetail/       # Device detail panel
│   │       ├── monitor/            # Real-time monitoring
│   │       ├── search/             # Filter/search bar
│   │       └── help/               # Help overlay
│   ├── config/
│   │   ├── config.go               # Viper configuration
│   │   ├── aliases.go              # Alias management
│   │   ├── devices.go              # Device registry
│   │   └── scenes.go               # Scene management
│   ├── output/                     # Structured output formatters
│   │   ├── format.go               # Output formatters (json, yaml, table, text)
│   │   └── table.go                # Table rendering
│   ├── plugins/
│   │   ├── loader.go               # Plugin discovery and loading
│   │   ├── executor.go             # Plugin execution
│   │   └── registry.go             # Plugin registry
│   ├── theme/
│   │   └── theme.go                # bubbletint theme integration
│   ├── testutil/
│   │   └── testutil.go             # Test utilities
│   └── version/
│       └── version.go              # Version info (ldflags)
├── pkg/
│   └── api/                        # Public API for plugins
├── docs/
│   ├── commands/                   # Command documentation
│   ├── architecture.md             # Architecture patterns reference [NEW]
│   ├── configuration.md            # Config file reference
│   ├── plugins.md                  # Plugin development guide
│   └── themes.md                   # Theming guide
├── examples/
│   ├── aliases/                    # Example alias configurations
│   ├── scripts/                    # Example automation scripts
│   └── plugins/                    # Example plugins
├── completions/
│   ├── shelly.bash
│   ├── shelly.zsh
│   ├── shelly.fish
│   └── shelly.ps1
├── themes/
│   └── default.yaml                # Default theme configuration
├── .github/
│   └── workflows/
│       ├── ci.yml                  # CI pipeline
│       ├── auto-tag.yml            # Auto-tagging on merge
│       └── release.yml             # GoReleaser workflow
├── .goreleaser.yaml
├── .golangci.yml
├── Makefile
├── README.md
├── CHANGELOG.md
└── PLAN.md                         # This file
```

- [x] Create all directories per structure above
- [x] Create placeholder files with package declarations

### 1.3 GitHub Actions CI/CD
- [x] Create `.github/workflows/ci.yml` (badges branch coverage, not Codecov)
- [x] Create `.github/workflows/auto-tag.yml`
- [x] Create `.github/workflows/release.yml`

### 1.4 GoReleaser Configuration
- [x] Create `.goreleaser.yaml`

### 1.5 Linting Configuration
- [x] Create `.golangci.yml` with strict settings

### 1.6 Makefile
- [x] Create `Makefile` with all targets

---

## Phase 2: Core CLI Infrastructure

### 2.1 Entry Point & Version
- [x] Create `cmd/shelly/main.go`
- [x] Create `internal/version/version.go`

### 2.2 Root Command
- [x] Create `internal/cli/root.go`
- [x] Implement config file loading via Viper

### 2.3 Configuration System
- [x] Create `internal/config/config.go`
- [x] Config structure:
  ```go
  type Config struct {
      // Global settings
      Output       string            // default output format
      Color        bool              // enable color output
      Theme        string            // bubbletint theme name
      APIMode      string            // "local" (default) | "cloud" | "auto"

      // Device discovery
      Discovery    DiscoveryConfig

      // Cloud settings (used when APIMode is "cloud" or "auto" with local fallback)
      Cloud        CloudConfig

      // Aliases
      Aliases      map[string]Alias

      // Device registry
      Devices      map[string]Device

      // Plugin settings
      Plugins      PluginsConfig
  }
  ```
  - `APIMode`:
    - `local` (default): Direct device communication via HTTP/WebSocket
    - `cloud`: Control via Shelly Cloud API
    - `auto`: Try local first, fall back to cloud if device unreachable
- [x] Create `internal/config/aliases.go`:
  - Alias struct (Name, Command, Shell bool)
  - Add, Remove, List, Get functions
  - Validate alias names (no conflicts with built-in commands)
- [x] Create `internal/config/devices.go`:
  - Device struct (Name, Address, Auth, Generation, etc.)
  - Register, Unregister, List, Get functions
  - Device groups support

### 2.4 Output Formatting
- [x] Create `internal/output/format.go`:
  - Formatter interface
  - JSON, YAML, Table, Text implementations
  - Auto-detect format from config/flags
- [x] Create `internal/output/table.go`:
  - Table renderer using lipgloss
  - Column alignment and wrapping
  - Header styling
  - Row alternation colors
- [x] Create `internal/output/color.go` (merged into messages.go):
  - Theme-aware color utilities
  - Status colors (success, warning, error, info)
  - Device state colors (online, offline, updating)

---

## Phase 3: Device Management Commands

### 3.1 Discovery Commands
- [x] `shelly discover` - Discover devices on network
  - Flags: --timeout, --network (subnet), --mdns, --ble, --coiot
  - Output discovered devices with name, IP, generation, type
  - Progress indicator during discovery
  - Filter by generation, type
- [x] `shelly discover mdns` - mDNS discovery only
- [x] `shelly discover ble` - BLE discovery only
  - Flags: --timeout, --bthome (include BTHome sensor broadcasts), --filter (device name prefix)
  - Shows BLE discovered devices with RSSI, connectable status, BTHome indicator
  - Gracefully handles BLE not supported on system
- [x] `shelly discover coiot` - CoIoT discovery (Gen1)
- [x] `shelly discover scan <subnet>` - Subnet scan (HTTP probing)

### 3.2 Device Registry Commands
- [x] `shelly device list` - List registered devices
  - Flags: --all, --online, --offline, --generation, --type
  - Table output with status indicators
- [x] `shelly device add <name> <address>` - Add device to registry
  - Flags: --auth, --generation (auto-detect if omitted)
  - Validate connectivity before adding
- [x] `shelly device remove <name>` - Remove device from registry
  - Confirmation prompt (--yes to skip)
- [x] `shelly device rename <old> <new>` - Rename device
- [x] `shelly device info <name|address>` - Show device details
  - Device info, config, status
  - Component list
  - Network info
- [x] `shelly device status <name|address>` - Show device status
  - Real-time status with auto-refresh option
- [x] `shelly device ping <name|address>` - Check device connectivity
- [x] `shelly device reboot <name|address>` - Reboot device
- [x] `shelly device factory-reset <name|address>` - Factory reset
  - Strong confirmation required (--yes --confirm)

### 3.3 Device Group Commands
- [x] `shelly group list` - List device groups
- [x] `shelly group create <name>` - Create group
- [x] `shelly group delete <name>` - Delete group
- [x] `shelly group add <group> <device>` - Add device to group
- [x] `shelly group remove <group> <device>` - Remove device from group
- [x] `shelly group members <name>` - List group members

---

## Phase 4: Device Control Commands

### 4.1 Switch Control (alias: sw)
- [x] `shelly switch list [device]` - List switches
- [x] `shelly switch on <device> [id]` - Turn switch on
  - Flags: --timer (auto-off after N seconds)
- [x] `shelly switch off <device> [id]` - Turn switch off
- [x] `shelly switch toggle <device> [id]` - Toggle switch
- [x] `shelly switch status <device> [id]` - Show switch status (alias: st)

### 4.2 Cover/Roller Control (aliases: cv, roller)
- [x] `shelly cover list [device]` - List covers
- [x] `shelly cover open <device> [id]` - Open cover
- [x] `shelly cover close <device> [id]` - Close cover
- [x] `shelly cover stop <device> [id]` - Stop cover
- [x] `shelly cover position <device> [id] <percent>` - Set position (aliases: pos, set)
- [x] `shelly cover calibrate <device> [id]` - Start calibration
- [x] `shelly cover status <device> [id]` - Show cover status (alias: st)

### 4.3 Light Control (aliases: lt, dim)
- [x] `shelly light list [device]` - List lights
- [x] `shelly light on <device> [id]` - Turn light on
  - Flags: --brightness, --transition
- [x] `shelly light off <device> [id]` - Turn light off
- [x] `shelly light toggle <device> [id]` - Toggle light
- [x] `shelly light set <device> [id]` - Set light parameters
  - Flags: --brightness, --transition
- [x] `shelly light status <device> [id]` - Show light status (alias: st)

### 4.4 RGB/RGBW Control (alias: color)
- [x] `shelly rgb on <device> [id]` - Turn RGB on
- [x] `shelly rgb off <device> [id]` - Turn RGB off
- [x] `shelly rgb set <device> [id>` - Set RGB color
  - Flags: --color (R,G,B), --hex, --brightness, --transition
- [x] `shelly rgb status <device> [id]` - Show RGB status (alias: st)
- [x] `shelly rgb toggle <device> [id]` - Toggle RGB
- [x] `shelly rgb list <device>` - List RGB components

### 4.5 Input Status
- [x] `shelly input list [device]` - List inputs
- [x] `shelly input status <device> [id]` - Show input status
- [x] `shelly input trigger <device> [id]` - Manually trigger input event

### 4.6 Batch Operations
- [x] `shelly batch on <devices...>` - Turn on multiple devices
- [x] `shelly batch off <devices...>` - Turn off multiple devices
- [x] `shelly batch toggle <devices...>` - Toggle multiple devices
- [x] `shelly batch command <command> <devices...>` - Run command on multiple devices
  - Flags: --parallel, --timeout, --continue-on-error

### 4.7 Scene Management
- [x] `shelly scene list` - List saved scenes
- [x] `shelly scene create <name>` - Create scene from current state
  - Flags: --devices (select which devices)
- [x] `shelly scene delete <name>` - Delete scene
- [x] `shelly scene activate <name>` - Activate scene
- [x] `shelly scene show <name>` - Show scene details
- [x] `shelly scene export <name> <file>` - Export scene to file
- [x] `shelly scene import <file>` - Import scene from file

---

## Phase 5: Configuration Commands

### 5.1 Device Configuration ✅
- [x] `shelly config get <device>` - Get device config (all components)
- [x] `shelly config get <device> <component>` - Get component config
- [x] `shelly config set <device> <component> <key>=<value>...` - Set config
- [x] `shelly config diff <device> <file>` - Compare config with file
- [x] `shelly config export <device> <file>` - Export config to file
- [x] `shelly config import <device> <file>` - Import config from file
  - Flags: --dry-run, --merge, --overwrite
- [x] `shelly config reset <device> [component]` - Reset to defaults

*Implementation details: See [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md#phase-5-configuration-commands)*

### 5.2 Network Configuration ✅
- [x] `shelly wifi status <device>` - Show WiFi status
- [x] `shelly wifi scan <device>` - Scan for networks
- [x] `shelly wifi set <device>` - Configure WiFi
  - Flags: --ssid, --password, --static-ip, --gateway, --dns
- [x] `shelly wifi ap <device>` - Configure AP mode (--clients to list connected clients)
- [x] `shelly ethernet status <device>` - Show ethernet status (Pro devices)
- [x] `shelly ethernet set <device>` - Configure ethernet


### 5.3 Cloud Configuration ✅
- [x] `shelly cloud status <device>` - Show cloud status
- [x] `shelly cloud enable <device>` - Enable cloud connection
- [x] `shelly cloud disable <device>` - Disable cloud connection


### 5.4 Auth Configuration ✅
- [x] `shelly auth status <device>` - Show auth status
- [x] `shelly auth set <device>` - Set auth credentials
  - Flags: --user, --password
- [x] `shelly auth disable <device>` - Disable auth


### 5.5 MQTT Configuration ✅
- [x] `shelly mqtt status <device>` - Show MQTT status
- [x] `shelly mqtt set <device>` - Configure MQTT
  - Flags: --server, --user, --password, --topic-prefix
- [x] `shelly mqtt disable <device>` - Disable MQTT


### 5.6 Webhook Configuration ✅
- [x] `shelly webhook list <device>` - List webhooks
- [x] `shelly webhook create <device>` - Create webhook
  - Flags: --event, --url, --name, --disable, --cid
- [x] `shelly webhook delete <device> <id>` - Delete webhook
- [x] `shelly webhook update <device> <id>` - Update webhook
  - Flags: --event, --url, --name, --enable, --disable


---

## Phase 6: Firmware Commands ✅

### 6.1 Firmware Status ✅
- [x] `shelly firmware check [device|--all]` - Check for updates
  - Show current version, available version, release notes summary
- [x] `shelly firmware status <device>` - Show firmware status

### 6.2 Firmware Updates ✅
- [x] `shelly firmware update <device>` - Update device firmware
  - Flags: --beta, --url (custom firmware), --yes (skip confirmation)
  - Progress indicator
  - Wait for device to come back online
- [x] `shelly firmware update --all` - Update all devices
  - Flags: --parallel, --staged (percentage-based rollout)
- [x] `shelly firmware rollback <device>` - Rollback to previous version

### 6.3 Firmware Download ✅
- [x] `shelly firmware download <device>` - Download firmware file
  - Flags: --output, --latest, --beta

*Implementation details: See [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md#phase-6-firmware-commands)*

---

## Phase 7: Script Commands (Gen2+) ✅

### 7.1 Script Management ✅
- [x] `shelly script list <device>` - List scripts (alias: ls)
- [x] `shelly script get <device> <id>` - Get script code (alias: code)
  - Flags: --status (show status instead of code)
- [x] `shelly script create <device>` - Create script (alias: new)
  - Flags: --code (inline), --file (from file), --enable, --name
- [x] `shelly script update <device> <id>` - Update script code (alias: up)
  - Flags: --code, --file, --name, --append, --enable
- [x] `shelly script delete <device> <id>` - Delete script (aliases: del, rm)
  - Flags: --yes (skip confirmation)
- [x] `shelly script start <device> <id>` - Start script (alias: run)
- [x] `shelly script stop <device> <id>` - Stop script
- [x] `shelly script eval <device> <id> <code>` - Evaluate code snippet (alias: exec)
- [x] `shelly script upload <device> <id> <file>` - Upload script from file (alias: put)
  - Flags: --append
- [x] `shelly script download <device> <id> <file>` - Download script to file (alias: save)

*Implementation details: See [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md#phase-7-script-commands-gen2)*

### 7.2 Script Library
- [ ] `shelly script-lib list` - List available script templates
- [ ] `shelly script-lib show <name>` - Show script template
- [ ] `shelly script-lib install <device> <name>` - Install template to device
  - Flags: --configure (interactive configuration)

---

## Phase 8: Schedule Commands ✅

### 8.1 Schedule Management ✅
- [x] `shelly schedule list <device>` - List schedules (alias: ls)
- [x] `shelly schedule create <device>` - Create schedule (alias: new)
  - Flags: --timespec (required), --calls (required), --enable
- [x] `shelly schedule update <device> <id>` - Update schedule (alias: up)
  - Flags: --timespec, --calls, --enable, --disable
- [x] `shelly schedule delete <device> <id>` - Delete schedule (aliases: del, rm)
  - Flags: --yes
- [x] `shelly schedule delete-all <device>` - Delete all schedules (alias: clear)
  - Flags: --yes
- [x] `shelly schedule enable <device> <id>` - Enable schedule
- [x] `shelly schedule disable <device> <id>` - Disable schedule

*Implementation details: See [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md#phase-8-schedule-commands)*

### 8.2 Schedule Helpers
- [ ] `shelly schedule sunrise <device>` - Create sunrise-based schedule
  - Flags: --offset, --action, --days
- [ ] `shelly schedule sunset <device>` - Create sunset-based schedule
- [ ] `shelly schedule timer <device>` - Create timer-based schedule

---

## Phase 9: Cloud Commands ✅

### 9.1 Cloud Authentication ✅
- [x] `shelly cloud login` - Authenticate with Shelly Cloud
  - Email/password authentication (interactive or via flags)
  - Saves token and server URL to config
- [x] `shelly cloud logout` - Clear cloud credentials
- [x] `shelly cloud auth-status` - Show cloud auth status (alias: whoami)
- [x] `shelly cloud token` - Show/manage access token

### 9.2 Cloud Device Management ✅
- [x] `shelly cloud devices` - List cloud-registered devices (aliases: ls, list)
- [x] `shelly cloud device <id>` - Show cloud device details (alias: get)
  - Flags: --status (show full status JSON)
- [x] `shelly cloud control <id> <action>` - Control via cloud
  - Actions: on, off, toggle, open, close, stop, position=N, brightness=N, light-on, light-off, light-toggle
  - Flags: --channel
- [x] `shelly cloud events` - Subscribe to real-time events (aliases: watch, subscribe)
  - WebSocket connection to Shelly Cloud
  - Output events as they arrive (text or JSON)
  - Flags: --device (filter), --event (filter), --format (text/json), --raw

*Implementation details: See [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md#phase-9-cloud-commands)*

---

## Phase 10: Backup & Restore Commands ✅

*Implementation details: See [docs/IMPLEMENTATION-NOTES.md](docs/IMPLEMENTATION-NOTES.md#phase-10-backup--restore)*

### 10.1 Backup Operations ✅
- [x] `shelly backup create <device> [file]` - Create device backup
  - Includes: config, scripts, schedules, webhooks, KVS
  - Flags: --encrypt (password-protected via AES-256-GCM)
- [x] `shelly backup restore <device> <file>` - Restore from backup
  - Flags: --dry-run, --skip-network, --decrypt
- [x] `shelly backup list` - List saved backups
- [x] `shelly backup export --all <directory>` - Backup all devices

### 10.2 Migration ✅
- [x] `shelly migrate <source> <target>` - Migrate config between devices
  - Flags: --dry-run, --force (different device types)
- [x] `shelly migrate validate <backup>` - Validate backup file
- [x] `shelly migrate diff <device> <backup>` - Show differences

---

## Phase 11: Monitoring Commands

> **Note:** Energy monitoring uses shelly-go `gen2/components/em.go`, `em1.go`, `pm.go`, `pm1.go` components.

### 11.1 Real-time Monitoring
- [x] `shelly monitor <device>` - Real-time status monitoring
  - Auto-refresh with configurable interval
  - Color-coded status changes
- [x] `shelly monitor power <device>` - Monitor power consumption
- [x] `shelly monitor events <device>` - Monitor device events
  - WebSocket subscription via shelly-go events package
- [x] `shelly monitor all` - Monitor all registered devices

### 11.2 Energy Monitoring (EM/EM1 Components) ✅
- [x] `shelly energy list <device>` - List energy meters (EM components)
  - Fully implemented and linted in `internal/cmd/energy/list/`
  - Lists both EM (3-phase) and EM1 (single-phase) components
  - Uses existing service methods from `internal/shelly/monitoring.go`
- [x] `shelly energy status <device> [id]` - Current power/energy status
  - Shows: voltage, current, power, energy, power factor, frequency
  - Fully implemented in `internal/cmd/energy/status/status.go` with auto-detection
  - Supports both EM (3-phase) and EM1 (single-phase) components
- [x] `shelly energy history <device> [id]` - Energy history (aliases: hist, events)
  - Flags: `--period` (hour, day, week, month), `--from`, `--to` (date range), `--limit`
  - Auto-detects EM vs EM1 component type
  - Calculates total kWh consumption from historical data
  - Comprehensive tests for time parsing and energy calculations
  - **Implementation:** Uses EMData/EM1Data components from shelly-go v0.1.5
- [x] `shelly energy export <device> [id]` - Export energy data (aliases: exp, dump)
  - Flags: `--format` (csv, json, yaml), `--output`, `--period`, `--from`, `--to`
  - Supports structured export (JSON/YAML) and CSV via HTTP endpoints
  - **Implementation:** Service methods with CSV URL generation (optimized to avoid unnecessary connections)
- [x] `shelly energy reset <device> [id]` - Reset energy counters
  - Fully implemented in `internal/cmd/energy/reset/reset.go`
  - Supports both EM and EM1 components with confirmation prompt

### 11.3 Power Monitoring (PM/PM1 Components) ✅
- [x] `shelly power list <device>` - List power meters (PM components)
- [x] `shelly power status <device> [id]` - Current power status
  - Shows: power (W), energy (Wh), voltage, current, frequency
  - Includes accumulated energy counters and return energy

**Note:** PM/PM1 components do not have historical data storage (unlike EM/EM1 which have EMData/EM1Data). PM also has no calibration method - calibration is only available for Cover and Thermostat components. The plan items for `power history` and `power calibrate` were incorrect assumptions.

### 11.4 Aggregated Energy Dashboard ✅
- [x] `shelly energy dashboard` - Summary of all energy meters (aliases: dash, summary)
  - Total power consumption
  - Per-device breakdown
  - Cost estimation via `--cost` and `--currency` flags
  - Concurrent device queries with errgroup
  - Supports EM, EM1, PM, PM1 components
- [x] `shelly energy compare` - Compare energy usage (aliases: cmp, diff)
  - Flags: `--devices`, `--period`, `--from`, `--to`
  - Historical energy comparison using EMData/EM1Data
  - Visual bar chart of energy distribution
  - Per-device percentage breakdown

### 11.5 Metrics Export ✅
- [x] `shelly metrics prometheus` - Start Prometheus exporter (aliases: prom)
  - Flags: `--port` (default 9090), `--devices`, `--interval` (default 15s)
  - HTTP server at /metrics with Prometheus format
  - Exports: shelly_power_watts, shelly_voltage_volts, shelly_current_amps, shelly_energy_wh, shelly_device_online
  - Collects from PM, PM1, EM, EM1 components concurrently
- [x] `shelly metrics json` - Output metrics as JSON
  - Flags: `--devices`, `--continuous`, `--interval`, `--output`
  - Structured JSON with timestamp, devices, components
  - For integration with other monitoring tools
- [x] `shelly metrics influxdb` - InfluxDB line protocol output (aliases: influx, line)
  - Flags: `--devices`, `--continuous`, `--interval`, `--output`, `--measurement`, `--tags`
  - InfluxDB line protocol format for time-series databases
  - Supports custom measurement names and additional tags

---

## Phase 12: Alias System (gh-style) ✅

### 12.1 Alias Commands ✅
- [x] `shelly alias list` - List all aliases (aliases: ls, l)
  - Supports JSON/YAML output formats
  - Table display with name, command, type columns
- [x] `shelly alias set <name> <command>` - Create alias (aliases: add, create)
  - Supports `--shell` flag for shell aliases
  - Auto-detects `!` prefix for shell commands
- [x] `shelly alias delete <name>` - Delete alias (aliases: del, rm, remove)
- [x] `shelly alias import <file>` - Import aliases from YAML file (alias: load)
  - Supports `--merge` to skip existing aliases
- [x] `shelly alias export [file]` - Export aliases to YAML file (aliases: save, dump)
  - Outputs to stdout if no file specified

### 12.2 Device Aliases
- [x] Device aliases handled via device registration (`shelly device add <name> <address>`)
  - No separate command needed - device names ARE the aliases

### 12.3 Alias Expansion ✅
- [x] Support argument interpolation with `$1`, `$2`, `$@` (in config/aliases.go)
- [x] Support shell command aliases (prefixed with `!`)
- [x] Implement alias expansion in command parsing (in root.go)
  - Intercepts args before cobra processes them
  - Expands command aliases and re-parses
  - Shell aliases execute directly via $SHELL

---

## Phase 13: Plugin/Extension System (gh-style) ✅

### 13.1 Plugin Infrastructure (PATH-based, gh-style) ✅
- [x] Create `internal/plugins/loader.go`:
  - Discover plugins in PATH and ~/.config/shelly/plugins/
  - Plugin naming convention: `shelly-*` executables
  - Validate plugin executables (exists, executable bit)
  - Version detection via `shelly-<name> --version`
- [x] Create `internal/plugins/executor.go`:
  - Execute plugins with argument forwarding
  - Environment variable injection:
    - SHELLY_CONFIG_PATH: Config file location
    - SHELLY_DEVICES_JSON: JSON of registered devices
    - SHELLY_OUTPUT_FORMAT: Current output format
    - SHELLY_NO_COLOR: Color disabled flag
    - SHELLY_VERBOSE: Verbose mode flag
  - Capture and display output
  - Handle exit codes
- [x] Create `internal/plugins/registry.go`:
  - Track installed plugins with metadata
  - Version tracking and update checking
  - Plugin manifest support (optional .shelly-plugin.yaml)

### 13.2 Plugin Commands ✅
- [x] `shelly extension list` - List installed extensions (aliases: ls, l)
  - Supports --all flag to show all discovered extensions
  - JSON/YAML output support
- [x] `shelly extension install <source>` - Install extension (alias: add)
  - Support local files, URLs, and GitHub repos (gh:user/repo)
  - GitHub integration via `internal/github/` package (downloads latest release binary)
  - Platform detection for correct binary (linux/darwin, amd64/arm64)
  - Archive extraction (tar.gz, zip)
  - Force reinstall with --force
- [x] `shelly extension remove <name>` - Remove extension (aliases: rm, uninstall, delete)
- [x] `shelly extension upgrade [name]` - Upgrade extension(s) (alias: update)
  - Supports --all flag
  - GitHub release version checking and automatic download
  - Uses shared `internal/github/` package for release handling
- [x] `shelly extension create <name>` - Scaffold new extension (aliases: new, init, scaffold)
  - Supports --lang (bash, go, python)
  - Creates full project structure with README, main file
- [x] `shelly extension exec <name> [args]` - Execute extension explicitly (alias: run)

### 13.3 Plugin SDK (Deferred)
- [ ] Create `pkg/api/` public package (future)
- [ ] Document plugin development in `docs/plugins.md` (future)
- [ ] Create example plugins in `examples/plugins/` (future)

---

## Phase 14: TUI Dashboard

> **READ WHEN IMPLEMENTING:** See [docs/architecture.md](docs/architecture.md) for full TUI architecture patterns, component structure, and code examples.

**Architecture:** gh-dash/BubbleTea with Elm Architecture (Model/Init/Update/View)

### 14.1 Tasks
- [x] Create `internal/tui/app.go` - Main app model
- [x] Create `internal/tui/keys.go` - Vim-style keyboard bindings
- [x] Create `internal/tui/styles.go` - Theme-aware styles via bubbletint

### 14.2 Components (each with model.go, view.go, update.go, styles.go)
- [x] `components/devicelist/` - Device table (bubbles/table)
- [x] `components/devicedetail/` - Device detail panel
- [x] `components/monitor/` - Real-time monitoring
- [x] `components/search/` - Filter bar (bubbles/textinput)
- [x] `components/statusbar/` - Bottom status bar
- [x] `components/help/` - Help overlay (glamour markdown)
- [x] `components/tabs/` - View switching tabs
- [x] `components/toast/` - Notifications
- [x] `components/cmdmode/` - Command mode input

### 14.3 Data Layer
- [x] Data fetching integrated into component models

### 14.4 Commands
- [x] `shelly dash` - Launch dashboard (--refresh, --filter, --view)
- [x] `shelly dash devices` - Device list view
- [x] `shelly dash monitor` - Monitoring view
- [x] `shelly dash events` - Event stream view

### 14.5 Features
- [x] Vim-style navigation (j/k, arrows)
- [x] Tab switching (1-4 or Tab/Shift+Tab)
- [x] Device actions (t=toggle, o=on, O=off, R=reboot)
- [x] Filter/search (/)
- [x] Help overlay (?)
- [x] Command mode (`:quit`, `:device`, `:filter`, `:theme`, `:view`, `:help`)
- [x] Runtime theme switching (via `:theme <name>` command)
- [x] Configurable keybindings via config.yaml (tui.keybindings section)

---

## Phase 15: Theme System

### 15.1 Theme Infrastructure
- [x] Create `internal/theme/theme.go`:
  - Integration with lrstanley/bubbletint
  - Theme loading from config
  - Runtime theme switching
- [x] Theme registry via bubbletint's DefaultRegistry (280+ built-in themes)
- [x] Theme structure defined via bubbletint v2 Tint type:
  ```go
  type Theme struct {
      Name        string
      Tint        tint.Tint        // bubbletint theme

      // Override colors
      StatusOK    lipgloss.Color
      StatusWarn  lipgloss.Color
      StatusError lipgloss.Color

      // Component styles
      Header      lipgloss.Style
      Table       TableStyles
      Panel       PanelStyles
  }
  ```

### 15.2 Theme Commands
- [x] `shelly theme list` - List available themes
- [x] `shelly theme set <name>` - Set theme
- [x] `shelly theme preview <name>` - Preview theme
- [x] `shelly theme current` - Show current theme
- [x] `shelly theme next/prev` - Cycle through themes
- [x] `shelly theme export <file>` - Export current theme
- [x] `shelly theme import <file>` - Import custom theme

### 15.3 Built-in Themes
- [x] Include themes from bubbletint (286 themes available):
  - Dracula (default)
  - Nord
  - Tokyo Night
  - GitHub Dark
  - Gruvbox
  - Catppuccin
  - One Dark
  - Solarized

---

## Phase 16: Shell Completions

### 16.1 Completion Generation
- [x] Create `internal/cli/completion/completion.go`:
  - Bash completion
  - Zsh completion
  - Fish completion
  - PowerShell completion
- [x] `shelly completion bash` - Output bash completions
- [x] `shelly completion zsh` - Output zsh completions
- [x] `shelly completion fish` - Output fish completions
- [x] `shelly completion powershell` - Output PowerShell completions
- [x] Generated completion files in `completions/` directory

### 16.2 Dynamic Completions
- [x] Implement dynamic completions for:
  - Device names (from registry) - CompleteDevices()
  - Device addresses (from discovery cache) - CompleteDiscoveredDevices()
  - Group names - CompleteGroups()
  - Script names (per device) - CompleteDeviceThenScriptID()
  - Schedule IDs (per device) - CompleteDeviceThenScheduleID()
  - Alias names - CompleteAliases()
  - Theme names - CompleteThemes()
  - Extension names - CompleteExtensions()
  - Scene names - CompleteScenes()
- [x] Cache completion data for performance (in-memory with TTL + filesystem cache for discovery)
- [x] Support completion descriptions

### 16.3 Completion Installation
- [x] `shelly completion install` - Install completions to shell config
  - Auto-detect shell
  - Add source line to rc file

---

## Phase 17: Update Command

### 17.0 Init Command ✅

> **INIT:** Before implementing, analyze and discuss:
> 1. What should `shelly init` do exactly? (First-run setup? Config generation? Device discovery?)
> 2. What's the optimal UX for the init flow? (Interactive wizard? Guided prompts? Non-interactive flags?)
> 3. How do other CLIs handle this? (gh auth login, docker init, npm init, git init)
> 4. What should be configured? (Default device, output format, theme, shell completions, cloud auth?)
> 5. Should it detect existing config and offer to migrate/update?

- [x] `shelly init` - First-run setup wizard (aliases: `setup`, `configure`)
  - Interactive guided setup for new users
  - Flags: `--non-interactive`, `--force` (overwrite existing config)
  - Creates config directory and example config file
  - Installs shell completions for detected shell
  - Offers to run device discovery

### 17.1 Update System ✅
- [x] Create `internal/cmd/update/update.go`:
  - Check for new releases via GitHub API
  - Download and verify checksums (SHA256)
  - Replace binary in-place with backup/restore
- [x] `shelly update` - Update CLI to latest version (aliases: `upgrade`, `self-update`)
  - Flags: `--check` (just check, don't update)
  - Flags: `--version <tag>` (specific version)
  - Flags: `--channel stable|beta` (release channel)
  - Flags: `--rollback` (revert to previous version)
  - Flags: `--yes` (skip confirmation)
  - Flags: `--include-pre` (include prereleases)
- [x] `shelly version` - Show current version
  - Include available update info (from cache)
  - Flags: `--short` (version only), `--json` (machine-readable), `--check` (check for updates)

### 17.2 Version Checking ✅
- [x] Check for updates on startup (configurable)
  - Rate limit to once per 24 hours
  - Cache check result in ~/.config/shelly/cache/latest-version
- [x] Display update notification if available
- [x] Respect `SHELLY_NO_UPDATE_CHECK` environment variable
- [x] Respect `--quiet` global flag to suppress notifications

---

## Phase 18: KVS Commands (Gen2+) ✅ Complete

### 18.1 KVS Management
- [x] `shelly kvs list <device>` - List KVS keys
  - Aliases: ls, l
  - Flags: --values (show values), --match (pattern filter)
- [x] `shelly kvs get <device> <key>` - Get KVS value
  - Aliases: g, read
  - Flags: --raw (output raw value only)
- [x] `shelly kvs set <device> <key> <value>` - Set KVS value
  - Aliases: s, write, put
  - Flags: --null (set null value)
  - Auto-parses JSON, numbers, booleans
- [x] `shelly kvs delete <device> <key>` - Delete KVS key
  - Aliases: del, rm, remove
  - Flags: --yes (skip confirmation)
- [x] `shelly kvs export <device> <file>` - Export all KVS data
  - Aliases: exp, save, dump
  - Flags: --format (json, yaml)
- [x] `shelly kvs import <device> <file>` - Import KVS data
  - Aliases: load, restore
  - Flags: --overwrite, --dry-run, --yes

---

## Phase 19: Advanced Features from shelly-manager

### 19.1 Configuration Templates ✅ Complete
- [x] `shelly template list` - List config templates
  - Aliases: ls, l
  - JSON/YAML output support
- [x] `shelly template create <name> <device>` - Create template from device
  - Aliases: new, save
  - Flags: --description, --include-wifi, --force
  - Captures device config with WiFi credentials sanitized by default
- [x] `shelly template apply <template> <device>` - Apply template
  - Aliases: set, push
  - Flags: --dry-run, --yes
  - Model compatibility warnings
- [x] `shelly template diff <template> <device>` - Compare template to device
  - Aliases: compare, cmp
  - Shows configuration differences in table format
- [x] `shelly template export <name> [file]` - Export template
  - Aliases: save, dump
  - Flags: --format (json, yaml)
  - Auto-detects format from file extension
- [x] `shelly template import <file> [name]` - Import template
  - Aliases: load
  - Flags: --force
  - Validates required fields
- [x] `shelly template delete <name>` - Delete template
  - Aliases: del, rm, remove
  - Flags: --yes

### 19.2 Export Formats ✅ Complete
- [x] `shelly export json <device> [file]` - Export as JSON
- [x] `shelly export yaml <device> [file]` - Export as YAML
- [x] `shelly export csv <devices...> [file]` - Export devices list as CSV
- [x] `shelly export ansible <devices...> [file]` - Export as Ansible inventory
- [x] `shelly export terraform <devices...> [file]` - Export as Terraform config

### 19.3 Provisioning ✅ Complete
- [x] `shelly provision wifi <device>` - Provision WiFi settings
  - Aliases: network, wlan
  - Interactive network scan and selection
  - Direct provisioning via --ssid and --password flags
  - --no-scan flag to skip scan and prompt for SSID
- [x] `shelly provision bulk <file>` - Bulk provision from config file
  - Aliases: batch, mass
  - YAML file with global and per-device WiFi settings
  - Parallel provisioning with --parallel flag
  - --dry-run for config validation
- [x] `shelly provision ble <device>` - BLE-based provisioning
  - Placeholder with workaround instructions
  - Guides users to use AP mode provisioning

### 19.4 Actions/Webhooks Helper ✅ Complete
- [x] `shelly action list <device>` - List available actions (Gen1)
  - Aliases: ls, show
  - Placeholder with workaround instructions (curl)
- [x] `shelly action set <device> <action> <url>` - Set action URL
  - Aliases: add, configure
  - Placeholder with workaround instructions
- [x] `shelly action clear <device> <action>` - Clear action URL
  - Aliases: delete, remove, rm
  - Placeholder with workaround instructions
- [x] `shelly action test <device> <action>` - Test action trigger
  - Aliases: trigger, fire
  - Documents that Gen1 actions trigger via state changes

---

## Phase 20: Convenience Commands

### 20.1 Quick Commands ✅ Complete
- [x] `shelly on <device>` - Quick on (detect switch/light/cover)
  - Aliases: turn-on, enable
  - Flags: --all (control all components)
- [x] `shelly off <device>` - Quick off
  - Aliases: turn-off, disable
  - Flags: --all (control all components)
- [x] `shelly toggle <device>` - Quick toggle
  - Aliases: flip, switch
  - Flags: --all (toggle all components)
- [x] `shelly status [device]` - Quick status (all or specific)
  - Aliases: st, state
  - Shows component states with colored output
- [x] `shelly reboot <device>` - Quick reboot
  - Aliases: restart
  - Flags: --delay, --yes
- [x] `shelly reset <device>` - Quick factory reset (with confirmation)
  - Aliases: factory-reset, wipe
  - Requires both --yes and --confirm flags for safety

### 20.2 Interactive Mode ✅ Complete
- [x] `shelly interactive` - Launch interactive REPL
  - Aliases: repl, i
  - Session state with active device
  - Built-in commands: help, devices, connect, disconnect, status, on, off, toggle, rpc, methods, info
  - Flags: --device, --no-prompt
- [x] `shelly shell <device>` - Interactive shell for device
  - Direct RPC method execution
  - Built-in commands: help, info, status, config, methods, components
  - Persistent connection to device

### 20.3 Debug Commands
- [ ] `shelly debug log <device>` - Get device debug log (Gen1)
- [ ] `shelly debug rpc <device> <method> [params]` - Raw RPC call
- [ ] `shelly debug coiot <device>` - Show CoIoT status
- [ ] `shelly debug websocket <device>` - WebSocket debug connection
- [ ] `shelly debug methods <device>` - List available RPC methods

---

## Phase 21: BTHome/Zigbee/LoRa Commands

### 21.1 BTHome Commands
- [ ] `shelly bthome list <device>` - List BTHome devices
- [ ] `shelly bthome add <device>` - Start BTHome discovery
- [ ] `shelly bthome remove <device> <id>` - Remove BTHome device
- [ ] `shelly bthome status <device> [id]` - BTHome device status

### 21.2 Zigbee Commands (Gen4 gateways)
- [ ] `shelly zigbee status <device>` - Zigbee network status
- [ ] `shelly zigbee discover <device>` - Discover Zigbee devices
- [ ] `shelly zigbee pair <device>` - Start pairing mode
- [ ] `shelly zigbee remove <device> <id>` - Remove Zigbee device
- [ ] `shelly zigbee list <device>` - List paired Zigbee devices

### 21.3 LoRa Commands
- [ ] `shelly lora status <device>` - LoRa add-on status
- [ ] `shelly lora config <device>` - Configure LoRa settings
- [ ] `shelly lora send <device> <data>` - Send LoRa message
- [ ] `shelly lora receive <device>` - Monitor incoming messages

---

## Phase 22: Matter Commands

### 22.1 Matter Management
- [ ] `shelly matter status <device>` - Matter status
- [ ] `shelly matter enable <device>` - Enable Matter
- [ ] `shelly matter disable <device>` - Disable Matter
- [ ] `shelly matter reset <device>` - Reset Matter config
- [ ] `shelly matter code <device>` - Show pairing code

---

## Phase 23: Gen1 Device Support

> **Note:** Gen1 devices use a different API than Gen2+. The shelly-go library provides full Gen1 support via the `gen1/` package.

### 23.1 Gen1 Client Integration
- [ ] Create `internal/client/gen1.go`:
  - Wrapper for shelly-go `gen1/` package
  - Auto-detection of device generation
  - Unified interface where possible

### 23.2 Gen1 Discovery
- [ ] Enhance `shelly discover coiot` for Gen1-specific info
- [ ] Add `--gen1-only` flag to discovery commands

### 23.3 Gen1 Control Commands
- [ ] Extend existing commands to support Gen1 devices:
  - `shelly switch on/off/toggle/status` - Use Gen1 relay API
  - `shelly cover open/close/stop/position` - Use Gen1 roller API
  - `shelly light on/off/set/status` - Use Gen1 light API (Bulbs, Duo)
  - `shelly rgb set/status` - Use Gen1 color API (RGBW, Bulb)

### 23.4 Gen1-Specific Commands
- [ ] `shelly gen1 settings <device>` - Get/set Gen1 settings
- [ ] `shelly gen1 actions <device>` - Manage Gen1 actions (URLs)
- [ ] `shelly gen1 status <device>` - Full Gen1 status dump
- [ ] `shelly gen1 ota <device>` - Gen1 OTA firmware update

### 23.5 CoIoT Real-time Updates
- [ ] `shelly gen1 coiot <device>` - Subscribe to CoIoT updates
- [ ] Integrate CoIoT into TUI monitoring view

---

## Phase 24: Sensor Commands

> **Note:** Covers environmental sensors available in shelly-go `gen2/components/`.

### 24.1 Temperature Sensor
- [ ] `shelly sensor temperature list <device>` - List temperature sensors
- [ ] `shelly sensor temperature status <device> [id]` - Current temperature
- [ ] `shelly sensor temperature history <device> [id]` - Temperature history
  - Flags: `--period`, `--format`

### 24.2 Humidity Sensor
- [ ] `shelly sensor humidity list <device>` - List humidity sensors
- [ ] `shelly sensor humidity status <device> [id]` - Current humidity

### 24.3 Flood Sensor
- [ ] `shelly sensor flood list <device>` - List flood sensors
- [ ] `shelly sensor flood status <device> [id]` - Flood detection status
- [ ] `shelly sensor flood test <device> [id]` - Test flood alarm

### 24.4 Smoke Sensor
- [ ] `shelly sensor smoke list <device>` - List smoke sensors
- [ ] `shelly sensor smoke status <device> [id]` - Smoke detection status
- [ ] `shelly sensor smoke test <device> [id]` - Test smoke alarm
- [ ] `shelly sensor smoke mute <device> [id]` - Mute alarm

### 24.5 Illuminance Sensor
- [ ] `shelly sensor illuminance list <device>` - List illuminance sensors
- [ ] `shelly sensor illuminance status <device> [id]` - Current light level

### 24.6 Voltmeter
- [ ] `shelly sensor voltmeter list <device>` - List voltmeters
- [ ] `shelly sensor voltmeter status <device> [id]` - Current voltage reading

### 24.7 Combined Sensor Status
- [ ] `shelly sensor status <device>` - All sensor readings in one view
- [ ] `shelly sensor monitor <device>` - Real-time sensor monitoring
  - Updates via WebSocket
  - TUI display with graphs

---

## Phase 25: Thermostat Commands

> **Note:** Thermostat support via shelly-go `gen2/components/thermostat.go`.

### 25.1 Thermostat Control
- [ ] `shelly thermostat list <device>` - List thermostats
- [ ] `shelly thermostat status <device> [id]` - Thermostat status
- [ ] `shelly thermostat set <device> [id]` - Set thermostat
  - Flags: `--target-temp`, `--mode`, `--enable`
- [ ] `shelly thermostat enable <device> [id]` - Enable thermostat
- [ ] `shelly thermostat disable <device> [id]` - Disable thermostat

### 25.2 Thermostat Schedules
- [ ] `shelly thermostat schedule list <device> [id]` - List schedules
- [ ] `shelly thermostat schedule set <device> [id]` - Set schedule
  - Interactive mode for schedule configuration

---

## Phase 26: Documentation

### 26.1 README.md
- [ ] Project overview and features
- [ ] Installation instructions (all methods)
- [ ] Quick start guide
- [ ] Configuration overview
- [ ] Link to full documentation

### 26.2 Command Documentation
- [ ] Generate `docs/commands/` from Cobra command help
- [ ] Include examples for each command
- [ ] Document all flags and options
- [ ] Add common use cases

### 26.3 Configuration Reference
- [ ] Create `docs/configuration.md`:
  - Config file format and location
  - All configuration options
  - Environment variables
  - Example configurations

### 26.4 Plugin Development Guide
- [ ] Create `docs/plugins.md`:
  - Plugin architecture overview
  - SDK documentation
  - Example plugin walkthrough
  - Publishing plugins

### 26.5 Theming Guide
- [ ] Create `docs/themes.md`:
  - Built-in themes
  - Custom theme creation
  - Theme file format
  - TUI component styling

### 26.6 Man Pages
- [ ] Generate man pages via Cobra
- [ ] Include in release artifacts

### 26.7 Documentation Site
- [ ] Create docs site using mkdocs-material or docusaurus
- [ ] Deploy to GitHub Pages
- [ ] Include:
  - Getting started guide
  - Command reference (generated from Cobra)
  - Configuration guide
  - Plugin development guide
  - TUI usage guide
  - Examples and tutorials
  - Terminal GIFs using Charm VHS

---

## Phase 27: Examples

### 27.1 Alias Examples
- [ ] Create `examples/aliases/`:
  - power-users.yaml - Power user aliases
  - shortcuts.yaml - Common shortcuts
  - automation.yaml - Automation-focused aliases

### 27.2 Script Examples
- [ ] Create `examples/scripts/`:
  - morning-routine.sh - Morning automation
  - away-mode.sh - Away mode setup
  - energy-report.sh - Energy reporting
  - bulk-update.sh - Bulk firmware update

### 27.3 Plugin Examples
- [ ] Create `examples/plugins/`:
  - shelly-notify - Desktop notifications
  - shelly-homekit - HomeKit bridge info
  - shelly-prometheus - Prometheus metrics

### 27.4 Config Examples
- [ ] Currently in `cfg/` with a full example and schema - needs update once here:
  - minimal.yaml - Minimal configuration
  - full.yaml - Full configuration with all options
  - multi-site.yaml - Multiple location setup

---

## Phase 28: Testing

> **READ WHEN IMPLEMENTING:** See [docs/testing.md](docs/testing.md) for full testing strategy, coverage targets, and code examples.

**Target:** 90%+ overall coverage

### 28.1 Tasks
- [ ] Implement unit tests per package (see docs/testing.md for targets)
- [ ] Create `internal/testutil/` with MockClient, MockServer, TestIOStreams, TestFactory
- [ ] Add integration tests with mock device server
- [ ] Add TUI tests using `charmbracelet/x/exp/teatest`
- [ ] Add E2E tests for CLI invocations
- [ ] Setup CI coverage reporting with threshold enforcement

---

## Phase 29: Innovative Commands

> These commands differentiate the CLI and showcase its power. They're the "wow factor" that makes this tool stand out.

### 29.1 User Experience Excellence

- [ ] `shelly feedback` - Integrated issue reporting to GitHub
  - Interactive mode or `--type bug|feature|device`
  - Auto-populates system info, CLI version, device context
  - Flags: `--attach-log`, `--screenshot`
- [ ] `shelly doctor` - Comprehensive diagnostic health check (inspired by `brew doctor`)
  - Checks: CLI version, config validity, network reachability, device health, firmware status, cloud auth
  - Flags: `--network`, `--devices`, `--full`
- [ ] `shelly context` - Context management (kubectl-style)
  - `shelly context list|use|create|export|import`
  - Switch between device groups/locations seamlessly
- [ ] `shelly history` - Command history with replay
  - `shelly history [--device]`, `shelly history replay <id>`, `shelly history export`
- [ ] `shelly diff` - Visual diff between device configurations
  - `shelly diff <device1> <device2>` or `shelly diff <device> <backup.json>`

### 29.2 Power User Features

- [ ] `shelly watch` - Real-time monitoring dashboard (BubbleTea TUI)
  - `shelly watch [devices...]`, `--power`, `--events`, `--interval`
  - Live updates via WebSocket
- [ ] `shelly simulate` - Dry run mode for commands
  - `shelly simulate <any-command>` - shows what would execute without changes
- [ ] `shelly record` - Macro recording and replay
  - `shelly record start|stop|list|play|export <name>`
- [ ] `shelly benchmark` - Device performance testing
  - Measures ping, API latency, toggle response times
  - `shelly benchmark <device> [--iterations N]`
- [ ] `shelly profile usage` - Analyze device usage patterns
  - Peak hours, average on-time, energy consumption, toggle frequency

### 29.3 Developer & Integration

- [ ] `shelly api` - Direct RPC access for developers
  - `shelly api <device> <method> [params]`
  - Flags: `--list-methods`, `--interactive` (REPL)
- [ ] `shelly webhook server` - Built-in local webhook receiver for testing
  - Auto-configures devices to send webhooks to this server
  - `shelly webhook server [--port 8080] [--log]`
- [ ] `shelly export terraform` - Infrastructure as Code export
  - `shelly export terraform|ansible|homeassistant --all`
- [ ] `shelly mqtt bridge` - Act as MQTT bridge for devices
  - `shelly mqtt bridge --broker mqtt://localhost:1883 [--ha-discovery]`
- [ ] `shelly test` - Integration testing for devices
  - `shelly test <device> [--stress] [--network] [--power]`

### 29.4 Smart Automation

- [ ] `shelly trigger` - Event-based automation rules
  - `shelly trigger create <name> --event <type> --action <command>`
  - Events: sunset, sunrise, device events, thresholds
- [ ] `shelly chain` - Command chaining with conditions
  - `shelly chain "cmd1" --then "cmd2" --on-error "notify"`
- [ ] `shelly schedule smart` - Intelligent scheduling
  - `--trigger sunset|sunrise [--offset +30m] [--condition "not weekend"]`

### 29.5 Fleet Management (Enterprise)

- [ ] `shelly fleet` - Fleet management for large deployments
  - `shelly fleet list|health|firmware|update|export|sync`
  - Staged rollouts: `shelly fleet update --staged 10%`
- [ ] `shelly audit` - Security audit for devices
  - Checks: default credentials, outdated firmware, network exposure
  - `shelly audit [--check-auth] [--check-firmware] [--check-network]`
- [ ] `shelly report` - Generate professional reports
  - `shelly report energy|devices|audit|usage --format pdf|html|json`

### 29.6 Quality of Life

- [ ] `shelly tips` - Contextual tips based on usage
  - `shelly tips [--command X] [--new]`
- [ ] `shelly undo` - Undo last action where possible
- [ ] `shelly favorite` - Quick access to frequent commands
  - `shelly favorite add|list|run|delete <name>`
  - Shortcut: `shelly @morning`
- [ ] `shelly notify` - Desktop notifications on events
  - `shelly notify on <event> [--threshold]`

### 29.7 Interactive Features

- [ ] `shelly wizard` - Guided setup wizards
  - `shelly wizard [discover|scene|automation]`
- [ ] `shelly learn` - Built-in interactive tutorial
  - `shelly learn [basics|automation|scripting]`

### 29.8 Convenience Shortcuts

- [ ] `shelly @<alias>` - Quick alias invocation prefix
- [ ] `shelly last` - Repeat last command
  - `shelly last [--device other-device]`
- [ ] `shelly clipboard` - Copy output to clipboard
  - `shelly <command> --clipboard`
- [ ] `shelly open` - Open device in browser
  - `shelly open <device> [--cloud] [--docs <component>]`
- [ ] `shelly qr` - Generate QR codes
  - `shelly qr <device> [--wifi] [--export file.png]`

### 29.9 Analysis & Insights

- [ ] `shelly stats` - Usage statistics
  - `shelly stats [devices|commands|energy] [--period]`
- [ ] `shelly trends` - Usage trend visualization
  - `shelly trends power|usage <device> [--export chart.svg]`
- [ ] `shelly compare` - Compare devices or time periods
  - `shelly compare <device1> <device2>` or `--this-week --last-week`
- [ ] `shelly forecast` - Usage/cost forecasting
  - `shelly forecast energy|cost --period month [--rate 0.15]`
- [ ] `shelly anomaly` - Anomaly detection
  - `shelly anomaly detect [--device X] [--alert]`

### 29.10 Networking & Connectivity

- [ ] `shelly ping` - Enhanced ping with latency stats
  - `shelly ping <device> [--all] [--continuous] [--graph]`
- [ ] `shelly trace` - Connection trace
  - `shelly trace <device> [--protocol]`
- [ ] `shelly network` - Network diagnostics
  - `shelly network scan|topology|bandwidth|latency-map`
- [ ] `shelly proxy` - Local proxy mode for remote access
  - `shelly proxy start [--port 8080]`

### 29.11 Security & Auth (Enhanced)

- [ ] `shelly auth rotate` - Rotate device credentials
- [ ] `shelly auth test` - Test auth connectivity
- [ ] `shelly auth export|import` - Encrypted credential export/import
- [ ] `shelly cert` - Certificate management
  - `shelly cert show|install|renew <device>`
- [ ] `shelly vault` - Secure credential storage
  - `shelly vault init|add|list|export`
- [ ] `shelly firewall` - Device firewall rules
  - `shelly firewall show|allow|deny <device> [subnet]`

### 29.12 Scripting & Automation

- [ ] `shelly cron` - Cron-style scheduling
  - `shelly cron "0 7 * * *" "scene activate morning"`
- [ ] `shelly wait` - Wait for condition
  - `shelly wait <device> --state on|--online [--timeout 30s]`
- [ ] `shelly if` - Conditional execution
  - `shelly if "<condition>" then "<command>"`
- [ ] `shelly loop` - Loop execution
  - `shelly loop N "command" [--delay 1s] [--until "time"]`
- [ ] `shelly pipeline` - Command pipelines
  - `shelly pipeline create|run|list <name>`

### 29.13 Data & Export

- [ ] `shelly dump` - Full data dump
  - `shelly dump <device> [--all] [--format json] [--schema]`
- [ ] `shelly import` - Bulk import
  - `shelly import devices.yaml|scenes.yaml|config.json`
- [ ] `shelly sync` - Cloud synchronization
  - `shelly sync cloud [--push|--pull]`
- [ ] `shelly archive` - Archive management
  - `shelly archive create|list|restore|--rotate N`

### 29.14 Debugging & Development

- [ ] `shelly debug trace` - Trace all API calls
- [ ] `shelly debug memory|crash-log|performance` - Device diagnostics
- [ ] `shelly mock` - Mock device mode for testing without hardware
  - `shelly mock create|list|scenario <name>`
- [ ] `shelly replay` - Replay recorded API calls
  - `shelly replay record|stop|play <file>`
- [ ] `shelly profile perf` - Performance profiling
  - `shelly profile perf <device>|network|cli`

### 29.15 Integration Commands

- [ ] `shelly homeassistant` - Home Assistant integration
  - `shelly homeassistant status|discover|export|mqtt-config`
- [ ] `shelly node-red` - Node-RED integration
  - `shelly node-red flows export|palette`
- [ ] `shelly openhab` - OpenHAB integration
  - `shelly openhab things|items`
- [ ] `shelly alexa` - Alexa integration info
  - `shelly alexa status|devices`
- [ ] `shelly google-home` - Google Home integration info
  - `shelly google-home status|devices`

### 29.16 Maintenance & Lifecycle

- [ ] `shelly maintenance` - Maintenance mode
  - `shelly maintenance enable|disable|schedule <device>`
- [ ] `shelly lifecycle` - Device lifecycle info
  - `shelly lifecycle <device> [--warranty|--age|--usage]`
- [ ] `shelly inventory` - Inventory management
  - `shelly inventory list|add|export|value`

### 29.17 Notifications & Alerts

- [ ] `shelly alert` - Alert management
  - `shelly alert create|list|test|snooze <name>`
  - `--condition "power > 1000W" --notify email:admin@...`
- [ ] `shelly notify setup` - Notification channel setup
  - `shelly notify setup email|slack|telegram|pushover|gotify`
- [ ] `shelly subscribe` - Event subscriptions
  - `shelly subscribe <device> --event <type>`

### 29.18 Documentation & Help

- [ ] `shelly man` - Manual pages
  - `shelly man <command> [--all] [--pdf]`
- [ ] `shelly examples` - Command examples
  - `shelly examples <command> [--scenario X] [--level advanced]`
- [ ] `shelly explain` - Explain what a command does
  - `shelly explain "switch on kitchen --timer 300"`
- [ ] `shelly cheatsheet` - Quick reference
  - `shelly cheatsheet [--pdf] [--poster]`

### 29.19 Meta & CLI Management

> **Note:** `shelly update` is defined in Phase 17. These are additional config/meta commands.

- [ ] `shelly config edit` - Open config in $EDITOR
- [ ] `shelly config validate|doctor|migrate` - Config management
- [ ] `shelly cache` - Cache management
  - `shelly cache clear|stats|warm [--discovery]`
- [ ] `shelly log` - CLI logging
  - `shelly log show|tail|level|export`
- [ ] `shelly telemetry` - Anonymous telemetry control
  - `shelly telemetry status|on|off|show`

### 29.20 Fun & Easter Eggs

- [ ] `shelly party` - Party mode (cycle colors on RGB devices)
  - `shelly party [--mode rainbow|disco] [--duration 5m]`
- [ ] `shelly sleep` - Gradual lights off
  - `shelly sleep [--duration 30m] [--except bedroom]`
- [ ] `shelly wake` - Gradual lights on
  - `shelly wake [--simulate sunrise] [--duration 15m]`
- [ ] `shelly demo` - Demo mode without real devices
  - `shelly demo [--scenario home] [--speed fast]`

---

## Phase 30: Polish & Release (FINAL)

> ⛔ **This is the FINAL phase.** Do not release until all prior phases are complete.

### 30.1 Code Quality
- [ ] Run golangci-lint with all enabled linters
- [ ] Fix all linter warnings
- [ ] Ensure consistent error messages
- [ ] Review public API surface
- [ ] Verify all commands have --help text
- [ ] Check for proper context propagation
- [ ] Verify all output respects `--no-color` and `--plain` flags
- [ ] Ensure all pipeable commands produce valid JSON/YAML

### 30.2 Performance
- [ ] Profile startup time
- [ ] Optimize config loading
- [ ] Cache discovery results
- [ ] Lazy-load TUI components

### 30.3 Accessibility
- [ ] Ensure TUI works with screen readers (where possible)
- [ ] Support `--no-color` for all output
- [ ] Provide text-only alternatives to TUI elements
- [ ] Verify exit codes are correct for scripting

### 30.4 Release Preparation
- [ ] Write CHANGELOG.md with all features
- [ ] Create GitHub release notes template
- [ ] Verify goreleaser builds all platforms
- [ ] Test installation instructions on all platforms
- [ ] Create professional ASCII banner/logo
- [ ] Final documentation review
- [ ] Create v1.0.0 tag

---

## Shelly-Go Library Coverage

> **Full details:** See [docs/shelly-go-coverage.md](docs/shelly-go-coverage.md)

**Summary:** 9 feature areas complete, ~25 remaining across Phases 5-25, plus 82 innovative commands in Phase 29.

| Priority | Features | Status |
|----------|----------|--------|
| **Critical** | Firmware, Cloud API | Phases 6, 9 |
| **High** | Energy, Scripts, Schedules, KVS, MQTT, Webhooks, WiFi | Phases 5, 7, 8, 11, 18 |
| **Medium** | BTHome, Zigbee, Gen1, Sensors, Thermostat | Phases 21, 23-25 |
| **Low** | Matter, LoRa, Z-Wave, ModBus, Virtual | Phases 21, 22 |
| **Innovative** | 82 new commands for QoL, fleet, automation, integrations | Phase 29 |

---

## Dependencies

> **READ WHEN IMPLEMENTING:** See [docs/dependencies.md](docs/dependencies.md) for full dependency list, usage guidelines, and code examples.

**Key dependencies:** shelly-go, cobra, viper, bubbletea, lipgloss, bubbletint, survey, errgroup

---

## Success Criteria

- [ ] All commands from shelly-manager implemented (where applicable)
- [ ] All shellyctl features implemented
- [ ] Full shelly-go library coverage via CLI
- [ ] TUI dashboard comparable to k9s/gh-dash
- [ ] Alias system matching gh CLI
- [ ] Plugin system matching gh CLI
- [ ] Shell completions for bash/zsh/fish/powershell
- [ ] Theme support via bubbletint (280+ themes)
- [ ] 90%+ test coverage
- [ ] Full documentation (godoc, user docs, examples)
- [ ] Automated releases via GitHub Actions
- [ ] Clean golangci-lint with strict settings
- [ ] Phase 29 innovative commands: `feedback`, `doctor`, `watch`, `api`, `audit`, `fleet`
- [ ] Home Assistant, Node-RED, OpenHAB integration commands
- [ ] Demo mode working without real devices (for showcasing)
- [ ] Professional branding (logo, ASCII banner)
