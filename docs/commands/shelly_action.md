## shelly action

Manage Gen1 device action URLs

### Synopsis

Manage action URLs for Gen1 Shelly devices.

Gen1 devices support action URLs that trigger HTTP requests when specific events
occur (e.g., button press, output state change). This is different from Gen2
webhooks which have a more structured API.

Common Gen1 actions include:
  - btn_on_url, btn_off_url: Button toggle actions
  - out_on_url, out_off_url: Output state change actions
  - roller_open_url, roller_close_url: Roller actions
  - longpush_url, shortpush_url: Button press duration actions

Note: For Gen2 devices, use 'shelly webhook' instead.

### Examples

```
  # List all action URLs for a device
  shelly action list living-room

  # Set an action URL
  shelly action set living-room out_on_url "http://homeserver/api/light-on"

  # Clear an action URL
  shelly action clear living-room out_on_url

  # Test an action
  shelly action test living-room out_on_url
```

### Options

```
  -h, --help   help for action
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
* [shelly action clear](shelly_action_clear.md)	 - Clear an action URL for a Gen1 device
* [shelly action list](shelly_action_list.md)	 - List action URLs for a Gen1 device
* [shelly action set](shelly_action_set.md)	 - Set an action URL for a Gen1 device
* [shelly action test](shelly_action_test.md)	 - Test/trigger an action on a Gen1 device

