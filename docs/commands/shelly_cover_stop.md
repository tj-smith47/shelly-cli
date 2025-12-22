## shelly cover stop

Stop cover

### Synopsis

Stop a cover/roller component on the specified device.

```
shelly cover stop <device> [flags]
```

### Examples

```
  # Stop cover movement
  shelly cover stop bedroom

  # Stop specific cover ID
  shelly cover halt bedroom --id 1
```

### Options

```
  -h, --help     help for stop
  -i, --id int   Cover component ID (default 0)
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

* [shelly cover](shelly_cover.md)	 - Control cover/roller components

