---
title: "shelly rgb toggle"
description: "shelly rgb toggle"
---

## shelly rgb toggle

Toggle rgb on/off

### Synopsis

Toggle a rgb component on or off on the specified device.

```
shelly rgb toggle <device> [flags]
```

### Examples

```
  # Toggle rgb
  shelly rgb toggle <device>

  # Toggle specific rgb by ID
  shelly rgb flip <device> --id 1

  # Toggle rgb by name
  shelly rgb toggle <device> --name "Kitchen Light"
```

### Options

```
  -h, --help          help for toggle
  -i, --id int        RGB component ID (default 0)
  -n, --name string   RGB name (alternative to --id)
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

* [shelly rgb](shelly_rgb.md)	 - Control RGB light components

