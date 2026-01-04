---
title: "shelly light"
description: "shelly light"
weight: 310
sidebar:
  collapsed: true
---

## shelly light

Control light components

### Synopsis

Control dimmable light components on Shelly devices.

### Examples

```
  # Turn on a light
  shelly light on kitchen

  # Set brightness to 50%
  shelly light set kitchen --brightness 50

  # Check light status
  shelly lt status bedroom
```

### Options

```
  -h, --help   help for light
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
* [shelly light list](shelly_light_list.md)	 - List light components
* [shelly light off](shelly_light_off.md)	 - Turn light off
* [shelly light on](shelly_light_on.md)	 - Turn light on
* [shelly light set](shelly_light_set.md)	 - Set light parameters
* [shelly light status](shelly_light_status.md)	 - Show light status
* [shelly light toggle](shelly_light_toggle.md)	 - Toggle light on/off

