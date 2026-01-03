---
title: "shelly sensor temperature"
description: "shelly sensor temperature"
---

## shelly sensor temperature

Manage temperature sensors

### Synopsis

Manage temperature sensors on Shelly devices.

Temperature sensors can be built-in or external (DS18B20).
Readings are provided in both Celsius and Fahrenheit.

### Examples

```
  # List temperature sensors
  shelly sensor temperature list living-room

  # Get temperature reading
  shelly sensor temperature status living-room
```

### Options

```
  -h, --help   help for temperature
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

* [shelly sensor](shelly_sensor.md)	 - Manage device sensors
* [shelly sensor temperature list](shelly_sensor_temperature_list.md)	 - List temperature sensors
* [shelly sensor temperature status](shelly_sensor_temperature_status.md)	 - Get temperature sensor status

