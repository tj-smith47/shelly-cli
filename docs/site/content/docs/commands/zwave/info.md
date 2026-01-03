---
title: "shelly zwave info"
description: "shelly zwave info"
---

## shelly zwave info

Show Z-Wave device information

### Synopsis

Show Z-Wave device information for a Shelly Wave model.

Displays device capabilities, supported protocols, and network topology options.

```
shelly zwave info <model> [flags]
```

### Examples

```
  # Show info for Wave 1PM
  shelly zwave info SNSW-001P16ZW

  # JSON output
  shelly zwave info SNSW-001P16ZW -o json
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

* [shelly zwave](shelly_zwave.md)	 - Z-Wave device utilities

