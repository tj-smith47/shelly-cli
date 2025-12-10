# Shelly CLI

[![CI](https://github.com/tj-smith47/shelly-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tj-smith47/shelly-cli/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/tj-smith47/shelly-cli/badges/coverage.json)](https://github.com/tj-smith47/shelly-cli/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/tj-smith47/shelly-cli.svg)](https://pkg.go.dev/github.com/tj-smith47/shelly-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/tj-smith47/shelly-cli)](https://goreportcard.com/report/github.com/tj-smith47/shelly-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful, intuitive command-line interface for managing Shelly smart home devices.

## Features

- **Full Shelly API Coverage** - Control all Gen1, Gen2, Gen3, and Gen4 devices
- **TUI Dashboard** - Interactive terminal dashboard inspired by k9s and gh-dash
- **Device Discovery** - Automatic discovery via mDNS, BLE, and CoIoT
- **Batch Operations** - Control multiple devices simultaneously
- **Scene Management** - Create and activate scenes
- **Firmware Management** - Check, update, and rollback firmware
- **Script Management** - Upload, edit, and manage device scripts (Gen2+)
- **Backup & Restore** - Full device configuration backup
- **Plugin System** - Extend functionality with custom plugins (gh-style)
- **Alias System** - Create shortcuts for common commands (gh-style)
- **Theme Support** - 280+ built-in themes via bubbletint
- **Shell Completions** - Bash, Zsh, Fish, and PowerShell

## Installation

### Homebrew (macOS/Linux)

```bash
brew install tj-smith47/tap/shelly
```

### Go Install

```bash
go install github.com/tj-smith47/shelly-cli/cmd/shelly@latest
```

### Download Binary

Download the latest release from the [releases page](https://github.com/tj-smith47/shelly-cli/releases).

## Quick Start

```bash
# Discover devices on your network
shelly discover

# Add a device to your registry
shelly device add living-room 192.168.1.100

# Control a switch
shelly switch on living-room
shelly switch off living-room
shelly switch toggle living-room

# Launch the TUI dashboard
shelly dash

# Get device status
shelly status living-room

# Update firmware
shelly firmware update living-room
```

## Configuration

Configuration is stored in `~/.config/shelly/config.yaml`. You can also use environment variables prefixed with `SHELLY_`.

```yaml
# Default output format (json, yaml, table, text)
output: table

# Enable color output
color: true

# Theme name (from bubbletint)
theme: dracula

# API mode: local (default), cloud, or auto
api_mode: local

# Device registry
devices:
  living-room:
    address: 192.168.1.100
    generation: 2
```

## Commands

| Command | Description |
|---------|-------------|
| `shelly discover` | Discover devices on the network |
| `shelly device` | Manage device registry |
| `shelly switch` | Control switches |
| `shelly cover` | Control covers/rollers |
| `shelly light` | Control lights |
| `shelly config` | Manage device configuration |
| `shelly firmware` | Firmware management |
| `shelly script` | Script management (Gen2+) |
| `shelly schedule` | Schedule management |
| `shelly backup` | Backup and restore |
| `shelly dash` | Launch TUI dashboard |
| `shelly alias` | Manage command aliases |
| `shelly extension` | Manage plugins |
| `shelly cloud` | Cloud API commands |

Run `shelly help` or `shelly <command> --help` for detailed usage.

## Shell Completions

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

## Plugins

Plugins are executables named `shelly-*` in your PATH or `~/.config/shelly/plugins/`.

```bash
# List installed plugins
shelly extension list

# Install a plugin
shelly extension install gh:user/shelly-myplugin

# Run a plugin
shelly myplugin [args]
```

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting a PR.

## License

MIT License - see [LICENSE](LICENSE) for details.
