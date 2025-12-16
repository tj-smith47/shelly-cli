<div align="center">

# Shelly CLI

[![CI](https://github.com/tj-smith47/shelly-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tj-smith47/shelly-cli/actions/workflows/ci.yml)
[![Docs](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/tj-smith47/shelly-cli/badges/docs.json)](https://github.com/tj-smith47/shelly-cli/actions/workflows/docs.yml)
<!-- [![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/tj-smith47/shelly-cli/badges/coverage.json)](https://github.com/tj-smith47/shelly-cli/actions) -->
[![Go Reference](https://pkg.go.dev/badge/github.com/tj-smith47/shelly-cli.svg)](https://pkg.go.dev/github.com/tj-smith47/shelly-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/tj-smith47/shelly-cli)](https://goreportcard.com/report/github.com/tj-smith47/shelly-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful, intuitive command-line interface for managing Shelly smart home devices.

[Installation](#installation) •
[Quick Start](#quick-start) •
[Documentation](#documentation) •
[Examples](examples/) •
[Contributing](CONTRIBUTING.md)

</div>

## Features

- **Full Shelly API Coverage** - Control all Gen1, Gen2, Gen3, and Gen4 devices
- **TUI Dashboard** - Interactive terminal dashboard inspired by k9s and gh-dash
- **Device Discovery** - Automatic discovery via mDNS, BLE, and CoIoT
- **Batch Operations** - Control multiple devices simultaneously
- **Scene Management** - Create and activate scenes across devices
- **Firmware Management** - Check, update, and manage firmware versions
- **Script Management** - Upload, edit, and manage device scripts (Gen2+)
- **Schedule Management** - Create and manage on-device schedules
- **Backup & Restore** - Full device configuration backup and restore
- **Plugin System** - Extend functionality with custom plugins (gh-style)
- **Alias System** - Create shortcuts for common commands
- **Theme Support** - 280+ built-in themes via bubbletint
- **Shell Completions** - Bash, Zsh, Fish, and PowerShell
- **Multiple Output Formats** - JSON, YAML, table, and CSV output
- **Energy Monitoring** - Track power consumption and energy usage
- **Smart Home Protocols** - BTHome, Zigbee, Matter, and LoRa support

## Installation

### Homebrew (macOS/Linux)

```bash
brew install tj-smith47/tap/shelly
```

### Go Install

```bash
go install github.com/tj-smith47/shelly-cli/cmd/shelly@latest
```

### Docker

```bash
# Pull the image
docker pull ghcr.io/tj-smith47/shelly-cli:latest

# Run a command
docker run --rm ghcr.io/tj-smith47/shelly-cli:latest device list

# With config mount
docker run --rm -v ~/.config/shelly:/home/shelly/.config/shelly ghcr.io/tj-smith47/shelly-cli:latest status kitchen
```

### Download Binary

Download the latest release from the [releases page](https://github.com/tj-smith47/shelly-cli/releases).

Available for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Quick Start

```bash
# Initialize the CLI (first-time setup)
shelly init

# Discover devices on your network
shelly discover

# Add a device to your registry
shelly device add living-room 192.168.1.100

# Control devices with quick commands
shelly on living-room
shelly off living-room
shelly toggle living-room

# Get device status
shelly status living-room

# Launch the TUI dashboard
shelly dash

# Update firmware
shelly firmware update living-room
```

## Command Overview

### Quick Commands

| Command | Description | Example |
|---------|-------------|---------|
| `on` | Turn device on (auto-detects type) | `shelly on kitchen` |
| `off` | Turn device off | `shelly off kitchen` |
| `toggle` | Toggle device state | `shelly toggle kitchen` |
| `status` | Show device status | `shelly status kitchen` |
| `reboot` | Reboot device | `shelly reboot kitchen` |
| `reset` | Factory reset (with confirmation) | `shelly reset kitchen --yes` |

### Device Control

| Command | Description | Example |
|---------|-------------|---------|
| `switch` | Control switch components | `shelly switch on kitchen --id 0` |
| `light` | Control light components | `shelly light set kitchen --brightness 50` |
| `rgb` | Control RGB lights | `shelly rgb set kitchen --color "#ff0000"` |
| `cover` | Control covers/rollers | `shelly cover open garage` |
| `thermostat` | Control thermostats | `shelly thermostat set living-room --temp 22` |
| `input` | Manage input components | `shelly input status kitchen` |
| `scene` | Manage scenes | `shelly scene activate movie-night` |
| `batch` | Batch device operations | `shelly batch on all-lights` |

### Device Management

| Command | Description | Example |
|---------|-------------|---------|
| `device` | Manage device registry | `shelly device list` |
| `discover` | Discover devices | `shelly discover` |
| `group` | Manage device groups | `shelly group create lights` |
| `backup` | Backup/restore configs | `shelly backup create kitchen` |
| `migrate` | Migrate configurations | `shelly migrate kitchen new-kitchen` |
| `schedule` | Manage schedules | `shelly schedule list kitchen` |
| `script` | Manage device scripts | `shelly script list kitchen` |

### Configuration

| Command | Description | Example |
|---------|-------------|---------|
| `config` | Manage device config | `shelly config get kitchen` |
| `wifi` | WiFi configuration | `shelly wifi status kitchen` |
| `mqtt` | MQTT configuration | `shelly mqtt enable kitchen` |
| `cloud` | Cloud connection | `shelly cloud status kitchen` |
| `auth` | Authentication | `shelly auth set kitchen --user admin` |
| `webhook` | Manage webhooks | `shelly webhook list kitchen` |
| `kvs` | Key-value storage | `shelly kvs list kitchen` |

### Smart Home Protocols

| Command | Description | Example |
|---------|-------------|---------|
| `bthome` | BTHome Bluetooth devices | `shelly bthome list gateway` |
| `zigbee` | Zigbee connectivity | `shelly zigbee pair gateway` |
| `matter` | Matter protocol | `shelly matter status kitchen` |
| `lora` | LoRa add-on | `shelly lora status kitchen` |

### Monitoring

| Command | Description | Example |
|---------|-------------|---------|
| `dash` | TUI dashboard | `shelly dash` |
| `energy` | Energy monitoring | `shelly energy status kitchen` |
| `power` | Power monitoring | `shelly power status kitchen` |
| `monitor` | Real-time monitoring | `shelly monitor kitchen` |
| `metrics` | Export metrics | `shelly metrics prometheus` |
| `sensor` | Sensor readings | `shelly sensor status kitchen` |

### Utility

| Command | Description | Example |
|---------|-------------|---------|
| `alias` | Command aliases | `shelly alias add on-all "batch on lights"` |
| `plugin` | Plugin management | `shelly plugin list` |
| `theme` | Theme management | `shelly theme list` |
| `completion` | Shell completions | `shelly completion bash` |
| `firmware` | Firmware management | `shelly firmware check kitchen` |
| `export` | Export data | `shelly export csv --all` |
| `debug` | Debug commands | `shelly debug rpc kitchen Shelly.GetInfo` |

## Configuration

Configuration is stored in `~/.config/shelly/config.yaml`. See [docs/configuration.md](docs/configuration.md) for complete reference.

### Example Configuration

```yaml
# Output settings
output: table
color: true
theme: dracula

# API mode: local (default), cloud, or auto
api_mode: local

# Discovery settings
discovery:
  timeout: 5s
  mdns: true
  ble: false
  coiot: true

# Device registry
devices:
  living-room:
    address: 192.168.1.100
    generation: 2
    auth:
      user: admin
      password: secret
  kitchen:
    address: 192.168.1.101
    generation: 3

# Device groups
groups:
  lights:
    devices: [living-room, kitchen, bedroom]
  downstairs:
    devices: [living-room, kitchen]

# Scenes
scenes:
  movie-night:
    actions:
      - device: living-room
        method: Switch.Set
        params: {id: 0, on: false}
      - device: kitchen
        method: Light.Set
        params: {id: 0, brightness: 20}

# Aliases
aliases:
  ll:
    command: device list
  morning:
    command: scene activate morning-routine
```

### Environment Variables

Environment variables override config file values for the current execution.

| Variable | Description | Default |
|----------|-------------|---------|
| `SHELLY_CONFIG` | Config file path | `~/.config/shelly/config.yaml` |
| `SHELLY_OUTPUT` | Output format | `table` |
| `SHELLY_NO_COLOR` | Disable colors (presence disables) | unset |
| `SHELLY_API_MODE` | API mode (`local`, `cloud`, `auto`) | `local` |
| `SHELLY_VERBOSE` | Enable verbose output | `false` |
| `SHELLY_QUIET` | Suppress non-essential output | `false` |
| `SHELLY_NO_UPDATE_CHECK` | Disable update notifications | unset |
| `SHELLY_CLOUD_EMAIL` | Cloud login email | - |
| `SHELLY_CLOUD_PASSWORD` | Cloud login password | - |
| `NO_COLOR` | Standard color disable (https://no-color.org) | unset |

**Note:** Nested config values use underscores: `SHELLY_CLOUD_ACCESS_TOKEN` maps to `cloud.access_token`.

## Output Formats

All commands support multiple output formats:

```bash
# Table (default, human-readable)
shelly device list

# JSON (machine-readable)
shelly device list -o json

# YAML
shelly device list -o yaml

# CSV (for spreadsheets)
shelly energy history kitchen -o csv > energy.csv

# Template (custom Go template)
shelly device list -o template --template '{{.Name}}: {{.Address}}'
```

### Piping and Scripting

```bash
# JSON output for jq processing
shelly device list -o json | jq '.[] | select(.online)'

# Script-friendly output
shelly status kitchen --plain  # Returns "on" or "off"

# Quiet mode for cron jobs
shelly batch on all-lights -q

# Export to file
shelly config export kitchen -o yaml > kitchen-backup.yaml
```

## TUI Dashboard

Launch the interactive terminal dashboard:

```bash
shelly dash
```

**Features:**
- Real-time device status
- Quick device control
- Power monitoring graphs
- Keyboard-driven navigation
- Customizable keybindings
- 280+ color themes

**Keyboard Shortcuts:**
- `j/k` or `↓/↑` - Navigate devices
- `Enter` - Select device
- `t` - Toggle selected device
- `o` - Turn on
- `f` - Turn off
- `r` - Refresh
- `/` - Filter devices
- `?` - Help
- `q` - Quit

## Plugin System

Extend functionality with plugins. See [docs/plugins.md](docs/plugins.md) for development guide.

```bash
# List plugins
shelly plugin list

# Install from GitHub
shelly plugin install gh:user/shelly-notify

# Install from file
shelly plugin install ./shelly-myext

# Create new plugin
shelly plugin create myext --lang go

# Run plugin
shelly myext [args]
```

### Example Plugin

```bash
# Install the example notification plugin
shelly plugin install examples/plugins/shelly-notify/shelly-notify

# Send notification when device changes state
shelly notify device kitchen
```

## Theme System

The CLI supports 280+ built-in themes. See [docs/themes.md](docs/themes.md) for customization.

```bash
# List available themes
shelly theme list

# Set theme
shelly theme set dracula

# Preview theme
shelly theme preview nord

# Show current theme
shelly theme current
```

### Popular Themes

- `dracula` (default) - Dark theme with vibrant colors
- `nord` - Arctic, bluish color palette
- `gruvbox` - Warm, retro color scheme
- `tokyo-night` - Modern dark theme
- `catppuccin` - Pastel color palette
- `monokai` - Classic syntax theme

### Custom Themes

Create custom themes in `~/.config/shelly/themes/`:

```yaml
# mytheme.yaml
name: mytheme
colors:
  foreground: "#f8f8f2"
  background: "#282a36"
  green: "#50fa7b"
  red: "#ff5555"
  yellow: "#f1fa8c"
  blue: "#6272a4"
  cyan: "#8be9fd"
  purple: "#bd93f9"
```

## Shell Completions

Generate completions for your shell:

```bash
# Bash
shelly completion bash > /etc/bash_completion.d/shelly

# Zsh
shelly completion zsh > "${fpath[1]}/_shelly"

# Fish
shelly completion fish > ~/.config/fish/completions/shelly.fish

# PowerShell
shelly completion powershell > shelly.ps1
```

## Aliases

Create shortcuts for common commands:

```bash
# Add alias
shelly alias add morning "scene activate morning-routine"

# List aliases
shelly alias list

# Run alias
shelly morning

# Remove alias
shelly alias remove morning
```

## Device Groups

Group devices for batch operations:

```bash
# Create group
shelly group create lights --devices living-room,kitchen,bedroom

# Control group
shelly batch on lights
shelly batch off lights

# List groups
shelly group list
```

## Scenes

Save and activate device states:

```bash
# Create scene from current state
shelly scene create movie-night --from living-room,kitchen

# Activate scene
shelly scene activate movie-night

# List scenes
shelly scene list
```

## Energy Monitoring

Track power consumption:

```bash
# Current power status
shelly energy status kitchen

# Historical data
shelly energy history kitchen --period day

# Export to CSV
shelly energy history kitchen -o csv > energy.csv

# Prometheus metrics
shelly metrics prometheus
```

## Backup and Restore

```bash
# Create backup
shelly backup create kitchen -o backup.json

# Restore backup
shelly backup restore kitchen backup.json

# List backups
shelly backup list
```

## Examples

The `examples/` directory contains ready-to-use configurations and scripts:

- **[examples/config/](examples/config/)** - Configuration templates (minimal, multi-site)
- **[examples/aliases/](examples/aliases/)** - Alias collections (shortcuts, automation, power-users)
- **[examples/scripts/](examples/scripts/)** - Shell scripts (presence detection, workstation sync, bulk updates)
- **[examples/plugins/](examples/plugins/)** - Plugin examples (desktop notifications)
- **[examples/deployments/](examples/deployments/)** - Kubernetes, Docker, and Docker Compose deployment examples

**We welcome script / manifest contributions!** If you've written useful automation scripts or integrations, please submit a PR to add them to the examples directory.

## Documentation

### Guides
- [Configuration Reference](docs/configuration.md) - Complete configuration options
- [Plugin Development](docs/plugins.md) - Create custom plugins
- [Theme Customization](docs/themes.md) - Theme system details

### Reference
- [Command Reference](docs/commands/) - Auto-generated docs for all 347 commands
- [Man Pages](docs/man/) - Unix manual pages
- [Architecture](docs/architecture.md) - Development patterns
- [Dependencies](docs/dependencies.md) - Library dependencies
- [Testing Strategy](docs/testing.md) - Test coverage approach

### Examples
- [Example Configurations](examples/config/) - Ready-to-use config files
- [Example Aliases](examples/aliases/) - Pre-built command shortcuts
- [Example Scripts](examples/scripts/) - Automation shell scripts
- [Example Plugins](examples/plugins/) - Plugin implementation examples
- [Deployment Examples](examples/deployments/) - Docker, Kubernetes, Prometheus

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting a PR.

### Development Setup

```bash
# Clone repository
git clone https://github.com/tj-smith47/shelly-cli.git
cd shelly-cli

# Install dependencies
go mod download

# Build
go build -o shelly ./cmd/shelly

# Run tests
go test ./...

# Lint
golangci-lint run ./...
```

## Support & Feedback

Found a bug? Have a feature request? Use the built-in feedback command:

```bash
# Submit feedback (opens GitHub issue form)
shelly feedback

# Run diagnostics for bug reports
shelly doctor
```

You can also [open an issue directly](https://github.com/tj-smith47/shelly-cli/issues) on GitHub.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built on [shelly-go](https://github.com/tj-smith47/shelly-go) library
- TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Themes by [bubbletint](https://github.com/lrstanley/bubbletint)
- Inspired by [gh CLI](https://github.com/cli/cli) and [k9s](https://github.com/derailed/k9s)
