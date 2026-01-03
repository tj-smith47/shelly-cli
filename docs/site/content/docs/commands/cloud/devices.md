---
title: "shelly cloud devices"
description: "shelly cloud devices"
---

## shelly cloud devices

List cloud-registered devices

### Synopsis

List all devices registered with your Shelly Cloud account.

Shows device ID, name, model, firmware version, and online status.

```
shelly cloud devices [flags]
```

### Examples

```
  # List all cloud devices
  shelly cloud devices

  # Output as JSON
  shelly cloud devices -o json
```

### Options

```
  -h, --help   help for devices
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

* [shelly cloud](shelly_cloud.md)	 - Manage cloud connection and Shelly Cloud API

