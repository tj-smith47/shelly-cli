---
title: "shelly qr"
description: "shelly qr"
weight: 480
sidebar:
  collapsed: true
---

## shelly qr

Generate device QR code

### Synopsis

Generate a QR code for a Shelly device.

The QR code can contain:
  - Device web interface URL (default)
  - WiFi network configuration (with --wifi flag)

By default, displays the QR code as ASCII art in the terminal.

```
shelly qr <device> [flags]
```

### Examples

```
  # Generate QR code for device web UI
  shelly qr kitchen-light

  # Generate WiFi configuration QR code
  shelly qr kitchen-light --wifi

  # Show only the content without QR display
  shelly qr kitchen-light --no-qr

  # JSON output with QR content
  shelly qr kitchen-light -o json
```

### Options

```
  -h, --help       help for qr
      --no-qr      Don't display QR code, just show content
      --size int   QR code size in pixels (for --save) (default 256)
      --wifi       Generate WiFi config QR content
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

