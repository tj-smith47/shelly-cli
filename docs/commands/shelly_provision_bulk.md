## shelly provision bulk

Bulk provision from config file

### Synopsis

Provision multiple devices from a YAML configuration file.

The config file specifies WiFi credentials and device-specific settings.
Provisioning is performed in parallel for efficiency.

Config file format:
  wifi:
    ssid: "MyNetwork"
    password: "secret"
  devices:
    - name: living-room
      address: 192.168.1.100  # optional, uses registered device if omitted
      device_name: "Living Room Light"  # optional device name to set
    - name: bedroom
      wifi:  # optional per-device WiFi override
        ssid: "OtherNetwork"
        password: "other-secret"

```
shelly provision bulk <config-file> [flags]
```

### Examples

```
  # Provision devices from config file
  shelly provision bulk devices.yaml

  # Dry run to validate config
  shelly provision bulk devices.yaml --dry-run

  # Limit parallel operations
  shelly provision bulk devices.yaml --parallel 2
```

### Options

```
      --dry-run        Validate config without provisioning
  -h, --help           help for bulk
      --parallel int   Maximum parallel provisioning operations (default 5)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly provision](shelly_provision.md)	 - Provision device settings

