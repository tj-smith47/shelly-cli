---
title: "shelly status"
description: "shelly status"
weight: 560
sidebar:
  collapsed: true
---

## shelly status

Show device status (quick overview)

### Synopsis

Show a quick status overview for a device or all registered devices.

If no device is specified, shows a summary of all registered devices
with their online/offline status and primary component state.

```
shelly status [device] [flags]
```

### Examples

```
  # Show status for a specific device
  shelly status living-room

  # Show status for all devices
  shelly status
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

