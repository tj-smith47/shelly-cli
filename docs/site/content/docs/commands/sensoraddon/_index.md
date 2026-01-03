---
title: "shelly sensoraddon"
description: "shelly sensoraddon"
weight: 560
sidebar:
  collapsed: true
---

## shelly sensoraddon

Manage Sensor Add-on peripherals

### Synopsis

Manage Sensor Add-on peripherals on Shelly Gen2+ devices.

The Sensor Add-on board allows connecting external sensors:
  - DS18B20: Dallas 1-Wire temperature sensors
  - DHT22: Temperature and humidity sensor
  - Digital inputs
  - Analog inputs

Supported devices: Plus1, Plus1PM, Plus2PM, PlusI4, Plus10V, PlusRGBWPM,
Dimmer0110VPM G3, Shelly1G3, Shelly1PMG3, Shelly2PMG3, ShellyI4G3

Note: Peripheral changes require a device reboot to take effect.

### Examples

```
  # List configured peripherals
  shelly sensoraddon list kitchen

  # Scan for OneWire devices
  shelly sensoraddon scan kitchen

  # Add a DS18B20 sensor
  shelly sensoraddon add kitchen ds18b20 --addr "40:255:100:6:199:204:149:177"

  # Add a DHT22 sensor
  shelly sensoraddon add kitchen dht22

  # Remove a peripheral
  shelly sensoraddon remove kitchen temperature:100
```

### Options

```
  -h, --help   help for sensoraddon
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly sensoraddon add](shelly_sensoraddon_add.md)	 - Add a peripheral
* [shelly sensoraddon list](shelly_sensoraddon_list.md)	 - List configured peripherals
* [shelly sensoraddon remove](shelly_sensoraddon_remove.md)	 - Remove a peripheral
* [shelly sensoraddon scan](shelly_sensoraddon_scan.md)	 - Scan for OneWire devices

