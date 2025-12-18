## shelly alias export

Export aliases to a YAML file

### Synopsis

Export all command aliases to a YAML file.

If no file is specified, outputs to stdout.

The output format is:
  aliases:
    name1: "command1"
    name2: "command2"
    shellalias: "!shell command"

```
shelly alias export [file] [flags]
```

### Examples

```
  # Export to file
  shelly alias export aliases.yaml

  # Export to stdout
  shelly alias export

  # Export and pipe to another command
  shelly alias export | cat
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
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly alias](shelly_alias.md)	 - Manage command aliases

