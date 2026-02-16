---
title: "shelly thermostat set"
description: "shelly thermostat set"
---

## shelly thermostat set

Set thermostat configuration

### Synopsis

Set thermostat configuration options.

Allows setting:
- Target temperature (--target)
- Operating mode (--mode): heat, cool, or auto
- Enable/disable state (--enable/--disable)

```
shelly thermostat set <device> [flags]
```

### Examples

```
  # Set target temperature to 22°C
  shelly thermostat set gateway --target 22

  # Set mode to heat
  shelly thermostat set gateway --mode heat

  # Set target and mode together
  shelly thermostat set gateway --target 21 --mode auto

  # Enable thermostat
  shelly thermostat set gateway --enable
```

### Options

```
      --disable        Disable thermostat
      --enable         Enable thermostat
  -h, --help           help for set
  -i, --id int         Thermostat component ID (default 0)
      --mode string    Operating mode (heat, cool, auto)
      --target float   Target temperature in Celsius
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

* [shelly thermostat](shelly_thermostat.md)	 - Manage thermostats

