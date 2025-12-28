---
title: "shelly virtual set"
description: "shelly virtual set"
---

## shelly virtual set

Set a virtual component value

### Synopsis

Set the value of a virtual component.

The key format is "type:id", for example "boolean:200" or "number:201".

For boolean components, use --toggle to flip the current value.
For button components, use "trigger" as the value to press the button.

```
shelly virtual set <device> <key> <value> [flags]
```

### Examples

```
  # Set a boolean to true
  shelly virtual set kitchen boolean:200 true

  # Toggle a boolean
  shelly virtual set kitchen boolean:200 --toggle

  # Set a number
  shelly virtual set kitchen number:201 25.5

  # Set text
  shelly virtual set kitchen text:202 "Hello World"

  # Set enum value
  shelly virtual set kitchen enum:203 "option1"

  # Trigger a button
  shelly virtual set kitchen button:204 trigger
```

### Options

```
  -h, --help     help for set
  -t, --toggle   Toggle boolean value instead of setting
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

* [shelly virtual](shelly_virtual.md)	 - Manage virtual components

