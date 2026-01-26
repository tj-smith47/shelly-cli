---
title: "shelly debug"
description: "shelly debug"
weight: 180
sidebar:
  collapsed: true
---

## shelly debug

Debug and diagnostic commands

### Synopsis

Debug and diagnostic commands for troubleshooting Shelly devices.

These commands provide low-level access to device communication protocols
and diagnostic information. Use them for debugging issues or exploring
device capabilities.

For direct API calls, use 'shelly api' instead.

WARNING: Some debug commands may affect device behavior. Use with caution.

### Examples

```
  # Get Gen1 device debug log
  shelly debug log living-room-gen1

  # Show CoIoT status
  shelly debug coiot living-room

  # Debug WebSocket connection
  shelly debug websocket living-room
```

### Options

```
  -h, --help   help for debug
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
* [shelly debug coiot](shelly_debug_coiot.md)	 - Show CoIoT/CoAP status or listen for multicast updates
* [shelly debug log](shelly_debug_log.md)	 - Get device debug log (Gen1)
* [shelly debug websocket](shelly_debug_websocket.md)	 - Debug WebSocket connection and stream events

