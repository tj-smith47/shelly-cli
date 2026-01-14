---
title: "shelly light off"
description: "shelly light off"
---

## shelly light off

Turn light off

### Synopsis

Turn off a light component on the specified device.

```
shelly light off <device> [flags]
```

### Examples

```
  # Turn off light
  shelly light off <device>

  # Turn off specific light by ID
  shelly light off <device> --id 1

  # Turn off light by name
  shelly light off <device> --name "Kitchen Light"
```

### Options

```
  -h, --help          help for off
  -i, --id int        Light component ID (default 0)
  -n, --name string   Light name (alternative to --id)
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

* [shelly light](shelly_light.md)	 - Control light components

