## shelly alias delete

Delete a alias

### Synopsis

Delete a saved alias permanently.

```
shelly alias delete <alias> [flags]
```

### Examples

```
  # Delete a alias (with confirmation)
  shelly alias delete my-alias

  # Delete without confirmation
  shelly alias delete my-alias --yes

  # Using alias
  shelly alias rm my-alias
```

### Options

```
  -h, --help   help for delete
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

* [shelly alias](shelly_alias.md)	 - Manage command aliases

