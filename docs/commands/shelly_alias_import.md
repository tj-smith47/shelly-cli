## shelly alias import

Import aliases from a YAML file

### Synopsis

Import command aliases from a YAML file.

By default, existing aliases with the same name are overwritten.
Use --merge to skip existing aliases instead of overwriting them.

The file format is:
  aliases:
    name1: "command1"
    name2: "command2"
    shellalias: "!shell command"

```
shelly alias import <file> [flags]
```

### Examples

```
  # Import aliases (overwrite existing)
  shelly alias import aliases.yaml

  # Import without overwriting existing aliases
  shelly alias import aliases.yaml --merge
```

### Options

```
  -h, --help    help for import
  -m, --merge   Merge with existing aliases (skip conflicts)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

* [shelly alias](shelly_alias.md)	 - Manage command aliases

