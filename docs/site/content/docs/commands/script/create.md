---
title: "shelly script create"
description: "shelly script create"
---

## shelly script create

Create a new script

### Synopsis

Create a new script on a Gen2+ Shelly device.

You can provide the script code inline with --code or from a file with --file.
Use --enable to automatically enable the script after creation.

```
shelly script create <device> [flags]
```

### Examples

```
  # Create an empty script
  shelly script create living-room --name "My Script"

  # Create with inline code
  shelly script create living-room --name "Hello" --code "print('Hello!');" --enable

  # Create from file
  shelly script create living-room --name "Auto Light" --file auto-light.js --enable
```

### Options

```
      --code string   Script code (inline)
      --enable        Enable script after creation
  -f, --file string   Script code file
  -h, --help          help for create
      --name string   Script name
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

