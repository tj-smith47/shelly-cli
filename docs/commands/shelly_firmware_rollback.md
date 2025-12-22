## shelly firmware rollback

Rollback to previous firmware

### Synopsis

Rollback device firmware to the previous version.

This is only available when the device supports rollback (typically after
a recent firmware update or when in safe mode).

```
shelly firmware rollback <device> [flags]
```

### Examples

```
  # Rollback firmware
  shelly firmware rollback living-room

  # Rollback without confirmation
  shelly firmware rollback living-room --yes
```

### Options

```
  -h, --help   help for rollback
  -y, --yes    Skip confirmation prompt
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

* [shelly firmware](shelly_firmware.md)	 - Manage device firmware

