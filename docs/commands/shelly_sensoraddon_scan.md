## shelly sensoraddon scan

Scan for OneWire devices

### Synopsis

Scan for OneWire devices on the Sensor Add-on bus.

Currently only DS18B20 OneWire temperature sensors are supported.

Note: This will fail if a DHT22 sensor is in use, as DHT22 shares
the same GPIOs as the OneWire bus.

```
shelly sensoraddon scan <device> [flags]
```

### Examples

```
  # Scan for OneWire devices
  shelly sensoraddon scan kitchen

  # JSON output
  shelly sensoraddon scan kitchen -o json
```

### Options

```
  -h, --help            help for scan
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

