## shelly sensor flood

Manage flood sensors

### Synopsis

Manage flood (water leak) sensors on Shelly devices.

Flood sensors detect water leaks and can trigger alarms with
different modes: disabled, normal, intense, or rain detection.

### Examples

```
  # List flood sensors
  shelly sensor flood list bathroom

  # Check flood status
  shelly sensor flood status bathroom

  # Test flood alarm
  shelly sensor flood test bathroom
```

### Options

```
  -h, --help   help for flood
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly sensor](shelly_sensor.md)	 - Manage device sensors
* [shelly sensor flood list](shelly_sensor_flood_list.md)	 - List flood sensors
* [shelly sensor flood status](shelly_sensor_flood_status.md)	 - Get flood sensor status
* [shelly sensor flood test](shelly_sensor_flood_test.md)	 - Test flood sensor

