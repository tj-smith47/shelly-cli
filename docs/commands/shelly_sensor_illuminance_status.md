## shelly sensor illuminance status

Get illuminance sensor status

### Synopsis

Get the current status of a illuminance sensor.

```
shelly sensor illuminance status <device> [flags]
```

### Examples

```
  # Get illuminance status
  shelly sensor illuminance status <device>

  # Get specific sensor
  shelly sensor illuminance status <device> --id 1

  # Output as JSON
  shelly sensor illuminance status <device> -o json
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
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly sensor illuminance](shelly_sensor_illuminance.md)	 - Manage illuminance sensors

