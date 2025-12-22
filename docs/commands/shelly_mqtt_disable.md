## shelly mqtt disable

Disable MQTT

### Synopsis

Disable MQTT on a device.

This disconnects the device from the MQTT broker and disables MQTT functionality.

```
shelly mqtt disable <device> [flags]
```

### Examples

```
  # Disable MQTT
  shelly mqtt disable living-room
```

### Options

```
  -h, --help   help for disable
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

* [shelly mqtt](shelly_mqtt.md)	 - Manage device MQTT configuration

