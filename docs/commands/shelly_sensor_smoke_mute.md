## shelly sensor smoke mute

Mute smoke alarm

### Synopsis

Mute an active smoke alarm.

The alarm will remain muted until the condition clears
and potentially re-triggers.

```
shelly sensor smoke mute <device> [flags]
```

### Examples

```
  # Mute smoke alarm
  shelly sensor smoke mute <device>

  # Mute specific sensor
  shelly sensor smoke mute <device> --id 1
```

### Options

```
  -h, --help     help for mute
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

* [shelly sensor smoke](shelly_sensor_smoke.md)	 - Manage smoke sensors

