## shelly action clear

Clear an action URL for a Gen1 device

### Synopsis

Clear (remove) an action URL for a Gen1 Shelly device.

This removes the configured URL for the specified action, disabling the
HTTP callback for that event.

Gen1 devices support various action event types:
  Output events:    out_on_url, out_off_url
  Button events:    btn1_on_url, btn1_off_url, btn2_on_url, btn2_off_url
  Input events:     input_on_url, input_off_url
  Push events:      longpush_url, shortpush_url, double_shortpush_url, triple_shortpush_url
  Roller events:    roller_open_url, roller_close_url, roller_stop_url
  Sensor events:    motion_url, no_motion_url, flood_detected_url, etc.

Gen2+ devices use webhooks instead. See 'shelly webhook delete'.

```
shelly action clear <device> <event> [flags]
```

### Examples

```
  # Clear output on action
  shelly action clear living-room out_on_url

  # Clear button long press action
  shelly action clear switch longpush_url

  # Clear action at a specific index
  shelly action clear relay out_on_url --index 1
```

### Options

```
  -h, --help        help for clear
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

