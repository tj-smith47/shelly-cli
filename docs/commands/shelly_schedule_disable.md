## shelly schedule disable

Disable a schedule

### Synopsis

Disable a schedule on a Gen2+ Shelly device.

```
shelly schedule disable <device> <id> [flags]
```

### Examples

```
  # Disable a schedule
  shelly schedule disable living-room 1
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
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly schedule](shelly_schedule.md)	 - Manage device schedules

