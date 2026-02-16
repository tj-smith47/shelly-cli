---
title: "shelly rgb set"
description: "shelly rgb set"
---

## shelly rgb set

Set RGB parameters

### Synopsis

Set parameters of an RGB light component on the specified device.

You can set color values (red, green, blue) brightness, and on/off state.
Values not specified will be left unchanged.

```
shelly rgb set <device> [flags]
```

### Examples

```
  # Set RGB color to red
  shelly rgb set living-room --red 255 --green 0 --blue 0

  # Set RGB with brightness
  shelly rgb color living-room -r 0 -g 255 -b 128 --brightness 75
```

### Options

```
  -b, --blue int         Blue value (0-255) (default -1)
      --brightness int   Brightness (0-100) (default -1)
  -g, --green int        Green value (0-255) (default -1)
  -h, --help             help for set
  -i, --id int           RGB component ID (default 0)
      --on               Turn on
  -r, --red int          Red value (0-255) (default -1)
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly rgb](shelly_rgb.md)	 - Control RGB light components

