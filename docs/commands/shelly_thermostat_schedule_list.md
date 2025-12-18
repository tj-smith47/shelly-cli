## shelly thermostat schedule list

List thermostat schedules

### Synopsis

List all schedules that control the thermostat.

By default, only shows schedules that target the thermostat component.
Use --all to show all device schedules.

```
shelly thermostat schedule list <device> [flags]
```

### Examples

```
  # List thermostat schedules
  shelly thermostat schedule list gateway

  # List schedules for specific thermostat ID
  shelly thermostat schedule list gateway --thermostat-id 1

  # List all device schedules
  shelly thermostat schedule list gateway --all

  # Output as JSON
  shelly thermostat schedule list gateway --json
```

### Options

```
      --all                 Show all device schedules
  -h, --help                help for list
      --json                Output as JSON
      --thermostat-id int   Filter by thermostat component ID
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

* [shelly thermostat schedule](shelly_thermostat_schedule.md)	 - Manage thermostat schedules

