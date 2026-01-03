---
title: "shelly thermostat schedule enable"
description: "shelly thermostat schedule enable"
---

## shelly thermostat schedule enable

Enable a schedule

### Synopsis

Enable a disabled schedule so it will run at its scheduled times.

```
shelly thermostat schedule enable <device> [flags]
```

### Examples

```
  # Enable schedule by ID
  shelly thermostat schedule enable gateway --id 1
```

### Options

```
  -h, --help     help for enable
      --id int   Schedule ID to enable (required)
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

