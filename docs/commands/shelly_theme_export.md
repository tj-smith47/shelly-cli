## shelly theme export

Export current theme

### Synopsis

Export the current theme configuration to a file.

Exports the base theme name, any custom color overrides, and the effective
colors (what you actually see). The exported file can be imported back with
'shelly theme import'.

If no file is specified, outputs to stdout.

```
shelly theme export [file] [flags]
```

### Examples

```
  # Export to file
  shelly theme export mytheme.yaml

  # Export to stdout
  shelly theme export
```

### Options

```
  -h, --help   help for export
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

* [shelly theme](shelly_theme.md)	 - Manage CLI color themes

