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
| Implement auth operations | `shelly/auth/` |
| Implement Modbus operations | `shelly/modbus/` |
| Implement provisioning operations | `shelly/provision/` |
| Implement network operations | `shelly/network/` (mqtt.go, ethernet.go, wifi.go) |
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
│   ├── version/        # shelly version
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
│   ├── flags.go        # Flag re-exports from flags/ subpackage
│   │                   #   BatchFlags, SceneFlags, ComponentFlags, etc.
│   │
│   ├── helpers.go      # Cobra helpers
│   │                   #   AddCommandsToGroup()
│   │
│   ├── log.go          # Logging utilities
│   │
│   ├── flags/          # Embeddable flag structs (subpackage)
│   │   ├── flags.go      # Core flag adders (AddTimeoutFlag, etc.)
│   │   ├── batch.go      # BatchFlags, AddBatchFlags()
│   │   ├── scene.go      # SceneFlags, AddSceneFlags()
│   │   ├── component.go  # ComponentFlags, QuickComponentFlags
│   │   ├── output.go     # OutputFlags, AddOutputFlags()
│   │   ├── confirm.go    # ConfirmFlags, AddConfirmFlags()
│   │   └── device.go     # DeviceListFlags, AddDeviceListFlags()
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
│   ├── manager.go      # Manager - config mutations
│   ├── devices.go      # Device registry
│   ├── aliases.go      # Alias management
│   │                   #   ExpandAliasArgs(), ExecuteShellAlias()
│   ├── scenes.go       # Scene management, ParseSceneFile()
│   ├── template.go     # Template management
│   ├── alerts.go       # Alert configuration
│   └── validation.go   # ValidateName() - shared validation
│
├── download/           # HTTP download utilities
│   └── download.go     # Download(), DownloadWithProgress()
│
├── github/             # GitHub integration
│   ├── releases.go     # Release checking, downloads
│   ├── updates.go      # Update checking, FindPreviousRelease()
│   └── install.go      # InstallRelease(), binary replacement
│
├── mock/               # Demo mode mock server
│   ├── server.go       # MockServer for demo/testing
│   ├── demo.go         # Demo mode utilities
│   ├── config.go       # Mock device configuration
│   ├── fixtures.go     # Test fixtures
│   ├── fleet.go        # Fleet mock data
│   └── discovery.go    # Mock discovery responses
│
├── ratelimit/          # Rate limiting with circuit breaker
│   ├── ratelimit.go    # RateLimiter, TokenBucket
│   ├── circuit.go      # CircuitBreaker
│   └── options.go      # Configuration options
│
├── telemetry/          # Opt-in anonymous usage analytics
│   └── telemetry.go    # Telemetry collection (disabled by default)
│
├── utils/              # Shared utilities (batch, device conversion)
│   ├── batch.go        # ResolveBatchTargets() - args/stdin/group/all
│   ├── device.go       # DiscoveredDeviceToConfig(), RegisterDiscoveredDevices()
│   │                   # UnmarshalJSON() - RPC response helper
│   └── must.go         # Must() - panic on error for init-time failures
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
│   │                   #   FormatOutput()
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
│   │
│   ├── table.go        # Table builder utilities
│   ├── backup.go       # FormatBackupsTable()
│   ├── benchmark.go    # Benchmark result formatters
│   ├── diff.go         # Config/script diff formatters
│   ├── prompt.go       # Prompt formatting helpers
│   ├── report.go       # Report formatters
│   ├── writer.go       # Writer utilities
│   │
│   ├── jsonfmt/        # JSON formatting with syntax highlighting
│   ├── yamlfmt/        # YAML formatting with syntax highlighting
│   ├── synfmt/         # Syntax highlighting utilities
│   └── tmplfmt/        # Go template formatting
│
├── plugins/            # Plugin system
│   ├── registry.go     # Registry - plugin management
│   ├── loader.go       # Plugin loading
│   ├── executor.go     # Plugin execution
│   ├── manifest.go     # Plugin metadata
│   ├── migrate.go      # Plugin migration
│   ├── upgrader.go     # Upgrader - plugin upgrade functionality
│   ├── hooks.go        # Plugin hooks
│   ├── types.go        # Shared types
│   └── scaffold/       # Plugin scaffolding
│
├── shelly/             # Business logic service layer
│   ├── shelly.go       # Service struct
│   │                   #   Connect(), WithConnection(), RawRPC()
│   │                   #   ResolveWithGeneration() - Gen1/Gen2 detection
│   │                   #   Service accessors for subpackages
│   │
│   ├── connection.go   # Connection management helpers
│   │                   #   DeviceClient, WithDevice(), WithDevices()
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
│   ├── config.go       # Config get/set, WiFi, BLE, cloud, webhooks
│   ├── wifi.go         # WiFi operations (WiFiStatusFull, WiFiConfigFull, etc.)
│   ├── backup.go       # Service methods using backup/ types
│   ├── schedule.go     # Schedule operations
│   ├── script.go       # Script operations
│   ├── firmware.go     # Firmware operations
│   ├── cloud.go        # Cloud operations
│   ├── monitoring.go   # MonitoringSnapshot, FetchAllSnapshots()
│   ├── kvs.go          # Service methods using kvs/ types
│   ├── template.go     # Template operations
│   ├── zigbee.go       # Zigbee operations
│   ├── bthome.go       # BTHome operations
│   ├── lora.go         # LoRa operations
│   ├── matter.go       # Matter operations
│   ├── migrate.go      # Migration operations
│   │
│   ├── auth/           # Authentication configuration
│   │   └── auth.go       # Service: GetStatus(), Set(), Disable()
│   │                     # Status type, CalculateHA1()
│   │
│   ├── automation/     # Scripts, schedules, webhooks
│   │   └── ...           # Automation-related operations
│   │
│   ├── backup/         # Backup domain types
│   │   └── types.go      # DeviceBackup, Options, RestoreOptions, DeviceInfo
│   │                     # Validate(), LoadAndValidate(), IsFile()
│   │
│   ├── component/      # Component operations
│   │   └── ...           # Switch, Light, Cover, RGB, Input
│   │
│   ├── device/         # Device operations
│   │   └── ...           # Reboot, Info, Status, List
│   │
│   ├── export/         # Export format builders
│   │   ├── ansible.go    # BuildAnsibleInventory(), AnsibleInventory
│   │   ├── terraform.go  # BuildTerraformConfig(), TerraformDevice
│   │   ├── backup.go     # BackupExporter, ScanBackupFiles(), WriteBackupFile()
│   │   └── energy.go     # FormatEMDataCSV(), FormatEM1DataCSV()
│   │
│   ├── firmware/       # Firmware checking and updates
│   │   └── ...           # Check, Update, Rollback, Cache
│   │
│   ├── kvs/            # KVS domain types
│   │   └── types.go      # Item, ListResult, GetResult, Export
│   │                     # ParseValue(), ParseImportFile()
│   │
│   ├── modbus/         # Modbus configuration
│   │   └── modbus.go     # Service: GetStatus(), GetConfig(), SetConfig()
│   │                     # Status, Config types
│   │
│   ├── network/        # Network-related services
│   │   ├── wifi.go       # WiFiService: GetStatusFull(), GetConfigFull(), etc.
│   │   ├── mqtt.go       # MQTTService: GetStatus(), GetConfig(), SetConfig()
│   │   ├── ethernet.go   # EthernetService: GetStatus(), GetConfig(), SetConfig()
│   │   ├── cloud.go      # CloudClient: Login(), GetAllDevices(), etc.
│   │   └── cloudcontrol.go # Cloud device control operations
│   │
│   ├── provision/      # Device provisioning
│   │   └── provision.go  # Service: GetDeviceInfoByAddress(), ConfigureWiFi()
│   │                     # GetBTHomeStatus(), StartBTHomeDiscovery()
│   │                     # DeviceInfo, BTHomeDiscovery types
│   │
│   ├── sensoraddon/    # Sensor add-on operations
│   │   └── ...           # Temperature/humidity sensor add-on handling
│   │
│   └── wireless/       # Wireless protocol services
│       └── ...           # BLE, BTHome, Matter, Zigbee, LoRa
│
├── term/               # Terminal presentation (composed displays)
│   ├── term.go         # Package doc, shared helpers (printTable, formatTemp)
│   ├── action.go       # DisplayActions - Gen1 action URL display
│   ├── alert.go        # DisplayAlerts*, DisplayAlertStatus
│   ├── audit.go        # DisplayAuditResults
│   ├── automation.go   # DisplayScript*, DisplaySchedule*, DisplayWebhook*
│   ├── backup.go       # DisplayBackupSummary, DisplayRestorePreview, etc.
│   ├── benchmark.go    # DisplayBenchmarkResults
│   ├── ble.go          # DisplayBLEProvisionResult
│   ├── bthome.go       # DisplayBTHomeDevices*, DisplayBTHomeSensor*
│   ├── component.go    # DisplaySwitch*, DisplayLight*, DisplayCover*, DisplayInput*
│   ├── config.go       # DisplayConfigTable, DisplaySceneList, DisplayAliasList
│   ├── debug.go        # PrintJSONResult, DisplayWebSocketEvent
│   ├── device.go       # DisplayDeviceStatus, DisplayAllSnapshots
│   ├── diff.go         # DisplayConfigDiffs, DisplayScriptDiffs, etc.
│   ├── discovery.go    # DisplayDiscoveredDevices, DisplayBLEDevices
│   ├── doctor.go       # DisplayDoctorResults, DisplayCheckResult*
│   ├── energy.go       # DisplayEnergyStatus, DisplayEnergyHistory
│   ├── event.go        # DisplayEvent, OutputEventJSON
│   ├── firmware.go     # DisplayFirmwareStatus, DisplayFirmwareInfo
│   ├── fleet.go        # DisplayFleetStatus, DisplayFleetHealth, DisplayFleetStats
│   ├── kvs.go          # DisplayKVS*
│   ├── network.go      # DisplayWiFi*, DisplayEthernet*, DisplayMQTT*, DisplayCloud*
│   ├── power.go        # DisplayPowerMetrics, DisplayDashboard, DisplayComparison
│   ├── repl.go         # REPL session display and command handling
│   ├── scene.go        # DisplaySceneDetails
│   ├── script.go       # DisplayScriptStatus, DisplayScriptCode
│   ├── sensor.go       # Generic sensor displays (partial application pattern)
│   │                   #   DisplayTemperature*, DisplayHumidity*, etc.
│   ├── shell.go        # Shell session display and command handling
│   └── version.go      # DisplayVersionInfo, DisplayUpdateAvailable, RunUpdateCheck
│
├── testutil/           # Testing utilities
│   ├── testutil.go     # Test helpers
│   ├── mockclient.go   # MockClient for testing
│   └── mockresolver.go # MockResolver for testing
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
│   └── components/     # TUI components (39 total)
│       ├── backup/       # Backup management
│       ├── batch/        # Batch operations
│       ├── ble/          # BLE device management
│       ├── cloud/        # Cloud integration
│       ├── cmdmode/      # Command mode input
│       ├── confirm/      # Confirmation dialogs
│       ├── devicedetail/ # Device detail view
│       ├── deviceinfo/   # Device info panel
│       ├── devicelist/   # Device list
│       ├── discovery/    # Device discovery
│       ├── energy/       # Energy dashboard
│       ├── energybars/   # Energy bar charts
│       ├── energyhistory/# Energy history view
│       ├── errorview/    # Error display
│       ├── events/       # Events stream
│       ├── firmware/     # Firmware management
│       ├── fleet/        # Fleet management
│       ├── form/         # Form inputs
│       ├── help/         # Help overlay
│       ├── inputs/       # Input components
│       ├── jsonviewer/   # JSON viewer
│       ├── kvs/          # KVS management
│       ├── loading/      # Loading indicators
│       ├── modal/        # Modal dialogs
│       ├── monitor/      # Real-time monitoring
│       ├── protocols/    # Protocol views
│       ├── provisioning/ # Device provisioning
│       ├── schedules/    # Schedule management
│       ├── scripts/      # Script management
│       ├── search/       # Search interface
│       ├── security/     # Security settings
│       ├── smarthome/    # Smart home integrations
│       ├── statusbar/    # Status bar
│       ├── system/       # System info
│       ├── tabs/         # Tab navigation
│       ├── toast/        # Toast notifications
│       ├── virtuals/     # Virtual components
│       ├── webhooks/     # Webhook management
│       └── wifi/         # WiFi configuration
│
├── webhook/            # Webhook server
│   └── server.go       # HTTP webhook receiver
│
├── wizard/             # Interactive setup wizard
│   ├── wizard.go       # Main wizard flow
│   ├── steps.go        # Wizard step definitions
│   ├── discovery.go    # Device discovery step
│   ├── check.go        # Validation checks
│   └── helpers.go      # Wizard utilities
│
└── version/            # Build-time version info
    ├── version.go      # Version, Commit, Date variables
    │                   # Get(), Short(), Long(), String()
    ├── cache.go        # ReadCachedVersion() - version cache management
    └── notification.go # ShowUpdateNotification() - update availability
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

### Embeddable Flag Structs

Flag structs in `cmdutil/flags/` reduce Options boilerplate:

```go
import "github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"

type Options struct {
    flags.BatchFlags     // Embed common batch flags
    flags.ComponentFlags // Embed component ID flag
    Factory *cmdutil.Factory
}

func NewCommand(f *cmdutil.Factory) *cobra.Command {
    opts := &Options{Factory: f}
    cmd := &cobra.Command{...}
    flags.AddBatchFlags(cmd, &opts.BatchFlags)
    flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Switch")
    return cmd
}
```

Available embeddable structs:
- `BatchFlags` - GroupName, All, Timeout, SwitchID, Concurrent
- `SceneFlags` - Timeout, Concurrent, DryRun
- `ComponentFlags` - ID (for component selection via --id/-i flag)
- `QuickComponentFlags` - ID with -1 default for "all components" (on/off/toggle)
- `OutputFlags` - Format (table/json/yaml display output)
- `ConfirmFlags` - Yes, Confirm (for destructive operations)
- `DeviceListFlags` - All, Group, Quiet, JSON (for device list filtering)

**Note:** For backward compatibility, these are also re-exported from `cmdutil` directly (e.g., `cmdutil.BatchFlags`). New code should import from `cmdutil/flags` directly.

**Dynamic helpers for custom defaults:**
- `AddOutputFlagsCustom(cmd, flags, default, allowed...)` - custom default and allowed values
- `AddOutputFlagsNamed(cmd, flags, name, short, default, allowed...)` - custom flag name

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

**Move ALL other private functions OUT of cmd/ using the above decision tree**

---

## See Also

- [configuration.md](configuration.md) - Configuration management
- [development.md](development.md) - Coding patterns and standards
- [testing.md](testing.md) - Testing strategy
- [plugins.md](plugins.md) - Plugin system
