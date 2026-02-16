---
title: "shelly link status"
description: "shelly link status"
---

## shelly link status

Show link status with derived device state

### Synopsis

Show the status of device links with resolved parent switch state.

When a linked child device is offline, its state is derived from the
parent switch state. If no device is specified, shows all links.

```
shelly link status [child-device] [flags]
```

### Examples

```
  # Show status of all links
  shelly link status

  # Show status for a specific linked device
  shelly link status bulb-duo

  # Output as JSON
  shelly link status -o json
```

### Options

```
  -h, --help   help for status
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

* [shelly link](shelly_link.md)	 - Manage device power links

