---
title: "shelly dash"
description: "shelly dash"
weight: 190
sidebar:
  collapsed: true
---

## shelly dash

Launch interactive TUI dashboard

### Synopsis

Launch the interactive TUI dashboard for managing Shelly devices.

The dashboard provides real-time monitoring, device control, and energy
tracking in a full-screen terminal interface.

Navigation:
  j/k or arrows  Navigate up/down
  h/l            Select component within device
  t              Toggle device/component
  o/O            Turn on/off
  R              Reboot device
  /              Filter devices
  :              Command mode
  ?              Show keyboard shortcuts
  q              Quit

```
shelly dash [flags]
```

### Examples

```
  # Launch dashboard with default settings
  shelly dash

  # Start with a device filter
  shelly dash --filter kitchen
```

### Options

```
      --filter string   Filter devices by name pattern
  -h, --help            help for dash
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
* [shelly dash devices](shelly_dash_devices.md)	 - Launch dashboard in devices view
* [shelly dash events](shelly_dash_events.md)	 - Launch dashboard in events view

