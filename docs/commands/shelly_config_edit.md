## shelly config edit

Open CLI config in editor

### Synopsis

Open the Shelly CLI configuration file in your default editor.

The editor is determined by:
  1. 'editor' setting in config file
  2. $EDITOR environment variable
  3. $VISUAL environment variable
  4. Falls back to common editors (nano, vim, vi)

This allows you to directly edit the configuration file, including
devices, aliases, groups, scenes, and other settings.

```
shelly config edit [flags]
```

### Examples

```
  # Open config in default editor
  shelly config edit

  # Set EDITOR and open
  EDITOR=nano shelly config edit
```

### Options

```
  -h, --help   help for edit
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

* [shelly config](shelly_config.md)	 - Manage CLI configuration

