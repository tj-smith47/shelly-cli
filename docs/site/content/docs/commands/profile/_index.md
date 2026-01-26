---
title: "shelly profile"
description: "shelly profile"
weight: 560
sidebar:
  collapsed: true
---

## shelly profile

Device profile information

### Synopsis

Query device profiles and capabilities.

Device profiles contain static information about Shelly device models,
including hardware capabilities, supported protocols, and resource limits.

Use these commands to:
- List all known device models
- Look up capabilities by model number
- Search for devices with specific features

### Examples

```
  # List all device profiles
  shelly profile list

  # Show info for a specific model
  shelly profile info SNSW-001P16EU

  # Search for dimming-capable devices
  shelly profile search dimmer
```

### Options

```
  -h, --help   help for profile
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
* [shelly profile info](shelly_profile_info.md)	 - Show device profile details
* [shelly profile list](shelly_profile_list.md)	 - List device profiles
* [shelly profile search](shelly_profile_search.md)	 - Search device profiles

