---
title: "shelly on"
description: "shelly on"
weight: 420
sidebar:
  collapsed: true
---

## shelly on

Turn on a device (auto-detects type)

### Synopsis

Turn on a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this opens them. For switches/lights/RGB, this turns them on.

By default, turns on all controllable components on the device.
Use --id to target a specific component (e.g., for multi-switch devices).

```
shelly on <device> [flags]
```

### Examples

```
  # Turn on all components on a device
  shelly on living-room

  # Turn on specific switch (for multi-switch devices)
  shelly on dual-switch --id 1

  # Open a cover
  shelly on bedroom-blinds
```

### Options

```
  -h, --help     help for on
      --id int   Component ID to control (omit to control all) (default -1)
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

