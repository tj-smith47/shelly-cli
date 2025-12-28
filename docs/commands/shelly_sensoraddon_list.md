## shelly sensoraddon list

List configured peripherals

### Synopsis

List all configured Sensor Add-on peripherals on a device.

```
shelly sensoraddon list <device> [flags]
```

### Examples

```
  # List peripherals
  shelly sensoraddon list kitchen

  # JSON output
  shelly sensoraddon list kitchen -o json
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format: table, json, yaml (default "table")
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

* [shelly sensoraddon](shelly_sensoraddon.md)	 - Manage Sensor Add-on peripherals

