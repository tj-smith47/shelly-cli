---
title: "shelly auth set"
description: "shelly auth set"
---

## shelly auth set

Set authentication credentials

### Synopsis

Set the password a device requires, turning authentication on.

Gen1 devices accept any username; without --user they get the user stored
for them, or admin. Gen2+ devices have a single user, admin, so --user can
only be left out or set to admin.

Once the device has the new password, every request to it needs that
password, so the new credentials are saved for a registered device. The
command then makes an authenticated request with the new password and fails
if the device does not accept it.

```
shelly auth set <device> [flags]
```

### Examples

```
  # Set the password (user admin)
  shelly auth set living-room --password secret

  # Read the password from stdin, keeping it out of shell history
  shelly auth set living-room --password-stdin < ~/.shelly-living-room-password

  # Set a custom username (Gen1 devices only)
  shelly auth set garage-gen1 --user myuser --password secret
```

### Options

```
  -h, --help              help for set
      --password string   New device password (or use --password-stdin)
      --password-stdin    Read the new device password from stdin
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

