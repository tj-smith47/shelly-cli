# shelly-notify

Desktop notifications for Shelly device events.

## Installation

```bash
# Install directly
shelly plugin install ./shelly-notify

# Or copy to plugins directory
cp shelly-notify ~/.config/shelly/plugins/
chmod +x ~/.config/shelly/plugins/shelly-notify
```

## Dependencies

- **Linux**: `notify-send` (package: `libnotify-bin` on Debian/Ubuntu, `libnotify` on Fedora)
- **macOS**: `osascript` (built-in)
- **Optional**: `jq` for device lookup and status parsing
- **Optional**: `curl` for device communication

## Usage

```bash
# Send custom notification
shelly notify send "Kitchen" "Light turned on"

# Check device status and notify
shelly notify device kitchen

# Check if device is online
shelly notify online kitchen

# Get power consumption notification
shelly notify power kitchen

# Test notification system
shelly notify test
```

## Commands

| Command | Description |
|---------|-------------|
| `send <title> <message>` | Send a custom notification |
| `device <device>` | Show device status notification |
| `online <device>` | Check if device is online/offline |
| `power <device>` | Notify current power consumption |
| `test` | Send a test notification |
| `help` | Show help |
| `version` | Show version |

## Environment Variables

Automatically provided by Shelly CLI:

- `SHELLY_DEVICES_JSON` - JSON of registered devices
- `SHELLY_CONFIG_PATH` - Path to config file
- `SHELLY_VERBOSE` - Enable verbose output
- `SHELLY_NO_COLOR` - Disable colored output

## Examples

```bash
# Cron job to notify if device goes offline
*/5 * * * * shelly notify online kitchen 2>&1 | logger -t shelly-notify

# Script integration
if shelly notify online kitchen; then
    echo "Kitchen device is up"
else
    echo "Kitchen device is down!"
fi
```
