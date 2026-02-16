---
title: "shelly power status"
description: "shelly power status"
---

## shelly power status

Show power meter status

### Synopsis

Show current status of a power meter component.

Displays real-time measurements including voltage, current, power,
frequency, and accumulated energy.

```
shelly power status <device> [id] [flags]
```

### Examples

```
  # Show power meter status
  shelly power status living-room

  # Show specific component by ID
  shelly power status living-room 0

  # Specify component type explicitly
  shelly power status living-room --type pm1

  # Output as JSON for scripting
  shelly power status living-room -o json
```

### Options

```
  -h, --help          help for status
      --type string   Component type (auto, pm, pm1) (default "auto")
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
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

* [shelly power](shelly_power.md)	 - Power meter operations (PM/PM1 components)

