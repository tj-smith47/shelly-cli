## shelly wifi scan

Scan for available WiFi networks

### Synopsis

Scan for available WiFi networks using a device.

The device will scan for nearby WiFi networks and report their SSID,
signal strength (RSSI), channel, and authentication type.

Note: Scanning may take several seconds to complete.

```
shelly wifi scan <device> [flags]
```

### Examples

```
  # Scan for networks
  shelly wifi scan living-room

  # Output as JSON
  shelly wifi scan living-room -o json
```

### Options

```
  -h, --help   help for scan
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

* [shelly wifi](shelly_wifi.md)	 - Manage device WiFi configuration

