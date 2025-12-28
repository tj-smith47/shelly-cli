---
title: "shelly device factory-reset"
description: "shelly device factory-reset"
---

## shelly device factory-reset

Factory reset a device

### Synopsis

Factory reset a Shelly device to its default settings.

WARNING: This will ERASE ALL settings on the device including:
- WiFi configuration
- Device name
- Authentication settings
- Schedules
- Scripts
- Webhooks

The device will return to AP mode and need to be reconfigured.

This command requires both --yes and --confirm flags for safety.

```
shelly device factory-reset <device> [flags]
```

### Examples

```
  # Factory reset with double confirmation
  shelly device factory-reset living-room --yes --confirm

  # Using aliases
  shelly device fr living-room --yes --confirm
  shelly device reset living-room --yes --confirm
  shelly device wipe living-room --yes --confirm

  # This will fail (safety measure)
  shelly device factory-reset living-room --yes
```

### Options

```
      --confirm   Double-confirm destructive operation
  -h, --help      help for factory-reset
  -y, --yes       Skip confirmation prompt
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly device](shelly_device.md)	 - Manage Shelly devices

