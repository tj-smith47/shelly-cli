## shelly cache show

Show cache information

### Synopsis

Show the cache directory path and its contents.

```
shelly cache show [flags]
```

### Examples

```
  # Show cache info
  shelly cache show
```

### Options

```
  -h, --help   help for show
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

* [shelly cache](shelly_cache.md)	 - Manage CLI cache

