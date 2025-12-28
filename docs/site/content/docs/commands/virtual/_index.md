---
title: "shelly virtual"
description: "shelly virtual"
weight: 690
sidebar:
  collapsed: true
---

## shelly virtual

Manage virtual components

### Synopsis

Manage virtual components on Shelly Gen2+ devices.

Virtual components allow you to create custom boolean, number, text, enum,
button, and group components that can be used in scripts and automations.

Virtual component IDs are automatically assigned in the range 200-299.

### Examples

```
  # List virtual components on a device
  shelly virtual list kitchen

  # Create a virtual boolean
  shelly virtual create kitchen boolean --name "Away Mode"

  # Get a virtual component value
  shelly virtual get kitchen boolean:200

  # Set a virtual component value
  shelly virtual set kitchen boolean:200 true

  # Delete a virtual component
  shelly virtual delete kitchen boolean:200
```

### Options

```
  -h, --help   help for virtual
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
* [shelly virtual create](shelly_virtual_create.md)	 - Create a virtual component
* [shelly virtual delete](shelly_virtual_delete.md)	 - Delete a virtual component
* [shelly virtual get](shelly_virtual_get.md)	 - Get a virtual component value
* [shelly virtual list](shelly_virtual_list.md)	 - List virtual components on a device
* [shelly virtual set](shelly_virtual_set.md)	 - Set a virtual component value

