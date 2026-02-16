---
title: "shelly zigbee status"
description: "shelly zigbee status"
---

## shelly zigbee status

Show Zigbee network status

### Synopsis

Show Zigbee network status for a Shelly device.

Displays the current Zigbee state including:
- Whether Zigbee is enabled
- Network state (not_configured, ready, steering, joined)
- EUI64 address (device's Zigbee identifier)
- PAN ID and channel when joined to a network
- Coordinator information

```
shelly zigbee status <device> [flags]
```

### Examples

```
  # Show Zigbee status
  shelly zigbee status living-room

  # Output as JSON
  shelly zigbee status living-room --json
```

### Options

```
  -f, --format string   Output format: text, json (default "text")
  -h, --help            help for status
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

* [shelly zigbee](shelly_zigbee.md)	 - Manage Zigbee connectivity

