# shelly-tasmota

A plugin for [shelly-cli](https://github.com/tj-smith47/shelly-cli) that provides full integration with Tasmota-based smart home devices.

## Features

- **Device Detection**: Automatically discover Tasmota devices during `shelly discover`
- **Status Monitoring**: View power state, sensors, and energy metrics
- **Device Control**: Control relays, switches, and lights with standard shelly-cli commands
- **Firmware Updates**: Check for and apply Tasmota firmware updates

## Supported Devices

Any device running [Tasmota](https://tasmota.github.io/docs/) firmware, including:

- Sonoff Basic, S20, S26, Pow series
- Shelly devices flashed with Tasmota
- Tuya-based smart plugs and switches
- ESP8266/ESP32 devices with Tasmota installed

## Installation

### From Source

```bash
cd examples/plugins/shelly-tasmota
go build -o shelly-tasmota .

# Install the plugin
shelly plugin install ./
```

### Manual Installation

1. Build the plugin binary:
   ```bash
   go build -o shelly-tasmota .
   ```

2. Copy to the plugins directory:
   ```bash
   mkdir -p ~/.config/shelly-cli/plugins/shelly-tasmota
   cp shelly-tasmota manifest.json ~/.config/shelly-cli/plugins/shelly-tasmota/
   ```

3. Verify installation:
   ```bash
   shelly plugin list
   ```

## Usage

### Device Discovery

Tasmota devices are automatically detected during discovery:

```bash
# Discover all devices (Shelly and Tasmota)
shelly discover

# Discover only Tasmota devices
shelly discover --platform tasmota

# Skip plugin detection (Shelly only)
shelly discover --skip-plugins
```

### Device Registration

Register a discovered Tasmota device:

```bash
# During discovery, select the device to register
shelly discover --register

# Or manually add with platform specified
shelly device add garage-plug 192.168.1.50 --platform tasmota
```

### Device Status

View status of a Tasmota device:

```bash
shelly device status garage-plug
```

Output:
```
Device: garage-plug
Platform: tasmota
Model: Sonoff Basic R3
Address: 192.168.1.50

  Status: Online

Components
COMPONENT   STATE
switch:0    on

Sensors
SENSOR      VALUE
wifi_rssi   -52 dBm

Energy
  Power:   45.3 W
  Voltage: 121.5 V
  Current: 0.372 A
  Total:   123.456 kWh
```

### Device Control

Control Tasmota devices with standard commands:

```bash
# Turn on
shelly switch on garage-plug

# Turn off
shelly switch off garage-plug

# Toggle
shelly switch toggle garage-plug

# Control specific relay on multi-channel devices
shelly switch on garage-plug --id 1
```

### Firmware Updates

Check and apply firmware updates:

```bash
# Check for updates on a specific device
shelly firmware check garage-plug

# Check all devices (including Tasmota)
shelly firmware check --all

# Interactive update workflow
shelly firmware updates

# Update all Tasmota devices to stable
shelly firmware updates --all --platform tasmota --yes

# Update to beta/development release
shelly firmware updates --devices garage-plug --beta --yes
```

## Firmware Update Workflow

The plugin integrates with the official Tasmota OTA server:

1. **Stable releases**: `http://ota.tasmota.com/tasmota/release/`
2. **Development builds**: `http://ota.tasmota.com/tasmota/`

The plugin automatically:
- Detects the chip type (ESP8266/ESP32)
- Uses the appropriate OTA URL
- Checks GitHub for the latest version
- Compares with the installed version

### Custom Firmware URL

For custom builds or specific variants:

```bash
shelly firmware update garage-plug --url http://example.com/custom-tasmota.bin.gz
```

## Authentication

For password-protected Tasmota devices, use the auth flags:

```bash
shelly device status garage-plug --auth-user admin --auth-pass secret
```

Or configure in `~/.config/shelly-cli/config.yaml`:

```yaml
devices:
  garage-plug:
    address: 192.168.1.50
    platform: tasmota
    auth:
      username: admin
      password: secret
```

## Component Types

The plugin supports these Tasmota component types:

| Component | Description |
|-----------|-------------|
| `switch`  | Relays, power outlets |
| `light`   | Dimmers, RGB bulbs |
| `sensor`  | Temperature, humidity sensors |
| `energy`  | Power monitoring devices |

## Troubleshooting

### Device Not Detected

1. Ensure the device is on the same network
2. Verify Tasmota web interface is accessible: `http://<device-ip>`
3. Check firewall settings for port 80

### Authentication Errors

1. Verify username and password in device config
2. For devices without authentication, ensure auth is disabled in Tasmota settings

### Firmware Update Fails

1. Ensure device has enough free memory
2. Check network connectivity
3. For ESP8266 devices, ensure you're using the correct variant (tasmota.bin.gz)

### GitHub Rate Limiting

The plugin queries GitHub API for version checks. Unauthenticated requests are limited to 60/hour. If you encounter rate limiting:

1. Wait for the rate limit to reset (1 hour)
2. Or run firmware checks less frequently

## Development

### Building

```bash
cd examples/plugins/shelly-tasmota
go build -o shelly-tasmota .
```

### Testing Hooks Manually

```bash
# Test detection
./shelly-tasmota detect --address 192.168.1.50

# Test status
./shelly-tasmota status --address 192.168.1.50

# Test control
./shelly-tasmota control --address 192.168.1.50 --action on --component switch --id 0

# Test firmware check
./shelly-tasmota check-updates --address 192.168.1.50

# Test firmware apply
./shelly-tasmota apply-update --address 192.168.1.50 --stage stable
```

### Hook Response Formats

All hooks return JSON. See `types.go` for the complete response schemas.

## Known Limitations

1. **No Real-Time Events**: Tasmota requires polling (no WebSocket support like Shelly)
2. **Rules/Berry**: Tasmota Rules and Berry scripting are not exposed via shelly-cli
3. **GPIO Configuration**: Template and GPIO configuration is not supported
4. **Backup/Restore**: Device backup/restore uses different format than Shelly

## License

MIT License - see the main shelly-cli repository for details.
