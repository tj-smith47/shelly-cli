---
title: "shelly mqtt set"
description: "shelly mqtt set"
---

## shelly mqtt set

Configure MQTT

### Synopsis

Configure MQTT settings for a device.

Set the MQTT broker address, credentials, and topic prefix for
integration with home automation systems.

```
shelly mqtt set <device> [flags]
```

### Examples

```
  # Configure MQTT with server and credentials
  shelly mqtt set living-room --server "mqtt://broker:1883" --user user --password pass

  # Configure with custom topic prefix
  shelly mqtt set living-room --server "mqtt://broker:1883" --topic-prefix "home/shelly"

  # Enable MQTT with existing configuration
  shelly mqtt set living-room --enable
```

### Options

```
      --enable                Enable MQTT
  -h, --help                  help for set
      --password string       MQTT password
      --server string         MQTT broker URL (e.g., mqtt://broker:1883)
      --topic-prefix string   MQTT topic prefix
      --user string           MQTT username
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

* [shelly mqtt](shelly_mqtt.md)	 - Manage device MQTT configuration

