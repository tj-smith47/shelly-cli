## shelly device config diff

Compare device configurations

### Synopsis

Compare configurations between two devices or a device and a backup file.

This command shows differences in configuration between:
  - Two live devices: shelly device config diff device1 device2
  - Device and backup: shelly device config diff device backup.json
  - Two backup files: shelly device config diff backup1.json backup2.json

Differences are shown with:
  + Added values (only in target)
  - Removed values (only in source)
  ~ Changed values (different between source and target)

```
shelly device config diff <source> <target> [flags]
```

### Examples

```
  # Compare two devices
  shelly device config diff kitchen-light bedroom-light

  # Compare device with backup file
  shelly device config diff living-room config-backup.json

  # Compare two backup files
  shelly device config diff backup1.json backup2.json

  # JSON output
  shelly device config diff device1 device2 -o json
```

### Options

```
  -h, --help   help for diff
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

* [shelly device config](shelly_device_config.md)	 - Manage device configuration

