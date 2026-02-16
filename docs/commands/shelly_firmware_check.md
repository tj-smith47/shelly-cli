## shelly firmware check

Check for firmware updates

### Synopsis

Check if firmware updates are available for a device.

Use --all to check all registered devices.

```
shelly firmware check [device] [flags]
```

### Examples

```
  # Check a specific device
  shelly firmware check living-room

  # Check all registered devices
  shelly firmware check --all
```

### Options

```
      --all    Check all registered devices
  -h, --help   help for check
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly firmware](shelly_firmware.md)	 - Manage device firmware

