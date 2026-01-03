## shelly auth export

Export device credentials

### Synopsis

Export device authentication credentials to a file.

Exports credentials from the CLI config for backup or transfer.
Use with auth import to restore credentials on another system.

```
shelly auth export [device...] [flags]
```

### Examples

```
  # Export all credentials
  shelly auth export --all -o credentials.json

  # Export specific devices
  shelly auth export kitchen bedroom -o creds.json
```

### Options

```
  -a, --all             Target all registered devices
  -h, --help            help for export
  -o, --output string   Output file path (default "credentials.json")
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly auth](shelly_auth.md)	 - Manage device authentication

