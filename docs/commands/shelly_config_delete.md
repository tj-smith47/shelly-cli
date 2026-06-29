## shelly config delete

Delete CLI configuration values

### Synopsis

Delete configuration values from the Shelly CLI config file.

Use dot notation for nested values (e.g., "discovery.timeout").
Multiple keys can be deleted at once.

If a key has nested child values, confirmation is required unless --yes is provided.

```
shelly config delete <key>... [flags]
```

### Examples

```
  # Delete a single setting
  shelly config delete discovery.timeout

  # Delete multiple settings
  shelly config delete discovery.timeout discovery.network

  # Delete a parent key with all children (with confirmation)
  shelly config delete discovery

  # Skip confirmation prompt
  shelly config delete discovery --yes

  # Using alias
  shelly config rm editor
```

### Options

```
  -h, --help   help for delete
  -y, --yes    Skip confirmation prompt
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

* [shelly config](shelly_config.md)	 - Manage CLI configuration

