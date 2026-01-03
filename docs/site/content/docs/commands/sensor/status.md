---
title: "shelly sensor status"
description: "shelly sensor status"
---

## shelly sensor status

Show all sensor readings

### Synopsis

Show all sensor readings from a Shelly device.

Displays a combined view of all available sensors including:
- Temperature (°C/°F)
- Humidity (%)
- Flood detection status
- Smoke detection status
- Illuminance (lux)
- Voltage readings

Only sensors present on the device will be shown.

```
shelly sensor status <device> [flags]
```

### Examples

```
  # Show all sensor readings
  shelly sensor status living-room

  # Output as JSON
  shelly sensor status living-room --json
```

### Options

```
  -f, --format string   Output format: text, json (default "text")
  -h, --help            help for status
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

