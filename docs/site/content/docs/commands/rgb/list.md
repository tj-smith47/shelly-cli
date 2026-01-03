---
title: "shelly rgb list"
description: "shelly rgb list"
---

## shelly rgb list

List rgb components

### Synopsis

List all RGB light components on the specified device with their current status.

RGB components control color-capable lights (RGBW, RGBW2, etc.). Each
component has an ID, optional name, state (ON/OFF), RGB color values,
brightness level, and power consumption if supported.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State (ON/OFF), Color (R:G:B), Brightness (%), Power

```
shelly rgb list <device> [flags]
```

### Examples

```
  # List all RGB components on a device
  shelly rgb list living-room

  # List RGB components with JSON output
  shelly rgb list living-room -o json

  # Get RGB lights that are currently ON
  shelly rgb list living-room -o json | jq '.[] | select(.output == true)'

  # Get current color values
  shelly rgb list living-room -o json | jq '.[] | {id, r: .rgb.r, g: .rgb.g, b: .rgb.b}'

  # Find lights set to pure red
  shelly rgb list living-room -o json | jq '.[] | select(.rgb.r == 255 and .rgb.g == 0 and .rgb.b == 0)'

  # Get brightness levels
  shelly rgb list living-room -o json | jq '.[] | {id, brightness}'

  # Short forms
  shelly rgb ls living-room
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

* [shelly rgb](shelly_rgb.md)	 - Control RGB light components

