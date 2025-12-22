## shelly sensor flood test

Test flood sensor

### Synopsis

Test the flood sensor on a Shelly device.

Note: The Flood component may not have a dedicated test method.
This command provides instructions for manual testing.

```
shelly sensor flood test <device> [flags]
```

### Examples

```
  # Test flood sensor
  shelly sensor flood test <device>
```

### Options

```
  -h, --help     help for test
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

* [shelly sensor flood](shelly_sensor_flood.md)	 - Manage flood sensors

