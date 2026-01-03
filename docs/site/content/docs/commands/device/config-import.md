---
title: "shelly device config import"
description: "shelly device config import"
---

## shelly device config import

Import configuration from a file

### Synopsis

Import device configuration from a JSON or YAML file.

By default, only specified keys are updated (merge mode). Use --overwrite
to replace the entire configuration.

```
shelly device config import <device> <file> [flags]
```

### Examples

```
  # Import configuration (merge mode)
  shelly config import living-room config-backup.json

  # Dry run - show what would change without applying
  shelly config import living-room config.json --dry-run

  # Overwrite entire configuration
  shelly config import living-room config.json --overwrite
```

### Options

```
      --dry-run     Show what would be changed without applying
  -h, --help        help for import
      --merge       Merge with existing configuration (default) (default true)
      --overwrite   Overwrite entire configuration
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

* [shelly device config](shelly_device_config.md)	 - Manage device configuration

