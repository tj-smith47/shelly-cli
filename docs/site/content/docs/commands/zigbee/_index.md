---
title: "shelly zigbee"
description: "shelly zigbee"
weight: 890
sidebar:
  collapsed: true
---

## shelly zigbee

Manage Zigbee connectivity

### Synopsis

Manage Zigbee connectivity on Shelly devices.

Zigbee support is available on Gen4 devices that can operate as
Zigbee end devices, connecting to Zigbee coordinators like
Home Assistant (ZHA), Zigbee2MQTT, or other compatible systems.

When operating in Zigbee mode, the device joins a Zigbee network
and can be controlled through the Zigbee coordinator instead of
or in addition to WiFi/HTTP control.

### Examples

```
  # Show Zigbee status
  shelly zigbee status living-room

  # Start pairing to join a network
  shelly zigbee pair living-room

  # List Zigbee-capable devices on network
  shelly zigbee list
```

### Options

```
  -h, --help   help for zigbee
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly zigbee list](shelly_zigbee_list.md)	 - List Zigbee-capable devices
* [shelly zigbee pair](shelly_zigbee_pair.md)	 - Start Zigbee network pairing
* [shelly zigbee remove](shelly_zigbee_remove.md)	 - Leave Zigbee network
* [shelly zigbee status](shelly_zigbee_status.md)	 - Show Zigbee network status

