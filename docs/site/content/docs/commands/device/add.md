---
title: "shelly device add"
description: "shelly device add"
---

## shelly device add

Add a device to the registry

### Synopsis

Add a Shelly device to the local registry.

The device will be verified and its generation/model auto-detected
unless --no-verify is specified.

The name is used as a friendly identifier for the device in other commands.
Names with spaces will be normalized to dashes (e.g., "Master Bathroom"
becomes "master-bathroom" as the key).

```
shelly device add <name> <address> [flags]
```

### Examples

```
  # Add a device (auto-detects generation and model)
  shelly device add kitchen 192.168.1.100

  # Add with authentication
  shelly device add secure-device 192.168.1.101 --auth admin:secret

  # Add without verification (offline)
  shelly device add offline-device 192.168.1.102 --no-verify --generation 2

  # Short form
  shelly dev add bedroom 192.168.1.103
```

### Options

```
      --auth string      Authentication credentials (user:pass)
  -g, --generation int   Device generation (auto-detected if omitted)
  -h, --help             help for add
      --no-verify        Skip connectivity check and auto-detection
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

