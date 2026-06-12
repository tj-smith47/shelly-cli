## shelly provision inspect

Read a device's persisted config at its factory WiFi AP

### Synopsis

Hop onto a device's factory WiFi access point, read the configuration it has
actually persisted (identity plus WiFi station settings), and return to the home
network.

Use this when a device configures but never appears on the LAN: it shows whether
the station SSID, key, and static IP took, and whether the device has associated
yet — answering "did the onboard / restore --to-ap actually write what I expected?"
without the device having to join the network first.

```
shelly provision inspect <ap-ssid> [flags]
```

### Examples

```
  # Inspect a Shelly bulb sitting at its factory AP
  shelly provision inspect ShellyBulbDuo-D0DCFF

  # Use a specific host IP on the AP subnet
  shelly provision inspect ShellyBulbDuo-D0DCFF --ap-ip 192.168.33.150

  # JSON output
  shelly provision inspect ShellyBulbDuo-D0DCFF -o json
```

### Options

```
      --ap-ip string   Static host IP to use on the device's AP subnet (default 192.168.33.133)
  -h, --help           help for inspect
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

* [shelly provision](shelly_provision.md)	 - Discover and provision new Shelly devices

