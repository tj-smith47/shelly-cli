## shelly config set

Set CLI configuration values

### Synopsis

Set configuration values in the Shelly CLI config file.

Use dot notation for nested values (e.g., "defaults.timeout=30s").

```
shelly config set <key>=<value>... [flags]
```

### Examples

```
  # Set default timeout
  shelly config set defaults.timeout=30s

  # Set output format
  shelly config set defaults.output=json

  # Set multiple values
  shelly config set defaults.timeout=30s defaults.output=json
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly config](shelly_config.md)	 - Manage CLI configuration

