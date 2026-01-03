---
title: "shelly matter"
description: "shelly matter"
weight: 340
sidebar:
  collapsed: true
---

## shelly matter

Manage Matter connectivity

### Synopsis

Manage Matter connectivity on Gen4+ Shelly devices.

Matter is a unified smart home connectivity standard that allows
devices from different manufacturers to work together. Shelly Gen4+
devices support Matter, enabling integration with Apple Home,
Google Home, Amazon Alexa, and other Matter-compatible controllers.

Key concepts:
- Fabric: A Matter network/ecosystem (e.g., Apple Home, Google Home)
- Commissioner: The app/controller that adds devices to a fabric
- Commissionable: Device is ready to be added to a fabric

Note: Matter support requires Gen4+ devices.

### Examples

```
  # Show Matter status
  shelly matter status living-room

  # Enable Matter on a device
  shelly matter enable living-room

  # Show pairing code for commissioning
  shelly matter code living-room

  # Reset Matter configuration
  shelly matter reset living-room --yes
```

### Options

```
  -h, --help   help for matter
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly matter code](shelly_matter_code.md)	 - Show Matter pairing code
* [shelly matter disable](shelly_matter_disable.md)	 - Disable Matter on a device
* [shelly matter enable](shelly_matter_enable.md)	 - Enable Matter on a device
* [shelly matter reset](shelly_matter_reset.md)	 - Reset Matter configuration
* [shelly matter status](shelly_matter_status.md)	 - Show Matter status

