---
title: "shelly modbus"
description: "shelly modbus"
weight: 400
sidebar:
  collapsed: true
---

## shelly modbus

Manage Modbus-TCP configuration

### Synopsis

Manage Modbus-TCP server on Shelly Gen2+ devices.

The Modbus-TCP server runs on port 502 when enabled, allowing integration
with industrial automation systems and SCADA software.

Device info registers (when enabled):
  30000: Device MAC (6 registers / 12 bytes)
  30006: Device model (10 registers / 20 bytes)
  30016: Device name (32 registers / 64 bytes)

Additional component-specific registers are documented per-component.

### Examples

```
  # Check Modbus status
  shelly modbus status kitchen

  # Enable Modbus
  shelly modbus enable kitchen

  # Disable Modbus
  shelly modbus disable kitchen
```

### Options

```
  -h, --help   help for modbus
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly modbus disable](shelly_modbus_disable.md)	 - Disable Modbus-TCP server
* [shelly modbus enable](shelly_modbus_enable.md)	 - Enable Modbus-TCP server
* [shelly modbus status](shelly_modbus_status.md)	 - Show Modbus-TCP status

