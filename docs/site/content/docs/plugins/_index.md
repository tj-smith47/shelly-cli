---
title: "Plugin System"
description: "Extend Shelly CLI with custom plugins"
weight: 50
---


Plugins extend Shelly CLI functionality. They are standalone executables that integrate seamlessly with the CLI.

> **Terminology:** "Plugin" is the primary term. "Extension" is an alias for compatibility.

## Quick Start

```bash
# Create a new plugin scaffold
shelly plugin create myext --lang bash

# Test locally
./shelly-myext/shelly-myext --help

# Install
shelly plugin install ./shelly-myext/shelly-myext

# Run
shelly myext --help
```

## How Plugins Work

Plugins are executable programs named `shelly-<name>` that:

1. **Discovered automatically** from `~/.config/shelly/plugins/` and `$PATH`
2. **Invoked transparently** - `shelly myext` runs `shelly-myext`
3. **Receive context** via environment variables (devices, config, theme)
4. **Output handled** - stdout/stderr passed through to user

```
User runs:     shelly myext --flag arg
CLI executes:  ~/.config/shelly/plugins/shelly-myext --flag arg
```

## Plugin Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `shelly plugin list` | `ls`, `l` | List installed plugins |
| `shelly plugin install <source>` | `add` | Install from file, URL, or GitHub |
| `shelly plugin remove <name>` | `rm`, `uninstall` | Remove a plugin |
| `shelly plugin upgrade [name]` | `update` | Upgrade plugin(s) |
| `shelly plugin create <name>` | `new`, `init` | Create plugin scaffold |
| `shelly plugin exec <name>` | `run` | Execute plugin explicitly |

## Environment Variables

Plugins receive these environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `SHELLY_CONFIG_PATH` | Path to config file | `~/.config/shelly/config.yaml` |
| `SHELLY_DEVICES_JSON` | JSON of registered devices | `{"kitchen": {"ip": "192.168.1.100"}}` |
| `SHELLY_OUTPUT_FORMAT` | Current output format | `table`, `json`, `yaml` |
| `SHELLY_NO_COLOR` | Color disabled | `1` if disabled |
| `SHELLY_VERBOSE` | Verbose mode | `1` if enabled |
| `SHELLY_QUIET` | Quiet mode | `1` if enabled |
| `SHELLY_API_MODE` | API mode | `local`, `cloud`, `auto` |
| `SHELLY_THEME` | Current theme name | `dracula` |

## Installation Sources

### Local File

```bash
shelly plugin install ./shelly-myext
shelly plugin install /path/to/shelly-myext
```

### GitHub Repository

```bash
shelly plugin install gh:user/shelly-myext
shelly plugin install github:user/shelly-myext
```

Downloads the latest release binary for your platform (linux/darwin, amd64/arm64).

### HTTP URL

```bash
shelly plugin install https://example.com/shelly-myext
```

## Creating Plugins

### Using the Scaffold Command

```bash
# Bash plugin (default)
shelly plugin create myext

# Go plugin
shelly plugin create myext --lang go

# Python plugin
shelly plugin create myext --lang python

# Custom output directory
shelly plugin create myext --output ~/projects
```

### Plugin Requirements

1. **Naming**: Must be named `shelly-<name>` (e.g., `shelly-notify`)
2. **Executable**: Must have executable permissions
3. **Version flag**: Should support `--version` for version detection
4. **Help flag**: Should support `--help` for usage info

### Bash Plugin Template

```bash
#!/usr/bin/env bash
# shelly-myext - Shelly CLI Plugin

set -euo pipefail

VERSION="0.1.0"

show_help() {
    cat << EOF
shelly-myext - Description of what this plugin does

Usage: shelly myext [command] [options]

Commands:
    help        Show this help message
    version     Show version information

Options:
    -h, --help      Show help
    -v, --version   Show version

Environment:
    SHELLY_DEVICES_JSON  - JSON of registered devices
    SHELLY_CONFIG_PATH   - Path to config file
EOF
}

main() {
    case "${1:-help}" in
        -h|--help|help)   show_help ;;
        -v|--version|version) echo "shelly-myext version $VERSION" ;;
        *)
            echo "Unknown command: $1" >&2
            exit 1
            ;;
    esac
}

main "$@"
```

### Go Plugin Template

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

const version = "0.1.0"

func main() {
    if len(os.Args) < 2 {
        showHelp()
        return
    }

    switch os.Args[1] {
    case "-h", "--help", "help":
        showHelp()
    case "-v", "--version", "version":
        fmt.Printf("shelly-myext version %s\n", version)
    case "devices":
        listDevices()
    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
        os.Exit(1)
    }
}

func showHelp() {
    fmt.Println(`shelly-myext - Description

Usage: shelly myext [command]

Commands:
    help      Show help
    version   Show version
    devices   List devices from environment`)
}

func listDevices() {
    devicesJSON := os.Getenv("SHELLY_DEVICES_JSON")
    if devicesJSON == "" {
        fmt.Println("No devices (SHELLY_DEVICES_JSON not set)")
        return
    }

    var devices map[string]any
    if err := json.Unmarshal([]byte(devicesJSON), &devices); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to parse devices: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Found %d device(s):\n", len(devices))
    for name := range devices {
        fmt.Printf("  - %s\n", name)
    }
}
```

### Python Plugin Template

```python
#!/usr/bin/env python3
"""shelly-myext - Shelly CLI Plugin"""

import json
import os
import sys

VERSION = "0.1.0"

def show_help():
    print("""shelly-myext - Description

Usage: shelly myext [command]

Commands:
    help      Show help
    version   Show version
    devices   List devices""")

def list_devices():
    devices_json = os.environ.get("SHELLY_DEVICES_JSON", "{}")
    try:
        devices = json.loads(devices_json)
    except json.JSONDecodeError:
        print("Failed to parse devices", file=sys.stderr)
        sys.exit(1)

    print(f"Found {len(devices)} device(s):")
    for name in devices:
        print(f"  - {name}")

def main():
    args = sys.argv[1:]
    cmd = args[0] if args else "help"

    if cmd in ("-h", "--help", "help"):
        show_help()
    elif cmd in ("-v", "--version", "version"):
        print(f"shelly-myext version {VERSION}")
    elif cmd == "devices":
        list_devices()
    else:
        print(f"Unknown command: {cmd}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
```

## Included Example: shelly-notify

The repository includes a complete example plugin at `examples/plugins/shelly-notify/`. This plugin sends desktop notifications for Shelly device events.

### Features

- Send custom notifications
- Check device status and notify
- Monitor device online/offline state
- Report power consumption
- Cross-platform (Linux notify-send, macOS osascript)

### Usage

```bash
# Install the example plugin
shelly plugin install examples/plugins/shelly-notify/shelly-notify

# Send a custom notification
shelly notify send "Kitchen" "Light turned on"

# Check device status
shelly notify device kitchen

# Check if device is online
shelly notify online kitchen

# Get power consumption notification
shelly notify power kitchen

# Test notification system
shelly notify test
```

See `examples/plugins/shelly-notify/README.md` for full documentation.

## Plugin Discovery

Plugins are discovered in this order (first match wins):

1. `~/.config/shelly/plugins/` - User plugin directory
2. Custom paths from config (`plugins.path` in config.yaml)
3. `$PATH` directories - System-wide plugins

```yaml
# config.yaml
plugins:
  path:
    - /opt/shelly-plugins
    - ~/custom-plugins
```

## Best Practices

### 1. Handle Missing Environment

```bash
# Don't assume SHELLY_DEVICES_JSON exists
devices="${SHELLY_DEVICES_JSON:-{}}"
```

### 2. Respect Output Format

```bash
if [[ "$SHELLY_OUTPUT_FORMAT" == "json" ]]; then
    echo '{"status": "ok"}'
else
    echo "Status: OK"
fi
```

### 3. Respect Color Settings

```bash
if [[ -z "$SHELLY_NO_COLOR" ]]; then
    GREEN='\033[0;32m'
    NC='\033[0m'
else
    GREEN=''
    NC=''
fi
echo -e "${GREEN}Success${NC}"
```

### 4. Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |

### 5. Version Output

Keep version output simple for parsing:

```
shelly-myext version 0.1.0
```

## Publishing Plugins

### GitHub Releases

1. Create a GitHub repository named `shelly-<name>`
2. Build binaries for each platform
3. Create a release with assets:
   - `shelly-myext-linux-amd64.tar.gz`
   - `shelly-myext-linux-arm64.tar.gz`
   - `shelly-myext-darwin-amd64.tar.gz`
   - `shelly-myext-darwin-arm64.tar.gz`

Users can then install with:
```bash
shelly plugin install gh:youruser/shelly-myext
```

### GoReleaser Example

```yaml
# .goreleaser.yaml
project_name: shelly-myext

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}-{{ .Os }}-{{ .Arch }}"
```

## Plugin Manifest System

Plugins are stored with metadata that enables automatic upgrades and source tracking.

### Directory Structure

```
~/.config/shelly/plugins/
├── shelly-myext/
│   ├── shelly-myext          # Binary executable
│   └── manifest.json         # Metadata file
├── shelly-another/
│   ├── shelly-another
│   └── manifest.json
└── .migrated                  # Migration marker
```

### Manifest Contents

Each plugin has a `manifest.json` tracking:

| Field | Description |
|-------|-------------|
| `schema_version` | Manifest format version |
| `name` | Plugin name (without `shelly-` prefix) |
| `version` | Semantic version |
| `installed_at` | Installation timestamp |
| `updated_at` | Last upgrade timestamp |
| `source.type` | Installation source: `github`, `url`, `local`, `unknown` |
| `source.url` | Source URL (GitHub or HTTP) |
| `source.ref` | Git tag/commit for GitHub sources |
| `binary.checksum` | SHA256 checksum for integrity |
| `binary.platform` | Platform (e.g., `linux-amd64`) |

### Example Manifest

```json
{
  "schema_version": "1",
  "name": "notify",
  "version": "1.2.0",
  "installed_at": "2024-12-15T10:30:00Z",
  "updated_at": "2024-12-15T10:30:00Z",
  "source": {
    "type": "github",
    "url": "https://github.com/user/shelly-notify",
    "ref": "v1.2.0",
    "asset": "shelly-notify-linux-amd64.tar.gz"
  },
  "binary": {
    "name": "shelly-notify",
    "checksum": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "platform": "linux-amd64",
    "size": 5242880
  }
}
```

### Upgrade Behavior by Source Type

| Source | Upgrade Behavior |
|--------|------------------|
| `github` | Checks GitHub releases, downloads newer version if available |
| `url` | Re-downloads from URL, replaces if checksum differs |
| `local` | Cannot auto-upgrade - reinstall manually |
| `unknown` | Migrated plugin - reinstall to enable auto-upgrade |

### Migration from Legacy Format

Plugins installed before the manifest system are automatically migrated on first CLI run. Migrated plugins have `source.type: "unknown"` - reinstall them to enable auto-upgrade:

```bash
# Check which plugins need reinstallation
shelly plugin list

# Reinstall from GitHub to enable upgrades
shelly plugin remove myext
shelly plugin install gh:user/shelly-myext
```

## Enhanced Plugin Capabilities

Plugins can declare capabilities and hooks to integrate deeply with shelly-cli. This enables features like device detection during discovery, device control via unified commands, and firmware updates.

### Capabilities

The `capabilities` field in the manifest declares what a plugin can do:

```json
{
  "capabilities": {
    "device_detection": true,
    "platform": "tasmota",
    "components": ["switch", "light", "sensor", "energy"],
    "firmware_updates": true,
    "hints": {
      "scene": "Tasmota uses Rules for automation. See: https://tasmota.github.io/docs/Rules/",
      "script": "Tasmota uses Berry scripting on ESP32. See: https://tasmota.github.io/docs/Berry/"
    }
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `device_detection` | boolean | Plugin can detect devices during `shelly discover` |
| `platform` | string | Platform name (e.g., "tasmota", "esphome") |
| `components` | array | Controllable component types: "switch", "light", "cover", "sensor", "energy" |
| `firmware_updates` | boolean | Plugin supports firmware update operations |
| `hints` | object | Helpful messages for unsupported commands (key=command, value=hint) |

### Hooks

The `hooks` field defines executable entry points that shelly-cli calls:

```json
{
  "hooks": {
    "detect": "./shelly-myext detect",
    "status": "./shelly-myext status",
    "control": "./shelly-myext control",
    "check_updates": "./shelly-myext check-updates",
    "apply_update": "./shelly-myext apply-update"
  }
}
```

#### Detect Hook

Called during `shelly discover` to probe if an address belongs to this platform.

**Input:**
```bash
./shelly-myext detect --address=192.168.1.100 [--auth-user=<user> --auth-pass=<pass>]
```

**Output:** JSON `DeviceDetectionResult`:
```json
{
  "detected": true,
  "platform": "tasmota",
  "device_id": "sonoff-basic-1234",
  "device_name": "Garage Light",
  "model": "Sonoff Basic R3",
  "firmware": "14.3.0",
  "mac": "AA:BB:CC:DD:EE:FF",
  "components": [
    {"type": "switch", "id": 0, "name": "Relay 1"}
  ]
}
```

**Exit codes:** 0 = detected, 1 = not this platform

#### Status Hook

Called to get device status.

**Input:**
```bash
./shelly-myext status --address=192.168.1.100 [--auth-user=<user> --auth-pass=<pass>]
```

**Output:** JSON `DeviceStatusResult`:
```json
{
  "online": true,
  "components": {
    "switch:0": {"output": true}
  },
  "sensors": {
    "wifi_rssi": -52
  },
  "energy": {
    "power": 45.3,
    "voltage": 121.5,
    "current": 0.372,
    "total": 123.456
  }
}
```

#### Control Hook

Called to execute device control commands.

**Input:**
```bash
./shelly-myext control --address=192.168.1.100 --action=<on|off|toggle> --component=<switch|light|cover> --id=<n> [--auth-user=<user> --auth-pass=<pass>]
```

**Output:** JSON `ControlResult`:
```json
{
  "success": true,
  "state": "on"
}
```

#### Check Updates Hook

Called to check for firmware updates.

**Input:**
```bash
./shelly-myext check-updates --address=192.168.1.100 [--auth-user=<user> --auth-pass=<pass>]
```

**Output:** JSON `FirmwareUpdateInfo`:
```json
{
  "current_version": "14.3.0",
  "latest_stable": "15.2.0",
  "latest_beta": "15.3.0b1",
  "has_update": true,
  "has_beta_update": true,
  "ota_url_stable": "http://ota.tasmota.com/tasmota/release/tasmota.bin.gz",
  "ota_url_beta": "http://ota.tasmota.com/tasmota/tasmota.bin.gz",
  "chip_type": "ESP8266"
}
```

#### Apply Update Hook

Called to apply a firmware update.

**Input:**
```bash
./shelly-myext apply-update --address=192.168.1.100 --stage=<stable|beta> [--url=<custom_ota_url>] [--auth-user=<user> --auth-pass=<pass>]
```

**Output:** JSON `UpdateResult`:
```json
{
  "success": true,
  "message": "Update initiated",
  "rebooting": true
}
```

### Device Lifecycle

When a plugin declares `device_detection: true`, it participates in the discovery flow:

1. **Discovery**: `shelly discover` scans the network
2. **Detection**: For addresses that aren't Shelly devices, each detection-capable plugin's `detect` hook is called
3. **Registration**: Detected devices are registered with `platform: "<plugin-platform>"`
4. **Command Routing**: When running commands on plugin-managed devices, the CLI routes to the appropriate plugin hook

### Environment Variables for Hooks

Hooks receive the standard plugin environment variables plus:

| Variable | Description |
|----------|-------------|
| `SHELLY_PLUGIN_DIR` | Directory where the plugin is installed |
| `SHELLY_CLI_VERSION` | CLI version for compatibility checks |

### Complete Manifest Example

```json
{
  "schema_version": "1",
  "name": "tasmota",
  "version": "1.0.0",
  "description": "Tasmota device support for shelly-cli",
  "installed_at": "2024-12-25T10:30:00Z",
  "source": {
    "type": "github",
    "url": "https://github.com/user/shelly-tasmota",
    "ref": "v1.0.0"
  },
  "binary": {
    "name": "shelly-tasmota",
    "checksum": "sha256:abc123...",
    "platform": "linux-amd64"
  },
  "minimum_shelly_version": "1.0.0",
  "capabilities": {
    "device_detection": true,
    "platform": "tasmota",
    "components": ["switch", "light", "sensor", "energy"],
    "firmware_updates": true,
    "hints": {
      "scene": "Tasmota uses Rules for automation",
      "script": "Tasmota uses Berry scripting on ESP32",
      "schedule": "Tasmota uses Timers for scheduling"
    }
  },
  "hooks": {
    "detect": "./shelly-tasmota detect",
    "status": "./shelly-tasmota status",
    "control": "./shelly-tasmota control",
    "check_updates": "./shelly-tasmota check-updates",
    "apply_update": "./shelly-tasmota apply-update"
  }
}
```

## Troubleshooting

### Plugin not found

```bash
# Check if plugin is installed
shelly plugin list

# Check plugin path
ls -la ~/.config/shelly/plugins/

# Verify executable permission
chmod +x ~/.config/shelly/plugins/shelly-myext
```

### Plugin crashes on load

```bash
# Run plugin directly to see errors
~/.config/shelly/plugins/shelly-myext --help

# Check for missing dependencies
ldd ~/.config/shelly/plugins/shelly-myext  # Linux
otool -L ~/.config/shelly/plugins/shelly-myext  # macOS
```

### Environment variables not set

```bash
# Debug environment
shelly plugin exec myext env | grep SHELLY_
```
