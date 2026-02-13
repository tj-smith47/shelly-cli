---
title: "shelly provision wifi"
description: "shelly provision wifi"
---

## shelly provision wifi

Interactive WiFi provisioning

### Synopsis

Provision WiFi settings interactively for a device.

By default, this command scans for available networks and prompts you to select one.
You can also provide SSID and password directly via flags.

```
shelly provision wifi <device> [flags]
```

### Examples

```
  # Interactive provisioning with network scan
  shelly provision wifi living-room

  # Direct provisioning with credentials
  shelly provision wifi living-room --ssid "MyNetwork" --password "secret"

  # Skip scan and prompt for SSID
  shelly provision wifi living-room --no-scan
```

### Options

```
  -h, --help              help for wifi
      --no-scan           Skip network scan, prompt for SSID
      --password string   WiFi password
      --ssid string       WiFi network name (skip selection)
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

* [shelly provision](shelly_provision.md)	 - Discover and provision new Shelly devices

