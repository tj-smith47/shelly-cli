---
title: "shelly modbus status"
description: "shelly modbus status"
---

## shelly modbus status

Show Modbus-TCP status

### Synopsis

Show the current Modbus-TCP server status and configuration.

```
shelly modbus status <device> [flags]
```

### Examples

```
  # Show Modbus status
  shelly modbus status kitchen

  # JSON output
  shelly modbus status kitchen -o json
```

### Options

```
  -h, --help            help for status
  -o, --output string   Output format: table, json, yaml (default "table")
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
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly modbus](shelly_modbus.md)	 - Manage Modbus-TCP configuration

