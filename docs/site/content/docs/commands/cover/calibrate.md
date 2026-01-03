---
title: "shelly cover calibrate"
description: "shelly cover calibrate"
---

## shelly cover calibrate

Calibrate cover

### Synopsis

Start calibration for a cover/roller component.

Calibration determines the open and close times for the cover.
The cover will move to both extremes during calibration.

```
shelly cover calibrate <device> [flags]
```

### Examples

```
  # Calibrate a cover
  shelly cover calibrate bedroom

  # Calibrate specific cover ID
  shelly cover cal bedroom --id 1
```

### Options

```
  -h, --help     help for calibrate
  -i, --id int   Cover component ID (default 0)
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

* [shelly cover](shelly_cover.md)	 - Control cover/roller components

