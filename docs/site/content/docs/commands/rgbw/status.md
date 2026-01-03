---
title: "shelly rgbw status"
description: "shelly rgbw status"
---

## shelly rgbw status

Show rgbw status

### Synopsis

Show the current status of a rgbw component on the specified device.

```
shelly rgbw status <device> [flags]
```

### Examples

```
  # Show rgbw status
  shelly rgbw status <device>

  # Show status with JSON output
  shelly rgbw st <device> -o json
```

### Options

```
  -h, --help     help for status
  -i, --id int   RGBW component ID (default 0)
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

* [shelly rgbw](shelly_rgbw.md)	 - Control RGBW LED outputs

