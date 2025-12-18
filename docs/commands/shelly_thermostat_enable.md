## shelly thermostat enable

Enable thermostat

### Synopsis

Enable a thermostat component.

Optionally set the operating mode when enabling:
- heat: Heating mode (opens valve when below target)
- cool: Cooling mode (opens valve when above target)
- auto: Automatic mode (device decides based on conditions)

```
shelly thermostat enable <device> [flags]
```

### Examples

```
  # Enable thermostat with current settings
  shelly thermostat enable gateway

  # Enable in heat mode
  shelly thermostat enable gateway --mode heat

  # Enable specific thermostat in auto mode
  shelly thermostat enable gateway --id 1 --mode auto
```

### Options

```
  -h, --help          help for enable
      --id int        Thermostat component ID
      --mode string   Operating mode (heat, cool, auto)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly thermostat](shelly_thermostat.md)	 - Manage thermostats

