---
title: "shelly switch list"
description: "shelly switch list"
---

## shelly switch list

List switch components

### Synopsis

List all switch components on the specified device with their current status.

Switch components control relay outputs (on/off). Each switch has an ID,
optional name, current state (ON/OFF), and power consumption if supported.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State (ON/OFF), Power (watts)

```
shelly switch list <device> [flags]
```

### Examples

```
  # List all switches on a device
  shelly switch list kitchen

  # List switches with JSON output
  shelly switch list kitchen -o json

  # Get switches that are currently ON
  shelly switch list kitchen -o json | jq '.[] | select(.output == true)'

  # Calculate total power consumption
  shelly switch list kitchen -o json | jq '[.[].power] | add'

  # Get switch IDs only
  shelly switch list kitchen -o json | jq -r '.[].id'

  # Check all switches across multiple devices
  for dev in kitchen bedroom living-room; do
    echo "=== $dev ==="
    shelly switch list "$dev" --no-color
  done

  # Short forms
  shelly switch ls kitchen
  shelly sw ls kitchen
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly switch](shelly_switch.md)	 - Control switch components

