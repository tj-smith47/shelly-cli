## shelly plugin remove

Remove an installed extension

### Synopsis

Remove an installed extension.

Only extensions installed in the user plugins directory can be removed.
Extensions found in PATH cannot be removed with this command.

```
shelly plugin remove <name> [flags]
```

### Examples

```
  # Remove an extension
  shelly extension remove myext
```

### Options

```
  -h, --help   help for remove
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

* [shelly plugin](shelly_plugin.md)	 - Manage CLI plugins

