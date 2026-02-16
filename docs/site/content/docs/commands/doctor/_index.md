---
title: "shelly doctor"
description: "shelly doctor"
weight: 220
sidebar:
  collapsed: true
---

## shelly doctor

Check system health and diagnose issues

### Synopsis

Run comprehensive diagnostics on the Shelly CLI setup.

Checks include:
  - CLI version and update availability
  - Configuration file validity
  - Registered devices and their reachability
  - Network connectivity
  - Cloud authentication status
  - Firmware update availability

Use --full for all checks including device connectivity tests.

```
shelly doctor [flags]
```

### Examples

```
  # Run basic diagnostics
  shelly doctor

  # Check network connectivity
  shelly doctor --network

  # Test all registered devices
  shelly doctor --devices

  # Full diagnostic suite
  shelly doctor --full
```

### Options

```
      --devices   Test device reachability
      --full      Run all diagnostic checks
  -h, --help      help for doctor
      --network   Check network connectivity
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

