# Shelly CLI TUI Guide

## Overview

The `shelly tui` command launches an interactive terminal interface for managing Shelly devices. The TUI provides real-time device monitoring, control, and configuration in a multi-panel layout.

## Launching

```bash
shelly tui [--filter <pattern>] [--refresh <duration>]
```

**Options:**
- `--filter`: Filter devices by name pattern (e.g., `--filter kitchen`)
- `--refresh`: Refresh interval for device status (default: 5s)

## Layout Modes

The TUI automatically adapts to terminal size:

| Terminal Width | Layout |
|----------------|--------|
| < 80 columns | **Narrow**: Panels stacked vertically |
| 80-120 columns | **Standard**: 3-column horizontal layout |
| > 120 columns | **Wide**: Extra space for device details |

## Tabs

The TUI has 5 main tabs, accessible via number keys 1-5:

| Key | Tab | Description |
|-----|-----|-------------|
| 1 | Dashboard | Device list, events, device info |
| 2 | Automation | Scripts, schedules, webhooks, KVS |
| 3 | Config | WiFi, system, cloud, inputs settings |
| 4 | Manage | Discovery, batch operations, firmware, backup |
| 5 | Fleet | Shelly Cloud fleet management |

## Navigation

### Global Keys

| Key | Action |
|-----|--------|
| `1-5` | Switch tabs |
| `Tab` | Cycle focus between panels |
| `Shift+Tab` | Cycle focus backwards |
| `?` | Toggle help overlay |
| `/` | Filter devices |
| `:` | Command mode |
| `q` | Quit |

### Panel Navigation

| Key | Action |
|-----|--------|
| `j` / `Down` | Move down in lists |
| `k` / `Up` | Move up in lists |
| `h` / `Left` | Navigate left / collapse |
| `l` / `Right` | Navigate right / expand |
| `Enter` | View details / confirm action |
| `Esc` | Close overlay / cancel / clear filter |

### Device Actions (Dashboard)

| Key | Action |
|-----|--------|
| `t` | Toggle device on/off |
| `o` | Turn on |
| `O` | Turn off (Shift+O) |
| `R` | Reboot device |
| `Enter` | View device JSON details |

## Dashboard Tab

The Dashboard displays three panels:

1. **Events Panel** (left): Real-time device events with timestamps
2. **Device List** (center): All configured devices with status indicators
3. **Device Info** (right): Selected device details including:
   - Connection status (online/offline)
   - Device model and firmware
   - Component status (switches, covers, sensors)
   - Energy readings (for supported devices)

**Status Indicators:**
- Green dot: Device online
- Red dot: Device offline
- Yellow dot: Device updating

## Automation Tab

Manage device automation with 5 panels:

1. **Scripts**: View, start, stop device scripts
2. **Editor**: Script code preview
3. **Schedules**: Manage time-based automations
4. **Webhooks**: Configure webhook notifications
5. **KVS**: Key-value store browser for script data

**Keys:**
- `1-5`: Jump to panel
- `Tab`: Cycle panels
- `Enter`: Edit selected item
- `n`: Create new item
- `d`: Delete selected

## Config Tab

Configure device settings across 4 panels:

1. **WiFi**: Network configuration, AP mode
2. **System**: Device name, timezone, debug settings
3. **Cloud**: Shelly Cloud connection settings
4. **Inputs**: Input button configuration

**Keys:**
- `1-4`: Jump to panel
- `Enter`: Edit setting
- `r`: Refresh configuration

## Manage Tab

Device management operations in 4 panels:

1. **Discovery**: Find new devices on network
   - `s` / `r`: Start scan
   - `a`: Add selected device
   - `1-3`: Switch discovery method (mDNS, HTTP, CoIoT)

2. **Batch**: Perform operations on multiple devices
   - `Space`: Toggle device selection
   - `a`: Select all
   - `n`: Select none
   - `x`: Execute operation
   - `1-5`: Select operation (On, Off, Toggle, Reboot, Check Firmware)

3. **Firmware**: Check and update device firmware
   - `c` / `r`: Check for updates
   - `Space`: Select device for update
   - `a`: Select all with updates
   - `u`: Update selected

4. **Backup**: Export and import device configurations
   - `1`: Export mode
   - `2`: Import mode
   - `x`: Execute export
   - `Enter`: Import selected backup

## Fleet Tab

Manage devices via Shelly Cloud (requires authentication):

1. **Devices**: Cloud-registered devices with status
2. **Groups**: Device groups
3. **Health**: Fleet health overview
4. **Operations**: Bulk operations (All On, All Off)

**Keys:**
- `r`: Refresh device list
- `Enter`: View device details
- `1-2`: Select operation
- `Enter`: Execute operation

## Error Handling

When errors occur, the TUI displays an error message with a retry hint:

```
Error: connection timeout
  Press 'r' to retry
```

All components support the `r` key to retry failed operations.

## Command Mode

Press `:` to enter command mode. Available commands:

| Command | Description |
|---------|-------------|
| `:q`, `:quit` | Quit TUI |
| `:refresh` | Force refresh all data |
| `:theme <name>` | Switch color theme |
| `:filter <pattern>` | Filter devices |
| `:clear` | Clear filter |

## Tips

1. **Quick Device Toggle**: Press `t` on any device in the list to toggle it
2. **View JSON**: Press `Enter` on a device to see raw JSON status
3. **Resize Friendly**: TUI automatically adjusts layout when terminal is resized
4. **Filter First**: Use `/` to filter to specific devices before batch operations
5. **Keyboard Only**: All features are accessible without a mouse

## Troubleshooting

**TUI appears garbled:**
- Resize your terminal window
- Ensure terminal supports 256 colors
- Try a different terminal emulator

**Devices not showing:**
- Verify devices are registered via `shelly discover --register` or `shelly init --device`
- Check network connectivity
- Try `shelly device list` to verify configuration

**Slow refresh:**
- Increase refresh interval with `--refresh 10s`
- Check device network connectivity
- Reduce number of configured devices

**Tab switching not working:**
- Ensure you're not in an input field (press `Esc` first)
- Check help overlay isn't open (press `?` to toggle)
