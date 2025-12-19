# Architecture Reference

This document defines **where code belongs** in the shelly-cli codebase. For **how to write code** (patterns, standards, conventions), see [development.md](development.md).

---

## Quick Reference

| I need to... | Put it in... |
|--------------|--------------|
| Define a cobra command | `cmd/<domain>/<action>/<action>.go` |
| Print formatted output to terminal | `term/` |
| Format data to string (no I/O) | `output/` |
| Implement device operations | `shelly/` |
| Export data to external format | `shelly/export/` |
| Define domain types | `model/` |
| Add CLI infrastructure (flags, runners) | `cmdutil/` |
| Add shared utilities | `utils/` |
| Manage configuration | `config/` |

---

## Directory Structure

```
internal/
├── branding/           # ASCII banner utilities
│   └── branding.go     # Banner, StyledBanner(), BannerLines()
│
├── browser/            # Cross-platform URL opening
│   └── browser.go      # Browser interface, New()
│
├── client/             # Low-level shelly-go SDK wrapper
│   ├── client.go       # Client struct, Gen2 connection
│   ├── detect.go       # Device detection
│   ├── gen1.go         # Gen1Client for legacy devices
│   ├── switch.go       # SwitchComponent accessor
│   ├── light.go        # LightComponent accessor
│   ├── rgb.go          # RGBComponent accessor
│   ├── cover.go        # CoverComponent accessor
│   ├── input.go        # InputComponent accessor
│   ├── kvs.go          # KVS accessor
│   └── thermostat.go   # ThermostatComponent accessor
│
├── cmd/                # Cobra commands ONLY
│   │                   # Each command: NewCommand(f *Factory) + run()
│   │                   # NO helper functions - move them elsewhere!
│   ├── root.go
│   ├── device/
│   │   ├── device.go   # Parent command
│   │   ├── info/       # shelly device info
│   │   ├── status/     # shelly device status
│   │   └── reboot/     # shelly device reboot
│   ├── switchcmd/      # shelly switch (on/off/toggle/status)
│   ├── light/          # shelly light (on/off/set/status)
│   ├── rgb/            # shelly rgb (on/off/set/status)
│   ├── cover/          # shelly cover (open/close/stop/status)
│   ├── sensor/         # shelly sensor (temp/humidity/flood/smoke)
│   ├── energy/         # shelly energy (status/history/export)
│   ├── config/         # shelly config (get/set/edit)
│   ├── backup/         # shelly backup (create/restore/list)
│   ├── export/         # shelly export (ansible/terraform)
│   ├── scene/          # shelly scene (activate/list/import)
│   ├── schedule/       # shelly schedule (list/create/delete)
│   ├── script/         # shelly script (list/run/stop)
│   ├── batch/          # shelly batch (on/off/command)
│   ├── discover/       # shelly discover (scan/mdns/ble)
│   ├── monitor/        # shelly monitor (power/status/events)
│   ├── versioncmd/     # shelly version
│   └── ...
│
├── cmdutil/            # Command infrastructure
│   ├── factory.go      # Factory (DI container)
│   │                   #   IOStreams(), Config(), ShellyService()
│   │                   #   ResolveAddress(), ExpandTargets()
│   │                   #   GetDevice(), GetGroup(), GetAlias()
│   │
│   ├── runner.go       # Execution helpers
│   │                   #   RunWithSpinner(), RunBatch()
│   │                   #   RunStatus[T](), RunList[T]()
│   │                   #   PrintResult[T](), PrintListResult[T]()
│   │
│   ├── flags.go        # Flag helpers
│   │                   #   AddComponentIDFlag(), AddBatchFlags()
│   │                   #   AddOutputFormatFlag(), AddTimeoutFlag()
│   │
│   ├── log.go          # Logging utilities
│   │
│   └── factories/      # Generic command factories
│       ├── component.go  # NewComponentCommand (switch/light/rgb on/off/toggle)
│       ├── batch.go      # NewBatchComponentCommand
│       ├── status.go     # NewStatusCommand[T]
│       ├── sensor.go     # NewSensorCommand[T]
│       ├── cover.go      # NewCoverCommand
│       ├── config.go     # NewConfigDeleteCommand, NewConfigListCommand
│       └── delete.go     # NewDeviceDeleteCommand
│
├── completion/         # Shell completion
│   └── completion.go   # Completers for bash/zsh/fish
│
├── config/             # Configuration management
│   ├── config.go       # Config struct, Load(), Save()
│   ├── manager.go      # Manager (916 lines) - config mutations
│   ├── devices.go      # Device registry
│   ├── aliases.go      # Alias management
│   ├── scenes.go       # Scene management, ParseSceneFile()
│   ├── template.go     # Template management
│   ├── alerts.go       # Alert configuration
│   └── validation.go   # ValidateName() - shared validation
│
├── github/             # GitHub integration
│   └── releases.go     # Release checking, downloads
│
├── utils/              # Shared utilities (batch, device conversion)
│   ├── batch.go        # ResolveBatchTargets() - args/stdin/group/all
│   └── device.go       # DiscoveredDeviceToConfig(), RegisterDiscoveredDevices()
│                       # UnmarshalJSON() - RPC response helper
│
├── iostreams/          # I/O abstraction
│   ├── iostreams.go    # IOStreams struct
│   │                   #   In, Out, ErrOut (io.Reader/Writer)
│   │                   #   IsStdinTTY(), IsStdoutTTY(), ColorEnabled()
│   │                   #   Printf(), Println(), Success(), Error()
│   │                   #   StartProgress(), StopProgress()
│   │
│   ├── color.go        # Color detection, theme handling
│   ├── progress.go     # Progress indicator management
│   ├── prompt.go       # Confirm(), Prompt() - user input
│   ├── debug.go        # DebugErr() - non-fatal error logging
│   ├── log.go          # Structured logging
│   └── multiwriter.go  # Multi-line concurrent output
│
├── model/              # Domain types
│   ├── device.go       # Device struct
│   │                   #   Name, Address, Generation, Type, Model, Auth
│   │
│   ├── component.go    # Component struct, *Status types, *Config types
│   │                   #   SwitchStatus, CoverStatus, LightStatus, RGBStatus
│   │
│   ├── sensor.go       # Sensor types
│   │                   #   AlarmSensor interface (generic)
│   │                   #   AlarmSensorReading, TemperatureReading
│   │                   #   HumidityReading, IlluminanceReading
│   │                   #   VoltmeterReading, DevicePowerReading
│   │
│   └── errors.go       # Error types
│
├── output/             # Pure formatters (data → string, NO IOStreams)
│   ├── format.go       # Output format routing
│   │                   #   Format type (JSON, YAML, Table, Text, Template)
│   │                   #   FormatOutput(), PrintJSON(), PrintYAML()
│   │                   #   FormatConfigValue(), FormatSize()
│   │
│   ├── state.go        # Boolean/state renderers
│   │                   #   Case enum (CaseLower, CaseTitle, CaseUpper)
│   │                   #   RenderOnOff(), RenderYesNo(), RenderOnline()
│   │                   #   RenderActive(), RenderAlarmState(), RenderMuteState()
│   │                   #   RenderUpdateStatus(), RenderGeneration()
│   │
│   ├── devices.go      # Device formatters
│   │                   #   FormatDiscoveredDevices(), FormatAlarmSensors()
│   │                   #   FormatPower(), FormatPowerColored(), FormatEnergy()
│   │                   #   FormatEMLines(), FormatEM1Line(), FormatPMLine()
│   │
│   ├── table.go        # Table builder utilities
│   │
│   └── backup.go       # FormatBackupsTable()
│
├── plugins/            # Plugin system
│   ├── registry.go     # Registry - plugin management
│   ├── loader.go       # Plugin loading
│   ├── executor.go     # Plugin execution
│   ├── manifest.go     # Plugin metadata
│   └── migrate.go      # Plugin migration
│
├── shelly/             # Business logic service layer
│   ├── shelly.go       # Service struct
│   │                   #   Connect(), WithConnection(), RawRPC()
│   │                   #   ResolveWithGeneration() - Gen1/Gen2 detection
│   │
│   ├── resolver.go     # DeviceResolver interface
│   ├── device.go       # Device operations
│   ├── devicedata.go   # DeviceData collection
│   ├── switch.go       # Switch operations
│   ├── light.go        # Light operations
│   ├── rgb.go          # RGB operations
│   ├── cover.go        # Cover operations
│   ├── input.go        # Input operations
│   ├── sensor.go       # Sensor operations
│   ├── power.go        # Power operations
│   ├── energy.go       # Energy meter operations
│   ├── quick.go        # QuickOn(), QuickOff(), QuickToggle()
│   ├── config.go       # Config get/set operations
│   ├── backup.go       # Backup/restore operations
│   ├── schedule.go     # Schedule operations
│   ├── script.go       # Script operations
│   ├── firmware.go     # Firmware operations
│   ├── cloud.go        # Cloud operations
│   ├── monitoring.go   # MonitoringSnapshot, FetchAllSnapshots()
│   ├── kvs.go          # KVS operations
│   ├── template.go     # Template operations
│   ├── zigbee.go       # Zigbee operations
│   ├── bthome.go       # BTHome operations
│   ├── lora.go         # LoRa operations
│   ├── matter.go       # Matter operations
│   └── migrate.go      # Migration operations
│   │
│   └── export/         # Export format builders
│       ├── ansible.go    # BuildAnsibleInventory(), AnsibleInventory
│       ├── terraform.go  # BuildTerraformConfig(), TerraformDevice
│       ├── backup.go     # BackupExporter, ScanBackupFiles(), WriteBackupFile()
│       └── energy.go     # FormatEMDataCSV(), FormatEM1DataCSV()
│
├── term/               # Terminal presentation (composed displays)
│   ├── term.go         # Package doc, shared helpers (printTable, formatTemp)
│   ├── power.go        # DisplayPowerMetrics, DisplayDashboard, DisplayComparison
│   ├── sensor.go       # Generic sensor displays (partial application pattern)
│   │                   #   DisplayTemperature*, DisplayHumidity*, etc.
│   ├── component.go    # DisplaySwitch*, DisplayLight*, DisplayCover*, DisplayInput*
│   ├── device.go       # DisplayDeviceStatus, DisplayAllSnapshots
│   ├── backup.go       # DisplayBackupSummary, DisplayRestorePreview, etc.
│   ├── network.go      # DisplayWiFi*, DisplayEthernet*, DisplayMQTT*, DisplayCloud*
│   ├── firmware.go     # DisplayFirmwareStatus, DisplayFirmwareInfo
│   ├── automation.go   # DisplayScript*, DisplaySchedule*, DisplayWebhook*
│   ├── config.go       # DisplayConfigTable, DisplaySceneList, DisplayAliasList
│   ├── kvs.go          # DisplayKVS*
│   ├── discovery.go    # DisplayDiscoveredDevices, DisplayBLEDevices
│   ├── diff.go         # DisplayConfigDiffs, DisplayScriptDiffs, etc.
│   ├── event.go        # DisplayEvent, OutputEventJSON
│   └── version.go      # DisplayVersionInfo, DisplayUpdateAvailable, RunUpdateCheck
│
├── testutil/           # Testing utilities
│   ├── testutil.go     # Test helpers
│   ├── mock_client.go  # MockClient for testing
│   └── mock_resolver.go # MockResolver for testing
│
├── theme/              # Theming (bubbletint)
│   └── theme.go        # Style builders
│                       #   Bold(), Dim(), Highlight()
│                       #   StatusOK(), StatusWarn(), StatusError()
│                       #   StyledPower(), StyledEnergy()
│                       #   FalseStyle (FalseError, FalseDim)
│
├── tui/                # Terminal UI (BubbleTea)
│   ├── app.go          # Main TUI application
│   ├── keys.go         # Key bindings
│   ├── styles.go       # TUI styles
│   └── components/     # TUI components
│       ├── devicelist/   # Device list component
│       ├── devicedetail/ # Device detail component
│       ├── energy/       # Energy dashboard
│       ├── monitor/      # Monitoring component
│       ├── events/       # Events component
│       ├── search/       # Search component
│       ├── tabs/         # Tab navigation
│       ├── statusbar/    # Status bar
│       ├── help/         # Help overlay
│       ├── toast/        # Toast notifications
│       └── cmdmode/      # Command mode
│
└── version/            # Build-time version info
    └── version.go      # Version, Commit, Date variables
                        # Get(), Short(), Long(), String()
```

---

## Decision Tree: "Where Does This Go?"

```
Is it a cobra command definition?
├── Yes → cmd/<domain>/<action>/<action>.go
│         (ONLY NewCommand + run, NO helpers!)
│
└── No → Does it write to IOStreams?
    │
    ├── Yes → term/
    │         (composed display functions wrapping output/)
    │
    └── No → Is it a pure formatter (data → string)?
        │
        ├── Yes → output/
        │         (format.go, state.go, devices.go, etc.)
        │
        └── No → Is it business logic (device operations)?
            │
            ├── Yes → shelly/
            │         (or shelly/export/ for external formats)
            │
            └── No → Is it a domain type?
                │
                ├── Yes → model/
                │
                └── No → Is it CLI infrastructure?
                    │
                    ├── Yes → cmdutil/
                    │         (factory, flags, runner, factories/)
                    │
                    └── No → Is it config-related?
                        │
                        ├── Yes → config/
                        │
                        └── No → utils/
                                  (shared utilities)
```

---

## Patterns

### Display vs Format

The separation between `output/` and `term/`:

| Package | Signature | Purpose |
|---------|-----------|---------|
| `output/` | `FormatX(data) → string` | Pure formatter, no I/O |
| `term/` | `DisplayX(ios, data)` | Prints to terminal |

**Example:**

```go
// output/devices.go - pure formatter (no IOStreams)
func FormatDiscoveredDevices(devices []discovery.DiscoveredDevice) *Table {
    // Returns table data structure
}

// term/discovery.go - IOStreams printer
func DisplayDiscoveredDevices(ios *iostreams.IOStreams, devices []discovery.DiscoveredDevice) {
    table := output.FormatDiscoveredDevices(devices)
    if table == nil {
        ios.NoResults("devices")
        return
    }
    table.PrintTo(ios.Out)
    ios.Count("device", len(devices))
}
```

### Generic Factories

Command factories in `cmdutil/factories/` reduce boilerplate:

```go
// Instead of writing 200+ lines per status command:
func NewCommand(f *cmdutil.Factory) *cobra.Command {
    return factories.NewStatusCommand(f, factories.StatusOpts[SwitchStatus]{
        Component:  "Switch",
        Aliases:    []string{"st", "s"},
        SpinnerMsg: "Fetching switch status...",
        Fetcher:    fetchStatus,
        Display:    displayStatus,
    })
}
```

Available factories:
- `NewComponentCommand()` - switch/light/rgb on/off/toggle
- `NewStatusCommand[T]()` - generic status with type safety
- `NewSensorCommand[T]()` - sensor commands with list/status
- `NewBatchComponentCommand()` - batch operations
- `NewCoverCommand()` - cover open/close/stop
- `NewConfigDeleteCommand()`, `NewConfigListCommand()`

### shelly/export/ Placement

Export builders live under `shelly/export/` because they:
1. Transform `shelly.DeviceData` → external format
2. Are tightly coupled to shelly types
3. Semantically represent "shelly data export"

**Contents:**
- `ansible.go` - Ansible inventory builder
- `terraform.go` - Terraform config builder
- `backup.go` - Backup file operations
- `energy.go` - Energy data CSV export

---

## Key Exports by Package

### cmdutil/runner.go
```go
RunWithSpinner(ctx, ios, msg, action) error
RunBatch(ctx, ios, svc, targets, concurrent, action) error
RunStatus[T](ctx, f, device, id, fetcher, display) error
RunList[T](ctx, f, device, fetcher, display) error
PrintResult[T](ios, format, data, displayFn) error
PrintListResult[T](ios, format, items, displayFn) error
```

### output/state.go
```go
type Case int // CaseLower, CaseTitle, CaseUpper

RenderOnOff(on bool, c Case, fs FalseStyle) string
RenderYesNo(value bool, c Case, fs FalseStyle) string
RenderOnline(online bool, c Case) string
RenderActive(active bool, c Case, fs FalseStyle) string
RenderAlarmState(alarm bool, msg string) string
RenderMuteState(muted bool) string
RenderUpdateStatus(available bool, version string) string
```

### model/sensor.go
```go
type AlarmSensor interface {
    GetID() int
    IsAlarm() bool
    IsMuted() bool
    GetErrors() []string
}

type TemperatureReading struct { ID int; TC, TF *float64; Errors []string }
type HumidityReading struct { ID int; RH *float64; Errors []string }
type IlluminanceReading struct { ID int; Lux *float64; Errors []string }
type VoltmeterReading struct { ID int; Voltage *float64; Errors []string }
```

---

## Rules for Commands (cmd/)

Commands in `cmd/` should contain **ONLY**:
1. `NewCommand(f *cmdutil.Factory) *cobra.Command` - constructor
2. `run(ctx, f, ...)` - execution logic (simple orchestration)

**Move these OUT of cmd/:**
- Display functions → `term/`
- Format functions → `output/`
- Data collection → `shelly/`
- Type definitions → `model/`
- Shared logic → `utils/` or `cmdutil/`

---

## See Also

- [development.md](development.md) - Coding patterns and standards
- [testing.md](testing.md) - Testing strategy
- [plugins.md](plugins.md) - Plugin system
- [configuration.md](configuration.md) - Configuration management
