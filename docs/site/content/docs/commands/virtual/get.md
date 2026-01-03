---
title: "shelly virtual get"
description: "shelly virtual get"
---

## shelly virtual get

Get a virtual component value

### Synopsis

Get the current value of a virtual component.

The key format is "type:id", for example "boolean:200" or "number:201".

```
shelly virtual get <device> <key> [flags]
```

### Examples

```
  # Get a boolean component
  shelly virtual get kitchen boolean:200

  # Get a number component with JSON output
  shelly virtual get kitchen number:201 -o json
```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly virtual](shelly_virtual.md)	 - Manage virtual components

