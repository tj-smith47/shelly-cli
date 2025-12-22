## shelly fleet health

Check device health

### Synopsis

Check the health status of devices in your fleet.

Reports devices that haven't been seen recently or have frequent
online/offline transitions indicating connectivity issues.

Requires an active fleet connection. Run 'shelly fleet connect' first.

```
shelly fleet health [flags]
```

### Examples

```
  # Check device health
  shelly fleet health

  # Custom threshold for "unhealthy"
  shelly fleet health --threshold 30m

  # JSON output
  shelly fleet health -o json
```

### Options

```
  -h, --help                 help for health
      --threshold duration   Time threshold for unhealthy status (default 10m0s)
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

* [shelly fleet](shelly_fleet.md)	 - Cloud-based fleet management

