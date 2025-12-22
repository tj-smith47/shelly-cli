## shelly sensor humidity

Manage humidity sensors

### Synopsis

Manage humidity sensors on Shelly devices.

Humidity sensors (DHT22, HTU21D, or similar) provide relative humidity readings.

### Examples

```
  # List humidity sensors
  shelly sensor humidity list living-room

  # Get humidity reading
  shelly sensor humidity status living-room
```

### Options

```
  -h, --help   help for humidity
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly sensor](shelly_sensor.md)	 - Manage device sensors
* [shelly sensor humidity list](shelly_sensor_humidity_list.md)	 - List humidity sensors
* [shelly sensor humidity status](shelly_sensor_humidity_status.md)	 - Get humidity sensor status

