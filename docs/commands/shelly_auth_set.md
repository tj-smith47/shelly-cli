## shelly auth set

Set authentication credentials

### Synopsis

Set authentication credentials for a device.

This enables authentication if not already enabled. The username defaults
to "admin" if not specified.

```
shelly auth set <device> [flags]
```

### Examples

```
  # Set credentials with default username
  shelly auth set living-room --password secret

  # Set credentials with custom username
  shelly auth set living-room --user myuser --password secret
```

### Options

```
  -h, --help              help for set
      --password string   Password for authentication (required)
      --realm string      Authentication realm (optional)
      --user string       Username for authentication (default "admin")
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly auth](shelly_auth.md)	 - Manage device authentication

