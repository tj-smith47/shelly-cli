---
title: "shelly fleet toggle"
description: "shelly fleet toggle"
---

## shelly fleet toggle

Toggle devices via cloud

### Synopsis

Toggle devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token

```
shelly fleet toggle [device...] [flags]
```

### Examples

```
  # Toggle specific device
  shelly fleet toggle device-id

  # Toggle all devices in a group
  shelly fleet toggle --group living-room

  # Toggle all relay devices
  shelly fleet toggle --all
```

### Options

```
  -a, --all                Target all registered devices
  -c, --concurrent int     Max concurrent operations (default 5)
  -g, --group string       Target device group
  -h, --help               help for toggle
  -s, --switch int         Switch component ID
  -t, --timeout duration   Timeout per device (default 10s)
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

* [shelly fleet](shelly_fleet.md)	 - Cloud-based fleet management

