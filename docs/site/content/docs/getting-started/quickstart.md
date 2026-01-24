---
title: "Quick Start"
description: "Get up and running with Shelly CLI in 5 minutes"
weight: 20
---

This guide will have you controlling your first Shelly device in 5 minutes.

## Prerequisites

- Shelly CLI installed ([Installation Guide](/docs/getting-started/installation/))
- A Shelly device connected to your network
- Your computer on the same network as the device

## Step 1: Initialize Configuration

Create the default configuration file:

```bash
shelly init
```

**Expected output:**
```
✓ Created configuration directory: ~/.config/shelly
✓ Created default configuration: ~/.config/shelly/config.yaml
✓ Shelly CLI initialized successfully!

Next steps:
  1. Run 'shelly discover --register' to find and register devices
  2. Or run 'shelly init --device <name>=<ip>' to manually add a device
  3. Run 'shelly --help' to see available commands
```

## Step 2: Discover Devices

Find Shelly devices on your network:

```bash
shelly discover
```

**Expected output:**
```
Discovering devices on local network...

Found 3 devices:

NAME                    IP              MODEL              GEN   ONLINE
shelly1-84CCA8ABC123    192.168.1.100   SHSW-1             1     ●
shellyplus1-ABC123      192.168.1.101   SNSW-001X16EU      2     ●
shellypro4pm-XYZ789     192.168.1.102   SNSW-104PM40EU     3     ●

To register devices: shelly discover --register
```

**Options:**
```bash
# Extend discovery timeout
shelly discover --timeout 10s

# Discover using BLE (for devices not yet on WiFi)
shelly discover ble

# Use a specific network range
shelly discover --network 10.0.0.0/24
```

## Step 3: Register Devices

### Option 1: Register During Discovery (Recommended)

Register all discovered devices automatically:

```bash
shelly discover --register
```

This will add all discovered devices to your registry with auto-generated names.

### Option 2: Manual Registration

Add specific devices during initialization:

```bash
shelly init --device living-room=192.168.1.101
```

**With authentication** (if device has a password):
```bash
shelly init --device living-room=192.168.1.101:admin:secret
```

**Multiple devices:**
```bash
shelly init --device living-room=192.168.1.101 --device kitchen=192.168.1.102
```

## Step 4: Control the Device

### Quick Commands

```bash
# Turn on
shelly on living-room

# Turn off
shelly off living-room

# Toggle state
shelly toggle living-room
```

### Check Status

```bash
shelly status living-room
```

**Expected output:**
```
Device: living-room (192.168.1.101)
Model:  Shelly Plus 1 (Gen2)
Online: ● Yes

Components:
  switch:0
    State:  ON
    Power:  45.2W
    Energy: 1.23 kWh
```

### JSON Output (for scripting)

```bash
shelly status living-room -o json
```

**Output:**
```json
{
  "name": "living-room",
  "address": "192.168.1.101",
  "model": "SNSW-001X16EU",
  "generation": 2,
  "online": true,
  "components": {
    "switch:0": {
      "output": true,
      "apower": 45.2,
      "aenergy": {
        "total": 1.23
      }
    }
  }
}
```

## Step 5: Launch the TUI Dashboard

For a visual interface:

```bash
shelly dash
```

**Keyboard shortcuts:**
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select device |
| `t` | Toggle device |
| `o` | Turn on |
| `f` | Turn off |
| `r` | Refresh |
| `/` | Filter devices |
| `?` | Help |
| `q` | Quit |

## Next Steps

- [Add Your First Device](/docs/getting-started/first-device/) - Detailed device setup
- [Configuration Reference](/docs/configuration/) - Customize your setup
- [Command Reference](/docs/commands/) - All available commands
- [Guides](/docs/guides/) - Aliases, scripting, automation
