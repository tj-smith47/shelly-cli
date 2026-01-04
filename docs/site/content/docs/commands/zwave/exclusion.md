---
title: "shelly zwave exclusion"
description: "shelly zwave exclusion"
---

## shelly zwave exclusion

Show exclusion instructions

### Synopsis

Show Z-Wave exclusion (unpairing) instructions for a device.

Exclusion modes:
  button - Manual exclusion using the S button (default)
  switch - Manual exclusion using the connected switch

```
shelly zwave exclusion <model> [flags]
```

### Examples

```
  # Show button-based exclusion (default)
  shelly zwave exclusion SNSW-001P16ZW

  # Show switch-based exclusion
  shelly zwave exclusion SNSW-001P16ZW --mode switch
```

### Options

```
  -h, --help          help for exclusion
      --mode string   Exclusion mode (button, switch) (default "button")
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

* [shelly zwave](shelly_zwave.md)	 - Z-Wave device utilities

