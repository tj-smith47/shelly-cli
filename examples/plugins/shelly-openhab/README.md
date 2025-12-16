# shelly-openhab

Export Shelly devices to OpenHAB configuration files.

## Installation

```bash
# Install directly
shelly plugin install ./shelly-openhab

# Or copy to plugins directory
cp shelly-openhab ~/.config/shelly/plugins/
chmod +x ~/.config/shelly/plugins/shelly-openhab
```

## Dependencies

- **jq** - JSON processing (`apt install jq` or `brew install jq`)
- **OpenHAB** 3.x or 4.x with [Shelly Binding](https://www.openhab.org/addons/bindings/shelly/)

## Usage

```bash
# Generate all configuration files
shelly openhab all

# Generate only Things file
shelly openhab things

# Generate only Items file
shelly openhab items

# List devices that would be exported
shelly openhab list

# Generate to specific directory
shelly openhab all /etc/openhab/conf/
```

## Commands

| Command | Description |
|---------|-------------|
| `things [file]` | Generate OpenHAB Things file |
| `items [file]` | Generate OpenHAB Items file |
| `all [dir]` | Generate both Things and Items files |
| `list` | List devices available for export |
| `help` | Show help |
| `version` | Show version |

## Generated Files

### shelly.things

```java
// Shelly Devices Things File
Thing shelly:shellyplus1:kitchen_light "Kitchen Light" @ "Shelly" [
    deviceIp="192.168.1.100",
    userId="",
    password=""
]
```

### shelly.items

```java
// Shelly Devices Items File
Group gShelly "Shelly Devices" <network>

Switch KITCHEN_LIGHT_Switch "Kitchen Light" <switch> (gShelly) {channel="shelly:*:kitchen_light:relay#output"}
Number:Power KITCHEN_LIGHT_Power "Kitchen Light Power [%.1f W]" <energy> (gShellyPower) {channel="shelly:*:kitchen_light:meter#currentWatts"}
```

## OpenHAB Setup

1. Install the Shelly Binding from OpenHAB UI:
   - Settings → Add-ons → Bindings → Search "Shelly" → Install

2. Generate configuration:
   ```bash
   shelly openhab all
   ```

3. Copy files to OpenHAB:
   ```bash
   sudo cp shelly.things /etc/openhab/conf/things/
   sudo cp shelly.items /etc/openhab/conf/items/
   ```

4. If devices have authentication enabled, edit `shelly.things`:
   ```java
   Thing shelly:shellyplus1:kitchen_light "Kitchen Light" @ "Shelly" [
       deviceIp="192.168.1.100",
       userId="admin",
       password="your-password"
   ]
   ```

5. Restart OpenHAB or reload configuration

## Supported Device Types

| Shelly Type | OpenHAB Thing Type |
|-------------|-------------------|
| Gen1 Relay/Switch | `shelly1` |
| Gen1 Plug | `shellyplug` |
| Gen1 Bulb | `shellybulb` |
| Gen1 RGBW2 | `shellyrgbw2` |
| Gen1 Dimmer | `shellydimmer` |
| Gen2+ Switch | `shellyplus1` |
| Gen2+ Plug | `shellyplugs` |
| Gen2+ Dimmer | `shellyplusdimmer` |
| Gen2+ Cover | `shellyplus2pm-roller` |

## Environment Variables

Automatically provided by Shelly CLI:

- `SHELLY_DEVICES_JSON` - JSON of registered devices
- `SHELLY_CONFIG_PATH` - Path to config file
- `SHELLY_NO_COLOR` - Disable colored output

## Links

- [OpenHAB Shelly Binding Docs](https://www.openhab.org/addons/bindings/shelly/)
- [OpenHAB Things Configuration](https://www.openhab.org/docs/configuration/things.html)
- [OpenHAB Items Configuration](https://www.openhab.org/docs/configuration/items.html)
