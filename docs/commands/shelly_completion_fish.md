## shelly completion fish

Generate fish completion script

### Synopsis

Generate the autocompletion script for fish.

To load completions in your current shell session:

  shelly completion fish | source

To load completions for every new session, execute once:

  shelly completion fish > ~/.config/fish/completions/shelly.fish

```
shelly completion fish
```

### Examples

```
  shelly completion fish > ~/.config/fish/completions/shelly.fish
```

### Options

```
  -h, --help   help for fish
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

* [shelly completion](shelly_completion.md)	 - Generate shell completion scripts

