# Shelly CLI - Comprehensive Implementation Plan

> **CRITICAL:** NEVER mark any task as "optional", "nice to have", "not necessary", or "future improvement". All tasks are REQUIRED. Do not defer work without explicit user approval. If context limits are reached, split the work into multiple sessions - do not skip tasks.

> **IMPORTANT:** All future session work derives from this PLAN.md file. No additional session files should be created. Check off completed work and add concise context notes only when necessary to save investigation time.

> **WORKFLOW REQUIREMENTS:**
> 1. **Lint after every file change:** Run `golangci-lint run <file>` immediately after creating or modifying any `.go` file. Address all issues before proceeding.
> 2. **Tests for every file:** Every new or modified `.go` file must have corresponding test coverage. Write tests immediately after implementation, not as a separate phase.
> 3. **Commit frequently:** Commit after completing each major section, large subsection, or BEFORE running any terminal commands that write to multiple files (e.g., `sed`, `gci`, bulk operations). Never leave work uncommitted.
> 4. **No nolint without approval:** Do not add `//nolint` directives without explicit user approval.

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

**Last Updated:** 2025-12-11
**Phase:** Phase 0 - Architecture Refactoring (Phase 0.1-0.3 COMPLETE, 0.4-0.7 remaining)
**Completed:** Phases 1-2 (full), Phase 0.1-0.3
**Partial:** Phases 3-4 (commands done, completions TBD), 12 (core done), 13 (core done), 15 (core done), 16 (basic done, dynamic TBD)
**Pending:** Phase 0.4-0.7, Phases 5-11, 14, 17-26
**Test Coverage:** ~25% average - TARGET: >90%

**Architecture Audit (2025-12-11):**
Comprehensive audit against industry standards (gh, kubectl, docker, jira-cli, gh-dash, k9s) revealed:
- **Critical:** 38+ duplicate command patterns need DRY refactoring
- **Critical:** Duplicate packages (`internal/output/` vs `internal/ui/`) need consolidation
- **Critical:** Context propagation missing (41+ files use `context.Background()`)
- **Critical:** Missing shelly-go features (firmware, cloud, energy, scripts, schedules, etc.)
- **High:** Batch operations should use `errgroup` instead of manual WaitGroup
- **High:** Test coverage at ~20% (target: 90%)
- See `docs/architecture.md` for detailed patterns from reference projects

**Session Notes:**
- CLI restructured from `internal/cli/` to `internal/cmd/` with subdirectory-per-command pattern
- All control commands (switch, cover, light, rgb, input, device, group, batch, scene) implemented with tests
- Phase 0.1-0.3 completed (IOStreams 92.3%, cmdutil 92.3%, package consolidation)
- Batch operations use `errgroup.SetLimit()` for concurrency
- All tests pass (48 test packages), golangci-lint passes with 0 issues

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

### 0.4 Context Propagation
Fix context handling throughout codebase:

- [ ] Update `internal/cmd/root.go` to setup cancellation-aware context
- [ ] Update all 41+ command run() functions to accept context parameter
- [ ] Use `cmd.Context()` instead of `context.Background()`
- [ ] Ensure Ctrl+C cancels in-flight HTTP requests

### 0.5 Concurrency Patterns
Replace manual patterns with errgroup and add multi-writer output:

- [ ] Refactor `internal/cmd/batch/on/on.go` to use errgroup + MultiWriter
- [ ] Refactor `internal/cmd/batch/off/off.go` to use errgroup + MultiWriter
- [ ] Refactor `internal/cmd/batch/toggle/toggle.go` to use errgroup + MultiWriter
- [ ] Refactor `internal/cmd/batch/command/command.go` to use errgroup + MultiWriter
- [ ] Refactor `internal/cmd/scene/activate/activate.go` to use errgroup + MultiWriter
- [ ] Refactor `internal/cmd/discover/scan/scan.go` to use MultiWriter for progress
- [ ] Refactor `internal/plugins/loader.go` to parallelize version detection

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
- [ ] Refactor all `on.go` commands (switch, light, rgb) to use `cmdutil.RunWithSpinner`
- [ ] Refactor all `off.go` commands to use `cmdutil.RunWithSpinner`
- [ ] Refactor all `toggle.go` commands to use `cmdutil.RunWithSpinner`
- [ ] Refactor all `status.go` commands to use `cmdutil.PrintResult`
- [ ] Refactor all `list.go` commands to use `cmdutil.PrintResult`
- [ ] Apply consistent flag patterns via `cmdutil.AddComponentIDFlag`

### 0.7 Test Coverage Foundation
Establish testing patterns for new packages:

- [x] Add comprehensive tests for `internal/iostreams/` (target: 90%+) - **92.3% achieved**
- [x] Add comprehensive tests for `internal/cmdutil/` (target: 90%+) - **92.3% achieved**
- [x] Add table-driven tests for runner patterns
- [ ] Update existing command tests to use new patterns

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

### 5.1 Device Configuration
- [ ] `shelly config get <device>` - Get device config (all components)
- [ ] `shelly config get <device> <component>` - Get component config
- [ ] `shelly config set <device> <component> <key>=<value>...` - Set config
- [ ] `shelly config diff <device> <file>` - Compare config with file
- [ ] `shelly config export <device> <file>` - Export config to file
- [ ] `shelly config import <device> <file>` - Import config from file
  - Flags: --dry-run, --merge, --overwrite
- [ ] `shelly config reset <device> [component]` - Reset to defaults

### 5.2 Network Configuration
- [ ] `shelly wifi status <device>` - Show WiFi status
- [ ] `shelly wifi scan <device>` - Scan for networks
- [ ] `shelly wifi set <device>` - Configure WiFi
  - Flags: --ssid, --password, --static-ip, --gateway, --dns
- [ ] `shelly wifi ap <device>` - Configure AP mode
- [ ] `shelly ethernet status <device>` - Show ethernet status (Pro devices)
- [ ] `shelly ethernet set <device>` - Configure ethernet

### 5.3 Cloud Configuration
- [ ] `shelly cloud status <device>` - Show cloud status
- [ ] `shelly cloud enable <device>` - Enable cloud connection
- [ ] `shelly cloud disable <device>` - Disable cloud connection

### 5.4 Auth Configuration
- [ ] `shelly auth status <device>` - Show auth status
- [ ] `shelly auth set <device>` - Set auth credentials
  - Flags: --user, --password
- [ ] `shelly auth disable <device>` - Disable auth

### 5.5 MQTT Configuration
- [ ] `shelly mqtt status <device>` - Show MQTT status
- [ ] `shelly mqtt set <device>` - Configure MQTT
  - Flags: --server, --user, --password, --topic-prefix
- [ ] `shelly mqtt disable <device>` - Disable MQTT

### 5.6 Webhook Configuration
- [ ] `shelly webhook list <device>` - List webhooks
- [ ] `shelly webhook create <device>` - Create webhook
  - Flags: --event, --url, --active
- [ ] `shelly webhook delete <device> <id>` - Delete webhook
- [ ] `shelly webhook update <device> <id>` - Update webhook

---

## Phase 6: Firmware Commands

### 6.1 Firmware Status
- [ ] `shelly firmware check [device|--all]` - Check for updates
  - Show current version, available version, release notes summary
- [ ] `shelly firmware status <device>` - Show firmware status
- [ ] `shelly firmware list` - List firmware versions for device type

### 6.2 Firmware Updates
- [ ] `shelly firmware update <device>` - Update device firmware
  - Flags: --beta, --url (custom firmware), --yes (skip confirmation)
  - Progress indicator
  - Wait for device to come back online
- [ ] `shelly firmware update --all` - Update all devices
  - Flags: --parallel, --staged (percentage-based rollout)
- [ ] `shelly firmware rollback <device>` - Rollback to previous version

### 6.3 Firmware Download
- [ ] `shelly firmware download <device-type> <version>` - Download firmware file
  - Flags: --output, --latest, --beta

---

## Phase 7: Script Commands (Gen2+)

### 7.1 Script Management
- [ ] `shelly script list <device>` - List scripts
- [ ] `shelly script get <device> <id>` - Get script code
- [ ] `shelly script create <device> <name>` - Create script
  - Flags: --code (inline), --file (from file), --enable
- [ ] `shelly script update <device> <id>` - Update script code
  - Flags: --code, --file, --name
- [ ] `shelly script delete <device> <id>` - Delete script
- [ ] `shelly script start <device> <id>` - Start script
- [ ] `shelly script stop <device> <id>` - Stop script
- [ ] `shelly script eval <device> <code>` - Evaluate code snippet
- [ ] `shelly script upload <device> <file>` - Upload script from file
- [ ] `shelly script download <device> <id> <file>` - Download script to file

### 7.2 Script Library
- [ ] `shelly script-lib list` - List available script templates
- [ ] `shelly script-lib show <name>` - Show script template
- [ ] `shelly script-lib install <device> <name>` - Install template to device
  - Flags: --configure (interactive configuration)

---

## Phase 8: Schedule Commands

### 8.1 Schedule Management
- [ ] `shelly schedule list <device>` - List schedules
- [ ] `shelly schedule create <device>` - Create schedule
  - Flags: --timespec, --calls, --enable
  - Interactive mode for complex schedules
- [ ] `shelly schedule update <device> <id>` - Update schedule
- [ ] `shelly schedule delete <device> <id>` - Delete schedule
- [ ] `shelly schedule delete-all <device>` - Delete all schedules
- [ ] `shelly schedule enable <device> <id>` - Enable schedule
- [ ] `shelly schedule disable <device> <id>` - Disable schedule

### 8.2 Schedule Helpers
- [ ] `shelly schedule sunrise <device>` - Create sunrise-based schedule
  - Flags: --offset, --action, --days
- [ ] `shelly schedule sunset <device>` - Create sunset-based schedule
- [ ] `shelly schedule timer <device>` - Create timer-based schedule

---

## Phase 9: Cloud Commands

### 9.1 Cloud Authentication
- [ ] `shelly cloud login` - Authenticate with Shelly Cloud
  - OAuth flow with browser redirect
  - Save token securely
- [ ] `shelly cloud logout` - Clear cloud credentials
- [ ] `shelly cloud status` - Show cloud auth status
- [ ] `shelly cloud token` - Show/refresh token (for debugging)

### 9.2 Cloud Device Management
- [ ] `shelly cloud devices` - List cloud-registered devices
- [ ] `shelly cloud device <id>` - Show cloud device details
- [ ] `shelly cloud control <id> <action>` - Control via cloud
- [ ] `shelly cloud events` - Subscribe to real-time events
  - Output events as they arrive
  - Flags: --filter, --format

---

## Phase 10: Backup & Restore Commands

### 10.1 Backup Operations
- [ ] `shelly backup create <device> [file]` - Create device backup
  - Includes: config, scripts, schedules, webhooks, KVS
  - Flags: --encrypt (password-protected)
- [ ] `shelly backup restore <device> <file>` - Restore from backup
  - Flags: --dry-run, --skip-network, --decrypt
- [ ] `shelly backup list` - List saved backups
- [ ] `shelly backup export --all <directory>` - Backup all devices

### 10.2 Migration
- [ ] `shelly migrate <source> <target>` - Migrate config between devices
  - Flags: --dry-run, --force (different device types)
- [ ] `shelly migrate validate <backup>` - Validate backup file
- [ ] `shelly migrate diff <device> <backup>` - Show differences

---

## Phase 11: Monitoring Commands

> **Note:** Energy monitoring uses shelly-go `gen2/components/em.go`, `em1.go`, `pm.go`, `pm1.go` components.

### 11.1 Real-time Monitoring
- [ ] `shelly monitor <device>` - Real-time status monitoring
  - Auto-refresh with configurable interval
  - Color-coded status changes
- [ ] `shelly monitor power <device>` - Monitor power consumption
- [ ] `shelly monitor events <device>` - Monitor device events
  - WebSocket subscription via shelly-go events package
- [ ] `shelly monitor all` - Monitor all registered devices

### 11.2 Energy Monitoring (EM/EM1 Components)
- [ ] `shelly energy list <device>` - List energy meters (EM components)
- [ ] `shelly energy status <device> [id]` - Current power/energy status
  - Shows: voltage, current, power, energy, power factor, frequency
- [ ] `shelly energy history <device> [id]` - Energy history
  - Flags: `--period` (minute, hour, day, week, month)
  - Flags: `--from`, `--to` (date range)
  - Flags: `--format` (json, csv, table)
- [ ] `shelly energy export <device> <file>` - Export energy data
  - Flags: `--period`, `--from`, `--to`, `--format`
- [ ] `shelly energy reset <device> [id]` - Reset energy counters

### 11.3 Power Monitoring (PM/PM1 Components)
- [ ] `shelly power list <device>` - List power meters (PM components)
- [ ] `shelly power status <device> [id]` - Current power status
  - Shows: power (W), energy (Wh), voltage, current
- [ ] `shelly power history <device> [id]` - Power history
- [ ] `shelly power calibrate <device> [id]` - Calibrate power meter

### 11.4 Aggregated Energy Dashboard
- [ ] `shelly energy dashboard` - Summary of all energy meters
  - Total power consumption
  - Per-device breakdown
  - Cost estimation (if rate configured)
- [ ] `shelly energy compare` - Compare energy usage
  - Flags: `--devices`, `--period`

### 11.5 Metrics Export
- [ ] `shelly metrics prometheus` - Start Prometheus exporter
  - Flags: `--port`, `--devices`, `--interval`
  - Exports: power, energy, voltage, current, temperature
- [ ] `shelly metrics json` - Output metrics as JSON
  - For integration with other monitoring tools
- [ ] `shelly metrics influxdb` - InfluxDB line protocol output
  - Flags: `--host`, `--database`, `--retention`

---

## Phase 12: Alias System (gh-style)

### 12.1 Alias Commands
- [x] `shelly alias list` - List all aliases
- [x] `shelly alias set <name> <command>` - Create alias
- [x] `shelly alias delete <name>` - Delete alias
- [x] `shelly alias import <file>` - Import aliases from file
- [x] `shelly alias export <file>` - Export aliases to file

### 12.2 Device Aliases
- [ ] `shelly alias device <name> <target>` - Alias device name (requires shelly-go)

### 12.3 Alias Expansion
- [x] Support argument interpolation with `$1`, `$2`, `$@`
- [x] Support shell command aliases (prefixed with `!`)
- [ ] Implement alias expansion in command parsing (requires device commands)

---

## Phase 13: Plugin/Extension System (gh-style)

### 13.1 Plugin Infrastructure (PATH-based, gh-style)
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

### 13.2 Plugin Commands
- [x] `shelly extension list` - List installed extensions
- [x] `shelly extension install <name|url>` - Install extension
  - Support GitHub repos: `gh:user/shelly-pluginname`
  - Support local files
- [x] `shelly extension remove <name>` - Remove extension
- [ ] `shelly extension upgrade [name]` - Upgrade extension(s)
- [x] `shelly extension create <name>` - Scaffold new extension
- [x] `shelly extension exec <name> [args]` - Execute extension explicitly
- [ ] `shelly extension browse` - Browse available extensions

### 13.3 Plugin SDK
- [ ] Create `pkg/api/` public package:
  - Config access functions
  - Device registry access
  - Output formatting utilities
  - Common types and interfaces
- [ ] Document plugin development in `docs/plugins.md`
- [ ] Create example plugins in `examples/plugins/`

---

## Phase 14: TUI Dashboard

> **READ WHEN IMPLEMENTING:** See [docs/architecture.md](docs/architecture.md) for full TUI architecture patterns, component structure, and code examples.

**Architecture:** gh-dash/BubbleTea with Elm Architecture (Model/Init/Update/View)

### 14.1 Tasks
- [ ] Create `internal/tui/app.go` - Main app model
- [ ] Create `internal/tui/keys.go` - Vim-style keyboard bindings
- [ ] Create `internal/tui/styles.go` - Theme-aware styles via bubbletint

### 14.2 Components (each with model.go, view.go, update.go, styles.go)
- [ ] `components/devicelist/` - Device table (bubbles/table)
- [ ] `components/devicedetail/` - Device detail panel
- [ ] `components/monitor/` - Real-time monitoring
- [ ] `components/search/` - Filter bar (bubbles/textinput)
- [ ] `components/statusbar/` - Bottom status bar
- [ ] `components/help/` - Help overlay (glamour markdown)
- [ ] `components/tabs/` - View switching tabs
- [ ] `components/toast/` - Notifications

### 14.3 Data Layer
- [ ] `data/devices.go` - Device fetching with caching
- [ ] `data/status.go` - Batch status updates
- [ ] `data/events.go` - WebSocket event stream

### 14.4 Commands
- [ ] `shelly dash` - Launch dashboard (--refresh, --filter, --view)
- [ ] `shelly dash devices` - Device list view
- [ ] `shelly dash monitor` - Monitoring view
- [ ] `shelly dash events` - Event stream view

### 14.5 Features
- [ ] Vim-style navigation (j/k/h/l, g/G)
- [ ] Command mode (`:quit`, `:device`, `:filter`, `:theme`)
- [ ] Runtime theme switching
- [ ] Configurable keybindings via config.yaml

---

## Phase 15: Theme System

### 15.1 Theme Infrastructure
- [x] Create `internal/theme/theme.go`:
  - Integration with lrstanley/bubbletint
  - Theme loading from config
  - Runtime theme switching
- [ ] Create `internal/theme/registry.go`:
  - Built-in theme definitions
  - Custom theme loading from files
- [ ] Define theme structure:
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
- [ ] `shelly theme export <file>` - Export current theme
- [ ] `shelly theme import <file>` - Import custom theme

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
- [ ] Implement dynamic completions for:
  - Device names (from registry)
  - Device addresses (from discovery cache)
  - Group names
  - Script names (per device)
  - Schedule IDs (per device)
  - Alias names
  - Theme names
  - Extension names
- [ ] Cache completion data for performance
- [x] Support completion descriptions

### 16.3 Completion Installation
- [x] `shelly completion install` - Install completions to shell config
  - Auto-detect shell
  - Add source line to rc file

---

## Phase 17: Self-Update Command

### 17.1 Update System
- [ ] Create `internal/cli/upgrade/upgrade.go`:
  - Check for new releases via GitHub API
  - Download and verify checksums
  - Replace binary in-place
- [ ] `shelly upgrade` - Upgrade to latest version
  - Flags: --check (just check, don't upgrade)
  - Flags: --version (specific version)
  - Flags: --pre-release (include pre-releases)
- [ ] `shelly version` - Show current version
  - Include available update info

### 17.2 Version Checking
- [ ] Check for updates on startup (configurable)
  - Rate limit to once per 24 hours
  - Cache check result
- [ ] Display update notification if available
- [ ] Respect `SHELLY_NO_UPDATE_CHECK` environment variable

---

## Phase 18: KVS Commands (Gen2+)

### 18.1 KVS Management
- [ ] `shelly kvs list <device>` - List KVS keys
- [ ] `shelly kvs get <device> <key>` - Get KVS value
- [ ] `shelly kvs set <device> <key> <value>` - Set KVS value
- [ ] `shelly kvs delete <device> <key>` - Delete KVS key
- [ ] `shelly kvs export <device> <file>` - Export all KVS data
- [ ] `shelly kvs import <device> <file>` - Import KVS data

---

## Phase 19: Advanced Features from shelly-manager

### 19.1 Configuration Templates
- [ ] `shelly template list` - List config templates
- [ ] `shelly template create <name>` - Create template from device
- [ ] `shelly template apply <name> <device>` - Apply template
  - Flags: --dry-run
- [ ] `shelly template diff <name> <device>` - Compare template to device
- [ ] `shelly template export <name> <file>` - Export template
- [ ] `shelly template import <file>` - Import template

### 19.2 Export Formats
- [ ] `shelly export json <device> [file]` - Export as JSON
- [ ] `shelly export yaml <device> [file]` - Export as YAML
- [ ] `shelly export csv <devices...> [file]` - Export devices list as CSV
- [ ] `shelly export ansible <devices...> [file]` - Export as Ansible inventory
- [ ] `shelly export terraform <devices...> [file]` - Export as Terraform config

### 19.3 Provisioning
- [ ] `shelly provision wifi <device>` - Provision WiFi settings
  - Interactive or via flags
- [ ] `shelly provision bulk <file>` - Bulk provision from config file
  - YAML file with device list and settings
  - Parallel provisioning
  - Progress reporting
- [ ] `shelly provision ble <device>` - BLE-based provisioning

### 19.4 Actions/Webhooks Helper
- [ ] `shelly action list <device>` - List available actions (Gen1)
- [ ] `shelly action set <device> <action> <url>` - Set action URL
- [ ] `shelly action clear <device> <action>` - Clear action URL
- [ ] `shelly action test <device> <action>` - Test action trigger

---

## Phase 20: Convenience Commands

### 20.1 Quick Commands
- [ ] `shelly on <device>` - Quick on (detect switch/light/cover)
- [ ] `shelly off <device>` - Quick off
- [ ] `shelly toggle <device>` - Quick toggle
- [ ] `shelly status [device]` - Quick status (all or specific)
- [ ] `shelly reboot <device>` - Quick reboot
- [ ] `shelly reset <device>` - Quick factory reset (with confirmation)

### 20.2 Interactive Mode
- [ ] `shelly interactive` - Launch interactive REPL
  - Command history
  - Tab completion
  - Persistent connection mode
- [ ] `shelly shell <device>` - Interactive shell for device
  - Execute RPC commands directly
  - Script REPL for Gen2+ devices

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

## Phase 22A: Gen1 Device Support

> **Note:** Gen1 devices use a different API than Gen2+. The shelly-go library provides full Gen1 support via the `gen1/` package.

### 22A.1 Gen1 Client Integration
- [ ] Create `internal/client/gen1.go`:
  - Wrapper for shelly-go `gen1/` package
  - Auto-detection of device generation
  - Unified interface where possible

### 22A.2 Gen1 Discovery
- [ ] Enhance `shelly discover coiot` for Gen1-specific info
- [ ] Add `--gen1-only` flag to discovery commands

### 22A.3 Gen1 Control Commands
- [ ] Extend existing commands to support Gen1 devices:
  - `shelly switch on/off/toggle/status` - Use Gen1 relay API
  - `shelly cover open/close/stop/position` - Use Gen1 roller API
  - `shelly light on/off/set/status` - Use Gen1 light API (Bulbs, Duo)
  - `shelly rgb set/status` - Use Gen1 color API (RGBW, Bulb)

### 22A.4 Gen1-Specific Commands
- [ ] `shelly gen1 settings <device>` - Get/set Gen1 settings
- [ ] `shelly gen1 actions <device>` - Manage Gen1 actions (URLs)
- [ ] `shelly gen1 status <device>` - Full Gen1 status dump
- [ ] `shelly gen1 ota <device>` - Gen1 OTA firmware update

### 22A.5 CoIoT Real-time Updates
- [ ] `shelly gen1 coiot <device>` - Subscribe to CoIoT updates
- [ ] Integrate CoIoT into TUI monitoring view

---

## Phase 22B: Sensor Commands

> **Note:** Covers environmental sensors available in shelly-go `gen2/components/`.

### 22B.1 Temperature Sensor
- [ ] `shelly sensor temperature list <device>` - List temperature sensors
- [ ] `shelly sensor temperature status <device> [id]` - Current temperature
- [ ] `shelly sensor temperature history <device> [id]` - Temperature history
  - Flags: `--period`, `--format`

### 22B.2 Humidity Sensor
- [ ] `shelly sensor humidity list <device>` - List humidity sensors
- [ ] `shelly sensor humidity status <device> [id]` - Current humidity

### 22B.3 Flood Sensor
- [ ] `shelly sensor flood list <device>` - List flood sensors
- [ ] `shelly sensor flood status <device> [id]` - Flood detection status
- [ ] `shelly sensor flood test <device> [id]` - Test flood alarm

### 22B.4 Smoke Sensor
- [ ] `shelly sensor smoke list <device>` - List smoke sensors
- [ ] `shelly sensor smoke status <device> [id]` - Smoke detection status
- [ ] `shelly sensor smoke test <device> [id]` - Test smoke alarm
- [ ] `shelly sensor smoke mute <device> [id]` - Mute alarm

### 22B.5 Illuminance Sensor
- [ ] `shelly sensor illuminance list <device>` - List illuminance sensors
- [ ] `shelly sensor illuminance status <device> [id]` - Current light level

### 22B.6 Voltmeter
- [ ] `shelly sensor voltmeter list <device>` - List voltmeters
- [ ] `shelly sensor voltmeter status <device> [id]` - Current voltage reading

### 22B.7 Combined Sensor Status
- [ ] `shelly sensor status <device>` - All sensor readings in one view
- [ ] `shelly sensor monitor <device>` - Real-time sensor monitoring
  - Updates via WebSocket
  - TUI display with graphs

---

## Phase 22C: Thermostat Commands

> **Note:** Thermostat support via shelly-go `gen2/components/thermostat.go`.

### 22C.1 Thermostat Control
- [ ] `shelly thermostat list <device>` - List thermostats
- [ ] `shelly thermostat status <device> [id]` - Thermostat status
- [ ] `shelly thermostat set <device> [id]` - Set thermostat
  - Flags: `--target-temp`, `--mode`, `--enable`
- [ ] `shelly thermostat enable <device> [id]` - Enable thermostat
- [ ] `shelly thermostat disable <device> [id]` - Disable thermostat

### 22C.2 Thermostat Schedules
- [ ] `shelly thermostat schedule list <device> [id]` - List schedules
- [ ] `shelly thermostat schedule set <device> [id]` - Set schedule
  - Interactive mode for schedule configuration

---

## Phase 23: Documentation

### 23.1 README.md
- [ ] Project overview and features
- [ ] Installation instructions (all methods)
- [ ] Quick start guide
- [ ] Configuration overview
- [ ] Link to full documentation

### 23.2 Command Documentation
- [ ] Generate `docs/commands/` from Cobra command help
- [ ] Include examples for each command
- [ ] Document all flags and options
- [ ] Add common use cases

### 23.3 Configuration Reference
- [ ] Create `docs/configuration.md`:
  - Config file format and location
  - All configuration options
  - Environment variables
  - Example configurations

### 23.4 Plugin Development Guide
- [ ] Create `docs/plugins.md`:
  - Plugin architecture overview
  - SDK documentation
  - Example plugin walkthrough
  - Publishing plugins

### 23.5 Theming Guide
- [ ] Create `docs/themes.md`:
  - Built-in themes
  - Custom theme creation
  - Theme file format
  - TUI component styling

### 23.6 Man Pages
- [ ] Generate man pages via Cobra
- [ ] Include in release artifacts

### 23.7 Documentation Site
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

## Phase 24: Examples

### 24.1 Alias Examples
- [ ] Create `examples/aliases/`:
  - power-users.yaml - Power user aliases
  - shortcuts.yaml - Common shortcuts
  - automation.yaml - Automation-focused aliases

### 24.2 Script Examples
- [ ] Create `examples/scripts/`:
  - morning-routine.sh - Morning automation
  - away-mode.sh - Away mode setup
  - energy-report.sh - Energy reporting
  - bulk-update.sh - Bulk firmware update

### 24.3 Plugin Examples
- [ ] Create `examples/plugins/`:
  - shelly-notify - Desktop notifications
  - shelly-homekit - HomeKit bridge info
  - shelly-prometheus - Prometheus metrics

### 24.4 Config Examples
- [ ] Create `examples/config/`:
  - minimal.yaml - Minimal configuration
  - full.yaml - Full configuration with all options
  - multi-site.yaml - Multiple location setup

---

## Phase 25: Testing

> **READ WHEN IMPLEMENTING:** See [docs/testing.md](docs/testing.md) for full testing strategy, coverage targets, and code examples.

**Target:** 90%+ overall coverage

### 25.1 Tasks
- [ ] Implement unit tests per package (see docs/testing.md for targets)
- [ ] Create `internal/testutil/` with MockClient, MockServer, TestIOStreams, TestFactory
- [ ] Add integration tests with mock device server
- [ ] Add TUI tests using `charmbracelet/x/exp/teatest`
- [ ] Add E2E tests for CLI invocations
- [ ] Setup CI coverage reporting with threshold enforcement

---

## Phase 26: Polish & Release

### 26.1 Code Quality
- [ ] Run golangci-lint with all enabled linters
- [ ] Fix all linter warnings
- [ ] Ensure consistent error messages
- [ ] Review public API surface
- [ ] Verify all commands have --help text
- [ ] Check for proper context propagation

### 26.2 Performance
- [ ] Profile startup time
- [ ] Optimize config loading
- [ ] Cache discovery results
- [ ] Lazy-load TUI components

### 26.3 Accessibility
- [ ] Ensure TUI works with screen readers (where possible)
- [ ] Support --no-color for all output
- [ ] Provide text-only alternatives to TUI elements

### 26.4 Release Preparation
- [ ] Write CHANGELOG.md with all features
- [ ] Create GitHub release notes template
- [ ] Verify goreleaser builds all platforms
- [ ] Test installation instructions
- [ ] Create v0.1.0 tag

---

## Shelly-Go Library Coverage

> **Full details:** See [docs/shelly-go-coverage.md](docs/shelly-go-coverage.md)

**Summary:** 9 feature areas complete, ~25 remaining across Phases 5-22C.

| Priority | Features | Status |
|----------|----------|--------|
| **Critical** | Firmware, Cloud API | Phases 6, 9 |
| **High** | Energy, Scripts, Schedules, KVS, MQTT, Webhooks, WiFi | Phases 5, 7, 8, 11, 18 |
| **Medium** | BTHome, Zigbee, Gen1, Sensors, Thermostat | Phases 21, 22A-C |
| **Low** | Matter, LoRa, Z-Wave, ModBus, Virtual | Phases 21, 22 |

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
