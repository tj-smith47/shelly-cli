---
title: "shelly sensoraddon add"
description: "shelly sensoraddon add"
---

## shelly sensoraddon add

Add a peripheral

### Synopsis

Add a Sensor Add-on peripheral to a device.

Peripheral types:
  ds18b20    - Dallas 1-Wire temperature sensor (requires --addr)
  dht22      - Temperature and humidity sensor
  digital_in - Digital input
  analog_in  - Analog input

Note: Changes require a device reboot to take effect.

```
shelly sensoraddon add <device> <type> [flags]
```

### Examples

```
  # Add a DS18B20 sensor
  shelly sensoraddon add kitchen ds18b20 --addr "40:255:100:6:199:204:149:177"

  # Add a DHT22 sensor
  shelly sensoraddon add kitchen dht22

  # Add with specific component ID
  shelly sensoraddon add kitchen digital_in --cid 101
```

### Options

```
      --addr string   Sensor address (required for DS18B20)
      --cid int       Component ID (auto-assigned if not specified)
  -h, --help          help for add
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly sensoraddon](shelly_sensoraddon.md)	 - Manage Sensor Add-on peripherals

