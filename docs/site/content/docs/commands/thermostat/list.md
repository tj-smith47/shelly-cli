---
title: "shelly thermostat list"
description: "shelly thermostat list"
---

## shelly thermostat list

List thermostats

### Synopsis

List all thermostat components on a Shelly device.

Thermostat components are typically found on Shelly BLU TRV (Thermostatic
Radiator Valve) devices connected via BLU Gateway. Each thermostat has
an ID, enabled state, and target temperature.

Use 'shelly thermostat status' for detailed readings including current
temperature. Use 'shelly thermostat set' to adjust target temperature.

Output is formatted as styled text by default. Use --json for
structured output suitable for scripting.

```
shelly thermostat list <device> [flags]
```

### Examples

```
  # List thermostats
  shelly thermostat list gateway

  # Output as JSON
  shelly thermostat list gateway --json

  # Get enabled thermostats only
  shelly thermostat list gateway --json | jq '.[] | select(.enabled == true)'

  # Get target temperatures
  shelly thermostat list gateway --json | jq '.[] | {id, target_c}'

  # Find thermostats set above 22°C
  shelly thermostat list gateway --json | jq '.[] | select(.target_c > 22)'

  # Count active thermostats
  shelly thermostat list gateway --json | jq '[.[] | select(.enabled)] | length'

  # Short form
  shelly thermostat ls gateway
```

### Options

```
  -f, --format string   Output format: text, json (default "text")
  -h, --help            help for list
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

