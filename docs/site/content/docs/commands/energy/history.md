---
title: "shelly energy history"
description: "shelly energy history"
---

## shelly energy history

Show energy consumption history

### Synopsis

Retrieve and display historical energy consumption data.

Shows voltage, current, power, and energy measurements stored by the
device over time (up to 60 days of 1-minute interval data).

Works with:
  - EM components (3-phase energy monitors)
  - EM1 components (single-phase energy monitors)

The device must have EMData or EM1Data components that store historical
measurements. Not all Shelly devices support historical data storage.

```
shelly energy history <device> [id] [flags]
```

### Examples

```
  # Show last 24 hours of energy data
  shelly energy history shelly-3em-pro

  # Show specific time range
  shelly energy history shelly-em --from "2025-01-01" --to "2025-01-07"

  # Show last week for specific component
  shelly energy history shelly-3em-pro 0 --period week

  # Limit number of records shown
  shelly energy history shelly-em --limit 100
```

### Options

```
      --from string     Start time (RFC3339 or YYYY-MM-DD)
  -h, --help            help for history
      --limit int       Limit number of data points (0 = no limit)
  -p, --period string   Time period (hour, day, week, month)
      --to string       End time (RFC3339 or YYYY-MM-DD)
      --type string     Component type (auto, em, em1) (default "auto")
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

* [shelly energy](shelly_energy.md)	 - Energy monitoring operations (EM/EM1 components)

