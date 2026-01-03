---
title: "shelly device alias"
description: "shelly device alias"
---

## shelly device alias

Manage device aliases

### Synopsis

Manage short aliases for devices.

Aliases allow you to reference devices by short, memorable names
in addition to their full registered names. For example, you can
create an alias "mb" for a device named "master-bathroom".

Aliases must be 1-32 characters, start with a letter or number,
and contain only letters, numbers, hyphens, and underscores.

```
shelly device alias <device> [alias] [flags]
```

### Examples

```
  # Add an alias to a device
  shelly device alias master-bathroom mb
  shelly device alias kitchen-light k

  # List aliases for a device
  shelly device alias master-bathroom --list

  # Remove an alias from a device
  shelly device alias master-bathroom --remove mb

  # Use an alias in commands
  shelly toggle mb
  shelly status k
```

### Options

```
  -h, --help            help for alias
  -l, --list            List all aliases for the device
  -r, --remove string   Remove the specified alias
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

