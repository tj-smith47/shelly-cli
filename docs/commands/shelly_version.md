## shelly version

Print version information

### Synopsis

Print the version of shelly CLI.

By default, shows version, commit, and build date.
Use --short for just the version number.
Use --json for machine-readable output.
Use --check to also check for available updates.

```
shelly version [flags]
```

### Examples

```
  # Show version info
  shelly version

  # Short version output
  shelly version --short

  # JSON output
  shelly version --json

  # Check for updates
  shelly version --check
```

### Options

```
  -c, --check   Check for available updates
  -h, --help    help for version
      --json    Output version info as JSON
  -s, --short   Print only the version number
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

