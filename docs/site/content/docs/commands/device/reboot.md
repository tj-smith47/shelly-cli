---
title: "shelly device reboot"
description: "shelly device reboot"
---

## shelly device reboot

Reboot device

### Synopsis

Reboot a Shelly device. Use --delay to set a delay in milliseconds.

```
shelly device reboot <device> [flags]
```

### Examples

```
  # Reboot a device
  shelly device reboot living-room

  # Reboot with confirmation skipped
  shelly device reboot living-room -y

  # Reboot with delay
  shelly device reboot living-room --delay 5000
```

### Options

```
  -d, --delay int   Delay in milliseconds before reboot
  -h, --help        help for reboot
  -y, --yes         Skip confirmation prompt
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

