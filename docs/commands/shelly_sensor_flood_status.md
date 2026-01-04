## shelly sensor flood status

Get flood sensor status

### Synopsis

Get the current status of a flood sensor.

```
shelly sensor flood status <device> [flags]
```

### Examples

```
  # Get flood status
  shelly sensor flood status <device>

  # Get specific sensor
  shelly sensor flood status <device> --id 1

  # Output as JSON
  shelly sensor flood status <device> -o json
```

### Options

```
  -h, --help     help for status
      --id int   Sensor ID (default 0)
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

* [shelly sensor flood](shelly_sensor_flood.md)	 - Manage flood sensors

