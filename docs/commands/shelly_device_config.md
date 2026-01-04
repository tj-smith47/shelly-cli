## shelly device config

Manage device configuration

### Synopsis

Manage device configuration settings.

Get, set, export, and import device configurations. Configuration includes
component settings, system parameters, and feature configurations.

### Examples

```
  # Get full device configuration
  shelly device config get living-room

  # Get specific component configuration
  shelly device config get living-room switch:0

  # Set configuration values
  shelly device config set living-room switch:0 name="Main Light"

  # Export configuration to file
  shelly device config export living-room config.json

  # Import configuration from file
  shelly device config import living-room config.json --dry-run

  # Compare configuration with a file
  shelly device config diff living-room config.json

  # Reset configuration to defaults
  shelly device config reset living-room switch:0
```

### Options

```
  -h, --help   help for config
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
* [shelly device config diff](shelly_device_config_diff.md)	 - Compare device configurations
* [shelly device config export](shelly_device_config_export.md)	 - Export device configuration to a file
* [shelly device config get](shelly_device_config_get.md)	 - Get device configuration
* [shelly device config import](shelly_device_config_import.md)	 - Import configuration from a file
* [shelly device config reset](shelly_device_config_reset.md)	 - Reset configuration to defaults
* [shelly device config set](shelly_device_config_set.md)	 - Set device configuration

