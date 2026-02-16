---
title: "shelly api methods"
description: "shelly api methods"
---

## shelly api methods

List available RPC methods (Gen2+ only)

### Synopsis

List all RPC methods available on a Shelly device.

This shows the methods you can call using 'shelly api <device> <Method>'.
Use --filter to search for specific methods by name.

Note: This command only works with Gen2+ devices as Gen1 devices don't
support RPC method introspection.

```
shelly api methods <device> [flags]
```

### Examples

```
  # List all methods
  shelly api methods living-room

  # Filter methods containing "Switch"
  shelly api methods living-room --filter Switch

  # Output as JSON
  shelly api methods living-room --json
```

### Options

```
      --filter string   Filter methods by name (case-insensitive)
  -f, --format string   Output format: text, json (default "text")
  -h, --help            help for methods
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

* [shelly api](shelly_api.md)	 - Execute API calls on Shelly devices

