# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Config system refactor (in progress)

## [1.0.0] - 2025-XX-XX

First stable release of the Shelly CLI - a production-ready command-line interface for Shelly smart home devices.

### Added

#### Core Infrastructure
- Factory pattern architecture with dependency injection for testability
- IOStreams abstraction for consistent, testable I/O operations
- TUI dashboard built with BubbleTea v2 (280+ themes via bubbletint)
- Plugin system with manifest support and GitHub integration
- Alias system with import/export functionality
- Shell completions for bash, zsh, fish, and PowerShell
- Self-update mechanism with `shelly update` command
- GitHub Actions CI/CD with GoReleaser for multi-platform builds

#### Device Management
- `device list` - List all registered and discovered devices
- `device info` - Show detailed device information
- `device status` - Display current device state
- `device reboot` - Reboot devices with optional delay
- `device factoryreset` - Factory reset with safety confirmations
- `device ui` - Open device web interface in browser
- `device ping` - Enhanced ping with latency statistics
- `on/off/toggle` - Quick control commands with auto-detection
- Gen1 and Gen2+ device support with automatic generation detection
- mDNS, BLE, and CoIoT discovery methods

#### Control Commands
- `switch on/off/toggle/status` - Switch control with multi-component support
- `light on/off/toggle/brightness/color` - Light control with dimming
- `rgb on/off/toggle/set` - RGB control with color temperature
- `cover open/close/stop/position` - Cover/roller shutter control
- `input list/status` - Input monitoring
- `batch on/off/toggle/command` - Multi-device batch operations

#### Energy & Monitoring
- `energy status` - Real-time power consumption display
- `energy history` - Historical energy data with date ranges
- `energy dashboard` - Interactive energy monitoring TUI
- `energy compare` - Compare energy usage across devices
- `metrics export` - Export to Prometheus, InfluxDB, or JSON formats
- `monitor all` - Real-time monitoring of all device states
- `monitor events` - Event stream monitoring

#### Automation & Scripting
- `script list` - List device scripts
- `script create` - Create new scripts with editor integration
- `script edit` - Edit existing scripts
- `script start/stop` - Control script execution
- `script delete` - Remove scripts from devices
- `schedule list` - List scheduled jobs
- `schedule create` - Create schedules with `@sunrise`/`@sunset` support
- `schedule delete` - Remove scheduled jobs
- `scene create` - Create scenes from current device states
- `scene activate` - Activate saved scenes
- `scene list/show/delete` - Scene management

#### Configuration & Backup
- `config get` - Get configuration values
- `config set` - Set configuration values
- `config export` - Export device configuration
- `config import` - Import configuration to devices
- `config reset` - Reset device configuration
- `config edit` - Open configuration in $EDITOR
- `backup create` - Create device backups
- `backup restore` - Restore from backups
- `backup list` - List available backups
- `template list/apply/diff/delete` - Configuration templates

#### Discovery & Network
- `discover scan` - Scan network for devices
- `discover ble` - Bluetooth LE discovery
- `discover coiot` - CoIoT/mDNS discovery
- `wifi scan` - Scan for WiFi networks
- `wifi ap` - Configure access point mode
- `wifi sta` - Configure station mode
- `cloud connect` - Connect to Shelly Cloud
- `cloud disconnect` - Disconnect from cloud
- `cloud status` - Cloud connection status
- `mqtt config/status` - MQTT configuration

#### Webhooks
- `webhook list` - List configured webhooks
- `webhook create` - Create new webhooks
- `webhook delete` - Remove webhooks
- `webhook server` - Built-in local webhook receiver for testing

#### Advanced Protocols
- `bthome list` - List BTHome devices
- `bthome add` - Add BTHome devices via discovery
- `bthome remove` - Remove BTHome devices
- `bthome status` - BTHome device status
- `zigbee status` - Zigbee network status (Gen4 gateways)
- `zigbee pair` - Start Zigbee pairing mode
- `zigbee remove` - Leave Zigbee network
- `zigbee list` - List Zigbee-capable devices
- `lora status` - LoRa add-on status
- `lora config` - Configure LoRa settings
- `lora send` - Send LoRa messages
- `matter enable/disable` - Matter protocol control
- `matter reset` - Reset Matter configuration
- `matter code` - Show Matter pairing code

#### Sensors & Climate
- `sensor temperature list/status` - Temperature sensors
- `sensor humidity list/status` - Humidity sensors
- `sensor flood list/status/test` - Flood detection
- `sensor smoke list/status/test/mute` - Smoke detection
- `sensor illuminance list/status` - Light level sensors
- `sensor voltmeter list/status` - Voltage monitoring
- `sensor status` - Combined sensor overview
- `thermostat status` - Thermostat current state
- `thermostat mode` - Set heating/cooling mode
- `thermostat target` - Set target temperature
- `thermostat schedule` - Manage thermostat schedules

#### Gen1 Device Support
- `gen1 relay on/off/toggle/status` - Gen1 relay control
- `gen1 roller open/close/stop/position/calibrate/status` - Gen1 roller control
- `gen1 light on/off/toggle/brightness/status` - Gen1 light control
- `gen1 color on/off/toggle/set/gain/status` - Gen1 RGBW control
- `gen1 status` - Full Gen1 status dump
- `gen1 settings` - Gen1 device settings
- `gen1 actions` - List Gen1 action URLs
- `gen1 ota check/update` - Gen1 firmware management
- `gen1 coiot` - Monitor Gen1 CoIoT broadcasts

#### Fleet & Enterprise
- `fleet connect` - Connect to Shelly Cloud fleet
- `fleet status` - Fleet-wide device status
- `fleet stats` - Aggregate statistics
- `fleet health` - Device health monitoring
- `fleet on/off/toggle` - Cloud-based group control
- `audit` - Security audit for devices
- `report` - Generate professional reports (energy, devices, audit)

#### Developer Tools
- `debug rpc` - Raw RPC calls to devices
- `debug methods` - List available RPC methods
- `debug coiot` - Show CoIoT status
- `debug websocket` - WebSocket debug information
- `debug log` - Device debug logs (Gen1)
- `mock create` - Create mock devices for testing
- `mock list` - List mock devices
- `mock scenario` - Load test scenarios
- `export terraform` - Export to Terraform format
- `export ansible` - Export to Ansible format
- `export homeassistant` - Export to Home Assistant format
- `interactive` - Launch interactive REPL
- `shell` - Device-specific interactive shell

#### User Experience
- `doctor` - Comprehensive diagnostic health check
- `diff` - Visual configuration comparison
- `benchmark` - Device performance testing
- `feedback` - Integrated GitHub issue reporting
- `qr` - Generate QR codes for devices
- `wait` - Wait for device conditions
- `sync cloud` - Cloud synchronization
- `cache clear/stats/warm` - Cache management
- `log show/tail/level/export` - CLI logging

#### Fun Commands
- `party` - RGB party mode with rainbow/disco effects
- `sleep` - Gradual lights off (bedtime mode)
- `wake` - Gradual lights on (sunrise simulation)

#### Security
- `auth rotate` - Rotate device credentials
- `auth test` - Test authentication
- `auth export/import` - Credential management
- `cert show/install` - Certificate management (Gen2+)
- `alert create/list/test/snooze/watch` - Alert management

#### Plugin System
- Plugin discovery and installation from GitHub
- Plugin manifest system for version tracking
- Automatic plugin updates
- Plugin execution with environment context

#### Theme System
- 280+ built-in themes from bubbletint
- Custom theme creation with YAML files
- Theme inheritance and color overrides
- Separate TUI theme configuration

### Documentation
- 347 auto-generated command documentation files
- Man pages for all commands
- Configuration reference guide
- Plugin development guide
- Theme customization guide
- Architecture documentation
- Testing strategy guide
- Example configurations, scripts, and plugins

### Technical Details
- Built with Go 1.25.5
- Uses `shelly-go` library v0.1.6 for device communication
- Cobra for CLI framework
- BubbleTea for TUI components
- Supports Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64)
