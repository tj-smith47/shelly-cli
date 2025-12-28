## shelly profile search

Search device profiles

### Synopsis

Search for device profiles by name, model, or feature.

Search by text query, or filter by capability or protocol.

```
shelly profile search <query> [flags]
```

### Examples

```
  # Search by name
  shelly profile search "plug"

  # Find devices with dimming
  shelly profile search --capability dimming

  # Find Z-Wave devices
  shelly profile search --protocol zwave

  # Combine filters
  shelly profile search --capability power_metering --protocol mqtt
```

### Options

```
      --capability string   Filter by capability (e.g., dimming, scripting, power_metering)
  -h, --help                help for search
  -o, --output string       Output format: table, json, yaml (default "table")
      --protocol string     Filter by protocol (e.g., mqtt, ble, zwave, matter)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly profile](shelly_profile.md)	 - Device profile information

