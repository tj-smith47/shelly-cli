---
title: "shelly link"
description: "shelly link"
weight: 340
sidebar:
  collapsed: true
---

## shelly link

Manage device power links

### Synopsis

Manage parent-child power relationships between devices.

Links define which switch controls the power to another device.
When a linked child device is offline, its state is derived from
the parent switch state. Control commands (on/off/toggle) automatically
proxy to the parent switch when the child is unreachable.

### Examples

```
  # Link a bulb to a switch (bulb is powered by switch:0)
  shelly link set bulb-duo bedroom-2pm

  # Link with a specific switch ID
  shelly link set garage-light garage-switch --switch-id 1

  # List all links
  shelly link list

  # Show link status with derived state
  shelly link status

  # Remove a link
  shelly link delete bulb-duo
```

### Options

```
  -h, --help   help for link
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
* [shelly link delete](shelly_link_delete.md)	 - Delete a link
* [shelly link list](shelly_link_list.md)	 - List links
* [shelly link set](shelly_link_set.md)	 - Set a device power link
* [shelly link status](shelly_link_status.md)	 - Show link status with derived device state

