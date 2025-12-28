---
title: "shelly light status"
description: "shelly light status"
---

## shelly light status

Show light status

### Synopsis

Show the current status of a light component on the specified device.

```
shelly light status <device> [flags]
```

### Examples

```
  # Show light status
  shelly light status <device>

  # Show status with JSON output
  shelly light st <device> -o json
```

### Options

```
  -h, --help     help for status
  -i, --id int   Light component ID (default 0)
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly light](shelly_light.md)	 - Control light components

