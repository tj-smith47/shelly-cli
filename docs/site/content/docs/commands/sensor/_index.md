---
title: "shelly sensor"
description: "shelly sensor"
weight: 570
sidebar:
  collapsed: true
---

## shelly sensor

Manage device sensors

### Synopsis

Manage environmental sensors on Shelly devices.

Supports reading from various sensor types available on Gen2+ devices:
- Device power (battery status)
- Flood sensors (water leak detection)
- Humidity sensors (DHT22, HTU21D)
- Illuminance sensors (light level)
- Smoke sensors (smoke detection with alarm)
- Temperature sensors (built-in or external DS18B20)
- Voltmeters (voltage measurement)

Use the status command to get a combined view of all sensors on a device,
or use specific subcommands for individual sensor types.

### Examples

```
  # Show all sensor readings
  shelly sensor status living-room

  # Get temperature reading
  shelly sensor temperature status living-room

  # Check flood sensor
  shelly sensor flood status bathroom

  # Mute smoke alarm
  shelly sensor smoke mute kitchen
```

### Options

```
  -h, --help   help for sensor
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly sensor devicepower](shelly_sensor_devicepower.md)	 - Manage device power sensors
* [shelly sensor flood](shelly_sensor_flood.md)	 - Manage flood sensors
* [shelly sensor humidity](shelly_sensor_humidity.md)	 - Manage humidity sensors
* [shelly sensor illuminance](shelly_sensor_illuminance.md)	 - Manage illuminance sensors
* [shelly sensor smoke](shelly_sensor_smoke.md)	 - Manage smoke sensors
* [shelly sensor status](shelly_sensor_status.md)	 - Show all sensor readings
* [shelly sensor temperature](shelly_sensor_temperature.md)	 - Manage temperature sensors
* [shelly sensor voltmeter](shelly_sensor_voltmeter.md)	 - Manage voltmeter sensors

