## shelly config delete

Delete CLI configuration values

### Synopsis

Delete configuration values from the Shelly CLI config file.

Use dot notation for nested values (e.g., "defaults.timeout").
Multiple keys can be deleted at once.

If a key has nested child values, confirmation is required unless --yes is provided.

```
shelly config delete <key>... [flags]
```

### Examples

```
  # Delete a single setting
  shelly config delete defaults.timeout

  # Delete multiple settings
  shelly config delete defaults.timeout defaults.output

  # Delete a parent key with all children (with confirmation)
  shelly config delete defaults

  # Skip confirmation prompt
  shelly config delete defaults --yes

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

* [shelly config](shelly_config.md)	 - Manage CLI configuration

