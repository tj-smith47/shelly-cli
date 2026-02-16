## shelly sensor smoke

Manage smoke sensors

### Synopsis

Manage smoke detection sensors on Shelly devices.

Smoke sensors provide alarm state detection and the ability
to mute active alarms.

### Examples

```
  # List smoke sensors
  shelly sensor smoke list kitchen

  # Check smoke status
  shelly sensor smoke status kitchen

  # Test smoke alarm
  shelly sensor smoke test kitchen

  # Mute active alarm
  shelly sensor smoke mute kitchen
```

### Options

```
  -h, --help   help for smoke
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
* [shelly sensor smoke list](shelly_sensor_smoke_list.md)	 - List smoke sensors
* [shelly sensor smoke mute](shelly_sensor_smoke_mute.md)	 - Mute smoke alarm
* [shelly sensor smoke status](shelly_sensor_smoke_status.md)	 - Get smoke sensor status
* [shelly sensor smoke test](shelly_sensor_smoke_test.md)	 - Test smoke sensor

