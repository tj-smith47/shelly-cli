---
title: "shelly cover open"
description: "shelly cover open"
---

## shelly cover open

Open cover

### Synopsis

Open a cover/roller component on the specified device.

```
shelly cover open <device> [flags]
```

### Examples

```
  # Open cover fully
  shelly cover open bedroom

  # Open cover for 5 seconds
  shelly cover up bedroom --duration 5

  # Open cover by name
  shelly cover open bedroom --name "Living Room Blinds"
```

### Options

```
  -d, --duration int   Duration in seconds (0 = full open)
  -h, --help           help for open
  -i, --id int         Cover component ID (default 0)
  -n, --name string    Cover name (alternative to --id)
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

* [shelly cover](shelly_cover.md)	 - Control cover/roller components

