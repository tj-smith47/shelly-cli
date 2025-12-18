## shelly dash monitor

Launch dashboard in monitor view

### Synopsis

Launch the TUI dashboard directly in the monitor view.

The monitor view shows real-time status updates for all devices,
including power consumption, temperature, and connectivity.

```
shelly dash monitor [flags]
```

### Examples

```
  # Launch monitor view
  shelly dash monitor

  # With filter
  shelly dash monitor --filter living

  # With fast refresh
  shelly dash monitor --refresh 2
```

### Options

```
      --filter string   Filter devices by name pattern
  -h, --help            help for monitor
      --refresh int     Data refresh interval in seconds (default 5)
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

* [shelly dash](shelly_dash.md)	 - Launch interactive TUI dashboard

