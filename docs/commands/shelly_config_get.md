## shelly config get

Get a CLI configuration value

### Synopsis

Get a configuration value from the Shelly CLI config file.

Use dot notation to access nested values (e.g., "defaults.timeout").
Without a key, shows all configuration values.

```
shelly config get [key] [flags]
```

### Examples

```
  # Get all settings
  shelly config get

  # Get default timeout
  shelly config get defaults.timeout

  # Output as JSON
  shelly config get -o json
```

### Options

```
  -h, --help   help for get
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

* [shelly config](shelly_config.md)	 - Manage CLI configuration

