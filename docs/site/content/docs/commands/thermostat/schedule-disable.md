---
title: "shelly thermostat schedule disable"
description: "shelly thermostat schedule disable"
---

## shelly thermostat schedule disable

Disable a schedule

### Synopsis

Disable a schedule so it will not run until re-enabled.

```
shelly thermostat schedule disable <device> [flags]
```

### Examples

```
  # Disable schedule by ID
  shelly thermostat schedule disable gateway --id 1
```

### Options

```
  -h, --help     help for disable
      --id int   Schedule ID to disable (required)
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

* [shelly thermostat schedule](shelly_thermostat_schedule.md)	 - Manage thermostat schedules

