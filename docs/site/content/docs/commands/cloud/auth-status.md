---
title: "shelly cloud auth-status"
description: "shelly cloud auth-status"
---

## shelly cloud auth-status

Show cloud authentication status

### Synopsis

Show the current Shelly Cloud authentication status.

Displays whether you're logged in, your email, and token validity.

```
shelly cloud auth-status [flags]
```

### Examples

```
  # Check authentication status
  shelly cloud auth-status

  # Also available as whoami
  shelly cloud whoami
```

### Options

```
  -h, --help   help for auth-status
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

* [shelly cloud](shelly_cloud.md)	 - Manage cloud connection and Shelly Cloud API

