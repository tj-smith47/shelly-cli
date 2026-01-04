---
title: "shelly input"
description: "shelly input"
weight: 290
sidebar:
  collapsed: true
---

## shelly input

Manage input components

### Synopsis

Manage input components on Shelly devices.

### Examples

```
  # List input components on a device
  shelly input list kitchen

  # Check input status
  shelly in status kitchen

  # Trigger an input action
  shelly input trigger kitchen --id 0
```

### Options

```
  -h, --help   help for input
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
* [shelly input list](shelly_input_list.md)	 - List input components
* [shelly input status](shelly_input_status.md)	 - Show input status
* [shelly input trigger](shelly_input_trigger.md)	 - Trigger input event

