---
title: "shelly action list"
description: "shelly action list"
---

## shelly action list

List action URLs for a Gen1 device

### Synopsis

List all configured action URLs for a Gen1 Shelly device.

Gen1 devices support various action types that trigger HTTP callbacks:
  - out_on_url, out_off_url: Output state change actions
  - btn_on_url, btn_off_url: Button toggle actions
  - longpush_url, shortpush_url: Button press duration actions
  - roller_open_url, roller_close_url, roller_stop_url: Roller actions

Gen2+ devices use webhooks instead. See 'shelly webhook list'.

```
shelly action list <device> [flags]
```

### Examples

```
  # List actions for a device
  shelly action list living-room

  # JSON output
  shelly action list living-room -o json
```

### Options

```
  -h, --help   help for list
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly action](shelly_action.md)	 - Manage Gen1 device action URLs

