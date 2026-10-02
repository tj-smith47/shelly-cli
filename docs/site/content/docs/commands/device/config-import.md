---
title: "shelly device config import"
description: "shelly device config import"
---

## shelly device config import

Import configuration from a file

### Synopsis

Import device configuration from a JSON or YAML file.

Keys present in the file are applied to the device; keys absent from the file
are left unchanged (the device merges the update — there is no whole-config
replace primitive). Capture a file in this format with 'shelly device config export'.

WiFi stations in the file: when the file's sys.device.mac is the device's own,
a station on the device's current network is written without a password (the
device keeps its key), and a changed network takes the password stored on this
host. A file from another device, or one with no MAC, never copies its station
address (ip, netmask, gw, nameserver, ipv4mode), and its network is written
only when the device is not already on it and this host has its password. A
station that cannot be written that way is left out with a warning; set it
with 'shelly wifi set'. --dry-run shows each station's planned write without
its password.

```
shelly device config import <device> <file> [flags]
```

### Examples

```
  # Import configuration
  shelly device config import living-room config-backup.json

  # Dry run - show what would change without applying
  shelly device config import living-room config.json --dry-run
```

### Options

```
      --dry-run   Show what would be changed without applying
  -h, --help      help for import
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
      --raw                     Print the exact device response(s) as a JSON array and suppress normal output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly device config](shelly_device_config.md)	 - Manage device configuration

