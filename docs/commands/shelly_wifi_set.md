## shelly wifi set

Configure WiFi connection

### Synopsis

Configure the WiFi station (client) connection for a device.

Set the SSID and password to connect to a WiFi network. Optionally configure
static IP settings instead of using DHCP.

```
shelly wifi set <device> [flags]
```

### Examples

```
  # Connect to a WiFi network
  shelly wifi set living-room --ssid "MyNetwork" --password "secret"

  # Configure static IP
  shelly wifi set living-room --ssid "MyNetwork" --password "secret" \
    --static-ip "192.168.1.50" --gateway "192.168.1.1" --netmask "255.255.255.0"

  # Disable WiFi station mode
  shelly wifi set living-room --disable
```

### Options

```
      --disable            Disable WiFi station mode
      --dns string         DNS server address (for static IP)
      --enable             Enable WiFi station mode
      --gateway string     Gateway address (for static IP)
  -h, --help               help for set
      --netmask string     Network mask (for static IP)
      --password string    WiFi password
      --ssid string        WiFi network name
      --static-ip string   Static IP address (uses DHCP if not set)
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

