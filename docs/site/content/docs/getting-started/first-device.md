---
title: "Adding Your First Device"
description: "Detailed guide to registering and configuring your first Shelly device"
weight: 30
---

This guide provides detailed instructions for adding your first Shelly device to the CLI.

## Prerequisites

Before you begin, ensure you have:

- Shelly CLI installed ([Installation Guide](/docs/getting-started/installation/))
- A Shelly device powered on and connected to your WiFi
- The device IP address or mDNS hostname
- Device authentication credentials (if password-protected)

## Finding Your Device

### Option 1: Use Discovery

The easiest way to find your device:

```bash
shelly discover
```

This scans your local network using mDNS and returns all discovered Shelly devices.

### Option 2: Check Your Router

Look in your router's DHCP client list for devices starting with "shelly".

### Option 3: Use the Shelly App

If you've already configured the device with the Shelly app:
1. Open the Shelly app
2. Tap on your device
3. Go to Settings → Device Information
4. Note the IP address

### Option 4: Device Access Point

New, unconfigured devices broadcast their own WiFi:
1. Connect to the device's AP (e.g., `shelly-plus-1-XXXXXX`)
2. The device is at `192.168.33.1`
3. Configure WiFi through the web UI first

## Registering the Device

### Basic Registration

```bash
shelly device add <name> <address>
```

**Example:**
```bash
shelly device add kitchen 192.168.1.100
```

**Rules for device names:**
- Use lowercase letters, numbers, and hyphens
- No spaces (use hyphens instead)
- Keep it short and descriptive
- Examples: `living-room`, `kitchen-light`, `garage-door`

### Registration with Authentication

If your device has a password set:

```bash
shelly device add kitchen 192.168.1.100 --user admin --password yourpassword
```

**Security tip:** Avoid passwords in shell history by using a prompt:

```bash
shelly device add kitchen 192.168.1.100 --user admin --password-stdin <<< "yourpassword"
```

### Registration with Generation Hint

If auto-detection fails, specify the generation:

```bash
shelly device add kitchen 192.168.1.100 --generation 2
```

| Generation | Devices |
|------------|---------|
| `1` | Shelly 1, 2, 2.5, Plug, Bulb, RGBW, Dimmer |
| `2` | Shelly Plus 1/2PM, Pro 1/2/4PM |
| `3` | Shelly Plus 1 Mini, i4 DC |
| `4` | Shelly Wall Display |

### Updating an Existing Device

To update credentials or address:

```bash
shelly device add kitchen 192.168.1.100 --force --user admin --password newpassword
```

The `--force` flag overwrites existing device configuration.

## Verifying Connection

After registration, verify the device responds:

```bash
# Get device info
shelly device info kitchen
```

**Expected output:**
```
Device: kitchen
Address: 192.168.1.100
Model: SNSW-001X16EU (Shelly Plus 1)
Generation: 2
Firmware: 1.0.0 (stable)
MAC: AABBCCDDEEFF
Uptime: 2d 5h 30m
WiFi: Connected (-45 dBm)
Cloud: Disabled
Auth: Enabled
```

```bash
# Check status
shelly status kitchen
```

```bash
# Test control
shelly toggle kitchen
```

## Configuration File

Your device is stored in `~/.config/shelly/config.yaml`:

```yaml
devices:
  kitchen:
    address: 192.168.1.100
    generation: 2
    model: SNSW-001X16EU
    auth:
      user: admin
      password: yourpassword
```

**Important:** Passwords are stored in plain text. Ensure the config file has restricted permissions:

```bash
chmod 600 ~/.config/shelly/config.yaml
```

## Troubleshooting

### "Device not found" or Connection Timeout

```bash
# Ping the device
ping 192.168.1.100

# Check if device responds to HTTP
curl http://192.168.1.100/shelly

# For Gen2+, try RPC endpoint
curl http://192.168.1.100/rpc/Shelly.GetDeviceInfo
```

**Possible causes:**
- Device on different subnet
- Firewall blocking connections
- Device in AP mode (not connected to your WiFi)

### Authentication Failed

```bash
# Verify credentials via curl
curl -u admin:yourpassword http://192.168.1.100/rpc/Shelly.GetDeviceInfo
```

**Possible causes:**
- Incorrect username (try `admin`)
- Incorrect password
- Auth not enabled on device (try without credentials)

### Generation Detection Failed

Explicitly specify the generation:

```bash
shelly device add kitchen 192.168.1.100 --generation 2
```

To determine generation:
- Check model number on the device
- Gen1: Model starts with `SH` (e.g., SHSW-1)
- Gen2+: Model starts with `SN` (e.g., SNSW-001X16EU)

### Device Shows Offline

```bash
# Check device status
shelly device info kitchen

# Try direct status
shelly status 192.168.1.100
```

**Possible causes:**
- IP address changed (use discovery to find new IP)
- Device rebooting
- Network issues

## Next Steps

Now that you've added your first device:

- [Create Device Groups](/docs/configuration/#device-groups) - Group multiple devices
- [Set Up Scenes](/docs/configuration/#scenes) - Create multi-device scenes
- [Configure Aliases](/docs/guides/aliases/) - Create command shortcuts
- [Explore the TUI](/docs/guides/tui-dashboard/) - Visual device control
