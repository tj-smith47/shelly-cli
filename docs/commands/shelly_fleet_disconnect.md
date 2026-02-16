## shelly fleet disconnect

Disconnect from Shelly Cloud hosts

### Synopsis

Disconnect from all connected Shelly Cloud hosts.

This command closes any active WebSocket connections to Shelly Cloud hosts.
Note: In CLI mode, connections are typically ephemeral per command. This
command is useful for explicitly verifying connectivity and cleanup.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token

```
shelly fleet disconnect [flags]
```

### Examples

```
  # Disconnect from all hosts
  shelly fleet disconnect
```

### Options

```
  -h, --help   help for disconnect
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

* [shelly fleet](shelly_fleet.md)	 - Cloud-based fleet management

