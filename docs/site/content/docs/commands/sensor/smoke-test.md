---
title: "shelly sensor smoke test"
description: "shelly sensor smoke test"
---

## shelly sensor smoke test

Test smoke sensor

### Synopsis

Test the smoke sensor on a Shelly device.

Note: The Smoke component may not have a dedicated test method.
This command provides instructions for manual testing.

```
shelly sensor smoke test <device> [flags]
```

### Examples

```
  # Test smoke sensor
  shelly sensor smoke test <device>
```

### Options

```
  -h, --help     help for test
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

