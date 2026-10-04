---
title: "shelly auth rotate"
description: "shelly auth rotate"
---

## shelly auth rotate

Rotate device credentials

### Synopsis

Rotate device authentication credentials.

This command sets a new password on the device, optionally generating a
secure random one. Gen1 devices accept any username, and without --user keep
the user stored for them; Gen2+ devices have a single user, admin.

The new credentials are saved for a registered device, and the command then
makes an authenticated request with the new password and fails if the
device does not accept it. A generated password is shown only with --show.

```
shelly auth rotate <device> [flags]
```

### Examples

```
  # Rotate with a new password
  shelly auth rotate living-room --password newSecret123

  # Read the new password from stdin, keeping it out of shell history
  shelly auth rotate living-room --password-stdin < ~/.shelly-living-room-password

  # Generate a random password
  shelly auth rotate living-room --generate

  # Generate and show the new password
  shelly auth rotate living-room --generate --show

  # Use specific password length
  shelly auth rotate living-room --generate --length 24
```

### Options

```
      --generate          Generate a random password
  -h, --help              help for rotate
      --length int        Generated password length (default 16)
      --password string   New password (or use --password-stdin or --generate)
      --password-stdin    Read the new password from stdin
      --show              Show the new password in output
      --user string       Username: any name on Gen1, where it defaults to the stored user or admin; Gen2+ devices allow only admin
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

