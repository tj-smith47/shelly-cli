---
title: "shelly auth rotate"
description: "shelly auth rotate"
---

## shelly auth rotate

Rotate device credentials

### Synopsis

Rotate device authentication credentials.

This command sets new authentication credentials on the device,
optionally generating a secure random password.

For security best practices:
  - Rotate credentials periodically
  - Use generated passwords (--generate)
  - Store credentials securely

```
shelly auth rotate <device> [flags]
```

### Examples

```
  # Rotate with a new password
  shelly auth rotate living-room --password newSecret123

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
      --password string   New password (or use --generate)
      --show              Show the new password in output
      --user string       Username for authentication (default "admin")
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

