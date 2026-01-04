## shelly log export

Export log file

### Synopsis

Export the log file to a specified location.

```
shelly log export [flags]
```

### Examples

```
  # Export to file
  shelly log export -o /tmp/shelly-debug.log

  # Export to stdout
  shelly log export
```

### Options

```
  -h, --help            help for export
  -o, --output string   Output file path
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly log](shelly_log.md)	 - Manage CLI logs

