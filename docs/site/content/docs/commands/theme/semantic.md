---
title: "shelly theme semantic"
description: "shelly theme semantic"
---

## shelly theme semantic

Show semantic color mappings

### Synopsis

Show the current semantic color mappings for the active theme.

Semantic colors provide consistent meaning across the CLI:
- Primary/Secondary: Main UI colors
- Success/Warning/Error/Info: Feedback colors
- Online/Offline/Updating/Idle: Device state colors
- TableHeader/TableCell/TableAltCell/TableBorder: Table styling

```
shelly theme semantic [flags]
```

### Examples

```
  # Show semantic colors
  shelly theme semantic

  # Output as JSON
  shelly theme semantic -o json
```

### Options

```
  -h, --help   help for semantic
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

* [shelly theme](shelly_theme.md)	 - Manage CLI color themes

