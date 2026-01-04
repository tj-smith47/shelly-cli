## shelly provision

Provision device settings

### Synopsis

Provision device settings including WiFi, network, and bulk configuration.

The provision commands provide an interactive workflow for setting up devices,
including WiFi network scanning and selection, bulk provisioning from config files,
and BLE-based provisioning for devices in AP mode.

### Examples

```
  # Interactive WiFi provisioning
  shelly provision wifi living-room

  # Bulk provision from config file
  shelly provision bulk devices.yaml

  # BLE-based provisioning for new device
  shelly provision ble 192.168.33.1
```

### Options

```
  -h, --help   help for provision
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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
* [shelly provision ble](shelly_provision_ble.md)	 - Provision a device via Bluetooth Low Energy
* [shelly provision bulk](shelly_provision_bulk.md)	 - Bulk provision from config file
* [shelly provision wifi](shelly_provision_wifi.md)	 - Interactive WiFi provisioning

