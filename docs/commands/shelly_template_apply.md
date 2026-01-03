## shelly template apply

Apply a template to a device

### Synopsis

Apply a saved configuration template to a device.

The template configuration will be merged with the device's current
settings. Use --dry-run to preview changes without applying them.

Note: Only devices of the same model/generation are fully compatible.

```
shelly template apply <template> <device> [flags]
```

### Examples

```
  # Apply a template to a device
  shelly template apply my-config bedroom

  # Preview changes without applying
  shelly template apply my-config bedroom --dry-run

  # Apply without confirmation
  shelly template apply my-config bedroom --yes
```

### Options

```
      --dry-run   Preview changes without applying
  -h, --help      help for apply
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly template](shelly_template.md)	 - Manage device configuration templates

