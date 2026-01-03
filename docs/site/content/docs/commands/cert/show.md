---
title: "shelly cert show"
description: "shelly cert show"
---

## shelly cert show

Show device TLS configuration

### Synopsis

Display TLS certificate configuration for a Gen2+ device.

```
shelly cert show <device> [flags]
```

### Examples

```
  # Show TLS config
  shelly cert show kitchen
```

### Options

```
  -h, --help   help for show
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

* [shelly cert](shelly_cert.md)	 - Manage device TLS certificates

