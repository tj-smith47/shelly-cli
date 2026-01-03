---
title: "shelly mqtt"
description: "shelly mqtt"
weight: 400
sidebar:
  collapsed: true
---

## shelly mqtt

Manage device MQTT configuration

### Synopsis

Manage MQTT configuration for devices.

Enable and configure MQTT for integration with home automation systems
like Home Assistant, OpenHAB, or custom MQTT brokers.

### Examples

```
  # Show MQTT status
  shelly mqtt status living-room

  # Configure MQTT broker
  shelly mqtt set living-room --server "mqtt://broker:1883" --user user --password pass

  # Disable MQTT
  shelly mqtt disable living-room
```

### Options

```
  -h, --help   help for mqtt
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
* [shelly mqtt disable](shelly_mqtt_disable.md)	 - Disable MQTT
* [shelly mqtt set](shelly_mqtt_set.md)	 - Configure MQTT
* [shelly mqtt status](shelly_mqtt_status.md)	 - Show MQTT status

