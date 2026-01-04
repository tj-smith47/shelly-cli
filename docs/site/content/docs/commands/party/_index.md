---
title: "shelly party"
description: "shelly party"
weight: 430
sidebar:
  collapsed: true
---

## shelly party

Party mode - flash lights!

### Synopsis

Start a party mode that rapidly toggles lights.

This is a fun command that makes your lights flash for a set duration.
For RGB lights, it cycles through random colors.

Use Ctrl+C to stop early.

```
shelly party [device...] [flags]
```

### Examples

```
  # Party with all devices for 30 seconds
  shelly party --all

  # Party with specific devices for 1 minute
  shelly party light-1 light-2 -d 1m

  # Fast strobe effect (200ms interval)
  shelly party --all -i 200ms
```

### Options

```
  -a, --all                 Target all registered devices
  -d, --duration duration   Party duration (default 30s)
  -h, --help                help for party
  -i, --interval duration   Toggle interval (default 500ms)
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

