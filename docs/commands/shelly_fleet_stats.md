## shelly fleet stats

View fleet statistics

### Synopsis

View aggregate statistics for your device fleet.

Requires an active fleet connection. Run 'shelly fleet connect' first.

```
shelly fleet stats [flags]
```

### Examples

```
  # View fleet statistics
  shelly fleet stats

  # JSON output
  shelly fleet stats -o json
```

### Options

```
  -h, --help   help for stats
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly fleet](shelly_fleet.md)	 - Cloud-based fleet management

