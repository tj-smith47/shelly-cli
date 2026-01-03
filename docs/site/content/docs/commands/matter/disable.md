---
title: "shelly matter disable"
description: "shelly matter disable"
---

## shelly matter disable

Disable Matter on a device

### Synopsis

Disable Matter connectivity on a Shelly device.

When Matter is disabled, the device will no longer be controllable
through Matter fabrics. Existing fabric pairings are preserved but
will not function until Matter is re-enabled.

To permanently remove all Matter pairings, use 'shelly matter reset'
instead.

```
shelly matter disable <device> [flags]
```

### Examples

```
  # Disable Matter
  shelly matter disable living-room

  # Re-enable later
  shelly matter enable living-room
```

### Options

```
  -h, --help   help for disable
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

* [shelly matter](shelly_matter.md)	 - Manage Matter connectivity

