## shelly log tail

Tail log file

### Synopsis

Show and optionally follow the log file in real-time.

```
shelly log tail [flags]
```

### Examples

```
  # Show last entries and follow
  shelly log tail -f

  # Just show last entries
  shelly log tail
```

### Options

```
  -f, --follow   Follow log output
  -h, --help     help for tail
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly log](shelly_log.md)	 - Manage CLI logs

