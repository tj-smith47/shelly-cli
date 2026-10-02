## shelly wifi set

Configure WiFi connection

### Synopsis

Configure the WiFi station (client) connection for a device.

Set the SSID and password to connect to a WiFi network. Optionally configure
static IP settings instead of using DHCP.

Without --password, a device that stays on the same network keeps the
password it has. A different network takes the password this host has stored
for it; with none, the command is refused. Use --open for a network that has
no password.

```
shelly wifi set <device> [flags]
```

### Examples

```
  # Connect to a WiFi network
  shelly wifi set living-room --ssid "MyNetwork" --password "secret"

  # Join a network this host knows, using its stored password
  shelly wifi set living-room --ssid "MyNetwork"

  # Join a network that has no password
  shelly wifi set living-room --ssid "GuestNet" --open

  # Configure static IP
  shelly wifi set living-room --ssid "MyNetwork" --password "secret" \
    --static-ip "192.168.1.50" --gateway "192.168.1.1" --netmask "255.255.255.0"

  # Change only the address; the gateway and netmask stay the device's
  shelly wifi set living-room --ssid "MyNetwork" --static-ip "192.168.1.51"

  # Disable WiFi station mode
  shelly wifi set living-room --disable
```

### Options

```
      --disable            Disable WiFi station mode
      --dns string         Static IPv4 nameserver (with --static-ip; default: the device's current one)
      --enable             Enable WiFi station mode
      --gateway string     Static IPv4 default gateway (with --static-ip; default: the device's current one)
  -h, --help               help for set
      --netmask string     Static IPv4 subnet mask (with --static-ip; default: the device's current one)
      --open               Join a network that has no password
      --password string    WiFi password for the network (when omitted and one is needed, the passphrase stored on this host for it is used)
      --ssid string        WiFi network name
      --static-ip string   Static IPv4 address (DHCP when not set; --gateway, --netmask and --dns default to the device's current ones)
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
      --raw                     Print the exact device response(s) as a JSON array and suppress normal output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly wifi](shelly_wifi.md)	 - Manage device WiFi configuration

