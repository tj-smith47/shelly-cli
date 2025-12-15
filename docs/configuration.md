# Configuration Reference

This document provides a complete reference for configuring the Shelly CLI.

## Configuration File

The CLI reads configuration from `~/.config/shelly/config.yaml` by default. You can specify a different location using the `--config` flag or `SHELLY_CONFIG` environment variable.

### Creating Configuration

```bash
# Initialize with defaults
shelly init

# Or create manually
mkdir -p ~/.config/shelly
touch ~/.config/shelly/config.yaml
```

## Configuration Options

### Global Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `output` | string | `table` | Default output format: `table`, `json`, `yaml`, `csv`, `template` |
| `color` | bool | `true` | Enable colored output |
| `theme` | string/object | `dracula` | Color theme (name or configuration block) |
| `api_mode` | string | `local` | API mode: `local`, `cloud`, or `auto` |
| `verbosity` | int | `0` | Verbosity level: 0=silent, 1=info, 2=debug, 3=trace |
| `quiet` | bool | `false` | Suppress non-essential output |

```yaml
output: table
color: true
theme: dracula
api_mode: local
verbosity: 0
quiet: false
```

### Logging Settings

Configure debug logging behavior for troubleshooting.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `log.json` | bool | `false` | Output logs in JSON format for machine parsing |
| `log.categories` | string | `""` | Filter logs by category (comma-separated) |

**Verbosity levels:**
- `0` (default): Silent - no debug output
- `1` (`-v`): Info level - basic operational logs
- `2` (`-vv`): Debug level - detailed diagnostic information
- `3` (`-vvv`): Trace level - maximum verbosity

**Log categories:** `network`, `config`, `device`, `api`, `auth`, `plugin`, `discovery`, `firmware`

```yaml
verbosity: 0

log:
  # JSON format for log aggregation tools
  json: false

  # Filter to specific categories (empty = all)
  # Example: "network,api" to only see network and API logs
  categories: ""
```

### Theme Configuration

The theme can be specified as a simple string or a configuration block.

#### Simple Theme Name

```yaml
theme: dracula
```

#### Theme with Color Overrides

```yaml
theme:
  name: dracula
  colors:
    foreground: "#f8f8f2"
    background: "#282a36"
    green: "#50fa7b"
    red: "#ff5555"
    yellow: "#f1fa8c"
    blue: "#6272a4"
    cyan: "#8be9fd"
    purple: "#bd93f9"
    bright_black: "#44475a"
```

#### External Theme File

```yaml
theme:
  file: ~/.config/shelly/themes/mytheme.yaml
```

### Discovery Settings

Configure device discovery behavior.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `discovery.timeout` | duration | `5s` | Discovery timeout |
| `discovery.mdns` | bool | `true` | Enable mDNS discovery |
| `discovery.ble` | bool | `false` | Enable BLE discovery |
| `discovery.coiot` | bool | `true` | Enable CoIoT discovery |
| `discovery.network` | string | auto | Default network for scanning |

```yaml
discovery:
  timeout: 5s
  mdns: true
  ble: false
  coiot: true
  network: 192.168.1.0/24
```

### Cloud Settings

Configure Shelly Cloud API access.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `cloud.enabled` | bool | `false` | Enable cloud API |
| `cloud.email` | string | - | Shelly Cloud email |
| `cloud.access_token` | string | - | Cloud API access token |
| `cloud.refresh_token` | string | - | Cloud API refresh token |
| `cloud.server_url` | string | - | Cloud server URL (usually auto-detected) |

```yaml
cloud:
  enabled: true
  email: user@example.com
  access_token: your-access-token
  refresh_token: your-refresh-token
```

**Note:** Use `shelly cloud login` to authenticate interactively.

### Device Registry

Register devices for easy access by name.

```yaml
devices:
  living-room:
    address: 192.168.1.100
    generation: 2
    model: SNSW-001X16EU
    auth:
      user: admin
      password: secret

  kitchen:
    address: 192.168.1.101
    generation: 3
    model: SNSW-102P16EU

  garage:
    address: shelly-garage.local
    generation: 2
```

#### Device Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `address` | string | yes | IP address, hostname, or mDNS name |
| `generation` | int | no | Device generation (1, 2, 3, or 4) |
| `model` | string | no | Device model identifier |
| `auth.user` | string | no | Authentication username |
| `auth.password` | string | no | Authentication password |

### Aliases

Create shortcuts for common commands.

```yaml
aliases:
  ll:
    name: ll
    command: device list
    shell: false

  morning:
    name: morning
    command: scene activate morning-routine
    shell: false

  backup-all:
    name: backup-all
    command: |
      for d in $(shelly device list -o json | jq -r '.[].name'); do
        shelly backup create "$d" -o "backups/$d.json"
      done
    shell: true
```

#### Alias Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `name` | string | - | Alias name |
| `command` | string | - | Command to execute |
| `shell` | bool | `false` | Execute via shell (enables pipes, variables) |

### Device Groups

Group devices for batch operations.

```yaml
groups:
  lights:
    name: lights
    devices:
      - living-room
      - kitchen
      - bedroom

  downstairs:
    name: downstairs
    devices:
      - living-room
      - kitchen
      - hallway

  outdoor:
    name: outdoor
    devices:
      - porch
      - garage
      - garden
```

### Scenes

Define scenes with multiple device actions.

```yaml
scenes:
  movie-night:
    name: movie-night
    description: Dim lights for movie watching
    actions:
      - device: living-room
        method: Switch.Set
        params:
          id: 0
          on: false
      - device: kitchen
        method: Light.Set
        params:
          id: 0
          brightness: 20
      - device: hallway
        method: Light.Set
        params:
          id: 0
          on: false

  morning-routine:
    name: morning-routine
    description: Wake up lights
    actions:
      - device: bedroom
        method: Light.Set
        params:
          id: 0
          on: true
          brightness: 50
      - device: kitchen
        method: Switch.Set
        params:
          id: 0
          on: true
```

#### Scene Action Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `device` | string | yes | Target device name |
| `method` | string | yes | RPC method to call |
| `params` | object | no | Method parameters |

### Alerts

Configure monitoring alerts for device conditions.

```yaml
alerts:
  kitchen-offline:
    name: kitchen-offline
    description: Alert when kitchen device goes offline
    device: kitchen
    condition: offline
    action: notify
    enabled: true
    created_at: "2025-01-15T10:00:00Z"

  high-power:
    name: high-power
    description: Alert on high power consumption
    device: heater
    condition: "power>2000"
    action: "webhook:http://example.com/alert"
    enabled: true
    snoozed_until: ""
    created_at: "2025-01-15T10:00:00Z"
```

#### Alert Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | yes | Unique alert identifier |
| `description` | string | no | Human-readable description |
| `device` | string | yes | Device to monitor |
| `condition` | string | yes | Trigger condition (e.g., `offline`, `power>100`, `temperature>30`) |
| `action` | string | yes | Action when triggered: `notify`, `webhook:URL`, or `command:CMD` |
| `enabled` | bool | yes | Whether alert is active |
| `snoozed_until` | string | no | RFC3339 timestamp if temporarily snoozed |
| `created_at` | string | yes | When alert was created |

#### Alert Commands

```bash
# Create an alert
shelly alert create kitchen-offline --device kitchen --condition offline

# List all alerts
shelly alert list

# Test an alert (dry-run)
shelly alert test kitchen-offline

# Snooze for 1 hour
shelly alert snooze kitchen-offline --duration 1h

# Clear snooze
shelly alert snooze kitchen-offline --clear
```

**Note:** The alert system stores configurations only. Active monitoring requires integration with `shelly monitor` or external scheduling.

### Templates

Store device configuration templates for provisioning.

```yaml
templates:
  switch-default:
    name: switch-default
    description: Default switch configuration
    model: SNSW-001X16EU
    generation: 2
    config:
      sys:
        device:
          name: ""
        ui_data: {}
      switch:0:
        name: Main
        initial_state: restore_last
        auto_on: false
        auto_off: false
```

### Plugin Settings

Configure the plugin system.

```yaml
plugins:
  enabled: true
  path:
    - ~/.config/shelly/plugins
    - /opt/shelly-plugins
    - ~/custom-plugins
```

### TUI Settings

Configure the TUI dashboard.

```yaml
tui:
  refresh_interval: 5  # seconds
  theme:
    name: nord  # Independent TUI theme
  keybindings:
    up: [k, up]
    down: [j, down]
    left: [h, left]
    right: [l, right]
    page_up: [ctrl+u, pgup]
    page_down: [ctrl+d, pgdn]
    home: [g, home]
    end: [G, end]
    enter: [enter]
    escape: [esc]
    refresh: [r, ctrl+r]
    filter: [/, f]
    command: [:]
    help: [?, F1]
    quit: [q, ctrl+c]
    toggle: [t, space]
    turn_on: [o]
    turn_off: [x]
    reboot: [R]
    tab: [tab]
    shift_tab: [shift+tab]
    view1: ["1"]
    view2: ["2"]
    view3: ["3"]
    view4: ["4"]
```

## Environment Variables

Configuration can be set via environment variables with the `SHELLY_` prefix. Nested config keys use underscores (e.g., `cloud.access_token` becomes `SHELLY_CLOUD_ACCESS_TOKEN`).

### User-Configurable Variables

| Variable | Config Equivalent | Description |
|----------|-------------------|-------------|
| `SHELLY_CONFIG` | - | Config file path (checked before default) |
| `SHELLY_OUTPUT` | `output` | Output format (`table`, `json`, `yaml`) |
| `SHELLY_NO_COLOR` | - | Disable colors (presence = disabled) |
| `SHELLY_THEME` | `theme` | Theme name |
| `SHELLY_API_MODE` | `api_mode` | API mode (`local`, `cloud`, `auto`) |
| `SHELLY_VERBOSITY` | `verbosity` | Verbosity level (0-3) |
| `SHELLY_QUIET` | `quiet` | Quiet mode (`true`/`false`) |
| `SHELLY_LOG_JSON` | `log.json` | JSON log output (`true`/`false`) |
| `SHELLY_LOG_CATEGORIES` | `log.categories` | Log category filter (comma-separated) |
| `SHELLY_NO_UPDATE_CHECK` | - | Disable update notifications (presence = disabled) |
| `SHELLY_CLOUD_ACCESS_TOKEN` | `cloud.access_token` | Cloud API token |
| `SHELLY_CLOUD_EMAIL` | - | Cloud login email (used by `shelly cloud login`) |
| `SHELLY_CLOUD_PASSWORD` | - | Cloud login password (used by `shelly cloud login`) |
| `NO_COLOR` | - | Standard color disable (https://no-color.org) |

### Plugin Environment Variables

These are set BY the CLI when executing plugins, not user-configurable:

| Variable | Description |
|----------|-------------|
| `SHELLY_CONFIG_PATH` | Path to config file |
| `SHELLY_DEVICES_JSON` | JSON string of registered devices |
| `SHELLY_OUTPUT_FORMAT` | Current output format |

## Global Flags

These flags are available on all commands and override config values.

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | - | Config file path |
| `--output` | `-o` | Output format |
| `--no-color` | - | Disable colors |
| `--quiet` | `-q` | Suppress non-essential output |
| `--verbose` | `-v` | Increase verbosity (stackable: `-v`, `-vv`, `-vvv`) |
| `--log-json` | - | Output logs in JSON format |
| `--log-categories` | - | Filter logs by category (comma-separated) |
| `--template` | - | Go template (with `-o template`) |

**Verbosity examples:**
```bash
# Info level - basic operation logs
shelly device list -v

# Debug level - detailed diagnostics
shelly device info kitchen -vv

# Trace level - maximum verbosity
shelly discover mdns -vvv

# Filter to specific categories
shelly device list -v --log-categories network,api

# JSON logs for parsing
shelly device list -vv --log-json
```

## Full Configuration Example

```yaml
# ~/.config/shelly/config.yaml

# Global settings
output: table
color: true
theme: dracula
api_mode: local
verbosity: 0
quiet: false

# Logging settings (for debugging)
log:
  json: false
  categories: ""

# Discovery settings
discovery:
  timeout: 5s
  mdns: true
  ble: false
  coiot: true
  network: 192.168.1.0/24

# Cloud settings (optional)
cloud:
  enabled: false
  email: ""
  access_token: ""
  refresh_token: ""

# Device registry
devices:
  living-room:
    address: 192.168.1.100
    generation: 2
    model: SNSW-001X16EU
    auth:
      user: admin
      password: secret

  kitchen:
    address: 192.168.1.101
    generation: 3

  bedroom:
    address: 192.168.1.102
    generation: 2

# Device groups
groups:
  all-lights:
    name: all-lights
    devices:
      - living-room
      - kitchen
      - bedroom

  downstairs:
    name: downstairs
    devices:
      - living-room
      - kitchen

# Scenes
scenes:
  movie-night:
    name: movie-night
    description: Dim lights for movies
    actions:
      - device: living-room
        method: Switch.Set
        params:
          id: 0
          on: false
      - device: kitchen
        method: Light.Set
        params:
          id: 0
          brightness: 20

  all-off:
    name: all-off
    description: Turn off all lights
    actions:
      - device: living-room
        method: Switch.Set
        params:
          id: 0
          on: false
      - device: kitchen
        method: Switch.Set
        params:
          id: 0
          on: false
      - device: bedroom
        method: Switch.Set
        params:
          id: 0
          on: false

# Aliases
aliases:
  ll:
    command: device list
  st:
    command: status
  off-all:
    command: batch off all-lights

# Alerts
alerts:
  kitchen-offline:
    name: kitchen-offline
    device: kitchen
    condition: offline
    action: notify
    enabled: true
    created_at: "2025-01-15T10:00:00Z"

# Templates
templates:
  switch-default:
    name: switch-default
    description: Default switch configuration
    model: SNSW-001X16EU
    generation: 2
    config: {}

# Plugin settings
plugins:
  enabled: true
  path:
    - ~/.config/shelly/plugins

# TUI settings
tui:
  refresh_interval: 5
  keybindings:
    quit: [q, ctrl+c]
    toggle: [t, space]
```

## Configuration Commands

```bash
# Show current configuration
shelly config show

# Get specific value
shelly config get output

# Set value
shelly config set output json
shelly config set theme nord

# Open config in editor
$EDITOR ~/.config/shelly/config.yaml
```

## Configuration Files Location

| Platform | Path |
|----------|------|
| Linux/macOS | `~/.config/shelly/config.yaml` |
| Windows | `%APPDATA%\shelly\config.yaml` |

### Directory Structure

```
~/.config/shelly/
├── config.yaml          # Main configuration
├── plugins/             # Installed plugins
│   ├── shelly-notify    # Plugin binary
│   └── ...
├── themes/              # Custom themes
│   └── mytheme.yaml
└── backups/             # Device backups (if using default path)
    ├── kitchen.json
    └── living-room.json
```

## Migration

If upgrading from an older version, the CLI will automatically migrate configuration where possible. Check the release notes for any manual migration steps required.
