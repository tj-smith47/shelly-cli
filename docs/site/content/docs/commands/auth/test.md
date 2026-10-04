---
title: "shelly auth test"
description: "shelly auth test"
---

## shelly auth test

Test authentication credentials

### Synopsis

Test authentication credentials against a device.

The device is asked for data it only returns to an authenticated caller
(Sys.GetStatus on Gen2+ devices, /settings on Gen1 devices), so a wrong
password fails the test. Without --password or --password-stdin the
credentials stored for the device are tested; with one of them the given
credentials are tested instead, and nothing is stored. The user defaults to
admin.

A device with authentication disabled accepts any credentials; the command
says so instead of reporting the password as correct.

Exit codes:
  0 - The device accepted the credentials
  1 - The device rejected the credentials, or could not be reached

```
shelly auth test <device> [flags]
```

### Examples

```
  # Test the credentials stored for the device
  shelly auth test living-room

  # Test other credentials
  shelly auth test living-room --user admin --password secret

  # Read the password from stdin (prompts without echo on a terminal)
  shelly auth test living-room --password-stdin < ~/.shelly-living-room-password

  # Quick test with short timeout
  shelly auth test living-room --timeout 5s
```

### Options

```
  -h, --help               help for test
      --password string    Password to test instead of the stored one
      --password-stdin     Read the password to test from stdin
      --timeout duration   Connection timeout (default 10s)
      --user string        Username for --password or --password-stdin (default admin)
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
      --raw                     Print the exact device response(s) as a JSON array and suppress normal output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly auth](shelly_auth.md)	 - Manage device authentication

