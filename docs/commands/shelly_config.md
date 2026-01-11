## shelly config

Manage CLI configuration

### Synopsis

Manage Shelly CLI configuration settings.

Get, set, and edit CLI preferences like default timeout, output format,
and theme settings.

For device configuration, use: shelly device config <subcommand>

### Examples

```
  # View all CLI settings
  shelly config get

  # Get specific setting
  shelly config get defaults.timeout

  # Set a value
  shelly config set defaults.output=json

  # Delete a setting
  shelly config delete defaults.timeout

  # Open config in editor
  shelly config edit

  # Reset to defaults
  shelly config reset
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly config delete](shelly_config_delete.md)	 - Delete CLI configuration values
* [shelly config edit](shelly_config_edit.md)	 - Open CLI config in editor
* [shelly config get](shelly_config_get.md)	 - Get a CLI configuration value
* [shelly config path](shelly_config_path.md)	 - Show configuration file path
* [shelly config reset](shelly_config_reset.md)	 - Reset CLI configuration to defaults
* [shelly config set](shelly_config_set.md)	 - Set CLI configuration values
* [shelly config show](shelly_config_show.md)	 - Display current CLI configuration

