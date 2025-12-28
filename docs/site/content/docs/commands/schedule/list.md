---
title: "shelly schedule list"
description: "shelly schedule list"
---

## shelly schedule list

List schedules on a device

### Synopsis

List all schedules on a Gen2+ Shelly device.

Shows schedule ID, enabled status, timespec (cron-like syntax), and the
RPC calls to execute. Schedules allow time-based automation of device
actions without external home automation systems.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Enabled, Timespec, Calls

```
shelly schedule list <device> [flags]
```

### Examples

```
  # List all schedules
  shelly schedule list living-room

  # Output as JSON
  shelly schedule list living-room -o json

  # Get enabled schedules only
  shelly schedule list living-room -o json | jq '.[] | select(.enable)'

  # Extract timespecs for enabled schedules
  shelly schedule list living-room -o json | jq -r '.[] | select(.enable) | .timespec'

  # Count total schedules
  shelly schedule list living-room -o json | jq length

  # Short form
  shelly sched ls living-room
```

### Options

```
  -h, --help   help for list
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

* [shelly schedule](shelly_schedule.md)	 - Manage device schedules

