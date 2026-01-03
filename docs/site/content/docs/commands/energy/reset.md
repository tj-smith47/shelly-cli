---
title: "shelly energy reset"
description: "shelly energy reset"
---

## shelly energy reset

Reset energy monitor counters

### Synopsis

Reset energy counters for an EM (3-phase) energy monitor.

Note: Only EM components support counter reset. EM1 components
do not have a reset capability.

```
shelly energy reset <device> [id] [flags]
```

### Examples

```
  # Reset all counters for EM component 0
  shelly energy reset shelly-3em-pro 0

  # Reset specific counter types
  shelly energy reset shelly-3em-pro 0 --types active,reactive

  # Reset with device alias
  shelly energy reset basement-em
```

### Options

```
  -h, --help            help for reset
      --types strings   Counter types to reset (leave empty for all)
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

* [shelly energy](shelly_energy.md)	 - Energy monitoring operations (EM/EM1 components)

