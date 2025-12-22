## shelly fleet status

View fleet device status

### Synopsis

View the status of all devices in your fleet.

Shows online/offline status and last seen time for each device
connected through Shelly Cloud.

Requires an active fleet connection. Run 'shelly fleet connect' first.

```
shelly fleet status [flags]
```

### Examples

```
  # View all device status
  shelly fleet status

  # Show only online devices
  shelly fleet status --online

  # Show only offline devices
  shelly fleet status --offline

  # JSON output
  shelly fleet status -o json
```

### Options

```
  -h, --help      help for status
      --offline   Show only offline devices
      --online    Show only online devices
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

* [shelly fleet](shelly_fleet.md)	 - Cloud-based fleet management

