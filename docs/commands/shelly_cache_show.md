## shelly cache show

Show cache statistics

### Synopsis

Display detailed information about the file cache.

Shows cache statistics including:
  - Total entries and size
  - Entries by data type
  - Device count
  - Expired entry count
  - Oldest and newest entries

```
shelly cache show [flags]
```

### Examples

```
  # Show cache statistics
  shelly cache show

  # Show cache stats in JSON format
  shelly cache show -o json
```

### Options

```
  -h, --help   help for show
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

* [shelly cache](shelly_cache.md)	 - Manage CLI cache

