---
title: "shelly auth test"
description: "shelly auth test"
---

## shelly auth test

Test authentication credentials

### Synopsis

Test authentication credentials against a device.

This command verifies that the provided credentials are valid
by attempting to connect to the device.

Exit codes:
  0 - Authentication successful
  1 - Authentication failed or error

```
shelly auth test <device> [flags]
```

### Examples

```
  # Test with provided credentials
  shelly auth test living-room --user admin --password secret

  # Test with configured credentials
  shelly auth test living-room

  # Quick test with short timeout
  shelly auth test living-room --timeout 5s
```

### Options

```
  -h, --help               help for test
      --password string    Password to test
      --timeout duration   Connection timeout (default 10s)
      --user string        Username to test
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

* [shelly auth](shelly_auth.md)	 - Manage device authentication

