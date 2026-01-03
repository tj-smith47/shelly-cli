## shelly mqtt status

Show MQTT status

### Synopsis

Show the MQTT connection status for a device.

Displays whether MQTT is enabled and if the device is connected to the broker.

```
shelly mqtt status <device> [flags]
```

### Examples

```
  # Show MQTT status
  shelly mqtt status living-room

  # Output as JSON
  shelly mqtt status living-room -o json
```

### Options

```
  -h, --help   help for status
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

* [shelly mqtt](shelly_mqtt.md)	 - Manage device MQTT configuration

