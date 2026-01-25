---
title: "shelly toggle"
description: "shelly toggle"
weight: 800
sidebar:
  collapsed: true
---

## shelly toggle

Toggle a device (auto-detects type)

### Synopsis

Toggle a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this toggles between open and close based on current state.

By default, toggles all controllable components on the device.
Use --id to target a specific component (e.g., for multi-switch devices).

```
shelly toggle <device> [flags]
```

### Examples

```
  # Toggle all components on a device
  shelly toggle living-room

  # Toggle specific switch (for multi-switch devices)
  shelly toggle dual-switch --id 1

  # Toggle a cover
  shelly toggle bedroom-blinds
```

### Options

```
  -h, --help     help for toggle
      --id int   Component ID to control (omit to control all) (default -1)
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

