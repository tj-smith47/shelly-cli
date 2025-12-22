## shelly theme set

Set the current theme

### Synopsis

Set the current CLI color theme.

Use --save to persist the theme to your configuration file. Without --save,
the theme is only applied for the current session.

```
shelly theme set <theme> [flags]
```

### Examples

```
  # Set theme for current session
  shelly theme set dracula

  # Set and save to config
  shelly theme set nord --save
```

### Options

```
  -h, --help   help for set
      --save   Save theme to configuration file
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly theme](shelly_theme.md)	 - Manage CLI color themes

