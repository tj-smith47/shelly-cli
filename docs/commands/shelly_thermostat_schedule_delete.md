## shelly thermostat schedule delete

Delete a thermostat schedule

### Synopsis

Delete a schedule from the device.

Use --id to specify the schedule ID to delete.
Use --all to delete all schedules (use with caution).

```
shelly thermostat schedule delete <device> [flags]
```

### Examples

```
  # Delete schedule by ID
  shelly thermostat schedule delete gateway --id 1

  # Delete all schedules
  shelly thermostat schedule delete gateway --all
```

### Options

```
      --all      Delete all schedules
  -h, --help     help for delete
      --id int   Schedule ID to delete
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

* [shelly thermostat schedule](shelly_thermostat_schedule.md)	 - Manage thermostat schedules

