## shelly virtual list

List virtual components on a device

### Synopsis

List all virtual components on a Shelly Gen2+ device.

Virtual components include boolean, number, text, enum, button, and group types.
Component IDs in the range 200-299 are reserved for virtual components.

```
shelly virtual list <device> [flags]
```

### Examples

```
  # List all virtual components
  shelly virtual list kitchen

  # JSON output
  shelly virtual list kitchen -o json
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format: table, json, yaml (default "table")
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
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly virtual](shelly_virtual.md)	 - Manage virtual components

