## shelly device info

Show device information

### Synopsis

Show detailed information about a device.

The device can be specified by its registered name or IP address.

```
shelly device info <device> [flags]
```

### Examples

```
  # Show info for a registered device
  shelly device info living-room

  # Show info by IP address
  shelly device info 192.168.1.100

  # Output as JSON
  shelly device info living-room -o json

  # Output as YAML
  shelly device info living-room -o yaml

  # Short form
  shelly dev info office-switch
```

### Options

```
  -h, --help   help for info
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

