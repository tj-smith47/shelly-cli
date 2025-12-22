## shelly sensor devicepower status

Get devicepower sensor status

### Synopsis

Get the current status of a devicepower sensor.

```
shelly sensor devicepower status <device> [flags]
```

### Examples

```
  # Get devicepower status
  shelly sensor devicepower status <device>

  # Get specific sensor
  shelly sensor devicepower status <device> --id 1

  # Output as JSON
  shelly sensor devicepower status <device> -o json
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
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly sensor devicepower](shelly_sensor_devicepower.md)	 - Manage device power sensors

