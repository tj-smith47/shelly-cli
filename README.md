<div align="center">

# Shelly CLI

<img src="assets/shelly-gopher.png" alt="Shelly Gopher" width="200">

[![ci](https://github.com/tj-smith47/shelly-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tj-smith47/shelly-cli/actions/workflows/ci.yml)
[![Docs](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/tj-smith47/shelly-cli/badges/docs.json)](https://github.com/tj-smith47/shelly-cli/actions/workflows/docs.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/tj-smith47/shelly-cli/badges/coverage.json)](https://github.com/tj-smith47/shelly-cli/actions)

[![Go Report Card](https://goreportcard.com/badge/github.com/tj-smith47/shelly-cli)](https://goreportcard.com/report/github.com/tj-smith47/shelly-cli)
[![Go Reference](https://pkg.go.dev/badge/github.com/tj-smith47/shelly-cli.svg)](https://pkg.go.dev/github.com/tj-smith47/shelly-cli)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A powerful, intuitive command-line interface for managing Shelly smart home devices.

[Installation](#installation) •
[Quick Start](#quick-start) •
[Highlights](#highlights) •
[Documentation](#documentation) •
[Contributing](CONTRIBUTING.md)

</div>

## Features

- **🎯 Full Shelly API Coverage** - Control all Gen1, Gen2, Gen3, and Gen4 devices
- **📊 TUI Dashboard** - Interactive terminal dashboard inspired by k9s and gh-dash
- **🔍 Device Discovery** - Automatic discovery via mDNS, BLE, and CoIoT
- **⚡ Batch Operations** - Control multiple devices simultaneously
- **🎬 Scene Management** - Create and activate scenes across devices
- **🔧 Firmware Management** - Check, update, and manage firmware versions
- **📜 Script Management** - Upload, edit, and manage device scripts (Gen2+)
- **⏰ Schedule Management** - Create and manage on-device schedules
- **💾 Backup & Restore** - Full device configuration backup and restore
- **🔌 Plugin System** - Extend functionality with custom plugins (gh-style)
- **🏷️ Alias System** - Create shortcuts for common commands
- **🎨 Theme Support** - 280+ built-in themes via bubbletint
- **🐚 Shell Completions** - Bash, Zsh, Fish, and PowerShell
- **📤 Multiple Output Formats** - JSON, YAML, table, and CSV output
- **⚡ Energy Monitoring** - Track power consumption and energy usage
- **🏠 Smart Home Protocols** - BTHome, Zigbee, Matter, and LoRa support

> **Built different:** gh-style architecture with factory pattern DI, 280+ themes, extensible plugin system, and comprehensive device coverage across all Shelly generations.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install tj-smith47/tap/shelly-cli
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

## Highlights

### TUI Dashboard

Launch an interactive terminal dashboard with real-time device status, quick control, and power monitoring graphs:

```bash
shelly dash
```

Keyboard-driven navigation with customizable keybindings and 280+ color themes. Press `?` in the dashboard for help.

### Wiring Diagrams

Print ASCII wiring diagrams for any Shelly device model directly in your terminal:

```bash
shelly diagram -m plus-2pm
shelly diagram -m pro-4pm -s detailed
shelly diagram -m dimmer-2 -g 1 -s compact
```

Supports schematic, compact, and detailed styles across all device topologies and generations.

### MCP Server (AI Assistant Integration)

Control your Shelly devices through AI assistants via the [Model Context Protocol](https://modelcontextprotocol.io):

```bash
# One-command setup for popular AI tools
shelly mcp claude enable
shelly mcp vscode enable
shelly mcp cursor enable
shelly mcp configure --gemini
```

See the [MCP documentation](docs/site/content/docs/guides/) for manual configuration and available tools.

### Plugin System

Extend functionality with gh-style plugins:

```bash
shelly plugin install gh:user/shelly-notify
shelly plugin create myext --lang go
```

See [docs/plugins.md](docs/plugins.md) for the development guide.

### Themes

280+ built-in color themes:

```bash
shelly theme set dracula
shelly theme preview nord
```

See [docs/themes.md](docs/themes.md) for customization and custom theme creation.

## Documentation

📖 **[Full Documentation Site](https://tj-smith47.github.io/shelly-cli/)** - Browse the complete documentation online.

### Guides
- [Configuration Reference](docs/configuration.md) - Complete configuration options and environment variables
- [Plugin Development](docs/plugins.md) - Create custom plugins
- [Theme Customization](docs/themes.md) - Theme system details

### Reference
- [Command Reference](docs/commands/) - Auto-generated documentation for all commands
- [Man Pages](docs/man/) - Unix manual pages
- [Architecture](docs/architecture.md) - Directory structure and placement guide
- [Development Guide](docs/development.md) - Development patterns and standards
- [Dependencies](docs/dependencies.md) - Library dependencies
- [Testing Strategy](docs/testing.md) - Test coverage approach

### Examples
- [Example Configurations](examples/config/) - Ready-to-use config files
- [Example Aliases](examples/aliases/) - Pre-built command shortcuts
- [Example Scripts](examples/scripts/) - Automation shell scripts
- [Example Plugins](examples/plugins/) - Plugin implementation examples
- [Deployment Examples](examples/deployments/) - Docker, Kubernetes, Prometheus

## Telemetry

Shelly CLI includes **opt-in** anonymous usage telemetry to help improve the CLI. This is disabled by default and must be explicitly enabled.

**What we collect (when enabled):**
- Command name (e.g., "device list", "switch on")
- Whether the command succeeded or failed
- CLI version, OS, and architecture
- Command execution duration

See the full payload definition: [`internal/telemetry/telemetry.go#Event`](internal/telemetry/telemetry.go)

**What we DO NOT collect:**
- Device information (names, addresses, configurations)
- IP addresses or network information
- Personal data of any kind
- Authentication credentials
- Command arguments or parameters

**How to enable/disable:**

```bash
# Enable during setup wizard
shelly init

# Enable via config command
shelly config set telemetry=true

# Disable
shelly config set telemetry=false
```

The telemetry data is sent to a self-hosted endpoint and is used solely for understanding which commands are most useful and identifying issues.

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

**We welcome script / manifest contributions!** If you've written useful automation scripts or integrations, please submit a PR to add them to the [examples](examples/) directory.

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

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built on [shelly-go](https://github.com/tj-smith47/shelly-go) library
- TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Themes by [bubbletint](https://github.com/lrstanley/bubbletint)
- Inspired by [gh CLI](https://github.com/cli/cli) and [k9s](https://github.com/derailed/k9s)

## Disclaimer

Generated entirely by Claude Opus 4.5 over many iterations 🤖

---

<div align="center">
<a href="https://www.buymeacoffee.com/tjsmith47" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>
</div>
