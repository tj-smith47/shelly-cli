---
title: "shelly batch toggle"
description: "shelly batch toggle"
---

## shelly batch toggle

Toggle switches

### Synopsis

Toggle switches multiple switch components simultaneously.

By default, targets switch component 0 on each device.
Use --switch to specify a different component ID.

Target devices can be specified multiple ways:
  - As arguments: device names or addresses
  - Via stdin: pipe device names (one per line or space-separated)
  - Via group: --group flag targets all devices in a group
  - Via all: --all flag targets all registered devices

Priority: explicit args > stdin > group > all

Stdin input supports comments (lines starting with #) and
blank lines are ignored, making it easy to use device lists
from files or other commands.

```
shelly batch toggle [device...] [flags]
```

### Examples

```
  # Toggle switches specific devices
  shelly batch toggle light-1 light-2

  # Toggle switches all devices in a group
  shelly batch toggle --group living-room

  # Toggle switches all registered devices
  shelly batch toggle --all

  # Control switch 1 on all devices in group
  shelly batch toggle --group bedroom --switch 1

  # Control concurrency and timeout
  shelly batch toggle --all --concurrent 10 --timeout 30s

  # Pipe device names from a file
  cat devices.txt | shelly batch toggle

  # Pipe from device list command
  shelly device list -o json | jq -r '.[].name' | shelly batch toggle
```

### Options

```
  -a, --all                Target all registered devices
  -c, --concurrent int     Max concurrent operations (default 5)
  -g, --group string       Target device group
  -h, --help               help for toggle
  -s, --switch int         Switch component ID
  -t, --timeout duration   Timeout per device (default 10s)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly batch](shelly_batch.md)	 - Execute commands on multiple devices

