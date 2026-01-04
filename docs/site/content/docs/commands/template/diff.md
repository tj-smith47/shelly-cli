---
title: "shelly template diff"
description: "shelly template diff"
---

## shelly template diff

Compare a template with a device

### Synopsis

Compare a saved configuration template with a device's current configuration.

Shows differences between the template and the device, highlighting
what would change if the template were applied.

```
shelly template diff <template> <device> [flags]
```

### Examples

```
  # Compare template with device
  shelly template diff my-config bedroom

  # Output as JSON
  shelly template diff my-config bedroom -o json
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

* [shelly template](shelly_template.md)	 - Manage device configuration templates

