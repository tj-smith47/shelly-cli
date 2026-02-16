---
title: "shelly cover position"
description: "shelly cover position"
---

## shelly cover position

Set cover position

### Synopsis

Set a cover/roller component to a specific position.

Position is specified as a percentage from 0 (closed) to 100 (open).

```
shelly cover position <device> <percent> [flags]
```

### Examples

```
  # Set cover to 50% open
  shelly cover position my-cover 50

  # Set cover 1 to fully open
  shelly cover position my-cover 100 --id 1
```

### Options

```
  -h, --help     help for position
  -i, --id int   Cover component ID (default 0)
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

