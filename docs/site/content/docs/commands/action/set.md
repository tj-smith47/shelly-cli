---
title: "shelly action set"
description: "shelly action set"
---

## shelly action set

Set an action URL for a Gen1 device

### Synopsis

Set an action URL for a Gen1 Shelly device.

Gen1 devices support various action event types:
  Output events:    out_on_url, out_off_url
  Button events:    btn1_on_url, btn1_off_url, btn2_on_url, btn2_off_url
  Input events:     input_on_url, input_off_url
  Push events:      longpush_url, shortpush_url, double_shortpush_url, triple_shortpush_url
  Roller events:    roller_open_url, roller_close_url, roller_stop_url
  Sensor events:    motion_url, no_motion_url, flood_detected_url, etc.
  System events:    overpower_url, overvoltage_url, overtemperature_url

Gen2+ devices use webhooks instead. See 'shelly webhook create'.

```
shelly action set <device> <event> <url> [flags]
```

### Examples

```
  # Set output on action
  shelly action set living-room out_on_url "http://homeserver/api/light-on"

  # Set button long press action
  shelly action set switch longpush_url "http://homeserver/api/dim-lights"

  # Set action at a specific index (for multi-channel devices)
  shelly action set relay out_on_url "http://server/trigger" --index 1

  # Set action but leave it disabled
  shelly action set switch out_on_url "http://server/test" --disabled
```

### Options

```
      --disabled    Disable the action (same as --enabled=false)
      --enabled     Enable the action (default true)
  -h, --help        help for set
      --index int   Action index (for multi-channel devices)
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

* [shelly action](shelly_action.md)	 - Manage Gen1 device action URLs

