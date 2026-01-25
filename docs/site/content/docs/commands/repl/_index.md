---
title: "shelly repl"
description: "shelly repl"
weight: 610
sidebar:
  collapsed: true
---

## shelly repl

Launch interactive REPL

### Synopsis

Launch an interactive REPL (Read-Eval-Print Loop) for Shelly CLI.

This provides a command-line shell where you can enter Shelly commands
without prefixing them with 'shelly'. It supports command history and
readline-style line editing (arrow keys, Ctrl+A/E, etc.).

Available commands in REPL:
  help             Show available commands
  devices          List registered devices
  connect <device> Set active device for subsequent commands
  disconnect       Clear active device
  status           Show status of active device
  on               Turn on all device components
  off              Turn off all device components
  toggle           Toggle all device components
  rpc <method>     Execute raw RPC call on active device
  exit, quit, q    Exit the REPL

To control a specific component (e.g., switch ID 1 on a multi-switch device),
use the rpc command with the appropriate method and parameters.

```
shelly repl [flags]
```

### Examples

```
  # Start REPL mode
  shelly repl

  # Start with a default device
  shelly repl --device living-room

  # Example session:
  > devices
  > connect living-room
  > status
  > on
  > exit

  # Control specific switch on multi-switch device (after connecting):
  > connect dual-switch
  > rpc Switch.Set {"id":1,"on":true}
  > rpc Switch.Toggle {"id":0}

  # Or without connecting first (device as first arg):
  > rpc dual-switch Switch.Set {"id":1,"on":true}
```

### Options

```
  -d, --device string   Default device to connect to
  -h, --help            help for repl
      --no-prompt       Disable interactive prompt (for scripting)
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

