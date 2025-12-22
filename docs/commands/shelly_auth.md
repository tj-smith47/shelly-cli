## shelly auth

Manage device authentication

### Synopsis

Manage device authentication settings.

Enable, configure, or disable authentication for local device access.
When authentication is enabled, a username and password are required
for all device operations.

### Examples

```
  # Show authentication status
  shelly auth status living-room

  # Set authentication credentials
  shelly auth set living-room --user admin --password secret

  # Disable authentication
  shelly auth disable living-room
```

### Options

```
  -h, --help   help for auth
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly auth disable](shelly_auth_disable.md)	 - Disable authentication
* [shelly auth export](shelly_auth_export.md)	 - Export device credentials
* [shelly auth import](shelly_auth_import.md)	 - Import device credentials
* [shelly auth rotate](shelly_auth_rotate.md)	 - Rotate device credentials
* [shelly auth set](shelly_auth_set.md)	 - Set authentication credentials
* [shelly auth status](shelly_auth_status.md)	 - Show authentication status
* [shelly auth test](shelly_auth_test.md)	 - Test authentication credentials

