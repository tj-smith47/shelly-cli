---
title: "shelly log show"
description: "shelly log show"
---

## shelly log show

Show recent log entries

### Synopsis

Show the most recent log entries from the CLI log file.

```
shelly log show [flags]
```

### Examples

```
  # Show last 50 lines (default)
  shelly log show

  # Show last 100 lines
  shelly log show -n 100
```

### Options

```
  -h, --help        help for show
  -n, --lines int   Number of lines to show (default 50)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

* [shelly log](shelly_log.md)	 - Manage CLI logs

