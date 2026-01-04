---
title: "shelly script upload"
description: "shelly script upload"
---

## shelly script upload

Upload script from file

### Synopsis

Upload script code to a device from a file.

By default, replaces the existing code. Use --append to add to existing code.

```
shelly script upload <device> <id> <file> [flags]
```

### Examples

```
  # Upload script from file
  shelly script upload living-room 1 script.js

  # Append code from file
  shelly script upload living-room 1 additions.js --append
```

### Options

```
      --append   Append to existing code
  -h, --help     help for upload
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

* [shelly script](shelly_script.md)	 - Manage device scripts

