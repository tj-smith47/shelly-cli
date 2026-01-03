---
title: "shelly firmware status"
description: "shelly firmware status"
---

## shelly firmware status

Show firmware status

### Synopsis

Show the current firmware status for a device.

Displays update status, available versions, and rollback availability.

```
shelly firmware status <device> [flags]
```

### Examples

```
  # Show firmware status
  shelly firmware status living-room
```

### Options

```
  -h, --help   help for status
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

* [shelly firmware](shelly_firmware.md)	 - Manage device firmware

