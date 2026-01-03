---
title: "shelly light set"
description: "shelly light set"
---

## shelly light set

Set light parameters

### Synopsis

Set parameters of a light component on the specified device.

You can set brightness and on/off state.
Values not specified will be left unchanged.

```
shelly light set <device> [flags]
```

### Examples

```
  # Set brightness to 50%
  shelly light set kitchen --brightness 50

  # Turn on and set brightness
  shelly light br kitchen -b 75 --on
```

### Options

```
  -b, --brightness int   Brightness (0-100) (default -1)
  -h, --help             help for set
  -i, --id int           Light component ID (default 0)
      --on               Turn on
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly light](shelly_light.md)	 - Control light components

