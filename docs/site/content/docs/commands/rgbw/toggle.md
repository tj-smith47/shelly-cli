---
title: "shelly rgbw toggle"
description: "shelly rgbw toggle"
---

## shelly rgbw toggle

Toggle rgbw on/off

### Synopsis

Toggle a rgbw component on or off on the specified device.

```
shelly rgbw toggle <device> [flags]
```

### Examples

```
  # Toggle rgbw
  shelly rgbw toggle <device>

  # Toggle specific rgbw by ID
  shelly rgbw flip <device> --id 1

  # Toggle rgbw by name
  shelly rgbw toggle <device> --name "Kitchen Light"
```

### Options

```
  -h, --help          help for toggle
  -i, --id int        RGBW component ID (default 0)
  -n, --name string   RGBW name (alternative to --id)
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

* [shelly rgbw](shelly_rgbw.md)	 - Control RGBW LED outputs

