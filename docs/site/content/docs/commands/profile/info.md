---
title: "shelly profile info"
description: "shelly profile info"
---

## shelly profile info

Show device profile details

### Synopsis

Show detailed information about a specific device model.

Displays hardware capabilities, supported protocols, components,
and resource limits for the specified device model.

```
shelly profile info <model> [flags]
```

### Examples

```
  # Show info for Shelly Plus 1PM
  shelly profile info SNSW-001P16EU

  # JSON output
  shelly profile info SNSW-001P16EU -o json
```

### Options

```
  -h, --help            help for info
  -o, --output string   Output format: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly profile](shelly_profile.md)	 - Device profile information

