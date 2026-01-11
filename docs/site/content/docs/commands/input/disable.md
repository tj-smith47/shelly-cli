---
title: "shelly input disable"
description: "shelly input disable"
---

## shelly input disable

Disable input component

### Synopsis

Disable an input component on a Shelly device.

When disabled, the input will not respond to physical button presses or
switch state changes. This can be useful for maintenance or to prevent
accidental triggers.

```
shelly input disable <device> [flags]
```

### Examples

```
  # Disable input on a device
  shelly input disable kitchen

  # Disable specific input by ID
  shelly input disable living-room --id 1

  # Using alias
  shelly input off bedroom
```

### Options

```
  -h, --help     help for disable
  -i, --id int   Input component ID (default 0)
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

* [shelly input](shelly_input.md)	 - Manage input components

