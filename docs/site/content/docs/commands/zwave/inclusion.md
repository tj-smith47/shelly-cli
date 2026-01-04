---
title: "shelly zwave inclusion"
description: "shelly zwave inclusion"
---

## shelly zwave inclusion

Show inclusion instructions

### Synopsis

Show Z-Wave inclusion (pairing) instructions for a device.

Inclusion modes:
  smart_start - Automatic inclusion via QR code (recommended)
  button      - Manual inclusion using the S button
  switch      - Manual inclusion using the connected switch

```
shelly zwave inclusion <model> [flags]
```

### Examples

```
  # Show SmartStart inclusion (default)
  shelly zwave inclusion SNSW-001P16ZW

  # Show button-based inclusion
  shelly zwave inclusion SNSW-001P16ZW --mode button

  # Show switch-based inclusion
  shelly zwave inclusion SNSW-001P16ZW --mode switch
```

### Options

```
  -h, --help          help for inclusion
      --mode string   Inclusion mode (smart_start, button, switch) (default "smart_start")
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

