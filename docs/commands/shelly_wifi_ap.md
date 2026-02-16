## shelly wifi ap

Configure WiFi access point

### Synopsis

Configure the WiFi access point (AP) mode for a device.

When enabled, the device creates its own WiFi network that other devices
can connect to. Use --clients to list connected clients.

```
shelly wifi ap <device> [flags]
```

### Examples

```
  # Enable access point with custom SSID
  shelly wifi ap living-room --enable --ssid "ShellyAP" --password "secret"

  # Disable access point
  shelly wifi ap living-room --disable

  # List connected clients
  shelly wifi ap living-room --clients
```

### Options

```
      --clients           List connected clients
      --disable           Disable access point
      --enable            Enable access point
  -h, --help              help for ap
      --password string   Access point password
      --ssid string       Access point SSID
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

* [shelly wifi](shelly_wifi.md)	 - Manage device WiFi configuration

