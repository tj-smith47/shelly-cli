# Shelly CLI - Comprehensive Implementation Plan

> **CRITICAL:** NEVER mark any task as "optional", "nice to have", "not necessary", or "future improvement". All tasks are REQUIRED. Do not defer work without explicit user approval. If context limits are reached, split the work into multiple sessions - do not skip tasks.

> **IMPORTANT:** All future session work derives from this PLAN.md file. No additional session files should be created. Check off completed work and add concise context notes only when necessary to save investigation time.

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

**Last Updated:** 2025-12-10
**Phase:** Blocked - waiting for shelly-go library release
**Completed:** Phases 1, 2, 12, 13, 15, 16
**Blocked:** Phases 3-11, 18-22 (require shelly-go)

Once shelly-go is published: `go get github.com/tj-smith47/shelly-go@latest`

---

## Phase 1: Project Foundation

### 1.1 Repository Setup
- [x] Initialize Go module: `go mod init github.com/tj-smith47/shelly-cli`
- [x] Configure `go.mod` with Go 1.25.5 minimum version
- [ ] Add shelly-go library dependency: `github.com/tj-smith47/shelly-go` (blocked - not released)
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
│   ├── cli/
│   │   ├── root.go                 # Root command
│   │   ├── device/                 # Device management commands
│   │   ├── control/                # Device control commands
│   │   ├── config/                 # Configuration commands
│   │   ├── discovery/              # Discovery commands
│   │   ├── firmware/               # Firmware commands
│   │   ├── script/                 # Script commands
│   │   ├── schedule/               # Schedule commands
│   │   ├── cloud/                  # Cloud commands
│   │   ├── alias/                  # Alias commands (gh-style)
│   │   ├── extension/              # Plugin/extension commands
│   │   └── completion/             # Shell completion
│   ├── tui/
│   │   ├── dash/                   # Main dashboard
│   │   ├── device/                 # Device detail view
│   │   ├── monitor/                # Real-time monitoring
│   │   ├── components/             # Reusable TUI components
│   │   └── theme/                  # Theme management
│   ├── config/
│   │   ├── config.go               # Viper configuration
│   │   ├── aliases.go              # Alias management
│   │   ├── devices.go              # Device registry
│   │   └── profiles.go             # User profiles
│   ├── output/
│   │   ├── format.go               # Output formatters (json, yaml, table, text)
│   │   ├── table.go                # Table rendering
│   │   └── color.go                # Color utilities
│   ├── plugins/
│   │   ├── loader.go               # Plugin discovery and loading
│   │   ├── executor.go             # Plugin execution
│   │   └── registry.go             # Plugin registry
│   └── version/
│       └── version.go              # Version info (ldflags)
├── pkg/
│   └── api/                        # Public API for plugins
├── docs/
│   ├── commands/                   # Command documentation
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
- [ ] Create `internal/config/aliases.go`:
  - Alias struct (Name, Command, Shell bool)
  - Add, Remove, List, Get functions
  - Validate alias names (no conflicts with built-in commands)
- [ ] Create `internal/config/devices.go`:
  - Device struct (Name, Address, Auth, Generation, etc.)
  - Register, Unregister, List, Get functions
  - Device groups support

### 2.4 Output Formatting
- [ ] Create `internal/output/format.go`:
  - Formatter interface
  - JSON, YAML, Table, Text implementations
  - Auto-detect format from config/flags
- [ ] Create `internal/output/table.go`:
  - Table renderer using lipgloss
  - Column alignment and wrapping
  - Header styling
  - Row alternation colors
- [ ] Create `internal/output/color.go`:
  - Theme-aware color utilities
  - Status colors (success, warning, error, info)
  - Device state colors (online, offline, updating)

---

## Phase 3: Device Management Commands

### 3.1 Discovery Commands
- [ ] `shelly discover` - Discover devices on network
  - Flags: --timeout, --network (subnet), --mdns, --ble, --coiot
  - Output discovered devices with name, IP, generation, type
  - Progress indicator during discovery
  - Filter by generation, type
- [ ] `shelly discover mdns` - mDNS discovery only
- [ ] `shelly discover ble` - BLE discovery only
- [ ] `shelly discover coiot` - CoIoT discovery (Gen1)
- [ ] `shelly discover scan <subnet>` - Subnet scan

### 3.2 Device Registry Commands
- [ ] `shelly device list` - List registered devices
  - Flags: --all, --online, --offline, --generation, --type
  - Table output with status indicators
- [ ] `shelly device add <name> <address>` - Add device to registry
  - Flags: --auth, --generation (auto-detect if omitted)
  - Validate connectivity before adding
- [ ] `shelly device remove <name>` - Remove device from registry
  - Confirmation prompt (--yes to skip)
- [ ] `shelly device rename <old> <new>` - Rename device
- [ ] `shelly device info <name|address>` - Show device details
  - Device info, config, status
  - Component list
  - Network info
- [ ] `shelly device status <name|address>` - Show device status
  - Real-time status with auto-refresh option
- [ ] `shelly device ping <name|address>` - Check device connectivity
- [ ] `shelly device reboot <name|address>` - Reboot device
- [ ] `shelly device factory-reset <name|address>` - Factory reset
  - Strong confirmation required (--yes --confirm)

### 3.3 Device Group Commands
- [ ] `shelly group list` - List device groups
- [ ] `shelly group create <name>` - Create group
- [ ] `shelly group delete <name>` - Delete group
- [ ] `shelly group add <group> <device>` - Add device to group
- [ ] `shelly group remove <group> <device>` - Remove device from group
- [ ] `shelly group members <name>` - List group members

---

## Phase 4: Device Control Commands

### 4.1 Switch Control
- [ ] `shelly switch list [device]` - List switches
- [ ] `shelly switch on <device> [id]` - Turn switch on
  - Flags: --timer (auto-off after N seconds)
- [ ] `shelly switch off <device> [id]` - Turn switch off
- [ ] `shelly switch toggle <device> [id]` - Toggle switch
- [ ] `shelly switch status <device> [id]` - Show switch status

### 4.2 Cover/Roller Control
- [ ] `shelly cover list [device]` - List covers
- [ ] `shelly cover open <device> [id]` - Open cover
- [ ] `shelly cover close <device> [id]` - Close cover
- [ ] `shelly cover stop <device> [id]` - Stop cover
- [ ] `shelly cover position <device> [id] <percent>` - Set position
- [ ] `shelly cover calibrate <device> [id]` - Start calibration
- [ ] `shelly cover status <device> [id]` - Show cover status

### 4.3 Light Control
- [ ] `shelly light list [device]` - List lights
- [ ] `shelly light on <device> [id]` - Turn light on
  - Flags: --brightness, --color-temp, --transition
- [ ] `shelly light off <device> [id]` - Turn light off
- [ ] `shelly light toggle <device> [id]` - Toggle light
- [ ] `shelly light set <device> [id]` - Set light parameters
  - Flags: --brightness, --color-temp, --rgb, --transition
- [ ] `shelly light status <device> [id]` - Show light status

### 4.4 RGB/RGBW Control
- [ ] `shelly rgb on <device> [id]` - Turn RGB on
- [ ] `shelly rgb off <device> [id]` - Turn RGB off
- [ ] `shelly rgb set <device> [id>` - Set RGB color
  - Flags: --red, --green, --blue, --white, --gain, --effect
- [ ] `shelly rgb status <device> [id]` - Show RGB status

### 4.5 Input Status
- [ ] `shelly input list [device]` - List inputs
- [ ] `shelly input status <device> [id]` - Show input status
- [ ] `shelly input trigger <device> [id]` - Manually trigger input event

### 4.6 Batch Operations
- [ ] `shelly batch on <devices...>` - Turn on multiple devices
- [ ] `shelly batch off <devices...>` - Turn off multiple devices
- [ ] `shelly batch toggle <devices...>` - Toggle multiple devices
- [ ] `shelly batch command <command> <devices...>` - Run command on multiple devices
  - Flags: --parallel, --timeout, --continue-on-error

### 4.7 Scene Management
- [ ] `shelly scene list` - List saved scenes
- [ ] `shelly scene create <name>` - Create scene from current state
  - Flags: --devices (select which devices)
- [ ] `shelly scene delete <name>` - Delete scene
- [ ] `shelly scene activate <name>` - Activate scene
- [ ] `shelly scene show <name>` - Show scene details
- [ ] `shelly scene export <name> <file>` - Export scene to file
- [ ] `shelly scene import <file>` - Import scene from file

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

### 11.1 Real-time Monitoring
- [ ] `shelly monitor <device>` - Real-time status monitoring
  - Auto-refresh with configurable interval
  - Color-coded status changes
- [ ] `shelly monitor power <device>` - Monitor power consumption
- [ ] `shelly monitor events <device>` - Monitor device events
- [ ] `shelly monitor all` - Monitor all registered devices

### 11.2 Energy Monitoring
- [ ] `shelly energy status <device>` - Current power status
- [ ] `shelly energy history <device>` - Energy history
  - Flags: --period (hour, day, week, month)
- [ ] `shelly energy export <device> <file>` - Export energy data

### 11.3 Metrics Export
- [ ] `shelly metrics prometheus` - Start Prometheus exporter
  - Flags: --port, --devices, --interval
- [ ] `shelly metrics json` - Output metrics as JSON
  - For integration with other monitoring tools

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

### 14.1 Dashboard Framework
- [ ] Create `internal/tui/dash/dash.go`:
  - Main dashboard model using Bubble Tea
  - Device list panel
  - Status overview panel
  - Quick actions panel
  - Keyboard navigation (vim-style: h/j/k/l)
- [ ] Create `internal/tui/components/`:
  - DeviceList component
  - StatusBar component
  - HelpBar component
  - SearchBar component
  - ActionMenu component
  - ProgressBar component
  - Toast/notification component

### 14.2 Dashboard Views (k9s-inspired)
- [ ] Device list view (default):
  - Table with columns: Name, IP, Type, Gen, Status, Power
  - Color-coded status indicators
  - Filter bar (/ to search)
  - Sort by column
- [ ] Device detail view:
  - Full device info
  - Component list
  - Recent events
  - Quick actions
- [ ] Monitoring view:
  - Real-time metrics
  - Sparkline graphs for power
  - Event log
- [ ] Group view:
  - Device groups
  - Bulk actions on groups

### 14.3 Dashboard Commands
- [ ] `shelly dash` - Launch main dashboard
  - Flags: --refresh (interval), --filter, --group
- [ ] `shelly dash devices` - Device list view
- [ ] `shelly dash monitor` - Monitoring view
- [ ] `shelly dash events` - Event stream view

### 14.4 Dashboard Navigation
- [ ] Implement keyboard shortcuts:
  - `q` / `Ctrl-C`: Quit
  - `?`: Help
  - `/`: Search/filter
  - `Enter`: Select/drill down
  - `Esc`: Back/cancel
  - `r`: Refresh
  - `d`: Device actions menu
  - `g`: Group actions menu
  - `s`: Settings/config
  - `1-9`: Quick switch views
  - `:`: Command mode (like vim/k9s)
- [ ] Implement command mode:
  - `:quit` - Exit
  - `:device <name>` - Go to device
  - `:group <name>` - Go to group
  - `:filter <pattern>` - Apply filter
  - `:refresh` - Force refresh
  - `:theme <name>` - Switch theme

### 14.5 Device Detail TUI
- [ ] Create `internal/tui/device/device.go`:
  - Device info panel
  - Components panel with status
  - Config panel
  - Actions panel
- [ ] Component views:
  - Switch view with toggle
  - Cover view with position slider
  - Light view with brightness/color controls
  - Input view with event history
  - Power view with graphs

### 14.6 Real-time Monitoring TUI
- [ ] Create `internal/tui/monitor/monitor.go`:
  - Multi-device power monitoring
  - Event stream panel
  - Status change notifications
  - WebSocket connection status

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

### 25.1 Unit Tests
- [ ] Test all command parsing and validation
- [ ] Test configuration loading and saving
- [ ] Test output formatters
- [ ] Test alias expansion
- [ ] Test plugin discovery and loading
- [ ] Test theme loading
- [ ] Test completion generation
- [ ] Target: 90%+ coverage

### 25.2 Integration Tests
- [ ] Test against mock Shelly devices (using testutil from shelly-go)
- [ ] Test discovery workflows
- [ ] Test backup/restore workflows
- [ ] Test TUI components (tea.Test)

### 25.3 E2E Tests
- [ ] Test CLI invocations
- [ ] Test config file scenarios
- [ ] Test plugin installation/execution
- [ ] Test completion scripts

### 25.4 Test Infrastructure
- [ ] Create `internal/testutil/`:
  - Mock device server
  - Test fixtures
  - Assertion helpers
- [ ] Use testing/synctest for time-dependent tests (Go 1.25)

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

## Dependencies

### Core Dependencies
```go
require (
    github.com/tj-smith47/shelly-go v0.1.0  // Shelly device library
    github.com/spf13/cobra v1.8.x           // CLI framework
    github.com/spf13/viper v1.18.x          // Configuration
    github.com/charmbracelet/bubbletea      // TUI framework
    github.com/charmbracelet/bubbles        // TUI components
    github.com/charmbracelet/lipgloss       // Styling (CLI + TUI)
    github.com/lrstanley/bubbletint         // Themes for ALL output (CLI + TUI)
    github.com/AlecAivazis/survey/v2        // Interactive prompts (confirmations, selections, inputs)
    github.com/briandowns/spinner           // Progress spinners for long operations
    github.com/olekukonko/tablewriter       // Table output (non-TUI)
    gopkg.in/yaml.v3                        // YAML support
)
```

### Theming Guidelines (bubbletint)
`github.com/lrstanley/bubbletint` provides theming for ALL CLI output, not just TUI:
- Table output colors
- Status indicators (success/warning/error)
- Device state colors (online/offline/updating)
- Spinner colors
- All lipgloss-styled output
- TUI dashboard components

### Spinner Guidelines (briandowns/spinner)
Use `github.com/briandowns/spinner` for long-running operations:
- Device discovery (mDNS, BLE, subnet scan)
- Firmware updates and downloads
- Backup/restore operations
- Bulk provisioning
- Cloud authentication
- Any operation taking >1 second

### Interactive Input Guidelines
All interactive user input (outside of TUI dashboard) uses `github.com/AlecAivazis/survey/v2`:
- Confirmations (e.g., factory reset, delete operations)
- Selections (e.g., choose device from list, select WiFi network)
- Text input (e.g., device names, credentials)
- Multi-select (e.g., choose devices for batch operations)
- Password input (e.g., auth credentials, encryption passwords)

### Development Dependencies
```go
require (
    github.com/stretchr/testify             // Testing
    github.com/golang/mock                  // Mocking (or use Go 1.25 testing features)
)
```

---

## Go 1.25.5 Features to Use

- [ ] `sync.WaitGroup.Go()` - Cleaner goroutine spawning
- [ ] `testing/synctest` - Virtualized time for tests
- [ ] Range over functions where applicable
- [ ] Swiss map implementation (automatic)
- [ ] GreenTea GC improvements (automatic)
- [ ] Container-aware GOMAXPROCS (automatic)

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
