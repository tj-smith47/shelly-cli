## shelly export csv

Export device list as CSV

### Synopsis

Export multiple devices as a CSV file.

The CSV includes device name, address, model, generation, and online status.
Use @all to export all registered devices.

If the last argument ends in .csv, it's treated as the output file.
Otherwise outputs to stdout.

```
shelly export csv <devices...> [file] [flags]
```

### Examples

```
  # Export to stdout
  shelly export csv living-room bedroom

  # Export all devices to file
  shelly export csv @all devices.csv

  # Export specific devices
  shelly export csv living-room bedroom kitchen devices.csv

  # Without header row
  shelly export csv @all --no-header
```

### Options

```
  -h, --help        help for csv
      --no-header   Omit CSV header row
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

* [shelly export](shelly_export.md)	 - Export fleet data for infrastructure tools

