---
title: "shelly auth import"
description: "shelly auth import"
---

## shelly auth import

Import device credentials

### Synopsis

Import device authentication credentials from a file.

Imports credentials that were previously exported with auth export.

```
shelly auth import <file> [flags]
```

### Examples

```
  # Import credentials
  shelly auth import credentials.json

  # Preview without applying
  shelly auth import credentials.json --dry-run
```

### Options

```
      --dry-run   Preview actions without executing
  -h, --help      help for import
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

* [shelly auth](shelly_auth.md)	 - Manage device authentication

