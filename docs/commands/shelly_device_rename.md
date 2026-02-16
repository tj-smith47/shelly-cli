## shelly device rename

Rename a device in the registry

### Synopsis

Rename a Shelly device in the local registry.

This updates the device's friendly name and updates any group memberships
to use the new name. The device's address and configuration are preserved.

```
shelly device rename <old-name> <new-name> [flags]
```

### Examples

```
  # Rename a device
  shelly device rename kitchen kitchen-light

  # Short form
  shelly dev mv bedroom master-bedroom
```

### Options

```
  -h, --help   help for rename
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

