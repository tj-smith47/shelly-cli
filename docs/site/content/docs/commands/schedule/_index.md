---
title: "shelly schedule"
description: "shelly schedule"
weight: 540
sidebar:
  collapsed: true
---

## shelly schedule

Manage device schedules

### Synopsis

Manage time-based schedules on Gen2+ Shelly devices.

Schedules allow you to execute RPC calls at specified times using
cron-like timespec expressions. Supports wildcards, ranges, and
special values like @sunrise and @sunset.

Note: Maximum 20 schedules per device.

### Examples

```
  # List schedules
  shelly schedule list living-room

  # Create a schedule to turn on at 8:00 AM every day
  shelly schedule create living-room --timespec "0 0 8 * *" \
    --calls '[{"method":"Switch.Set","params":{"id":0,"on":true}}]'

  # Create a schedule for sunset
  shelly schedule create living-room --timespec "@sunset" \
    --calls '[{"method":"Switch.Set","params":{"id":0,"on":false}}]'

  # Enable/disable a schedule
  shelly schedule enable living-room 1
  shelly schedule disable living-room 1

  # Delete a schedule
  shelly schedule delete living-room 1
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly schedule create](shelly_schedule_create.md)	 - Create a new schedule
* [shelly schedule delete](shelly_schedule_delete.md)	 - Delete a schedule
* [shelly schedule delete-all](shelly_schedule_delete-all.md)	 - Delete all schedules
* [shelly schedule disable](shelly_schedule_disable.md)	 - Disable a schedule
* [shelly schedule enable](shelly_schedule_enable.md)	 - Enable a schedule
* [shelly schedule list](shelly_schedule_list.md)	 - List schedules on a device
* [shelly schedule update](shelly_schedule_update.md)	 - Update a schedule

