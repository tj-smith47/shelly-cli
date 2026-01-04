## shelly migrate diff

Show differences between device and backup

### Synopsis

Show the differences between a device's current state and a backup file.

This helps you understand what would change if you restored the backup.

```
shelly migrate diff <device> <backup-file> [flags]
```

### Examples

```
  # Show differences
  shelly migrate diff living-room backup.json
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

* [shelly migrate](shelly_migrate.md)	 - Migrate configuration between devices

