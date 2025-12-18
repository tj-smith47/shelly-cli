## shelly thermostat override

Override target temperature

### Synopsis

Temporarily override the target temperature.

Override mode sets a different target temperature for a specified
duration. After the override expires, the thermostat returns to
its normal schedule.

This is useful for temporary temperature adjustments without
modifying the permanent schedule.

```
shelly thermostat override <device> [flags]
```

### Examples

```
  # Override to 25°C for 30 minutes
  shelly thermostat override gateway --target 25 --duration 30m

  # Override to 20°C for 2 hours
  shelly thermostat override gateway --target 20 --duration 2h

  # Override with device defaults
  shelly thermostat override gateway

  # Cancel active override
  shelly thermostat override gateway --cancel
```

### Options

```
      --cancel              Cancel active override
  -d, --duration duration   Override duration (e.g., 30m, 2h)
  -h, --help                help for override
      --id int              Thermostat component ID
  -t, --target float        Target temperature in Celsius
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

