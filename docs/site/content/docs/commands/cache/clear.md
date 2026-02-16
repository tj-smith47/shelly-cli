---
title: "shelly cache clear"
description: "shelly cache clear"
---

## shelly cache clear

Clear the cache

### Synopsis

Clear cached device data.

Use flags to specify what to clear:
  --all      Clear all cached data (requires confirmation)
  --device   Clear cache for a specific device
  --type     Clear cache for a specific data type (requires --device)
  --expired  Clear only expired entries

```
shelly cache clear [flags]
```

### Examples

```
  # Clear all cache
  shelly cache clear --all

  # Clear cache for a specific device
  shelly cache clear --device kitchen

  # Clear cache for specific device and type
  shelly cache clear --device kitchen --type firmware

  # Clear only expired entries
  shelly cache clear --expired
```

### Options

```
  -a, --all             Clear all cached data
  -d, --device string   Clear cache for specific device
  -e, --expired         Clear only expired entries
  -h, --help            help for clear
  -t, --type string     Clear cache for specific data type (requires --device)
  -y, --yes             Skip confirmation prompt
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

* [shelly cache](shelly_cache.md)	 - Manage CLI cache

