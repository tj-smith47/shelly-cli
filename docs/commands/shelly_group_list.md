## shelly group list

List groups

### Synopsis

List all configured groups.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

```
shelly group list [flags]
```

### Examples

```
  # List all groups
  shelly group list

  # Output as JSON
  shelly group list -o json

  # Output as YAML
  shelly group list -o yaml
```

### Options

```
  -h, --help   help for list
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

* [shelly group](shelly_group.md)	 - Manage device groups

