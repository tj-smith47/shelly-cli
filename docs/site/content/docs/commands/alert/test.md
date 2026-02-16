---
title: "shelly alert test"
description: "shelly alert test"
---

## shelly alert test

Test an alert by triggering it

### Synopsis

Test an alert by manually triggering its action.

This simulates the alert condition being met and executes the configured action.

```
shelly alert test <name> [flags]
```

### Examples

```
  # Test an alert
  shelly alert test kitchen-offline
```

### Options

```
  -h, --help   help for test
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

* [shelly alert](shelly_alert.md)	 - Manage monitoring alerts

