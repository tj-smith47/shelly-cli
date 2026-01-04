---
title: "shelly thermostat schedule create"
description: "shelly thermostat schedule create"
---

## shelly thermostat schedule create

Create a thermostat schedule

### Synopsis

Create a new schedule for thermostat control.

The schedule will execute at the specified time and set the thermostat
to the configured target temperature, mode, or enabled state.

Timespec format (cron-like):
  "ss mm hh DD WW" - seconds, minutes, hours, day of month, weekday

  Wildcards: * (any), ranges: 1-5, lists: 1,3,5, steps: 0-59/10
  Special: @sunrise, @sunset (with optional +/- offset in minutes)

Examples:
  "0 0 8 * *"     - Every day at 8:00 AM
  "0 0 7 * 1-5"   - Weekdays at 7:00 AM
  "0 30 22 * *"   - Every day at 10:30 PM
  "0 0 6 * 0,6"   - Weekends at 6:00 AM
  "@sunrise"      - At sunrise
  "@sunset-30"    - 30 minutes before sunset

```
shelly thermostat schedule create <device> [flags]
```

### Examples

```
  # Set temperature to 22°C at 7:00 AM on weekdays
  shelly thermostat schedule create gateway --target 22 --time "0 0 7 * 1-5"

  # Set temperature to 18°C at 10:00 PM every day
  shelly thermostat schedule create gateway --target 18 --time "0 0 22 * *"

  # Switch to heat mode at sunrise
  shelly thermostat schedule create gateway --mode heat --time "@sunrise"

  # Disable thermostat at midnight
  shelly thermostat schedule create gateway --disable --time "0 0 0 * *"

  # Create a disabled schedule (won't run until enabled)
  shelly thermostat schedule create gateway --target 20 --time "0 0 9 * *" --disabled
```

### Options

```
      --disable             Disable the thermostat
      --enable              Enable the thermostat
      --enabled             Whether the schedule itself is enabled (default true)
  -h, --help                help for create
      --mode string         Thermostat mode (heat, cool, auto)
      --target float        Target temperature in Celsius
      --thermostat-id int   Thermostat component ID
  -t, --time string         Schedule timespec (required)
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

* [shelly thermostat schedule](shelly_thermostat_schedule.md)	 - Manage thermostat schedules

