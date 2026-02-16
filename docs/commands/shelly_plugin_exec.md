## shelly plugin exec

Execute an extension

### Synopsis

Execute an extension explicitly with the given arguments.

This is useful when the extension name conflicts with a built-in command
or when you want to explicitly invoke an extension.

```
shelly plugin exec <name> [args...] [flags]
```

### Examples

```
  # Run an extension
  shelly extension exec myext --some-flag

  # Run with arguments
  shelly extension exec myext arg1 arg2
```

### Options

```
  -h, --help   help for exec
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

* [shelly plugin](shelly_plugin.md)	 - Manage CLI plugins

