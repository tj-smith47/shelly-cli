---
title: "shelly script delete"
description: "shelly script delete"
---

## shelly script delete

Delete a script

### Synopsis

Delete a script from a Gen2+ Shelly device.

This permanently removes the script and its code from the device.

```
shelly script delete <device> <script-id> [flags]
```

### Examples

```
  # Delete a script
  shelly script delete <device> 1

  # Delete without confirmation
  shelly script delete <device> 1 --yes
```

### Options

```
  -h, --help   help for delete
  -y, --yes    Skip confirmation prompt
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

* [shelly script](shelly_script.md)	 - Manage device scripts

