## shelly profile list

List device profiles

### Synopsis

List all known Shelly device profiles.

Optionally filter by generation or series.

```
shelly profile list [flags]
```

### Examples

```
  # List all profiles
  shelly profile list

  # List Gen2 devices
  shelly profile list --gen gen2

  # List Pro series devices
  shelly profile list --series pro

  # JSON output
  shelly profile list -o json
```

### Options

```
      --gen string      Filter by generation (gen1, gen2, gen3, gen4)
  -h, --help            help for list
  -o, --output string   Output format: table, json, yaml (default "table")
      --series string   Filter by series (classic, plus, pro, mini, blu, wave)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly profile](shelly_profile.md)	 - Device profile information

