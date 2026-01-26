---
title: "shelly monitor"
description: "shelly monitor"
weight: 480
sidebar:
  collapsed: true
---

## shelly monitor

Real-time device monitoring

### Synopsis

Real-time monitoring of Shelly devices.

Monitor device status, power consumption, and events in real-time
with automatic refresh and color-coded status changes.

### Examples

```
  # Monitor device status
  shelly monitor status kitchen

  # Monitor power consumption
  shelly mon power living-room

  # Monitor events from a device
  shelly monitor events kitchen
```

### Options

```
  -h, --help   help for monitor
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
* [shelly monitor all](shelly_monitor_all.md)	 - Monitor all registered devices
* [shelly monitor events](shelly_monitor_events.md)	 - Monitor device events in real-time
* [shelly monitor power](shelly_monitor_power.md)	 - Monitor power consumption in real-time
* [shelly monitor status](shelly_monitor_status.md)	 - Monitor device status in real-time

