---
title: "shelly sensor smoke mute"
description: "shelly sensor smoke mute"
---

## shelly sensor smoke mute

Mute smoke alarm

### Synopsis

Mute an active smoke alarm.

The alarm will remain muted until the condition clears
and potentially re-triggers.

```
shelly sensor smoke mute <device> [flags]
```

### Examples

```
  # Mute smoke alarm
  shelly sensor smoke mute <device>

  # Mute specific sensor
  shelly sensor smoke mute <device> --id 1
```

### Options

```
  -h, --help     help for mute
      --id int   Sensor ID (default 0)
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

* [shelly sensor smoke](shelly_sensor_smoke.md)	 - Manage smoke sensors

