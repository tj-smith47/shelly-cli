---
title: "shelly thermostat calibrate"
description: "shelly thermostat calibrate"
---

## shelly thermostat calibrate

Calibrate thermostat valve

### Synopsis

Initiate valve calibration on a thermostat.

Calibration helps the thermostat learn the full range of valve
movement. This should be performed:
- After initial installation
- If the valve behavior seems incorrect
- After physical maintenance on the valve

The calibration process takes a few minutes. The valve will
move through its full range to determine open/close positions.

```
shelly thermostat calibrate <device> [flags]
```

### Examples

```
  # Calibrate thermostat
  shelly thermostat calibrate gateway

  # Calibrate specific thermostat
  shelly thermostat calibrate gateway --id 1
```

### Options

```
  -h, --help     help for calibrate
  -i, --id int   Thermostat component ID (default 0)
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

