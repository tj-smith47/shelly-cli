---
title: "shelly energy compare"
description: "shelly energy compare"
---

## shelly energy compare

Compare energy usage between devices

### Synopsis

Compare energy consumption across multiple devices for a specified time period.

Shows each device's total energy consumption, average power, and percentage
of the total consumption. Useful for identifying high-energy consumers.

By default, compares all registered devices. Use --devices to specify a subset.

```
shelly energy compare [flags]
```

### Examples

```
  # Compare all devices for the last day
  shelly energy compare

  # Compare specific devices for the last week
  shelly energy compare --devices kitchen,living-room,garage --period week

  # Compare for a specific date range
  shelly energy compare --from "2025-01-01" --to "2025-01-07"

  # Output as JSON
  shelly energy compare -o json
```

### Options

```
      --devices strings   Devices to compare (default: all registered)
      --from string       Start time (RFC3339 or YYYY-MM-DD)
  -h, --help              help for compare
  -p, --period string     Time period (hour, day, week, month) (default "day")
      --to string         End time (RFC3339 or YYYY-MM-DD)
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

* [shelly energy](shelly_energy.md)	 - Energy monitoring operations (EM/EM1 components)

