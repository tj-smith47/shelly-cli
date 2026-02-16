---
title: "shelly cover close"
description: "shelly cover close"
---

## shelly cover close

Close cover

### Synopsis

Close a cover/roller component on the specified device.

```
shelly cover close <device> [flags]
```

### Examples

```
  # Close cover fully
  shelly cover close bedroom

  # Close cover for 5 seconds
  shelly cover down bedroom --duration 5

  # Close cover by name
  shelly cover close bedroom --name "Living Room Blinds"
```

### Options

```
  -d, --duration int   Duration in seconds (0 = full close)
  -h, --help           help for close
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

