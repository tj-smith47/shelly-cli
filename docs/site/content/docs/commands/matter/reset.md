---
title: "shelly matter reset"
description: "shelly matter reset"
---

## shelly matter reset

Reset Matter configuration

### Synopsis

Reset all Matter settings on a Shelly device.

This command:
- Unpairs the device from all Matter fabrics
- Erases all Matter credentials and data
- Returns Matter to factory default state

Unlike 'shelly device factory-reset', this only affects Matter settings.
WiFi, device name, and other configurations are preserved.

After reset, the device must be re-commissioned to any Matter fabrics.

```
shelly matter reset <device> [flags]
```

### Examples

```
  # Reset Matter (with confirmation)
  shelly matter reset living-room

  # Reset without confirmation
  shelly matter reset living-room --yes
```

### Options

```
  -h, --help   help for reset
  -y, --yes    Skip confirmation prompt
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

* [shelly matter](shelly_matter.md)	 - Manage Matter connectivity

