---
title: "shelly rgb"
description: "shelly rgb"
weight: 510
sidebar:
  collapsed: true
---

## shelly rgb

Control RGB light components

### Synopsis

Control RGB light components on Shelly devices.

### Examples

```
  # Turn on RGB light
  shelly rgb on living-room

  # Set RGB color to red
  shelly rgb set living-room --red 255 --green 0 --blue 0

  # Check RGB status
  shelly rgb status living-room
```

### Options

```
  -h, --help   help for rgb
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
* [shelly rgb list](shelly_rgb_list.md)	 - List rgb components
* [shelly rgb off](shelly_rgb_off.md)	 - Turn rgb off
* [shelly rgb on](shelly_rgb_on.md)	 - Turn rgb on
* [shelly rgb set](shelly_rgb_set.md)	 - Set RGB parameters
* [shelly rgb status](shelly_rgb_status.md)	 - Show rgb status
* [shelly rgb toggle](shelly_rgb_toggle.md)	 - Toggle rgb on/off

