## shelly virtual delete

Delete a virtual component

### Synopsis

Delete a virtual component from a Shelly Gen2+ device.

The key format is "type:id", for example "boolean:200" or "number:201".

This action cannot be undone.

```
shelly virtual delete <device> <key> [flags]
```

### Examples

```
  # Delete a virtual component
  shelly virtual delete kitchen boolean:200

  # Skip confirmation
  shelly virtual delete kitchen boolean:200 --yes
```

### Options

```
      --confirm   Double-confirm destructive operation
  -h, --help      help for delete
  -y, --yes       Skip confirmation prompt
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

* [shelly virtual](shelly_virtual.md)	 - Manage virtual components

