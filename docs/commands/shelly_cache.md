## shelly cache

Manage CLI cache

### Synopsis

Manage the Shelly CLI file cache.

The cache stores device data for faster access:
  - Device information and configuration
  - Firmware update status
  - Automation settings (schedules, webhooks, scripts)
  - Protocol configurations (MQTT, Modbus)

Cache data has type-specific TTLs and is shared between CLI and TUI.

### Examples

```
  # Show cache statistics
  shelly cache show

  # Clear all cache
  shelly cache clear

  # Clear cache for specific device
  shelly cache clear --device kitchen

  # Clear only expired entries
  shelly cache clear --expired
```

### Options

```
  -h, --help   help for cache
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly cache clear](shelly_cache_clear.md)	 - Clear the cache
* [shelly cache show](shelly_cache_show.md)	 - Show cache statistics

