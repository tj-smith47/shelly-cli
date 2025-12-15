# Shelly CLI - Comprehensive Implementation Plan

> ⛔ **STOP. READ [CLAUDE.md](CLAUDE.md) FIRST.** All mandatory workflow rules are there.

---

## Session Quick Start

**For LLM sessions:** This section contains everything needed to start working immediately.

### What This Project Is
A production-ready Cobra CLI for Shelly smart home devices, targeting adoption by ALLTERCO Robotics as the official Shelly CLI. Built on the `shelly-go` library (at `/db/appdata/shelly-go`).

### Before You Code
1. **Read [CLAUDE.md](CLAUDE.md)** - Mandatory workflow rules (factory pattern, IOStreams, error handling)
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

**Last Updated:** 2025-12-15 | **Current Phase:** 27 - Examples | **shelly-go:** v0.1.6

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
| 13.1 | Plugin System | ✅ Complete |
| 13.2 | Plugin Manifest System | ✅ Complete |
| 14 | TUI Dashboard | ✅ Complete |
| 15 | Theme System | ✅ Complete |
| 16 | Shell Completions | ✅ Complete |
| 17 | Update Command | ✅ Complete |
| 18-20 | Advanced Features (Monitor, Debug) | ✅ Complete |
| 21 | BTHome/Zigbee/LoRa Commands | ✅ Complete |
| 22 | Matter Commands | ✅ Complete |
| 23 | Gen1 Device Support | ✅ Complete |
| 24 | Sensor Commands | ✅ Complete |
| 25 | Thermostat Commands | ✅ Complete |
| 26 | Documentation | ✅ Complete |
| 27 | Examples | ⏳ Pending |
| 28 | Testing (90%+ coverage) | ⏳ Pending |
| 29 | Innovative Commands (~25) | ⏳ Pending |
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
| `--clipboard` | | bool | false | Copy command output to clipboard (TODO: implement) |

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

## Phase 0: Architecture Refactoring ✅ COMPLETE

See `.claude/COMPLETED.md` for details.

---

## Phases 1-6: Foundation & Core Commands ✅ COMPLETE

See `.claude/COMPLETED.md` for details.

---

## Phase 7: Script Commands ✅ COMPLETE (except templates)

### 7.1 Script Management ✅ COMPLETE

### 7.2 Script Templates
- [ ] `shelly script template list` - List available script templates
- [ ] `shelly script template show <name>` - Show script template
- [ ] `shelly script template install <device> <name>` - Install template to device
  - Flags: --configure (interactive configuration)

---

## Phase 8: Schedule Commands ✅ COMPLETE

See `.claude/COMPLETED.md` for details. (Note: `--timespec "@sunrise"` and `"@sunset"` already supported)

---

## Phases 9-12: Cloud, Backup, Monitoring, Alias ✅ COMPLETE

See `.claude/COMPLETED.md` for details.

---

## Phase 13: Plugin System ✅ COMPLETE

See `.claude/COMPLETED.md` for details and `docs/plugins.md` for plugin development guide.

Example plugin: `examples/plugins/shelly-notify/`

### 13.2 Plugin Manifest System ✅ Complete

See `docs/plugins.md` "Plugin Manifest System" section for documentation.

- [x] Create `internal/plugins/manifest.go` with Manifest, Source, Binary types
- [x] Create `internal/plugins/migrate.go` for automatic migration of existing plugins
- [x] Update `internal/cmd/plugin/install/` to create manifest on install
- [x] Update `internal/cmd/plugin/upgrade/` to read manifest for source info
- [x] Update `internal/cmd/plugin/list/` to show manifest metadata (version, source, install date)
- [x] Update `internal/cmd/plugin/remove/` to remove entire plugin directory

---

## Phases 14-19: TUI, Themes, Completions, Update, KVS, Advanced Features ✅ COMPLETE

See `.claude/COMPLETED.md` for details.

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

**Code Audit Notes (Phase 20.2):**
- Duplication identified between `on/off/toggle` quick commands and `interactive` control helpers
- Both implement component control logic (switch/light/rgb/cover on/off/toggle)
- Patterns differ slightly: quick commands use spinner + return first error; interactive uses inline feedback
- **Recommended refactor:** Extract shared `ComponentController` helper to `internal/control/` package
- Address in Phase 28 (Testing) with comprehensive test coverage

### 20.3 Debug Commands ✅ Complete
- [x] `shelly debug log <device>` - Get device debug log (Gen1)
  - Aliases: logs, debug-log
  - Shows workaround instructions for Gen1 devices (curl http://<ip>/debug/log)
  - Gen2+ redirected to debug rpc commands
- [x] `shelly debug rpc <device> <method> [params]` - Raw RPC call
  - Aliases: call, invoke
  - Flags: --raw (unformatted output)
  - Accepts JSON params as third argument
- [x] `shelly debug coiot <device>` - Show CoIoT status
  - Aliases: coap
  - Flags: --json
  - Extracts coiot/device/sys sections from Sys.GetConfig
- [x] `shelly debug websocket <device>` - WebSocket debug connection
  - Aliases: ws, events
  - Shows Ws.GetConfig and Ws.GetStatus (Gen2+)
  - Falls back to Sys.GetConfig for ws section
- [x] `shelly debug methods <device>` - List available RPC methods
  - Aliases: list-methods, lm
  - Flags: --filter (filter by name), --json
  - Groups methods by namespace

**Code Audit Notes (Phase 20.3):**
- JSON printing pattern (`json.MarshalIndent` + print) appears in rpc.go, methods.go, coiot.go, websocket.go
- websocket.go extracts `printJSONResult` helper; similar pattern could be shared across debug commands
- Sys.GetConfig section extraction appears in both coiot.go and websocket.go (fallback)
- **Recommended refactor:** Extract shared `debugutil` package with `PrintJSON()` and `ExtractConfigSection()` helpers
- Address in Phase 28 (Testing) with comprehensive test coverage

---

## Phase 21: BTHome/Zigbee/LoRa Commands

### 21.1 BTHome Commands
- [x] `shelly bthome list <device>` - List BTHome devices
  - Aliases: ls, devices
  - Flags: --json
  - Parses Shelly.GetStatus for bthomedevice:* components
- [x] `shelly bthome add <device>` - Start BTHome discovery
  - Aliases: discover, scan
  - Flags: --duration (scan time), --addr (add specific device), --name
  - Uses BTHome.StartDeviceDiscovery or BTHome.AddDevice
- [x] `shelly bthome remove <device> <id>` - Remove BTHome device
  - Aliases: rm, delete, del
  - Flags: --yes (skip confirmation)
  - Uses BTHome.DeleteDevice
- [x] `shelly bthome status <device> [id]` - BTHome device status
  - Aliases: st, info
  - Flags: --json
  - Shows BTHome.GetStatus (component) or BTHomeDevice.GetStatus (specific device)

### 21.2 Zigbee Commands (Gen4 gateways)
- [x] `shelly zigbee status <device>` - Zigbee network status
  - Aliases: st, info
  - Flags: --json
  - Shows Zigbee.GetConfig and Zigbee.GetStatus
- [x] `shelly zigbee pair <device>` - Start pairing mode
  - Aliases: join, connect
  - Flags: --timeout (pairing timeout)
  - Enables Zigbee and starts Zigbee.StartNetworkSteering
- [x] `shelly zigbee remove <device>` - Leave Zigbee network
  - Aliases: leave, disconnect, rm
  - Flags: --yes (skip confirmation)
  - Disables Zigbee which causes device to leave network
- [x] `shelly zigbee list` - List Zigbee-capable devices
  - Aliases: ls, devices
  - Flags: --json
  - Scans config devices for Zigbee support via Zigbee.GetConfig

### 21.3 LoRa Commands
- [x] `shelly lora status <device>` - LoRa add-on status
  - Aliases: st, info
  - Flags: --id (component ID), --json
  - Shows LoRa.GetConfig and LoRa.GetStatus (RSSI, SNR)
- [x] `shelly lora config <device>` - Configure LoRa settings
  - Aliases: set, configure
  - Flags: --id, --freq (Hz), --bw (bandwidth), --dr (data rate), --power (dBm)
  - Uses LoRa.SetConfig
- [x] `shelly lora send <device> <data>` - Send LoRa message
  - Aliases: tx, transmit
  - Flags: --id, --hex (data is hex)
  - Base64 encodes data for LoRa.SendBytes API

**Code Audit Notes (Phase 21):**
- JSON marshal/unmarshal pattern for conn.Call() results is repeated across all files
- Complexity warnings (gocyclo) in list.go and status.go due to conditional rendering logic
- Similar struct patterns for config/status responses could be consolidated
- **Recommended refactor:** Extract shared `rpcutil.UnmarshalResult()` helper
- Address in Phase 28 (Testing) with comprehensive test coverage

---

## Phase 22: Matter Commands

### 22.1 Matter Management
- [x] `shelly matter status <device>` - Matter status
  - Aliases: st, info
  - Flags: --json
  - Shows Matter.GetConfig (enabled) and Matter.GetStatus (commissionable, fabrics count)
- [x] `shelly matter enable <device>` - Enable Matter
  - Aliases: on, activate
  - Uses Matter.SetConfig with enable=true
- [x] `shelly matter disable <device>` - Disable Matter
  - Aliases: off, deactivate
  - Uses Matter.SetConfig with enable=false
- [x] `shelly matter reset <device>` - Reset Matter config
  - Aliases: factory-reset, unpair
  - Flags: --yes (skip confirmation)
  - Uses Matter.FactoryReset
- [x] `shelly matter code <device>` - Show pairing code
  - Aliases: pairing, qr
  - Flags: --json
  - Tries Matter.GetCommissioningCode, falls back to web UI hint

**Code Audit Notes (Phase 22):**
- Follows same marshal/unmarshal pattern as Phase 21 commands
- Complexity in code.go due to fallback logic for commissioning info
- Matter API is simpler (enable/disable/reset) compared to Zigbee/BTHome
- Address in Phase 28 (Testing) with comprehensive test coverage

---

## Phase 23: Gen1 Device Support ✅

> **Note:** Gen1 devices use a different API than Gen2+. Uses direct HTTP REST API calls (Gen1 doesn't use RPC).

### 23.1 Gen1 Client Integration ✅
- [x] Create `internal/client/gen1.go`:
  - Wrapper for shelly-go `gen1/` package
  - Auto-detection of device generation
  - Unified interface where possible

### 23.2 Gen1 Discovery ✅
- [x] Enhance `shelly discover coiot` for Gen1-specific info
- [x] Add `--gen1-only` flag to discovery commands

### 23.3 Gen1 Control Commands ✅
- [x] `shelly gen1 relay on/off/toggle/status` - Gen1 relay control
- [x] `shelly gen1 roller open/close/stop/position/calibrate/status` - Gen1 roller shutter control
- [x] `shelly gen1 light on/off/toggle/brightness/status` - Gen1 light control (Bulbs, Duo)
- [x] `shelly gen1 color on/off/toggle/set/gain/status` - Gen1 RGBW control (RGBW2, Bulb)

### 23.4 Gen1-Specific Commands ✅
- [x] `shelly gen1 status <device>` - Full Gen1 status dump
  - Aliases: st, info
  - Flags: --json
  - Shows WiFi, relays, meters, rollers, temperature, uptime from /status endpoint
- [x] `shelly gen1 settings <device>` - Show Gen1 device settings
  - Aliases: config, cfg
  - Flags: --json
  - Shows device info, name, firmware, cloud, CoIoT, MQTT, relay settings from /settings endpoint
- [x] `shelly gen1 actions <device>` - List Gen1 action URLs
  - Aliases: urls, webhooks
  - Flags: --json
  - Parses action URLs from relays and inputs in settings
- [x] `shelly gen1 ota check/update` - Gen1 OTA firmware management
  - check: Check for firmware updates
  - update: Apply firmware update with optional custom URL

### 23.5 CoIoT Real-time Updates ✅
- [x] `shelly gen1 coiot` - Monitor Gen1 devices via CoIoT multicast
  - Aliases: monitor, watch, listen
  - Flags: --timeout, --follow
  - Listens on 224.0.1.187:5683 for device broadcasts
- [ ] Integrate CoIoT into TUI monitoring view (Phase 29)

**Code Audit Notes (Phase 23):**
- HTTP fetch pattern (`fetchSettings`, `fetchStatus`) duplicated across all gen1 commands
- Similar address resolution, http:// prefix handling, auth setup, response body close pattern
- **Recommended refactor:** Extract shared `gen1http` package with `FetchEndpoint(ctx, ios, device, path)` helper
- Display helper functions follow consistent pattern: `display<Section>(ios, data)`
- Type assertion pattern with `hasX` variables for errcheck compliance
- Address in Phase 28 (Testing) with comprehensive test coverage

---

## Phase 24: Sensor Commands ✅

> **Note:** Covers environmental sensors available in shelly-go `gen2/components/`.

### 24.1 Temperature Sensor ✅
- [x] `shelly sensor temperature list <device>` - List temperature sensors
- [x] `shelly sensor temperature status <device> [id]` - Current temperature
- [ ] ~~`shelly sensor temperature history <device> [id]` - Temperature history~~ **SKIPPED** - No historical data API exists for sensors (unlike EMData for energy meters)

### 24.2 Humidity Sensor ✅
- [x] `shelly sensor humidity list <device>` - List humidity sensors
- [x] `shelly sensor humidity status <device> [id]` - Current humidity

### 24.3 Flood Sensor ✅
- [x] `shelly sensor flood list <device>` - List flood sensors
- [x] `shelly sensor flood status <device> [id]` - Flood detection status
- [x] `shelly sensor flood test <device> [id]` - Test flood alarm (provides guidance, no API method)

### 24.4 Smoke Sensor ✅
- [x] `shelly sensor smoke list <device>` - List smoke sensors
- [x] `shelly sensor smoke status <device> [id]` - Smoke detection status
- [x] `shelly sensor smoke test <device> [id]` - Test smoke alarm (provides guidance, no API method)
- [x] `shelly sensor smoke mute <device> [id]` - Mute alarm

### 24.5 Illuminance Sensor ✅
- [x] `shelly sensor illuminance list <device>` - List illuminance sensors
- [x] `shelly sensor illuminance status <device> [id]` - Current light level (with human-readable descriptions)

### 24.6 Voltmeter ✅
- [x] `shelly sensor voltmeter list <device>` - List voltmeters
- [x] `shelly sensor voltmeter status <device> [id]` - Current voltage reading

### 24.7 Combined Sensor Status ✅
- [x] `shelly sensor status <device>` - All sensor readings in one view
- [ ] `shelly sensor monitor <device>` - Real-time sensor monitoring (Phase 29 TUI improvements)

**Code Audit Notes (Phase 24):**
- Factory pattern with `*cmdutil.Factory`
- IOStreams via `f.IOStreams()` methods
- JSON marshal/unmarshal pattern for API responses
- Component key prefix parsing from `Shelly.GetStatus` response
- Helper functions for display to reduce cyclomatic complexity
- Test commands provide guidance when no API method exists

---

## Phase 25: Thermostat Commands ✅ COMPLETE

See `.claude/COMPLETED.md` for details.

---

## Phase 26: Documentation ✅ Complete

### 26.1 README.md ✅
- [x] Project overview and features
- [x] Installation instructions (all methods)
- [x] Quick start guide
- [x] Configuration overview
- [x] Link to full documentation

### 26.2 Command Documentation ✅
- [x] Generate `docs/commands/` from Cobra command help (296 files)
- [x] Include examples for each command
- [x] Document all flags and options
- [x] Created `cmd/docgen/` tool for regeneration

### 26.3 Configuration Reference ✅
- [x] Create `docs/configuration.md`:
  - Config file format and location
  - All configuration options
  - Environment variables
  - Example configurations

### 26.4 Plugin Development Guide ✅
- [x] `docs/plugins.md` already existed and was comprehensive
- [x] Added Plugin Manifest System section
- [x] Plugin architecture, SDK, examples, publishing all documented

### 26.5 Theming Guide ✅
- [x] Create `docs/themes.md`:
  - Built-in themes (280+)
  - Custom theme creation
  - Theme file format
  - TUI component styling

### 26.6 Man Pages ✅
- [x] Generate man pages via Cobra (296 files in `docs/man/`)
- [x] Created `cmd/mangen/` tool for regeneration
- [x] Added `make manpages` target

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
- [ ] Currently in `cfg/` with a full example and schema
  - Migrate config example to `examples/config/` (replaces one or both of next 2), leave schema
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

- [x] `shelly feedback` - Integrated issue reporting to GitHub
  - Interactive mode or `--type bug|feature|device`
  - Auto-populates system info, CLI version, device context
  - Flags: `--attach-log`, `--screenshot`
- [x] `shelly doctor` - Comprehensive diagnostic health check (inspired by `brew doctor`)
  - Checks: CLI version, config validity, network reachability, device health, firmware status, cloud auth
  - Flags: `--network`, `--devices`, `--full`
- [x] `shelly diff` - Visual diff between device configurations
  - `shelly diff <device1> <device2>` or `shelly diff <device> <backup.json>`

### 29.2 Power User Features

- [x] `shelly benchmark` - Device performance testing
  - Measures ping, API latency, toggle response times
  - `shelly benchmark <device> [--iterations N]`

### 29.3 Developer & Integration

- [x] `shelly webhook server` - Built-in local webhook receiver for testing
  - Auto-configures devices to send webhooks to this server
  - `shelly webhook server [--port 8080] [--log]`
- [x] `shelly export terraform` - Infrastructure as Code export
  - `shelly export terraform|ansible|homeassistant --all`


### 29.5 Fleet Management (Enterprise)

- [x] `shelly fleet` - Cloud-based fleet management (uses integrator.FleetManager)
  - `shelly fleet connect [--host]` - Connect to Shelly Cloud hosts
  - `shelly fleet status` - Fleet-wide device status and health
  - `shelly fleet stats` - Aggregate statistics (online/offline counts)
  - `shelly fleet health` - Device health monitoring (unhealthy devices)
  - `shelly fleet on|off|toggle --group <name>` - Group control via cloud
  - Note: Different from `batch` - uses cloud WebSocket, not local HTTP
- [x] `shelly audit` - Security audit for devices
  - Checks: default credentials, outdated firmware, network exposure
  - `shelly audit [--check-auth] [--check-firmware] [--check-network]`
- [x] `shelly report` - Generate professional reports
  - `shelly report energy|devices|audit|usage --format pdf|html|json`


### 29.8 Convenience Shortcuts

- [x] `shelly open` - Open device in browser
  - `shelly open <device> [--cloud] [--docs <component>]`
- [x] `shelly qr` - Generate QR codes
  - `shelly qr <device> [--wifi] [--export file.png]`

### 29.10 Networking & Connectivity

- [x] `shelly ping` - Enhanced ping with latency stats
  - `shelly ping <device> [--all] [--continuous] [--graph]`

### 29.11 Security & Auth (Enhanced)

- [x] `shelly auth rotate` - Rotate device credentials
- [x] `shelly auth test` - Test auth connectivity
- [x] `shelly auth export|import` - Credential export/import
- [x] `shelly cert` - Certificate management (Gen2+ Shelly.PutUserCA)
  - `shelly cert show|install <device>`

### 29.12 Scripting & Automation

- [x] `shelly wait` - Wait for condition
  - `shelly wait <device> --state on|--online [--timeout 30s]`

### 29.13 Data & Export

- [x] `shelly sync` - Cloud synchronization
  - `shelly sync cloud [--push|--pull]`

### 29.14 Debugging & Development

- [x] `shelly mock` - Mock device mode for testing without hardware
  - `shelly mock create|list|scenario <name>`


### 29.17 Notifications & Alerts

- [x] `shelly alert` - Alert management
  - `shelly alert create|list|test|snooze <name>`
  - `--condition "power > 1000W" --notify email:admin@...`


### 29.19 Meta & CLI Management

> **Note:** `shelly update` is defined in Phase 17. These are additional config/meta commands.

- [x] `shelly config edit` - Open config in $EDITOR
- [x] `shelly cache` - Cache management
  - `shelly cache clear|stats|warm [--discovery]`
- [x] `shelly log` - CLI logging
  - `shelly log show|tail|level|export`

### 29.20 Fun & Easter Eggs

- [x] `shelly party` - Party mode (cycle colors on RGB devices)
  - `shelly party [--mode rainbow|disco] [--duration 5m]`
- [x] `shelly sleep` - Gradual lights off
  - `shelly sleep [--duration 30m] [--except bedroom]`
- [x] `shelly wake` - Gradual lights on
  - `shelly wake [--simulate sunrise] [--duration 15m]`

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

### 30.5 Documentation Site (Post-Release)
- [ ] Create docs site using Hugo
- [ ] Deploy to GitHub Pages
- [ ] Add `shelly docsite` command - opens documentation site in browser
- [ ] Include:
  - Getting started guide
  - Command reference (generated from Cobra)
  - Configuration guide
  - Plugin development guide
  - TUI usage guide
  - Examples and tutorials
  - Terminal GIFs using Charm VHS

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
| **Innovative** | ~25 commands for QoL, fleet, security, scripting | Phase 29 |

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
- [ ] Phase 29 innovative commands: `feedback`, `doctor`, `watch`, `audit`, `fleet`
- [ ] Home Assistant, Node-RED, OpenHAB integration commands
- [ ] Demo mode working without real devices (for showcasing)
- [ ] Professional branding (logo, ASCII banner)
