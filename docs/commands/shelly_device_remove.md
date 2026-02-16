## shelly device remove

Remove a device from the registry

### Synopsis

Remove a Shelly device from the local registry.

The device will also be removed from any groups it belongs to.
This does not affect the physical device itself.

```
shelly device remove <name> [flags]
```

### Examples

```
  # Remove a device
  shelly device remove kitchen

  # Remove with force (skip confirmation)
  shelly device remove kitchen --force

  # Short form
  shelly dev rm bedroom
```

### Options

```
  -f, --force   Skip confirmation prompt
  -h, --help    help for remove
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

