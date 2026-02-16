## shelly wifi status

Show WiFi status

### Synopsis

Show the current WiFi status for a device.

Displays connection status, IP address, SSID, signal strength (RSSI),
and number of clients connected to the access point (if enabled).

```
shelly wifi status <device> [flags]
```

### Examples

```
  # Show WiFi status
  shelly wifi status living-room

  # Output as JSON
  shelly wifi status living-room -o json
```

### Options

```
  -h, --help   help for status
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

