---
title: "shelly thermostat schedule"
description: "shelly thermostat schedule"
---

## shelly thermostat schedule

Manage thermostat schedules

### Synopsis

Manage time-based schedules for thermostat control.

Schedules allow automatic temperature adjustments at specific times.
You can set different target temperatures for different times of day
or days of the week.

Schedule timespec format (cron-like):
  "ss mm hh DD WW" - seconds, minutes, hours, day of month, weekday

Special values:
  @sunrise  - At sunrise (with optional offset like @sunrise+30)
  @sunset   - At sunset (with optional offset like @sunset-15)

Examples:
  "0 0 8 * 1-5"   - 8:00 AM on weekdays
  "0 30 22 * *"   - 10:30 PM every day
  "0 0 6 * 0,6"   - 6:00 AM on weekends

### Examples

```
  # List all thermostat schedules
  shelly thermostat schedule list gateway

  # Create a morning schedule (22°C at 7:00 AM on weekdays)
  shelly thermostat schedule create gateway --target 22 --time "0 0 7 * 1-5"

  # Create a night schedule (18°C at 10:00 PM every day)
  shelly thermostat schedule create gateway --target 18 --time "0 0 22 * *"

  # Delete a schedule
  shelly thermostat schedule delete gateway --id 1
```

### Options

```
  -h, --help   help for schedule
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

* [shelly thermostat](shelly_thermostat.md)	 - Manage thermostats
* [shelly thermostat schedule create](shelly_thermostat_schedule_create.md)	 - Create a thermostat schedule
* [shelly thermostat schedule delete](shelly_thermostat_schedule_delete.md)	 - Delete a thermostat schedule
* [shelly thermostat schedule disable](shelly_thermostat_schedule_disable.md)	 - Disable a schedule
* [shelly thermostat schedule enable](shelly_thermostat_schedule_enable.md)	 - Enable a schedule
* [shelly thermostat schedule list](shelly_thermostat_schedule_list.md)	 - List thermostat schedules

