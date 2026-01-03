## shelly sensoraddon remove

Remove a peripheral

### Synopsis

Remove a Sensor Add-on peripheral from a device.

The component key format is "type:id", for example "temperature:100" or "input:101".

Note: Changes require a device reboot to take effect.

```
shelly sensoraddon remove <device> <component> [flags]
```

### Examples

```
  # Remove a peripheral
  shelly sensoraddon remove kitchen temperature:100

  # Skip confirmation
  shelly sensoraddon remove kitchen input:101 --yes
```

### Options

```
      --confirm   Double-confirm destructive operation
  -h, --help      help for remove
  -y, --yes       Skip confirmation prompt
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

* [shelly sensoraddon](shelly_sensoraddon.md)	 - Manage Sensor Add-on peripherals

