## shelly plugin list

List installed extensions

### Synopsis

List installed extensions.

Extensions are external plugins that add new commands to the CLI. They
are standalone executables named 'shelly-*' that can be installed from
git repositories or created locally.

By default, only shows extensions installed in the user plugins directory.
Use --all to show all discovered extensions including those in PATH.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: Name, Version, Source (github/url/local/unknown), Path

```
shelly plugin list [flags]
```

### Examples

```
  # List installed extensions
  shelly extension list

  # List all discovered extensions
  shelly extension list --all

  # Output as JSON
  shelly extension list -o json

  # Get extension names only
  shelly extension list -o json | jq -r '.[].name'

  # Find extensions without versions
  shelly extension list -o json | jq '.[] | select(.version == "")'

  # Short form
  shelly ext ls
```

### Options

```
  -a, --all    List all discovered extensions, not just installed ones
  -h, --help   help for list
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

