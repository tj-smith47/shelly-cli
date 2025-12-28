---
title: "shelly script stop"
description: "shelly script stop"
---

## shelly script stop

Stop a running script

### Synopsis

Stop a running script on a Gen2+ Shelly device.

```
shelly script stop <device> <id> [flags]
```

### Examples

```
  # Stop a script
  shelly script stop living-room 1
```

### Options

```
  -h, --help   help for stop
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

